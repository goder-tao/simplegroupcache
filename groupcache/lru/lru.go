// @Description:
// @Author: tao
// @Update: 2021/10/08 14:49
package lru

import (
	"container/list"
	"errors"
)

// Groupcache结构体
type cache struct {
	// cache的最大容量
	maxBytes int64
	// cache的当前容量
	nBytes int64
	// lru list, 双向链表实现淘汰和将刚刚访问过的节点插入头节点
	ll *list.List
	// map实现O(1)复杂度找到list节点
	cache map[string]*list.Element
	// 淘汰函数
	OnEvicted func(key string, value Value)
}

// cache缓存项, list的一个节点
type entry struct {
	key   string
	value Value
}

// 用一个接口允许值是任何数据结构
type Value interface {
	Len() int
}

// 默认的淘汰回调方法
func defaultEvicted(key string, value Value) {

}

// 查找
func (c *cache) Get(key string) (Value, error) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		e := ele.Value.(*entry)
		return e.value, nil
	} else {
		return nil, errors.New("no such key-value")
	}
}

// 删除节点队列尾部节点
func (c *cache) Remove() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *cache) RemoveNode(key string) Value {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		return ele.Value.(*entry).value
	} else {
		return nil
	}
}

// 新增节点
func (c *cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // k-v已经存在了，覆盖
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else { // k-v还没存在，新建
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nBytes += int64(ele.Value.(*entry).value.Len() + len(ele.Value.(*entry).key))
	}

	// 超过cache的最大size进行淘汰
	for c.nBytes > c.maxBytes {
		c.Remove()
	}
}
