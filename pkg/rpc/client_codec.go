package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"net/rpc"

	"github.com/vmihailenco/msgpack/v5"
)

type clientCodecMsgpack struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
}

func NewClientCodecMsgpack(conn io.ReadWriteCloser) clientCodecMsgpack {
	buf := bufio.NewWriter(conn)
	return clientCodecMsgpack{
		rwc:    conn,
		dec:    msgpack.NewDecoder(conn),
		enc:    msgpack.NewEncoder(buf),
		encBuf: buf,
	}
}

func (c clientCodecMsgpack) WriteRequest(r *rpc.Request, body any) error {
	if err := c.enc.Encode(r); err != nil {
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		return err
	}
	return c.encBuf.Flush()
}

func (c clientCodecMsgpack) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c clientCodecMsgpack) ReadResponseBody(body any) error {
	return c.dec.Decode(body)
}

func (c clientCodecMsgpack) Close() error {
	return c.rwc.Close()
}

type clientCodecGob struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func NewClientCodecGob(conn io.ReadWriteCloser) clientCodecGob {
	buf := bufio.NewWriter(conn)
	return clientCodecGob{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}
}

func (c clientCodecGob) WriteRequest(r *rpc.Request, body any) error {
	if err := c.enc.Encode(r); err != nil {
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		return err
	}
	return c.encBuf.Flush()
}

func (c clientCodecGob) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c clientCodecGob) ReadResponseBody(body any) error {
	return c.dec.Decode(body)
}

func (c clientCodecGob) Close() error {
	return c.rwc.Close()
}
