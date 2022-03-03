package hbtp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
)

type Context struct {
	sended  bool
	conn    net.Conn
	control int32
	cmd     string
	args    url.Values
	hds     []byte
	bds     []byte

	hdrq *Map
	hdrs *Map
	data *Map
}

func (c *Context) Sended() bool {
	return c.sended
}
func (c *Context) Conn(ownership ...bool) net.Conn {
	defer func() {
		if len(ownership) > 0 && ownership[0] {
			c.conn = nil
		}
	}()
	return c.conn
}
func (c *Context) ReqHeader() *Map {
	if c.hdrq == nil {
		c.hdrq = NewMaps(c.hds)
	}
	return c.hdrq
}
func (c *Context) ResHeader() *Map {
	if c.hdrs == nil {
		c.hdrs = NewMap()
	}
	return c.hdrs
}
func (c *Context) Control() int32 {
	return c.control
}
func (c *Context) Command() string {
	return c.cmd
}
func (c *Context) Args() url.Values {
	return c.args
}
func (c *Context) HeadBytes() []byte {
	return c.hds
}
func (c *Context) BodyBytes() []byte {
	return c.bds
}

func (c *Context) response(code int32, hds []byte, bds []byte) error {
	if c.sended {
		return errors.New("already send")
	}
	info := &resInfoV1{
		Code:    code,
		LenHead: uint32(len(hds)),
		LenBody: uint32(len(bds)),
	}
	bts, err := FlcStruct2Byte(info)
	if err != nil {
		return err
	}
	c.sended = true
	ctx := context.Background()
	err = TcpWrite(ctx, c.conn, bts)
	if err != nil {
		return err
	}
	if info.LenHead > 0 {
		err = TcpWrite(ctx, c.conn, hds)
		if err != nil {
			return err
		}
	}
	if info.LenBody > 0 {
		err = TcpWrite(ctx, c.conn, bds)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Context) ResBytes(code int32, bdbts []byte, hds ...[]byte) error {
	var hdbts []byte
	if len(hds) > 0 {
		hdbts = hds[0]
	} else if c.hdrs != nil {
		hdbts = c.hdrs.ToBytes()
	}
	return c.response(code, hdbts, bdbts)
}
func (c *Context) ResString(code int32, s string, hds ...[]byte) error {
	return c.ResBytes(code, []byte(s), hds...)
}
func (c *Context) ResStringf(code int32, s string, o ...interface{}) error {
	if len(o) > 0 {
		s = fmt.Sprintf(s, o...)
	}
	return c.ResBytes(code, []byte(s))
}

func (c *Context) ResJson(code int32, body interface{}, hds ...[]byte) error {
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
		c.data = NewMap()
	}
	c.data.Set(k, data)
}
func (c *Context) GetData(k string) (interface{}, bool) {
	if c.data == nil {
		c.data = NewMap()
	}
	return c.data.Get(k)
}

func ParseContext(ctx context.Context, conn net.Conn, egn *Engine) (*Context, error) {
	info := &msgInfo{}
	infoln := FlcStructSizeof(info)
	lmtm := egn.GetLmtTm()
	ctx, _ = context.WithTimeout(ctx, lmtm.TmOhther)
	bts, err := TcpRead(ctx, conn, uint(infoln))
	if err != nil {
		return nil, err
	}
	err = FlcByte2Struct(bts, info)
	if err != nil {
		return nil, err
	}
	if info.Version != 1 {
		return nil, errors.New("not found version")
	}
	lmtx := egn.GetlmtMax(info.Control)
	if uint64(info.LenCmd+info.LenArg) > lmtx.MaxOhther {
		return nil, errors.New("bytes1 out limit!!")
	}
	if uint64(info.LenHead) > lmtx.MaxHeads {
		return nil, errors.New("bytes2 out limit!!")
	}
	if uint64(info.LenBody) > lmtx.MaxBodys {
		return nil, errors.New("bytes3 out limit!!")
	}
	rt := &Context{
		conn:    conn,
		control: info.Control,
	}
	ctx, _ = context.WithTimeout(ctx, lmtm.TmHeads)
	if info.LenCmd > 0 {
		bts, err = TcpRead(ctx, conn, uint(info.LenCmd))
		if err != nil {
			return nil, err
		}
		rt.cmd = string(bts)
	}
	if info.LenArg > 0 {
		bts, err = TcpRead(ctx, conn, uint(info.LenArg))
		if err != nil {
			return nil, err
		}
		args, err := url.ParseQuery(string(bts))
		if err == nil {
			rt.args = args
		}
	}
	if info.LenHead > 0 {
		rt.hds, err = TcpRead(ctx, conn, uint(info.LenHead))
		if err != nil {
			return nil, err
		}
	}

	ctx, _ = context.WithTimeout(ctx, lmtm.TmBodys)
	if info.LenBody > 0 {
		rt.bds, err = TcpRead(ctx, conn, uint(info.LenBody))
		if err != nil {
			return nil, err
		}
	}
	return rt, nil
}
