package lib

import "github.com/pkg/errors"

// CachingStore is a Store with two layers - one being a more expensive, slower
// to access authoritative source of data, and the other being a local cache.
type CachingStore struct {
	Authority Store
	Cache     Store

	KnownMissing map[string]struct{}
}

// NewCachingStore returns a fully functional implementation of Store, using two
// stores.
func NewCachingStore(authority, cache Store) *CachingStore {
	return &CachingStore{
		Authority:    authority,
		Cache:        cache,
		KnownMissing: make(map[string]struct{}),
	}
}

// Get is a layered implementation of the Store's Get method.
func (l *CachingStore) Get(key string) (value string, found bool, err error) {
	if _, missing := l.KnownMissing[key]; missing {
		return
	}

	if value, found, err = l.Cache.Get(key); found || err != nil {
		err = errors.Wrap(err, "could not retrieve value from cache")
		return
	}

	if value, found, err = l.Authority.Get(key); err != nil {
		err = errors.Wrap(err, "could not retrieve value from authority")
		return
	}

	if found {
		err = l.cache(key, value)
	} else {
		l.KnownMissing[key] = struct{}{}
	}

	return
}

// Set is a layered implementation of the Store's Set method.
func (l *CachingStore) Set(key string, value string) error {
	delete(l.KnownMissing, key)

	if err := l.Authority.Set(key, value); err != nil {
		return errors.Wrap(err, "could not set value in authority")
	}

	return l.cache(key, value)
}

func (l *CachingStore) cache(key string, value string) error {
	return errors.Wrap(l.Cache.Set(key, value), "could not set value in cache")
}
