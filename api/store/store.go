package store

import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"hash/fnv"
	appLog "k8s.io/klog/v2"
	"time"
)

// Store is a wrapper around a LevelDB instance.
type Store struct {
	db     *leveldb.DB
	path   string
	window time.Duration
}

// Init creates a new Store.
func (s *Store) Init(path string) error {
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return err
	}

	s.db = db
	s.path = path
	s.window = time.Minute * 6

	return nil
}

// DB returns the underlying leveldb database.
func (s *Store) DB() *leveldb.DB {
	return s.db
}

// Close closes the store.
func (s *Store) Close() {
	s.db.Close()
}

// Save a new set of samples to the database.
// args: key, value
func (s *Store) Save(k string, data []byte) error {
	err := s.db.Put([]byte(k), data, nil)
	if err != nil {
		appLog.Error(err)
	}
	//appLog.Infof("Record saved : %s", k)
	return nil
}

// GetByKeyPrefix returns the drift(s) value for a given key prefix
func (s *Store) GetByKeyPrefix(k string) ([]byte, error) {
	var entries []byte
	cnt := 0
	appLog.Infof("Get record by key prefix: %s", k)

	iter := s.db.NewIterator(util.BytesPrefix([]byte(k)), nil)

	for iter.Next() {
		entries = append(entries, iter.Value()...)
		cnt++
	}

	// release
	iter.Release()
	err := iter.Error()
	if err != nil {
		appLog.Errorf("error releasing iterator: %s", err)
		return nil, err
	}
	appLog.Infof("Status: Retrieved %d records from store", cnt)
	return entries, nil
}

// SaveMetrics saves the metric samples to the database
func (s *Store) SaveMetrics(d map[string][]*model.Sample) (string, error) {

	var prefix string
	var cnt int

	//for each node (n) key: nodeName, value (v): []byte
	for n, v := range d {
		//for each model.Sample (ms) key: metricName, value: []byte
		for _, ms := range v {
			if MetricLabel.IsValid(*ms) {
				//appLog.Infof("metric: %s", time.UnixMilli(ms.Timestamp.UnixNano()/int64(time.Millisecond)))
				z := getUniqueKey(ms)
				d, err := ms.MarshalJSON()
				if err != nil {
					err = errors.Wrap(err, "failed to marshal ms metric")
					appLog.Error(err)
					return "", err
				}

				prefix = fmt.Sprintf("/%s/%s", n, z)

				err = s.Save(prefix, d)
				if err != nil {
					err = errors.Wrap(err, "failed to save metrics scrape record")
					appLog.Error(err)
				}
				cnt++
				//err = s.Save(fmt.Sprintf("/key/%s", prefix), []byte(fmt.Sprintf("%s", prefix)))
			}
		}
	}
	appLog.Infof(fmt.Sprintf("saved %v metric samples", cnt))
	return prefix, nil
}

func getUniqueKey(sample *model.Sample) string {
	h := fnv.New64a()
	// Hash of Timestamp
	h.Write([]byte(time.Now().String()))
	key := hex.EncodeToString(h.Sum(nil))

	prefix := fmt.Sprintf("%s/%s/%s/%s", string(sample.Metric["namespace"]), string(sample.Metric["pod"]), sample.Metric["__name__"], sample.Metric["container"])
	key = fmt.Sprintf("%s/%v", prefix, key)
	return key
}

func getPartialPrefix(sample *model.Sample) string {
	prefix := fmt.Sprintf("%s/%s/%s/%s", string(sample.Metric["namespace"]), string(sample.Metric["pod"]), sample.Metric["__name__"], sample.Metric["container"])
	return prefix
}
