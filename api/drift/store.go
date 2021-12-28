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

type Store struct {
	db     *leveldb.DB
	path   string
	window time.Duration
}

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

func (s *Store) DB() *leveldb.DB {
	return s.db
}
func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Save(key string, data []byte) error {
	err := s.db.Put([]byte(key), data, nil)
	if err != nil {
		appLog.Error(err)
	}
	appLog.Infof("Drift: saved using key: %s", key)
	return nil
}

func (s *Store) GetDriftByKey(key string) (interface{}, error) {

	drift := &PodDrift{}
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return drift, err
	}

	err = json.Unmarshal(data, &drift)

	return drift, nil
}

func (s *Store) GetDriftByKeyPrefix(keyPrefix string) ([]byte, error) {
	appLog.Infof("get drift by key prefix: %s", keyPrefix)
	var iter iterator.Iterator

	iter = s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)

	drifts, err := s.GetDrifts(iter)
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

func (s *Store) GetDrifts(iter iterator.Iterator) ([]byte, error) {
	var entries []byte
	cnt := 0

	for iter.Next() {
		entries = append(entries, iter.Value()...)
		cnt++
	}

	appLog.Infof("Returning drift count: %d", cnt)
	return entries, nil
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
				break
			}
			return nil, err
		}
		samples = append(samples, v...)
	}
	return samples, nil
}
