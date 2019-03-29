package main

import "sync"

type store interface {
	get(key string) (value string, found bool, err error)
	set(key string, value string) error
}

type inMemoryStore struct {
	data map[string]string
	lock *sync.RWMutex
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		data: make(map[string]string),
		lock: new(sync.RWMutex),
	}
}

func (s *inMemoryStore) get(key string) (value string, found bool, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, found = s.data[key]
	return
}

func (s *inMemoryStore) set(key string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = value
	return nil
}
