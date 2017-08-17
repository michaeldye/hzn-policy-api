package api

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
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

	router.HandleFunc("/status", statusAPI).Methods("GET", "HEAD", "OPTIONS")

	glog.Infof("Gonna listen on port: %v", listenOn)

	// will run in a greenthread, the function this is in will return
	go func() {
		http.ListenAndServe(listenOn, router)
	}()
}

// status is an API server status return type
type status struct {
	Online    bool `json:"online"`
	FileCount int  `json:"file_count"`
}

func newStatus(fileCount int) *status {
	return &status{
		Online:    true,
		FileCount: fileCount,
	}
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

func statusAPI(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "HEAD":
		glog.V(5).Infof("HEAD: %v", r)
	case "OPTIONS":
		glog.V(5).Infof("OPTIONS: %v", r)
		w.Header().Set("Allow", "HEAD, OPTIONS, GET")
		w.WriteHeader(http.StatusOK)
	case "GET":
		glog.V(5).Infof("GET: %v", r)

		wrapper := make(map[string]interface{}, 0)
		wrapper["status"] = newStatus(66)

		writeResponse(w, wrapper, http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
