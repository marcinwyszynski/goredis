package lib

// Store is capable of storing and retrieving elements.
type Store interface {
	Get(key string) (value string, found bool, err error)
	Set(key string, value string) error
}
