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
	lmts  map[int32]*LmtMaxConfig
	notfn ConnFun

	lmtTm  *LmtTmConfig
	lmtMax *LmtMaxConfig
}

func NewEngine(ctx context.Context) *Engine {
	c := &Engine{
		fns:    make(map[int32][]ConnFun),
		lmts:   make(map[int32]*LmtMaxConfig),
		lmtTm:  MakeLmtTmCfg(),
		lmtMax: MakeLmtMaxCfg(),
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	return c
}
func (c *Engine) SetLmtTm(lmt *LmtTmConfig) {
	c.lmtTm = lmt
}
func (c *Engine) SetlmtMax(lmt *LmtMaxConfig) {
	c.lmtMax = lmt
}
func (c *Engine) GetLmtTm() *LmtTmConfig {
	return c.lmtTm
}
func (c *Engine) GetlmtMax(k int32) *LmtMaxConfig {
	rt, ok := c.lmts[k]
	if ok {
		return rt
	}
	return c.lmtMax
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
			Debugf("Engine handleConn recover:%+v", err)
			Debugf("%s", string(debug.Stack()))
		}
	}()

	res, err := ParseContext(c.ctx, conn, c)
	if err != nil {
		Debugf("Engine handleConn ParseContext err:%+v", err)
		conn.Close()
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

func (c *Engine) RegFun(control int32, fn ConnFun, lmtMax ...*LmtMaxConfig) bool {
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
	if len(lmtMax) > 0 {
		c.lmts[control] = lmtMax[0]
	}
	return true
}
func (c *Engine) RegParamFun(control int32, fn interface{}, lmtMax ...*LmtMaxConfig) bool {
	return c.RegFun(control, ParamFunHandle(fn), lmtMax...)
}
func (c *Engine) RegGrpcFun(control int32, rpc IRPCRoute, lmtMax ...*LmtMaxConfig) bool {
	return c.RegFun(control, GrpcFunHandle(rpc), lmtMax...)
}
