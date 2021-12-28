package provider

import (
	"fmt"
	"github.com/gorilla/mux"
	appLog "k8s.io/klog/v2"
	"net/http"
)

func driftHandler(s *Store) func(http.ResponseWriter, *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {

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

		appLog.Infof("driftHandler: %v", prefix)
		resp, err := s.GetDriftByKeyPrefix(prefix)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		//j, err := json.Marshal(resp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("JSON Error - %v", err)))
			if err != nil {
				//fmt.Printf("Error cannot write response: %v", err)
			}
		}

		_, err = w.Write(resp)
		if err != nil {
			fmt.Printf("Error cannot write response: %v", err)
		}
	}

	return fn
}
