package provider

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"k8s.io/klog/v2"
	"net/http"
)

func statsSummaryHandler(s *Store) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		resp, err := getStatsSummary(s, vars["Name"])
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("Node Metrics Error - %v", err.Error())))
			if err != nil {
				klog.Errorf("Error cannot write response: %v", err)
			}
		}

		j, err := json.Marshal(resp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("JSON Error - %v", err.Error())))
			if err != nil {
				klog.Errorf("Error cannot write response: %v", err)
			}
		}

		_, err = w.Write(j)
		if err != nil {
			klog.Errorf("Error cannot write response: %v", err)
		}
	}

	return fn
}

func getStatsSummary(s *Store, nodeName string) ([]SummaryStats, error) {

	var results []SummaryStats
	var summaryStats SummaryStats
	var iter iterator.Iterator
	cnt := 0

	if len(nodeName) > 0 {
		iter = s.db.NewIterator(util.BytesPrefix([]byte(fmt.Sprintf("%s-", nodeName))), nil)

	} else {
		iter = s.db.NewIterator(nil, nil)
	}

	for iter.Next() {
		//if cnt < 26 {
		err := json.Unmarshal(iter.Value(), &summaryStats)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		results = append(results, summaryStats)
		cnt++
		//}
	}

	klog.Infof("Status: Retrieved %d records from s")

	//release
	iter.Release()
	err := iter.Error()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return results, nil
}
