package filter


type Filter interface {
	Put(key string)
	Contain(key string) bool
}

