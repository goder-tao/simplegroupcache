package codec

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"io"
	"simplecache/groupcache/pb"
)

type ProtoBufCodec struct {
	conn io.ReadWriteCloser
	isClosed bool
}

var _ Codec = (*ProtoBufCodec)(nil)

func (p ProtoBufCodec) Close() error {
	return p.conn.Close()
}

func (p ProtoBufCodec) ReadHeader(h *RequestHeader) error {
	rsp := &pb.RPCRequest{}
	if err := proto.Unmarshal(nil, rsp); err != nil {
		return err
	}
	if len(rsp.Err) != 0 {
		return errors.New(rsp.Err)
	}
	h.Name = rsp.Name
	h.Key = rsp.Key
	h.Err = rsp.Err
	return nil
}

func (p ProtoBufCodec) ReaderBody(b interface{}) error {
	rsp := &pb.RPCResponse{}
	if err := proto.Unmarshal(nil, rsp); err != nil {
		return err
	}
	b = rsp.Value
	return nil
}

func (p ProtoBufCodec) Write(h *RequestHeader, b interface{}) error {
	prt := &pb.Request{
		Member: h.Name,
		Key: h.Key,
	}
	binData, err := proto.Marshal(prt)
	if err != nil{
		return err
	}
	_, err = p.conn.Write(binData)
	if err != nil {
		return err
	}
	return nil
}

func (p ProtoBufCodec)IsClosed() bool {
	return p.isClosed
}

func NewProtoBuf(conn io.ReadWriteCloser) Codec {
	return ProtoBufCodec{
		conn: conn,
		isClosed: false,
	}
}