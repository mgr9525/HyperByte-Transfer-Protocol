package hbtp

import (
	"context"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

type ConnFun func(res *Context)

type Engine struct {
	ctx  context.Context
	cncl context.CancelFunc
	lsr  net.Listener

	fnlk  sync.Mutex
	fns   map[int32][]ConnFun
	notfn ConnFun

	conf Config
}

func NewEngine(ctx context.Context) *Engine {
	c := &Engine{
		fns:  make(map[int32][]ConnFun),
		conf: MakeConfig(),
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	return c
}
func (c *Engine) Config(conf Config) {
	c.conf = conf
}
func (c *Engine) NotFoundFun(fn ConnFun) {
	c.notfn = fn
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
	Infof("hbtp run on:%s", host)
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

	res, err := ParseContext(c.ctx, conn, c.conf)
	if err != nil {
		Debugf("Engine handleConn ParseContext err:%+v", err)
		return
	}
	c.recoverCallMapfn(res)
	if res.conn != nil {
		res.conn.Close()
	}
}
func (c *Engine) recoverCallMapfn(res *Context) {
	defer func() {
		if err := recover(); err != nil {
			Debugf("Engine recoverCallMapfn recover:%+v", err)
			Debugf("%s", string(debug.Stack()))
		}
	}()

	c.fnlk.Lock()
	fns, ok := c.fns[res.Control()]
	c.fnlk.Unlock()
	if ok {
		for _, fn := range fns {
			if res.Sended() {
				break
			}
			fn(res)
		}
	} else if c.notfn != nil {
		c.notfn(res)
	} else {
		res.ResString(ResStatusNotFound, "Not Found Control Function")
	}
	if !res.Sended() {
		res.ResString(ResStatusErr, "Unknown")
	}
}

func (c *Engine) RegFun(control int32, fn ConnFun) bool {
	c.fnlk.Lock()
	defer c.fnlk.Unlock()
	_, ok := c.fns[control]
	if ok || fn == nil {
		Debugf("Engine RegFun err:control(%d) is exist", control)
		return false
	}
	fns := c.fns[control]
	fns = append(fns, fn)
	c.fns[control] = fns
	return true
}
func (c *Engine) RegParamFun(control int32, fn interface{}) bool {
	return c.RegFun(control, ParamFunHandle(fn))
}
func (c *Engine) RegGrpcFun(control int32, rpc IRPCRoute) bool {
	return c.RegFun(control, GrpcFunHandle(rpc))
}
