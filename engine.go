package hbtp

import (
	"context"
	"fmt"
	"github.com/mgr9525/HyperByte-Transfer-Protocol/utils"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

//返回true则连接不会关闭
type ConnFun func(res *Context)

type Engine struct {
	ctx context.Context
	lsr *net.TCPListener

	fnlk sync.Mutex
	fns  map[int]ConnFun

	tmsHead time.Duration
	tmsBody time.Duration
	maxHead uint
	maxBody uint
}

func NewEngine(ctx context.Context) *Engine {
	et := &Engine{
		ctx: ctx,
		fns: make(map[int]ConnFun),

		tmsHead: conf.tmsHead,
		tmsBody: conf.tmsBody,
		maxHead: conf.maxHead,
		maxBody: conf.maxBody,
	}
	return et
}
func (c *Engine) Run(host string) error {
	addr, _ := net.ResolveTCPAddr("tcp", host)
	lsr, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	c.lsr = lsr
	//go func() {
	println("hbtp run on:" + host)
	for !utils.CheckContext(c.ctx) {
		c.runAcp()
		time.Sleep(time.Millisecond * 100)
	}
	//}()
	return nil
}
func (c *Engine) runAcp() {
	defer func() {
		if err := recover(); err != nil {
			println(fmt.Sprintf("Engine runAcp recover:%+v", err))
			println(fmt.Sprintf("%s", string(debug.Stack())))
		}
	}()
	if c.lsr == nil {
		time.Sleep(time.Millisecond)
		return
	}
	conn, err := c.lsr.AcceptTCP()
	if err != nil {
		println(fmt.Sprintf("runAcp AcceptTCP err:%+v", err))
		return
	}
	go c.handleConn(conn)
}
func (c *Engine) handleConn(conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			println(fmt.Sprintf("Engine runAcp recover:%+v", err))
			println(fmt.Sprintf("%s", string(debug.Stack())))
		}
	}()
	needclose := true
	defer func() {
		if needclose {
			conn.Close()
		}
	}()

	ctx, _ := context.WithTimeout(c.ctx, time.Second*5)
	bts, err := utils.TcpRead(ctx, conn, 2)
	if err != nil {
		println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
		return
	}
	if bts[0] != 0x8e || bts[1] != 0x8f {
		println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
		return
	}
	bts, err = utils.TcpRead(ctx, conn, 4)
	if err != nil {
		println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
		return
	}
	mcode := int(utils.BigByteToInt(bts))
	bts, err = utils.TcpRead(ctx, conn, 4)
	if err != nil {
		println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
		return
	}
	hdln := uint(utils.BigByteToInt(bts))
	if hdln > c.maxHead {
		println(fmt.Sprintf("Engine handleConn handleRead head size out max:%d/%d", hdln, c.maxHead))
		return
	}
	ctx, _ = context.WithTimeout(c.ctx, c.tmsHead)
	var hdbts []byte
	if hdln > 0 {
		hdbts, err = utils.TcpRead(ctx, conn, hdln)
		if err != nil {
			println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
			return
		}
	}

	bts, err = utils.TcpRead(ctx, conn, 4)
	if err != nil {
		println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
		return
	}
	bdln := uint(utils.BigByteToInt(bts))
	if bdln > c.maxBody {
		println(fmt.Sprintf("Engine handleConn handleRead body size out max:%d/%d", bdln, c.maxBody))
		return
	}
	ctx, _ = context.WithTimeout(c.ctx, c.tmsBody)
	var bdbts []byte
	if bdln > 0 {
		bdbts, err = utils.TcpRead(ctx, conn, bdln)
		if err != nil {
			println(fmt.Sprintf("Engine handleConn handleRead err:%+v", err))
			return
		}
	}

	needclose = c.recoverCallMapfn(mcode, &Context{
		clve: true,
		conn: conn,
		hds:  hdbts,
		bds:  bdbts,
	})
}
func (c *Engine) recoverCallMapfn(mcode int, res *Context) (rt bool) {
	rt = false
	defer func() {
		if err := recover(); err != nil {
			rt = false
			println(fmt.Sprintf("Engine recoverCallMapfn recover:%+v", err))
			println(fmt.Sprintf("%s", string(debug.Stack())))
		}
	}()

	c.fnlk.Lock()
	fn, ok := c.fns[mcode]
	c.fnlk.Unlock()
	if ok && fn != nil {
		fn(res)
	}
	return res.clve
}

func (c *Engine) RegFun(mcode int, fn ConnFun) bool {
	c.fnlk.Lock()
	defer c.fnlk.Unlock()
	_, ok := c.fns[mcode]
	if ok || fn == nil {
		println(fmt.Sprintf("Engine RegFun err:code(%d) is exist", mcode))
		return false
	}
	c.fns[mcode] = fn
	return true
}

func (c *Engine) SetMaxHeadLen(n uint) {
	c.maxHead = n
}
func (c *Engine) SetMaxBodyLen(n uint) {
	c.maxBody = n
}

func (c *Engine) ReadHeadTimeout(n time.Duration) {
	c.tmsHead = n
}
func (c *Engine) ReadBodyTimeout(n time.Duration) {
	c.tmsBody = n
}
