package peer

import (
	"simplecache/groupcache/lru"
)

// PeerPicker 通过某种方式pick到peer的接口，eg通过一致性hash获取到target peer的接口
type PeerPicker interface {
	Pick(key string) (peer PeerGetter, ok bool)
}

// 从远程peer查询peer缓存的接口，最核心参数就是namespace和key，编解码的细节交给codec完成
type PeerGetter interface {
	Get(name, key string) (lru.ByteValue, error)
}