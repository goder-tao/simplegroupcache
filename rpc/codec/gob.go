package codec

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	enc *gob.Encoder
	dec *gob.Decoder
	buf *bufio.Writer
}

var _ Codec = (*GobCodec)(nil)

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(h *RequestHeader) error {
	return g.dec.Decode(h)
}

func (g *GobCodec) ReaderBody(b interface{}) error {
	return g.dec.Decode(b)
}

func (g *GobCodec) Write(h *RequestHeader, b interface{}) error {
	defer func() {
		err := g.buf.Flush()
		if err != nil{
			_ = g.conn.Close()
		}
	}()
	err := g.enc.Encode(h)
	if err != nil {
		return err
	}
	err = g.enc.Encode(b)
	return err
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf: buf,
		enc: gob.NewEncoder(bufio.NewWriter(buf)),
		dec: gob.NewDecoder(bufio.NewReader(conn)),
	}
}

