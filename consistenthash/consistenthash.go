// @Description: 一致性hash算法实现请求的负载均衡
// @Author: tao
// @Update: 2021/10/10 18:44
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 依赖注入，允许自定义的哈希
type Hash func(data []byte) uint32

type Map struct {
	// 计算节点的hash值
	hash Hash
	// 虚拟节点倍数，实际中受节点的性能影响，不一定都是一个相同的值
	replicas int
	// 哈希环
	vKeys []int
	// 虚拟节点和真实节点的映射
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加一系列节点，并映射成replicas个虚拟节点
func (m* Map) Add(keys ...string)  {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			vHash := int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.vKeys = append(m.vKeys, vHash)
			m.hashMap[vHash] = key
		}
	}
	sort.Ints(m.vKeys)
}

// 根据key获得hash环上的最合适的节点
func (m *Map) Get(key string) string {
	if len(m.vKeys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.vKeys), func(i int) bool {
		return m.vKeys[i] >= hash
	})

	return m.hashMap[m.vKeys[idx%len(m.vKeys)]]
}