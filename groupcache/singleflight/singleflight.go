package singleflight

import "sync"

type call struct {
	wg sync.WaitGroup		// wg来通知删除key
	ch chan int				// 由channel来通知并发阻塞的其他g
	val interface{}
	err error
}

// Group 保存将所有并发一个key的请求用一个call实例来表示
type Group struct {
	mu sync.Mutex      // 并发保护map
	m map[string]*call // 保存并发同一个key的call
}

func New() *Group {
	return &Group{
		m: make(map[string]*call),
	}
}

// DoOnce make sure that fn do only once
func (g *Group) DoOnce(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		<-c.ch
		return c.val, c.err
	}

	c := new(call)
	g.m[key] = c
	c.ch = make(chan int)
	g.mu.Unlock()

	c.val, c.err = fn()
	// notify other g that wait for the same key
	close(c.ch)

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

