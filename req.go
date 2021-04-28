package hbtp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

type Request struct {
	clve bool
	conn net.Conn
	fs   int
	code int
	hds  []byte
	bds  []byte

	ctx  context.Context
	cncl context.CancelFunc

	hdr  *Header
	hdrs *Header
}

func (c *Request) ReqHeader() *Header {
	if c.hdr != nil {
		return c.hdr
	}
	c.hdr = &Header{}
	return c.hdr
}
func (c *Request) ResHeader() (*Header, error) {
	if c.hds == nil {
		return nil, errors.New("is do?")
	}
	if c.hdrs != nil {
		return c.hdrs, nil
	}
	hdr, err := ParseHeader(c.hds)
	if err != nil {
		return nil, err
	}
	c.hdrs = hdr
	return hdr, nil
}
func (c *Request) ResCode() int {
	return c.code
}
func (c *Request) ResHeadBytes() []byte {
	return c.hds
}
func (c *Request) ResBodyBytes() []byte {
	return c.bds
}
func (c *Request) ResBodyJson(bd interface{}) error {
	if c.bds == nil {
		return errors.New("is do?")
	}
	return json.Unmarshal(c.bds, bd)
}

func NewRequest(addr string, fs int, timeout ...time.Duration) (*Request, error) {
	tmo := time.Second * 5
	if len(timeout) > 0 {
		tmo = timeout[0]
	}
	conn, err := net.DialTimeout("tcp", addr, tmo)
	if err != nil {
		return nil, err
	}
	cli := &Request{
		clve: true,
		conn: conn,
		fs:   fs,
	}
	//cli.handleConn()
	return cli, nil
}
func NewRPCReq(addr string, fs int, path string, timeout ...time.Duration) (*Request, error) {
	req, err := NewRequest(addr, fs, timeout...)
	if err != nil {
		return nil, err
	}
	req.ReqHeader().Path = path
	return req, nil
}
func (c *Request) SetContext(ctx context.Context) {
	if ctx == nil {
		return
	}
	c.ctx = ctx
	c.cncl = nil
}
func (c *Request) send(bds []byte, hds ...[]byte) error {
	var hd []byte
	if len(hds) > 0 {
		hd = hds[0]
	} else if c.hdr != nil {
		hd = c.hdr.Bytes()
	}

	_, err := c.conn.Write([]byte{0x8e, 0x8f})
	if err != nil {
		return err
	}
	ctrls := BigIntToByte(int64(c.fs), 4)
	hdln := BigIntToByte(int64(len(hd)), 4)
	contln := BigIntToByte(int64(len(bds)), 4)
	if _, err := c.conn.Write(ctrls); err != nil {
		return err
	}
	if _, err := c.conn.Write(hdln); err != nil {
		return err
	}
	if EndContext(c.ctx) {
		return errors.New("context dead")
	}
	if hd != nil {
		if _, err := c.conn.Write(hd); err != nil {
			return err
		}
	}
	if EndContext(c.ctx) {
		return errors.New("context dead")
	}
	if _, err := c.conn.Write(contln); err != nil {
		return err
	}
	if bds != nil {
		if _, err := c.conn.Write(bds); err != nil {
			return err
		}
	}
	return nil
}
func (c *Request) Res() error {
	if c.ctx == nil {
		return errors.New("need do some thing")
	}
	bts, err := TcpRead(c.ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	c.code = int(BigByteToInt(bts))
	bts, err = TcpRead(c.ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	hln := uint(BigByteToInt(bts))
	if hln > conf.maxHead {
		println(fmt.Sprintf("Request Res head size out max:%d/%d", hln, conf.maxHead))
		return errors.New("head len out max")
	}
	var hdbts []byte
	if hln > 0 {
		hdbts, err = TcpRead(c.ctx, c.conn, hln)
		if err != nil {
			return err
		}
	}
	bts, err = TcpRead(c.ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	bln := uint(BigByteToInt(bts))
	if bln > conf.maxBody {
		println(fmt.Sprintf("Request Res body size out max:%d/%d", bln, conf.maxBody))
		return errors.New("body len out max")
	}
	var bdbts []byte
	if bln > 0 {
		bdbts, err = TcpRead(c.ctx, c.conn, bln)
		if err != nil {
			return err
		}
	}
	c.hds = hdbts
	c.bds = bdbts
	return nil
}
func (c *Request) DoNoRes(ctx context.Context, body interface{}, hds ...[]byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx, c.cncl = context.WithCancel(ctx)

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
func (c *Request) Do(ctx context.Context, body interface{}, hds ...[]byte) error {
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
		c.clve = !ownership[0]
	}
	return c.conn
}
func (c *Request) Close() error {
	if c.clve {
		return c.conn.Close()
	}
	if c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
	return nil
}
