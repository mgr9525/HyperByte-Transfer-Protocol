package hbtp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mgr9525/HyperByte-Transfer-Protocol/utils"
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
	ctrls := utils.BigIntToByte(int64(c.fs), 4)
	if _, err := c.conn.Write(ctrls); err != nil {
		return err
	}
	hdln := utils.BigIntToByte(int64(len(hd)), 4)
	if _, err := c.conn.Write(hdln); err != nil {
		return err
	}
	if hd != nil {
		if _, err := c.conn.Write(hd); err != nil {
			return err
		}
	}
	contln := utils.BigIntToByte(int64(len(bds)), 4)
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
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	bts, err := utils.TcpRead(ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	c.code = int(utils.BigByteToInt(bts))
	bts, err = utils.TcpRead(ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	hln := uint(utils.BigByteToInt(bts))
	if hln > conf.maxHead {
		println(fmt.Sprintf("Request Res head size out max:%d/%d", hln, conf.maxHead))
		return errors.New("head len out max")
	}
	ctx, _ = context.WithTimeout(context.Background(), conf.tmsHead)
	var hdbts []byte
	if hln > 0 {
		hdbts, err = utils.TcpRead(ctx, c.conn, hln)
		if err != nil {
			return err
		}
	}
	bts, err = utils.TcpRead(ctx, c.conn, 4)
	if err != nil {
		println(fmt.Sprintf("Request Res err:%+v", err))
		return err
	}
	bln := uint(utils.BigByteToInt(bts))
	if bln > conf.maxBody {
		println(fmt.Sprintf("Request Res body size out max:%d/%d", bln, conf.maxBody))
		return errors.New("body len out max")
	}
	ctx, _ = context.WithTimeout(context.Background(), conf.tmsBody)
	var bdbts []byte
	if bln > 0 {
		bdbts, err = utils.TcpRead(ctx, c.conn, bln)
		if err != nil {
			return err
		}
	}
	c.hds = hdbts
	c.bds = bdbts
	return nil
}
func (c *Request) DoNoRes(body interface{}, hds ...[]byte) error {
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
	return c.send(bdbts, hds...)
}
func (c *Request) Do(body interface{}, hds ...[]byte) error {
	err := c.DoNoRes(body, hds...)
	if err != nil {
		return err
	}
	return c.Res()
}
func (c *Request) Conn() net.Conn {
	return c.conn
}
func (c *Request) GetConn() net.Conn {
	c.clve = false
	return c.conn
}
func (c *Request) Close() error {
	if c.clve {
		return c.conn.Close()
	}
	return nil
}
