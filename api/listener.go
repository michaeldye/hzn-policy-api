package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/open-horizon/anax/policy"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"reflect"
	"strings"
	"syscall"
	// "github.com/cloudflare/cfssl/errors"
	// "io"
	"errors"
)

// glog Info level guidelines:
// 3 = standard, most info
// 4 = more info, kinda debug
// 5 = debug
// 6 = trace

func authenticateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			token = tokens[0]
			token = strings.TrimPrefix(token, "Bearer ")
		}

		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}


	})
}

// Listen sets up an HTTP server and listens on given interface and port (ex: "0.0.0.0:8080")
func Listen(listenOn string) {
	router := mux.NewRouter()

	router.HandleFunc("/status", statusAPI).Methods("GET", "HEAD", "OPTIONS")
	router.HandleFunc("/policy/{id:[0-9A-Za-z.-]+}", policyAPI).Methods("GET", "HEAD", "OPTIONS", "POST", "DELETE")
	router.HandleFunc("/policies", policiesAPI).Methods("GET", "HEAD", "OPTIONS", "POST")
	router.HandleFunc("/policies/names", policiesNamesAPI).Methods("GET", "HEAD", "OPTIONS")

	glog.Infof("Listening on port: %v", listenOn)

	recovery := handlers.RecoveryHandler()(router)

	// will run in a greenthread, the function this is in will return
	go func() {
		http.ListenAndServe(listenOn, recovery)
	}()
}

// status is an API server status return type
type status struct {
	Online    bool `json:"online"`
	FileCount int  `json:"file_count"`
}

func newStatus(online bool, fileCount int) *status {
	return &status{
		Online:    online,
		FileCount: fileCount,
	}
}

type policyNameList struct {
	Policies []string `json:"policies"`
}

type policyList struct {
	Policies map[string]policy.Policy `json:"policies"`
}

func policyNameListFactory() (*policyNameList) {
	return &policyNameList{Policies: []string{}}
}

func policyListFactory() (*policyList) {
	return &policyList{Policies: map[string]policy.Policy{}}
}

