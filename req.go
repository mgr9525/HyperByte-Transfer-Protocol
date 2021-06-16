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
	own     bool
	addr    string
	timeout time.Duration
	conn    net.Conn
	control int32
	cmd     string
	args    url.Values

	code  int32
	hdsrs []byte
	bdsrs []byte

	ctx  context.Context
	cncl context.CancelFunc

	hdrq *Map
	hdrs *Map

	started time.Time
}

func (c *Request) ReqHeader() *Map {
	if c.hdrq == nil {
		c.hdrq = NewMap()
	}
	return c.hdrq
}
func (c *Request) ResHeader() (*Map, error) {
	if c.hdsrs == nil {
		return nil, errors.New("is do?")
	}
	if c.hdrs == nil {
		c.hdrs = NewMaps(c.hdsrs)
	}
	return c.hdrs, nil
}
func (c *Request) ResCode() int32 {
	return c.code
}
func (c *Request) ResHeadBytes() []byte {
	return c.hdsrs
}
func (c *Request) ResBodyBytes() []byte {
	return c.bdsrs
}
func (c *Request) ResBodyJson(bd interface{}) error {
	if c.bdsrs == nil {
		return errors.New("is do?")
	}
	return json.Unmarshal(c.bdsrs, bd)
}

func NewRequest(addr string, control int32, timeout ...time.Duration) *Request {
	tmo := time.Second * 30
	if len(timeout) > 0 {
		tmo = timeout[0]
	}
	cli := &Request{
		conf:    MakeConfig(),
		own:     true,
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
		own:     true,
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
	var err error
	c.started = time.Now()
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
	} else if c.hdrq != nil {
		hd = c.hdrq.ToBytes()
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
	bts, err := Struct2ByteLen(info, lenMsgInfo)
	if err != nil {
		return err
	}
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
func (c *Request) Res() error {
	if c.ctx == nil || c.conn == nil {
		return errors.New("need do some thing")
	}
	info := &resInfoV1{}
	infoln := SizeOf(info)
	bts, err := TcpRead(c.ctx, c.conn, uint(infoln))
	if err != nil {
		return err
	}
	err = Byte2Struct(bts, info)
	if err != nil {
		return err
	}
	c.code = info.Code
	if info.LenHead > 0 {
		c.hdsrs, err = TcpRead(c.ctx, c.conn, uint(info.LenHead))
		if err != nil {
			return err
		}
	}
	if info.LenBody > 0 {
		c.bdsrs, err = TcpRead(c.ctx, c.conn, uint(info.LenBody))
		if err != nil {
			return err
		}
	}
	return nil
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
func (c *Request) Do(ctx context.Context, body interface{}, hds ...interface{}) error {
	err := c.DoNoRes(ctx, body, hds...)
	if err != nil {
		return err
	}
	return c.Res()
}

/*
	if ownership is `true`,the conn is never close!
	so you need close manual.
*/
func (c *Request) Conn(ownership ...bool) net.Conn {
	if len(ownership) > 0 {
		c.own = !ownership[0]
	}
	return c.conn
}
func (c *Request) Close() error {
	if c.own && c.conn != nil {
		return c.conn.Close()
	}
	if c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
	return nil
}
