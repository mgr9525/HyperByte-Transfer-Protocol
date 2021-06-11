package hbtp

import (
	"context"
	"net"
	"net/url"
	"runtime/debug"
	"sync"
	"time"
)

//返回true则连接不会关闭
type ConnFun func(res *Context)

type Engine struct {
	ctx  context.Context
	cncl context.CancelFunc
	lsr  net.Listener

	fnlk sync.Mutex
	fns  map[int32]ConnFun

	conf Config
}

func NewEngine(ctx context.Context) *Engine {
	c := &Engine{
		fns:  make(map[int32]ConnFun),
		conf: MakeConfig(),
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	return c
}
func (c *Engine) Config(conf Config) {
	c.conf = conf
}
func (c *Engine) Stop() {
	if c.lsr != nil {
		c.lsr.Close()
	}
	if c.cncl != nil {
		c.cncl()
		c.cncl = nil
	}
}
func (c *Engine) Run(host string) error {
	//addr, _ := net.ResolveTCPAddr("tcp", host)
	lsr, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	c.lsr = lsr
	//go func() {
	println("hbtp run on:" + host)
	for !EndContext(c.ctx) {
		c.runAcp()
	}
	//}()
	return nil
}
func (c *Engine) runAcp() {
	defer func() {
		if err := recover(); err != nil {
			Debugf("Engine runAcp recover:%+v", err)
			Debugf("%s", string(debug.Stack()))
		}
	}()
	if c.lsr == nil {
		time.Sleep(time.Millisecond * 100)
		return
	}
	conn, err := c.lsr.Accept()
	if err != nil {
		Debugf("runAcp AcceptTCP err:%+v", err)
		return
	}
	go c.handleConn(conn)
}
func (c *Engine) handleConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			Debugf("Engine runAcp recover:%+v", err)
			Debugf("%s", string(debug.Stack()))
		}
	}()
	needclose := true
	defer func() {
		if needclose {
			conn.Close()
		}
	}()

	info := &msgInfo{}
	infoln, _ := Size4Struct(info)
	ctx, _ := context.WithTimeout(c.ctx, c.conf.TmsInfo)
	bts, err := TcpRead(ctx, conn, uint(infoln))
	if err != nil {
		Debugf("Engine handleConn handleRead err:%+v", err)
		return
	}
	err = Byte2Struct(bts, info)
	if err != nil {
		return
	}
	rtctx := &Context{
		clve:    true,
		conn:    conn,
		control: info.Control,
	}
	ctx, _ = context.WithTimeout(c.ctx, c.conf.TmsHead)
	if info.LenCmd > 0 {
		bts, err = TcpRead(ctx, conn, uint(info.LenCmd))
		if err != nil {
			Debugf("Engine handleConn handleRead err:%+v", err)
			return
		}
		rtctx.cmd = string(bts)
	}
	if info.LenArg > 0 {
		bts, err = TcpRead(ctx, conn, uint(info.LenArg))
		if err != nil {
			Debugf("Engine handleConn handleRead err:%+v", err)
			return
		}
		args, err := url.ParseQuery(string(bts))
		if err == nil {
			rtctx.args = args
		}
	}
	if info.LenHead > 0 {
		rtctx.hds, err = TcpRead(ctx, conn, uint(info.LenHead))
		if err != nil {
			Debugf("Engine handleConn handleRead err:%+v", err)
			return
		}
	}

	ctx, _ = context.WithTimeout(c.ctx, c.conf.TmsBody)
	if info.LenBody > 0 {
		rtctx.bds, err = TcpRead(ctx, conn, uint(info.LenBody))
		if err != nil {
			Debugf("Engine handleConn handleRead err:%+v", err)
			return
		}
	}

	needclose = c.recoverCallMapfn(rtctx)
}
func (c *Engine) recoverCallMapfn(res *Context) (rt bool) {
	rt = false
	defer func() {
		if err := recover(); err != nil {
			rt = false
			Debugf("Engine recoverCallMapfn recover:%+v", err)
			Debugf("%s", string(debug.Stack()))
		}
	}()

	c.fnlk.Lock()
	fn, ok := c.fns[res.control]
	c.fnlk.Unlock()
	if ok && fn != nil {
		fn(res)
	}
	return res.clve
}

func (c *Engine) RegFun(control int32, fn ConnFun) bool {
	c.fnlk.Lock()
	defer c.fnlk.Unlock()
	_, ok := c.fns[control]
	if ok || fn == nil {
		Debugf("Engine RegFun err:control(%d) is exist", control)
		return false
	}
	c.fns[control] = fn
	return true
}
func (c *Engine) RegParamFun(control int32, fn interface{}) bool {
	return c.RegFun(control, paramFunHandle(fn))
}
func (c *Engine) RegGrpcFun(control int32, rpc IRPCRoute) bool {
	return c.RegFun(control, grpcFunHandle(rpc))
}
