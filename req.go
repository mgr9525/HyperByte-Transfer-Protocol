package hbtp

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"time"
)

type Request struct {
	conf    Config
	addr    string
	timeout time.Duration
	conn    net.Conn
	sended  bool
	control int32
	cmd     string
	args    url.Values

	ctx  context.Context
	cncl context.CancelFunc

	header *Map
}

func (c *Request) Header() *Map {
	if c.header == nil {
		c.header = NewMap()
	}
	return c.header
}

func NewRequest(addr string, control int32, timeout ...time.Duration) *Request {
	tmo := time.Second * 30
	if len(timeout) > 0 {
		tmo = timeout[0]
	}
	cli := &Request{
		conf:    MakeConfig(),
		addr:    addr,
		timeout: tmo,
		control: control,
	}
	return cli
}
func NewConnRequest(conn net.Conn, control int32, timeout ...time.Duration) *Request {
	tmo := time.Second * 30
	if len(timeout) > 0 {
		tmo = timeout[0]
	}
	cli := &Request{
		conf:    MakeConfig(),
		conn:    conn,
		timeout: tmo,
		control: control,
	}
	return cli
}
func (c *Request) Config(conf Config) *Request {
	c.conf = conf
	return c
}
func (c *Request) SetContext(ctx context.Context) *Request {
	if ctx == nil {
		return c
	}
	c.ctx = ctx
	c.cncl = nil
	return c
}
func (c *Request) Timeout(tmo time.Duration) *Request {
	c.timeout = tmo
	return c
}
func (c *Request) Command(cmd string) *Request {
	c.cmd = cmd
	return c
}
func (c *Request) Args(args url.Values) *Request {
	c.args = args
	return c
}
func (c *Request) SetArg(k, v string) *Request {
	if c.args == nil {
		c.args = url.Values{}
	}
	c.args.Set(k, v)
	return c
}
func (c *Request) write(bts []byte) (int, error) {
	if c.ctx == nil || c.conn == nil {
		return 0, errors.New("do is call?")
	}
	n, err := c.conn.Write(bts)
	if err != nil {
		return 0, err
	}
	if EndContext(c.ctx) {
		return 0, errors.New("ctx is end")
	}
	return n, nil
}
func (c *Request) send(bds []byte, hds ...interface{}) error {
	if c.sended {
		return errors.New("already send")
	}
	var err error
	if c.conn == nil {
		if c.addr == "" {
			return errors.New("addr is empty")
		}
		c.conn, err = net.DialTimeout("tcp", c.addr, c.conf.TmsInfo)
		if err != nil {
			return err
		}
	}
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	c.ctx, c.cncl = context.WithTimeout(c.ctx, c.timeout)

	var hd []byte
	if len(hds) > 0 {
		switch hds[0].(type) {
		case []byte:
			hd = hds[0].([]byte)
		case string:
			hd = []byte(hds[0].(string))
		default:
			hd, err = json.Marshal(hds[0])
			if err != nil {
				return err
			}
		}
	} else if c.header != nil {
		hd = c.header.ToBytes()
	}

	var args string
	if c.args != nil {
		args = c.args.Encode()
	}
	info := &msgInfo{
		Version: 1,
		Control: c.control,
		LenCmd:  uint16(len(c.cmd)),
		LenArg:  uint16(len(args)),
		LenHead: uint32(len(hd)),
		LenBody: uint32(len(bds)),
	}
	bts, err := FlcStruct2Byte(info)
	if err != nil {
		return err
	}
	c.sended = true
	_, err = c.write(bts)
	if err != nil {
		return err
	}
	if info.LenCmd > 0 {
		_, err = c.write([]byte(c.cmd))
		if err != nil {
			return err
		}
	}
	if info.LenArg > 0 {
		_, err = c.write([]byte(args))
		if err != nil {
			return err
		}
	}
	if info.LenHead > 0 {
		_, err = c.write(hd)
		if err != nil {
			return err
		}
	}
	if info.LenBody > 0 {
		_, err = c.write(bds)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Request) Res() (*Response, error) {
	if !c.sended {
		return nil, errors.New("not send")
	}
	if c.ctx == nil || c.conn == nil {
		return nil, errors.New("need do some thing")
	}
	info := &resInfoV1{}
	infoln := FlcStructSizeof(info)
	bts, err := TcpRead(c.ctx, c.conn, uint(infoln))
	if err != nil {
		return nil, err
	}
	err = FlcByte2Struct(bts, info)
	if err != nil {
		return nil, err
	}
	if uint64(info.LenHead) > MaxHeads {
		return nil, errors.New("bytes2 out limit!!")
	}
	if uint64(info.LenBody) > MaxBodys {
		return nil, errors.New("bytes3 out limit!!")
	}
	rt := &Response{code: info.Code}
	if info.LenHead > 0 {
		rt.heads, err = TcpRead(c.ctx, c.conn, uint(info.LenHead))
		if err != nil {
			return nil, err
		}
	}
	if info.LenBody > 0 {
		rt.bodys, err = TcpRead(c.ctx, c.conn, uint(info.LenBody))
		if err != nil {
			return nil, err
		}
	}
	return rt, nil
}
func (c *Request) DoNoRes(ctx context.Context, body interface{}, hds ...interface{}) error {
	c.ctx = ctx
	var err error
	var bdbts []byte
	if body != nil {
		switch body.(type) {
		case []byte:
			bdbts = body.([]byte)
		case string:
			bdbts = []byte(body.(string))
		default:
			bdbts, err = json.Marshal(body)
			if err != nil {
				return err
			}
		}
	}
	return c.send(bdbts, hds...)
}
func (c *Request) Do(ctx context.Context, body interface{}, hds ...interface{}) (*Response, error) {
	err := c.DoNoRes(ctx, body, hds...)
	if err != nil {
		return nil, err
	}
	return c.Res()
}

/*
	if ownership is `true`,the conn is never close!
	so you need close manual.
*/
func (c *Request) Conn(ownership ...bool) net.Conn {
	defer func() {
		if len(ownership) > 0 && ownership[0] {
			c.conn = nil
		}
	}()
	return c.conn
}
func (c *Request) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	if c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
	return nil
}
