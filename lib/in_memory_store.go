package lib

import "sync"

type inMemoryStore struct {
	data map[string]string
	lock *sync.RWMutex
}

// NewInMemoryStore returns an in-memory implementation of Store.
func NewInMemoryStore() Store {
	return &inMemoryStore{
		data: make(map[string]string),
		lock: new(sync.RWMutex),
	}
}

func (s *inMemoryStore) Get(key string) (value string, found bool, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, found = s.data[key]
	return
}

func (s *inMemoryStore) Set(key string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = value
	return nil
}
