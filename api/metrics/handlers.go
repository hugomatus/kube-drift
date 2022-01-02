package metrics

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"
)

func metricsHandler(s *store.Store) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		prefixKey := handleVars(r)

		appLog.Infof("metricsHandler: %s", prefixKey)
		resp, err := s.GetByKeyPrefix(prefixKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(resp)
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

	if vars["container"] != "" {
		prefixKey = fmt.Sprintf("/%s/%s/%s/%s", vars["name"], vars["namespace"], vars["podname"], vars["container"])
	}

	if vars["metric"] != "" {
		prefixKey = fmt.Sprintf("/%s/%s/%s/%s/%s", vars["name"], vars["namespace"], vars["podname"], vars["container"], vars["metric"])
	}
	return prefixKey
}
