package store

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"hash/fnv"
	"io"
	appLog "k8s.io/klog/v2"
	"strings"
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

	appLog.Infof("get drift by key prefix: %s", k)

	iter := s.db.NewIterator(util.BytesPrefix([]byte(k)), nil)

	drifts, err := s.getRecords(iter)
	if err != nil {
		appLog.Errorf("error getting drift: %s", err)
		return drifts, err
	}

	// release
	iter.Release()
	err = iter.Error()
	if err != nil {
		appLog.Errorf("error releasing iterator: %s", err)
		return nil, err
	}
	return drifts, nil
}

func (s *Store) getRecords(i iterator.Iterator) ([]byte, error) {
	var entries []byte
	cnt := 0

	for i.Next() {
		entries = append(entries, i.Value()...)
		cnt++
	}

	appLog.Infof("Record count: %d", cnt)
	return entries, nil
}

// GetMetrics returns the metrics for a given key prefix
func (s *Store) GetMetrics(k string) ([]*model.Sample, error) {
	var results []*model.Sample
	var iter iterator.Iterator
	cnt := 0

	if len(k) > 0 {
		iter = s.db.NewIterator(util.BytesPrefix([]byte(k)), nil)

	} else {
		iter = s.db.NewIterator(nil, nil)
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
				z := getUniqueKey()
				d, err := ms.MarshalJSON()
				if err != nil {
					err = errors.Wrap(err, "failed to marshal ms metric")
					appLog.Error(err)
					return "", err
				}

				prefix = fmt.Sprintf("/%s/%s/%v", n, getPartialPrefix(ms), z)

				err = s.Save(prefix, d)
				if err != nil {
					err = errors.Wrap(err, "failed to save metrics scrape record")
					appLog.Error(err)
				}
				cnt++
			}
		}
	}
	appLog.Infof(fmt.Sprintf("Total: Metric Sample Records=%v", cnt))
	return prefix, nil
}

// DecodeResponse decodes the response from the prometheus samples
func DecodeResponse(d []byte) ([]*model.Sample, error) {

	ioReaderData := strings.NewReader(string(d))
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
				break
			}
			return nil, err
		}
		samples = append(samples, v...)
	}
	return samples, nil
}

func getUniqueKey() string {
	h := fnv.New64a()
	// Hash of Timestamp
	h.Write([]byte(time.Now().String()))
	key := hex.EncodeToString(h.Sum(nil))
	return key
}

func getPartialPrefix(sample *model.Sample) string {
	prefix := fmt.Sprintf("%s/%s/%s/%s", string(sample.Metric["namespace"]), string(sample.Metric["pod"]), sample.Metric["__name__"], sample.Metric["container"])
	return prefix
}
