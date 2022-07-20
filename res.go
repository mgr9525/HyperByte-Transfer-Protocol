package hbtp

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
)

type Response struct {
	conn   net.Conn
	code   int32
	heads  []byte
	bodys  []byte
	bdok   sync.Mutex
	bdln   uint32
	header *Map
}

func (c *Response) Header() (*Map, error) {
	if c.heads == nil {
		return nil, errors.New("is do?")
	}
	if c.header == nil {
		c.header = NewMap()
		err := json.Unmarshal(c.heads, &c.header)
		if err != nil {
			return nil, err
		}
	}
	return c.header, nil
}
func (c *Response) Code() int32 {
	return c.code
}
func (c *Response) HeadBytes() []byte {
	return c.heads
}
func (c *Response) BodyBytes(ctxs ...context.Context) []byte {
	c.bdok.Lock()
	defer c.bdok.Unlock()
	if c.bodys == nil && c.bdln > 0 {
		ctx := context.Background()
		if len(ctxs) > 0 && ctxs[0] != nil {
			ctx = ctxs[0]
		}
		bds, err := TcpRead(ctx, c.conn, uint(c.bdln))
		if err != nil {
			println("get_bodys tcp read err:" + err.Error())
		} else {
			c.bodys = bds
		}
	}
	return c.bodys
}
func (c *Response) BodyJson(bd interface{}) error {
	bds := c.BodyBytes(nil)
	if bds == nil {
		return errors.New("is do?")
	}
	return json.Unmarshal(bds, bd)
}
