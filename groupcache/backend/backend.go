package backend

type Backend interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
}