func writeResponse(writer http.ResponseWriter, outModel interface{}, successStatusCode int) {
	serial, err := json.Marshal(outModel)

	if err != nil {
		glog.Error(err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(successStatusCode)
	writer.Header().Set("Content-Type", "application/json")

	if _, err := writer.Write(serial); err != nil {
		glog.Error(err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
	}
}

func getPolicyList(location string) (*policyList, error) {
	pl := *policyListFactory()

	files, _ := ioutil.ReadDir(location)
	for _, f := range files {
		if (strings.HasSuffix(f.Name(), ".policy")) {
			pname := strings.TrimSuffix(f.Name(), ".policy")
			pol, err := policy.ReadPolicyFile(fmt.Sprintf("/tmp/policy.d/%s.policy", pname))
			if err != nil {
				return nil, err
			}
			pl.Policies[pname] = *pol
		}
	}

	return &pl, nil
}

func getPolicyNameList(location string) (*policyNameList, error) {
	pl := policyNameListFactory()

	files, err := ioutil.ReadDir(location)
        if err == nil {
		for _, f := range files {
			if (strings.HasSuffix(f.Name(), ".policy")) {
				/* this is the worst of all possible timelines -walter */
				pl.Policies = append(pl.Policies, strings.TrimSuffix(f.Name(), ".policy"))
			}
		}
	} else {
		return nil, err
	}
	return pl, nil
}

func writePolicy(location string, pol policy.Policy)(error) {
	jbod, err := json.Marshal(pol)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(location, jbod, 0600)
	return err
}

func deletePolicy(pname string)(error) {
	pname = strings.TrimSuffix(pname, ".policy")
	ppath := policyToPath(pname, "/tmp/policy.d/")
	err := syscall.Unlink(ppath)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func policyToPath(pname string, location string)(string) {
	return fmt.Sprintf("%s/%s.policy", location, pname)
}

func strInList(sortedlist []string, val string)(bool) {
	ix := sort.SearchStrings(sortedlist, val)
	if ix >= len(sortedlist) {
		return false
	} else if sortedlist[ix] == val {
		return true
	}
	return false
}

func setPolicies(location string, newpol *policyList)(error) {
	/* get list of policies on disk */
	diskpols, err := getPolicyNameList("/tmp/policy.d/")
	sort.Strings(diskpols.Policies)

	if err != nil {
		return err
	}

	deferredErr := false
	for pname, pol := range newpol.Policies {
		if strInList(diskpols.Policies, pname) {
			/* compare existing policy to the new one */
			oldpol, err := policy.ReadPolicyFile(fmt.Sprintf("/tmp/policy.d/%s.policy", pname))
			if err != nil {
				deferredErr = true
			} else {
				if reflect.DeepEqual(pol, oldpol) {
					glog.V(5).Infof("Skipping identical policy '%s'", pname)
				} else {
					glog.V(5).Infof("Replacing policy '%s'", pname)
					err := writePolicy(fmt.Sprintf("/tmp/policy.d/%s.policy", pname), pol)
					if err != nil {
						glog.V(5).Infof("Policy write failed for updated policy '%s'", pname)
						deferredErr = true
					}
				}
			}
		} else {
			/* write new policy not previously on disk */
			err := writePolicy(fmt.Sprintf("/tmp/policy.d/%s.policy", pname), pol)
			if err != nil {
				glog.V(5).Infof("Policy write failed for new policy '%s'", pname)
				deferredErr = true
			}
		}
	}

	/* remove policies that don't exist in the provided set */
	for _, pname := range diskpols.Policies {
		if _, ok := newpol.Policies[pname]; !ok {
			glog.V(5).Infof("Removing policy '%s'", pname)
			deletePolicy(pname)
		}
	}

	if deferredErr {
		return errors.New(fmt.Sprintf("Policy batch process had partial failure- see logs"))
	}
	return nil

}

/* APIs */
func statusAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "HEAD":
		glog.V(5).Infof("HEAD: %v", r)
		w.WriteHeader(http.StatusOK)
	case "OPTIONS":
		glog.V(5).Infof("OPTIONS: %v", r)
		w.Header().Set("Allow", "HEAD, OPTIONS, GET")
		w.WriteHeader(http.StatusOK)
	case "GET":
		glog.V(5).Infof("GET: %v", r)

		diskpols, err := getPolicyNameList("/tmp/policy.d/")
		if err != nil {
			wrapper := make(map[string]interface{}, 0)
			wrapper["status"] = newStatus(false, 0)

			writeResponse(w, wrapper, http.StatusInternalServerError)
		} else {
			numpol := len(diskpols.Policies)
			wrapper := make(map[string]interface{}, 0)
			wrapper["status"] = newStatus(true, numpol)

			writeResponse(w, wrapper, http.StatusOK)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func policyAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	switch r.Method {
	case "HEAD":
		glog.V(5).Infof("HEAD: %v", r)
		w.WriteHeader(http.StatusOK)
	case "OPTIONS":
		glog.V(5).Infof("OPTIONS: %v", r)
		w.Header().Set("Allow", "HEAD, OPTIONS, GET, POST, DELETE")
		w.WriteHeader(http.StatusOK)
	case "GET":
		glog.V(5).Infof("GET: %v", r)

		if id, exists := vars["id"]; !exists {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			glog.V(5).Infof("Some id: %s", id)

			id = strings.TrimSuffix(id, ".policy")

			pol, err := policy.ReadPolicyFile(fmt.Sprintf("/tmp/policy.d/%s.policy", id))

			if err == nil {
				writeResponse(w, pol, http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	case "POST":
		glog.V(5).Infof("POST: %v", r)

		if id, exists := vars["id"]; !exists {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			var pol policy.Policy
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				glog.V(5).Infof("Problem reading request body")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				err := json.Unmarshal(body, &pol)
				if err != nil {
					glog.V(5).Infof("Policy marshal error: %s", err)
					w.WriteHeader(http.StatusBadRequest)
				} else {
					err := writePolicy(fmt.Sprintf("/tmp/policy.d/%s.policy", id), pol)
					if err != nil {
						glog.V(5).Infof("Policy write failed")
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusOK)
					}
				}
			}
		}
	case "DELETE":
		glog.V(5).Infof("DELETE: %v", r)
		if id, exists := vars["id"]; !exists {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			err := deletePolicy(id)
			if err != nil {
				glog.V(5).Infof("Problem deleting policy file '%s': %s", id, err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func policiesAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "HEAD":
		glog.V(5).Infof("HEAD: %v", r)
		w.WriteHeader(http.StatusOK)
	case "OPTIONS":
		glog.V(5).Infof("OPTIONS: %v", r)
		w.Header().Set("Allow", "HEAD, OPTIONS, GET, POST")
		w.WriteHeader(http.StatusOK)
	case "GET":
		glog.V(5).Infof("GET: %v", r)
		pl, err := getPolicyList("/tmp/policy.d")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			writeResponse(w, pl.Policies, http.StatusOK)
		}
	case "POST":
		glog.V(5).Infof("POST: %v", r)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.V(5).Infof("Problem reading request body")
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			plist := policyListFactory()
			err := json.Unmarshal(body, &plist.Policies)
			if err != nil {
				glog.V(5).Infof("Problem reading request body")
				w.WriteHeader(http.StatusBadRequest)
			} else {
				err := setPolicies("/tmp/policy.d/", plist)
				if err != nil {
					glog.V(5).Infof("Problem reading request body")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func policiesNamesAPI(w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infof("policiesnamesapi: %v", r)

	switch r.Method {
	case "HEAD":
		glog.V(5).Infof("HEAD: %v", r)
		w.WriteHeader(http.StatusOK)
	case "OPTIONS":
		glog.V(5).Infof("OPTIONS: %v", r)
		w.Header().Set("Allow", "HEAD, OPTIONS, GET, POST")
		w.WriteHeader(http.StatusOK)
	case "GET":
		glog.V(5).Infof("GET: %v", r)
		pl, err := getPolicyNameList("/tmp/policy.d")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			writeResponse(w, pl.Policies, http.StatusOK)
		}
	case "POST":
		glog.V(5).Infof("POST: %v", r)
		w.WriteHeader(http.StatusOK)

		wrapper := make(map[string]interface{}, 0)
		wrapper["status"] = newStatus(66)

		writeResponse(w, wrapper, http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
