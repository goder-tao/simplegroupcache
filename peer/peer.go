package peer

// 远程节点处获取，找到PeerGetter，每个peer有一个独一无二的key --> ip地址
type PeerPicker interface {
	Pick(key string) (peer PeerGetter, ok bool)
}

//
type PeerGetter interface {
	Get(member string, key string) ([]byte, error)
}
