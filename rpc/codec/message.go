package codec

// message 提供一切和传输报文相关的定义

// InfoHeader 提供一些在C/S间协商rpc协议的字段
type InfoHeader struct {
	GobType string
}

// RequestHeader 保存rpc调用的一些参数
type RequestHeader struct {
	// 资源定位符: /__rpc__/namespace/key   ->    最终取的是key的value
	Name string
	Key string
	Err string
}
