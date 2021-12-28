package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

func cadvisorHandler(s *Store) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		prefixKey := getKeyPrefix(r)

		klog.Infof("cadvisorHandler: %s", prefixKey)
		resp, err := getStatsSummary(s, prefixKey)
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

func getStatsSummary(s *Store, keyPrefix string) ([]*model.Sample, error) {
	var results []*model.Sample
	var iter iterator.Iterator
	cnt := 0

	if len(keyPrefix) > 0 {
		iter = s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)

	} else {
		iter = s.db.NewIterator(nil, nil)
	}

	for iter.Next() {
		result := model.Sample{}
		err := json.Unmarshal(iter.Value(), &result)

		if err != nil {
			klog.Error(err)
			return nil, err
		}
		results = append(results, &result)
		cnt++
	}

	klog.Infof("Status: Retrieved %d records from store", cnt)

	//release
	iter.Release()
	err := iter.Error()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return results, nil
}

func DecodeResponse(data []byte) ([]*model.Sample, error) {

	ioReaderData := strings.NewReader(string(data))
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(ioReaderData)

	if err != nil {
		return nil, err
	}
	dec := expfmt.NewDecoder(buf, expfmt.FmtText)
	decoder := expfmt.SampleDecoder{
		Dec:  dec,
		Opts: &expfmt.DecodeOptions{},
	}

	var samples []*model.Sample
	for {
		var v model.Vector
		if err := decoder.Decode(&v); err != nil {

			if err == io.EOF {
				// Expected loop termination condition.
				break
			}
			return nil, err
		}
		samples = append(samples, v...)
	}
	return samples, nil
}

func getKeyPrefix(r *http.Request) string {
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
