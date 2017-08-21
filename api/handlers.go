package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/open-horizon/anax/policy"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"syscall"
)

type PolicyHandler struct {
	PolicyDir string
}

/* Status */
func (ph *PolicyHandler) statusGET(w http.ResponseWriter, r *http.Request) {
	diskpols, err := ph.getPolicyNameList()
	if err != nil {
		wrapper := make(map[string]interface{}, 0)
		wrapper["status"] = statusFactory(false, 0)

		writeResponse(w, wrapper, http.StatusInternalServerError)
	} else {
		numpol := len(diskpols.Policies)
		wrapper := make(map[string]interface{}, 0)
		wrapper["status"] = statusFactory(true, numpol)

		writeResponse(w, wrapper, http.StatusOK)
	}
}

func (ph *PolicyHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
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
		ph.statusGET(w,r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

/* Policy Names  */
func (ph *PolicyHandler) policiesNamesHandler(w http.ResponseWriter, r *http.Request) {
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
		pl, err := ph.getPolicyNameList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			writeResponse(w, pl.Policies, http.StatusOK)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

/* Policies */
func (ph *PolicyHandler) policiesGET(w http.ResponseWriter, r *http.Request) {
	pl, err := ph.getPolicyList()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		writeResponse(w, pl.Policies, http.StatusOK)
	}
}

func (ph *PolicyHandler) policiesPOST(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.V(5).Infof("Problem reading request body")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		newpols := policyListFactory()
		err := json.Unmarshal(body, &newpols.Policies)
		if err != nil {
			glog.V(5).Infof("Problem reading request body")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			err := ph.setPolicies(newpols)
			if err != nil {
				glog.V(5).Infof("Problem reading request body")
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}

func (ph *PolicyHandler) policiesHandler(w http.ResponseWriter, r *http.Request) {
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
		ph.policiesGET(w,r)
	case "POST":
		glog.V(5).Infof("POST: %v", r)
		ph.policiesPOST(w,r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

/* Policy */
func (ph *PolicyHandler) policyGET(w http.ResponseWriter, r *http.Request, rv map[string]string) {
	if id, exists := rv["id"]; !exists {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		glog.V(5).Infof("Some id: %s", id)
		pol, err := ph.readPolicy(id)
		if err == nil {
			writeResponse(w, pol, http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func (ph *PolicyHandler) policyPOST(w http.ResponseWriter, r *http.Request, rv map[string]string) {
	if id, exists := rv["id"]; !exists {
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
				err := ph.writePolicy(id, pol)
				if err != nil {
					glog.V(5).Infof("Policy write failed")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}

func (ph *PolicyHandler) policyDELETE(w http.ResponseWriter, r *http.Request, rv map[string]string) {
	if id, exists := rv["id"]; !exists {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		err := ph.deletePolicy(id)
		if err != nil {
			glog.V(5).Infof("Problem deleting policy file '%s': %s", id, err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (ph *PolicyHandler) policyHandler(w http.ResponseWriter, r *http.Request) {
	rv := mux.Vars(r)

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
		ph.policyGET(w,r,rv)
	case "POST":
		glog.V(5).Infof("POST: %v", r)
		ph.policyPOST(w,r,rv)
	case "DELETE":
		glog.V(5).Infof("DELETE: %v", r)
		ph.policyDELETE(w,r,rv)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

/* Util */
func (ph *PolicyHandler) deletePolicy(pname string)(error) {
	ppath := ph.policyNameToPath(pname)
	err := syscall.Unlink(ppath)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (ph *PolicyHandler) policyPathToName(ppath string) (string) {
	basename := path.Base(ppath)
	return strings.TrimSuffix(basename, ".policy")
}

func (ph *PolicyHandler) policyNameToPath(pname string) (string) {
	pname = strings.TrimSuffix(pname, ".policy")
	return fmt.Sprintf("%s/%s.policy", ph.PolicyDir, pname)
}

func (ph *PolicyHandler) setPolicies(newpol *policyList)(error) {
	/* get list of policies on disk */
	diskpols, err := ph.getPolicyNameList()
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
					err := ph.writePolicy(pname, pol)
					if err != nil {
						glog.V(5).Infof("Policy write failed for updated policy '%s'", pname)
						deferredErr = true
					}
				}
			}
		} else {
			/* write new policy not previously on disk */
			err := ph.writePolicy(pname, pol)
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
			ph.deletePolicy(pname)
		}
	}

	if deferredErr {
		return errors.New(fmt.Sprintf("Policy batch process had partial failure- see logs"))
	}
	return nil
}

func (ph *PolicyHandler) getPolicyNameList() (*policyNameList, error) {
	pl := policyNameListFactory()
	files, err := ioutil.ReadDir(ph.PolicyDir)
	if err == nil {
		for _, f := range files {
			if (strings.HasSuffix(f.Name(), ".policy")) {
				/* this is the worst of all possible timelines -walter */
				pl.Policies = append(pl.Policies, ph.policyPathToName(f.Name()))
			}
		}
	} else {
		return nil, err
	}
	return pl, nil
}

func (ph *PolicyHandler) writePolicy(pname string, pol policy.Policy)(error) {
	ppath := ph.policyNameToPath(pname)
	jbod, err := json.Marshal(pol)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(ppath, jbod, 0600)
	return err
}

func (ph *PolicyHandler) getPolicyList() (*policyList, error) {
	pl := *policyListFactory()

	files, _ := ioutil.ReadDir(ph.PolicyDir)
	for _, f := range files {
		if (strings.HasSuffix(f.Name(), ".policy")) {
			pname := ph.policyPathToName(f.Name())
			pol, err := ph.readPolicy(f.Name())
			if err != nil {
				return nil, err
			}
			pl.Policies[pname] = *pol
		}
	}

	return &pl, nil
}

func (ph *PolicyHandler) readPolicy(pname string) (*policy.Policy, error) {
	ppath := ph.policyNameToPath(pname)
	return policy.ReadPolicyFile(ppath)
}
