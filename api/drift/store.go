package provider

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
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

func (s *Store) Save(data []byte) (string, error) {
	return "key", nil
}

func (s *Store) GetDriftByKey(key string) (KubeDrift, error) {

	var drift KubeDrift

	//iter = c.db.NewIterator(nil, nil)
	data, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return drift, err
	}

	drift = Deserialize(data)

	return drift, nil
}

func (s *Store) GetDriftByKeyPrefix(keyPrefix string) ([]KubeDrift, error) {

	var iter iterator.Iterator
	//iter = c.db.NewIterator(nil, nil)
	iter = s.db.NewIterator(util.BytesPrefix([]byte(keyPrefix)), nil)

	entries, drifts, err := s.GetDrifts(iter)
	if err != nil {
		return drifts, err
	}

	return entries, nil
}

func (s *Store) GetDrifts(iter iterator.Iterator) ([]KubeDrift, []KubeDrift, error) {
	var entries []KubeDrift
	var entry KubeDrift
	cnt := 0

	for iter.Next() {
		entry = Deserialize(iter.Value())
		entries = append(entries, entry)
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
	err := s.db.Put([]byte(drift.GetKey()), []byte(drift.Serialize()), nil)
	if err != nil {
		return err
	}
	return nil
}
