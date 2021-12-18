package store

import (
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

type Store struct {
	db     *leveldb.DB
	path   string
	window time.Duration
}

func (s *Store) New(path string) error {
	return nil
}

func (s *Store) Save(data []byte) (string, error) {
	return "key", nil
}

func (s *Store) GetByKeyPrefix(prefix string) ([]byte, error) {
	return []byte("data"), nil
}
