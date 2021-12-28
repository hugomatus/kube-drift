package provider

import (
	"bytes"
	"encoding/json"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
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

// New creates a new Store.
func (s *Store) New(path string) error {
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
	appLog.Infof("Drift: saved using k: %s", k)
	return nil
}

// GetDriftByKey returns the drift value for a given key
func (s *Store) GetDriftByKey(k string) (interface{}, error) {

	drift := &PodDrift{}
	data, err := s.db.Get([]byte(k), nil)
	if err != nil {
		return drift, err
	}

	err = json.Unmarshal(data, &drift)

	return drift, nil
}

// GetDriftByKeyPrefix returns the drift(s) value for a given key prefix
func (s *Store) GetDriftByKeyPrefix(k string) ([]byte, error) {
	appLog.Infof("get drift by key prefix: %s", k)
	var iter iterator.Iterator

	iter = s.db.NewIterator(util.BytesPrefix([]byte(k)), nil)

	drifts, err := s.getDrifts(iter)
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

func (s *Store) getDrifts(i iterator.Iterator) ([]byte, error) {
	var entries []byte
	cnt := 0

	for i.Next() {
		entries = append(entries, i.Value()...)
		cnt++
	}

	appLog.Infof("Returning drift count: %d", cnt)
	return entries, nil
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
