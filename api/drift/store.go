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

func (s *Store) Save(drift KubeDrift) error {
	drift.SetKey() //fmt.Sprintf("%s/%s/%s", event, p.Namespace, p.UID)
	data, err := json.Marshal(drift)

	if err != nil {
		return err
	}

	err = s.db.Put([]byte(drift.GetKey()), data, nil)
	if err != nil {
		klog.Error(err)
	}
	klog.Infof("saved drift: %s", drift.GetKey())
	return nil
}

func (s *Store) GetDriftByKey(key string) (KubeDrift, error) {

	drift := KubeDrift{}

	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return drift, err
	}

	err = json.Unmarshal(data, &drift)
	//drift = Deserialize(data)

	return drift, nil
}

func (s *Store) GetDriftByKeyPrefix(keyPrefix string) ([]KubeDrift, error) {
	klog.Infof("get drift by key prefix: %s", keyPrefix)
	var iter iterator.Iterator
	//iter = c.db.NewIterator(nil, nil)
	iter = s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)

	entries, drifts, err := s.GetDrifts(iter)
	if err != nil {
		klog.Errorf("error getting drift: %s", err)
		return drifts, err
	}

	return entries, nil
}

func (s *Store) GetDrifts(iter iterator.Iterator) ([]KubeDrift, []KubeDrift, error) {
	var entries []KubeDrift
	cnt := 0

	for iter.Next() {
		drift := KubeDrift{}
		json.Unmarshal(iter.Value(), &drift)
		entries = append(entries, drift)
		cnt++
	}

	// release
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, nil, err
	}
	return entries, nil, nil
}

func (s *Store) SaveDrift(drift KubeDrift) error {
	data, err := json.Marshal(drift)
	if err != nil {
		return err
	}
	err = s.db.Put([]byte(drift.GetKey()), data, nil)
	if err != nil {
		return err
	}
	return nil
}
