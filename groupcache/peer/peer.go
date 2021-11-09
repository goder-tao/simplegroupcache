package peer

import "simplecache/groupcache/pb"

// 远程节点处获取，找到PeerGetter，每个peer有一个独一无二的key --> ip地址
type PeerPicker interface {
	Pick(key string) (peer PeerGetter, ok bool)
}

// 通过http调用远程peer查询peer缓存的接口
type PeerGetter interface {
	Get(req *pb.Request) (*pb.Response, error)
}
