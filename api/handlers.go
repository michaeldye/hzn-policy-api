package api
/*

import (
	"net/http"
	"github.com/golang/glog"
)

type PolicyHandler struct {
	PolicyDir string
}

func (ph PolicyHandler) statusAPI(w http.ResponseWriter, r *http.Request) {
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
*/
