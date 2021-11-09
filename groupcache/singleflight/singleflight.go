package singleflight

import "sync"

type call struct {
	wg sync.WaitGroup			// call.val 和 call.err 获取过程中由一个wg来保护
	val interface{}
	err error
}

// Group 保存将所有并发一个key的请求用一个call实例来表示
type Group struct {
	mu sync.Mutex      // 并发保护map
	m map[string]*call // 保存并发同一个key的call
}

// DoOnce make sure that fn do only once
func (g *Group) DoOnce(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

