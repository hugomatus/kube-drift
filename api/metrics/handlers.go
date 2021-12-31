package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"
)

func cadvisorHandler(s *store.Store) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		prefixKey := handleVars(r)

		appLog.Infof("cadvisorHandler: %s", prefixKey)
		resp, err := s.GetMetrics(prefixKey)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("Node Metrics Error - %v", err.Error())))
			if err != nil {
				appLog.Errorf("Error cannot write response: %v", err)
			}
		}

		j, err := json.Marshal(resp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("JSON Error - %v", err.Error())))
			if err != nil {
				appLog.Errorf("Error cannot write response: %v", err)
			}
		}

		_, err = w.Write(j)
		if err != nil {
			appLog.Errorf("Error cannot write response: %v", err)
		}
	}

	return fn
}

func handleVars(r *http.Request) string {
	vars := mux.Vars(r)
	prefixKey := fmt.Sprintf("/%s", vars["name"])
	//{name}/{namespace}/{podname}/{metric}

	if vars["namespace"] != "" {
		prefixKey = fmt.Sprintf("/%s/%s", vars["name"], vars["namespace"])
	}
	if vars["podname"] != "" {
		prefixKey = fmt.Sprintf("/%s/%s/%s", vars["name"], vars["namespace"], vars["podname"])
	}

	if vars["metric"] != "" {
		prefixKey = fmt.Sprintf("/%s/%s/%s/%s", vars["name"], vars["namespace"], vars["podname"], vars["metric"])
	}
	return prefixKey
}
