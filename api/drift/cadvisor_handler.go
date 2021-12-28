package drift

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api/store"
	"github.com/prometheus/common/model"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	appLog "k8s.io/klog/v2"
	"net/http"
)

func cadvisorHandler(s *store.Store) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		prefixKey := handleVars(r)

		appLog.Infof("cadvisorHandler: %s", prefixKey)
		resp, err := getMetrics(s, prefixKey)
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

func getMetrics(s *store.Store, k string) ([]*model.Sample, error) {
	var results []*model.Sample
	var iter iterator.Iterator
	cnt := 0

	if len(k) > 0 {
		iter = s.DB().NewIterator(util.BytesPrefix([]byte(k)), nil)

	} else {
		iter = s.DB().NewIterator(nil, nil)
	}

	for iter.Next() {
		result := model.Sample{}
		err := json.Unmarshal(iter.Value(), &result)

		if err != nil {
			appLog.Error(err)
			return nil, err
		}
		results = append(results, &result)
		cnt++
	}

	appLog.Infof("Status: Retrieved %d records from store", cnt)

	//release
	iter.Release()
	err := iter.Error()
	if err != nil {
		appLog.Error(err)
		return nil, err
	}

	return results, nil
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
