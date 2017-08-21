package api

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"github.com/open-horizon/anax/policy"
	"net/http"
)

// glog Info level guidelines:
// 3 = standard, most info
// 4 = more info, kinda debug
// 5 = debug
// 6 = trace

// Listen sets up an HTTP server and listens on given interface and port (ex: "0.0.0.0:8080")
func Listen(listenOn string) {
	router := mux.NewRouter()

	ph := &PolicyHandler{"/tmp/policy.d/", "blerf"}

	router.HandleFunc("/status", ph.statusHandler).Methods("GET", "HEAD", "OPTIONS")
	router.HandleFunc("/policy/{id:[0-9A-Za-z.-]+}", ph.policyHandler).Methods("GET", "HEAD", "OPTIONS", "POST", "DELETE")
	router.HandleFunc("/policies", ph.policiesHandler).Methods("GET", "HEAD", "OPTIONS", "POST")
	router.HandleFunc("/policies/names", ph.policiesNamesHandler).Methods("GET", "HEAD", "OPTIONS")

	glog.Infof("Listening on port: %v", listenOn)

	authmiddleware := ph.authenticateHandler(router)
	recoverymiddleware := handlers.RecoveryHandler()(authmiddleware)

	// will run in a greenthread, the function this is in will return
	go func() {
		http.ListenAndServe(listenOn, recoverymiddleware)
	}()
}

// status is an API server status return type
type status struct {
	Online      bool `json:"online"`
	PolicyCount int  `json:"policy_count"`
}

type policyNameList struct {
	Policies []string `json:"policies"`
}

type policyList struct {
	Policies map[string]policy.Policy `json:"policies"`
}

func statusFactory(online bool, policyCount int) *status {
	return &status{
		Online:      online,
		PolicyCount: policyCount,
	}
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

