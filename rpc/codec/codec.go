package codec

import "io"

const (
	GOB = "gob"
	PROTOBUF = "protobuf"
)

type CodecFunc func(conn io.ReadWriteCloser) Codec

// Codec 编解码接口
type Codec interface {
	io.Closer
	ReadHeader(h *RequestHeader) error
	ReaderBody(b interface{}) error
	Write(h *RequestHeader, b interface{}) error
}

var CodecFunMap map[string]CodecFunc

func init() {
	CodecFunMap = make(map[string]CodecFunc)
	CodecFunMap[GOB] = NewGobCodec
	CodecFunMap[PROTOBUF] = NewProtoBuf
}