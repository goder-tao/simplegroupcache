// @Description: group
// @Author: tao
// @Date: 2021/10/08 19:14
package group

import (
	"errors"
	"fmt"
	"simplecache/groupcache/lru"
	"simplecache/groupcache/peer"
	"simplecache/groupcache/singleflight"
	"simplecache/util"
	"sync"
)

// 缓存未命中时的回调方法，由用户来提供实现
type Getter interface {
	Get(key string) ([]byte, error)
}

// 接口型函数, 调用接口的方法相当于调用自己，让一个函数能够当成回调传入
// 可替换多种不同的函数，只要保持和接口定义的类型一致即可，IOP编程
type GetFunc func(key string) ([]byte, error)

func (g GetFunc) Get(key string) ([]byte, error) {
	return g(key)
}

// 一种类型的缓存, 多种类型的缓存构成了group，在同一个节点上
type Member struct {
	name   string
	getter Getter
	mCache *lru.Cache
	picker peer.PeerPicker
	loader *singleflight.Group
}

type Group struct {
	mu sync.RWMutex
	group map[string]*Member
}

// 提供一个默认的group
var defaultGroup Group = Group{
	group: make(map[string]*Member),
}

func NewMember(name string, cacheSize int64, getter Getter, onEvicted func(key string, value lru.Value)) *Member {
	if getter == nil {
		panic("getter is nil")
	}
	defaultGroup.mu.Lock()
	defer defaultGroup.mu.Unlock()
	m := &Member{
		name:   name,
		getter: getter,
		mCache: lru.NewCache(cacheSize, onEvicted),
		loader: &singleflight.Group{},
	}
	defaultGroup.group[name] = m
	return m
}

func GetMember(name string) *Member {
	defaultGroup.mu.RLock()
	m := defaultGroup.group[name]
	defaultGroup.mu.RUnlock()
	return m
}


// Group的方法
func (g *Group) NewMember(name string, cacheSize int64, getter Getter, onEvicted func(key string, value lru.Value)) *Member {
	if getter == nil {
		panic("getter is nil")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	m := &Member{
		name:   name,
		getter: getter,
		mCache: lru.NewCache(cacheSize, onEvicted),
	}
	g.group[name] = m
	return m
}

func (g *Group) GetMember(name string) *Member {
	g.mu.RLock()
	m := g.group[name]
	g.mu.RUnlock()
	return m
}

// PeerGet 实现远程peer通过网络请求当前cache server的功能
// -----> 缓存中获取 ---yes--->  直接返回
//            | no
//             --------->  底层存储获取
func (m *Member)PeerGet(key string) (lru.ByteValue, error) {
	v, err := m.mCache.Get(key)
	if err == nil {
		fmt.Println("remote cache hit")
		return v, nil
	}
	return m.getLocally(key)
}


// Member的方法
func (m *Member) RegisterPeers(picker peer.PeerPicker) {
	if m.picker != nil {
		fmt.Printf("%v has picker aleady", m.name)
		return
	}
	m.picker = picker
}

// Get 流程：
// -------------> 本地Cache获取 ----------是-------------> 获取缓存的数据
//                     | 否
//                     |-----> 远程peer获取 --------是-------> 获取成功 -------->返回数据
//                                 | 否                         | 失败
//                                 |-----> 磁盘io读取数据 -----> io数据
func (m *Member) Get(key string) (lru.ByteValue, error) {
	if len(key) == 0 {
		return lru.ByteValue{}, errors.New("invalid key")
	}
	if v, err := m.mCache.Get(key); err == nil {
		fmt.Println("local cache hit")
		return v, nil
	}
	return m.load(key)
}

func (m *Member) load(key string) (lru.ByteValue, error) {
	view, err := m.loader.DoOnce(key, func() (interface{}, error) {
		if m.picker != nil {
			// 根据key在hash环上的映射，到peerGetter
			if peer, ok := m.picker.Pick(key); ok {
				// 从远程节点查当前节点(name)需要的key对应的数据
				resp, err := peer.Get(m.name, key)
				if err == nil {
					return resp, nil
				} else {
					util.Error("Failed to get from remote peer, "+err.Error())
				}
			}
		}
		return m.getLocally(key)
	})

	if err != nil {
		return lru.ByteValue{}, err
	} else {
		return view.(lru.ByteValue), nil
	}
}

// getLocally 实现在本地调用底层存储获取数据
func (m *Member) getLocally(key string) (lru.ByteValue, error) {
	 bytes, err := m.getter.Get(key)
	 if err != nil {
		return lru.ByteValue{}, err
	}

	value := lru.ByteValue{B: bytes}
	m.mCache.Add(key, value)
	return value, nil
}
