package provider

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"k8s.io/klog/v2"
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

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Save(key string, data []byte) error {
	err := s.db.Put([]byte(key), data, nil)
	if err != nil {
		klog.Error(err)
	}
	klog.Infof("saved drift: %s", key)
	return nil
}

func (s *Store) GetDriftByKey(key string) (interface{}, error) {

	drift := &PodDrift{}

	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return drift, err
	}

	err = json.Unmarshal(data, &drift)
	//drift = Deserialize(data)

	return drift, nil
}

func (s *Store) GetDriftByKeyPrefix(keyPrefix string) ([]byte, error) {
	klog.Infof("get drift by key prefix: %s", keyPrefix)
	var iter iterator.Iterator
	//iter = c.db.NewIterator(nil, nil)
	iter = s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)

	drifts, err := s.GetDrifts(iter)
	if err != nil {
		klog.Errorf("error getting drift: %s", err)
		return drifts, err
	}

	// release
	iter.Release()
	err = iter.Error()
	if err != nil {
		klog.Errorf("error releasing iterator: %s", err)
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

	klog.Infof("Returning drift count: %d", cnt)
	return entries, nil
}
