package drift

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"
)

func driftHandler(s *store.Store) func(http.ResponseWriter, *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {

		prefix := handleDriftVars(r)

		appLog.Infof("driftHandler: %v", prefix)
		resp, err := s.GetByKeyPrefix(prefix)

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

func handleDriftVars(r *http.Request) string {
	vars := mux.Vars(r)

	kind := vars["kind"]
	namespace := vars["namespace"]
	templateHash := vars["template-hash"]

	prefix := fmt.Sprintf("/%s/%s", kind, namespace)

	if namespace == "" {
		prefix = fmt.Sprintf("/%s", kind)
	}

	if namespace != "" && templateHash != "" {
		prefix = fmt.Sprintf("/%s/%s/%s", kind, namespace, templateHash)
	}
	return prefix
}
