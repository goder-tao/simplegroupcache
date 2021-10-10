// @Description: 支持并发的Cache
// @Author: tao
// @Date: 2021/10/08 17:23
package lru

import (
	"container/list"
	"sync"
)

type Cache struct {
	mu      sync.Mutex
	lru     *cache
	maxSize int64
	OnEvicted func(key string, value Value)
}

// 可自定义的淘汰回调方法
func NewCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	if onEvicted == nil {
		return &Cache{
			maxSize: maxBytes,
			OnEvicted: defaultEvicted,
		}
	} else {
		return &Cache{
			maxSize: maxBytes,
			OnEvicted: onEvicted,
		}
	}
}

func (c *Cache) Add(key string, value ByteValue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = &cache{
			maxBytes: c.maxSize,
			ll: list.New(),
			cache: make(map[string]*list.Element),
			OnEvicted: c.OnEvicted,
		}
	}
	c.lru.Add(key, value)
}

func (c *Cache) Remove() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = &cache{
			maxBytes: c.maxSize,
			ll: list.New(),
			cache: make(map[string]*list.Element),
			OnEvicted: c.OnEvicted,
		}
	}
	c.lru.Remove()
}

func (c *Cache) Get(key string) (ByteValue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = &cache{
			maxBytes: c.maxSize,
			ll: list.New(),
			cache: make(map[string]*list.Element),
			OnEvicted: c.OnEvicted,
		}
	}
	v, err := c.lru.Get(key)
	if err != nil {
		return ByteValue{}, err
	} else {
		return v.(ByteValue), nil
	}

}

