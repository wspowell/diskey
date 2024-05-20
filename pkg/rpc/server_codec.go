package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net/rpc"

	"github.com/vmihailenco/msgpack/v5"
)

type serverCodecMsgpack struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
	closed bool
}

func NewServerCodecMsgpack(conn io.ReadWriteCloser) *serverCodecMsgpack {
	buf := bufio.NewWriter(conn)
	return &serverCodecMsgpack{
		rwc:    conn,
		dec:    msgpack.NewDecoder(conn),
		enc:    msgpack.NewEncoder(buf),
		encBuf: buf,
	}
}

func (c *serverCodecMsgpack) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *serverCodecMsgpack) ReadRequestBody(body any) error {
	return c.dec.Decode(body)
}

func (c *serverCodecMsgpack) WriteResponse(r *rpc.Response, body any) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Msgpack couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: msgpack error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a msgpack problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: msgpack error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *serverCodecMsgpack) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

// Modified from net/rpc/server.go:serverCodecGob.
type serverCodecGob struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func NewServerCodecGob(conn io.ReadWriteCloser) *serverCodecGob {
	buf := bufio.NewWriter(conn)
	return &serverCodecGob{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
	}
}

func (c *serverCodecGob) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *serverCodecGob) ReadRequestBody(body any) error {
	return c.dec.Decode(body)
}

func (c *serverCodecGob) WriteResponse(r *rpc.Response, body any) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a gob problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *serverCodecGob) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}
