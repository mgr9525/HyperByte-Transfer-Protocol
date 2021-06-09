package hbtp

import (
	"encoding/json"
	"errors"
	"net"
)

type Context struct {
	clve bool
	conn net.Conn
	code int
	hds  []byte
	bds  []byte

	hdr  *Header
	hdrs *Header

	data Mp
}

func (c *Context) Conn(ownership ...bool) net.Conn {
	if len(ownership) > 0 {
		c.clve = !ownership[0]
	}
	return c.conn
}
func (c *Context) ReqHeader() (*Header, error) {
	if c.hds == nil {
		return nil, errors.New("no head byte")
	}
	if c.hdr != nil {
		return c.hdr, nil
	}
	hdr, err := ParseHeader(c.hds)
	if err != nil {
		return nil, err
	}
	c.hdr = hdr
	return hdr, nil
}
func (c *Context) ResHeader() *Header {
	if c.hdrs != nil {
		return c.hdrs
	}
	c.hdrs = &Header{}
	return c.hdrs
}
func (c *Context) Code() int {
	return c.code
}
func (c *Context) HeadBytes() []byte {
	return c.hds
}
func (c *Context) BodyBytes() []byte {
	return c.bds
}

func (c *Context) response(control int, hds []byte, conts []byte) error {
	ctrls := BigIntToByte(int64(control), 4)
	//wtr:=bufio.NewWriter(conn)
	_, err := c.conn.Write(ctrls)
	if err != nil {
		return err
	}
	hdln := BigIntToByte(int64(len(hds)), 4)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(hdln)
	if hds != nil {
		_, err = c.conn.Write(hds)
		if err != nil {
			return err
		}
	}
	contln := BigIntToByte(int64(len(conts)), 4)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(contln)
	if conts != nil {
		_, err = c.conn.Write(conts)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Context) ResBytes(code int, bdbts []byte, hds ...[]byte) error {
	var hdbts []byte
	if len(hds) > 0 {
		hdbts = hds[0]
	} else if c.hdrs != nil {
		hdbts = c.hdrs.Bytes()
	}
	return c.response(code, hdbts, bdbts)
}
func (c *Context) ResString(code int, s string, hds ...[]byte) error {
	return c.ResBytes(code, []byte(s), hds...)
}

func (c *Context) ResJson(code int, body interface{}, hds ...[]byte) error {
	var err error
	var bdbts []byte
	if body != nil {
		switch body.(type) {
		case []byte:
			bdbts = body.([]byte)
		default:
			bdbts, err = json.Marshal(body)
			if err != nil {
				return err
			}
		}
	}
	return c.ResBytes(code, bdbts, hds...)
}
func (c *Context) SetData(k string, data interface{}) {
	if c.data == nil {
		c.data = Mp{}
	}
	c.data[k] = data
}
func (c *Context) GetData(k string) (interface{}, bool) {
	if c.data == nil {
		c.data = Mp{}
	}
	v, ok := c.data[k]
	return v, ok
}
