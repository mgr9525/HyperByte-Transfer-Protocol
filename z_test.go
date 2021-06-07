package hbtp

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	eg := NewEngine(context.Background())
	eg.RegFun(1, ParamFunHandle(helloWorldFun))
	eg.RegFun(2, RPCFunHandle(&testFuns{}))
	go func() {
		time.Sleep(time.Second * 10)
		eg.Stop()
	}()
	if err := eg.Run(":7030"); err != nil {
		println("engine run err:", err.Error())
	}
}

func helloWorldFun(c *Context, hdr *Header) {
	host, _ := hdr.GetString("host")
	println("helloWorldFun header host:" + host)
	println("helloWorldFun body:" + string(c.BodyBytes()))
	c.ResHeader().Set("cookie", "1234567")
	c.ResString(ResStatusOk, "ok")
}

func TestRequest(t *testing.T) {
	req, err := NewRequest("localhost:7030", 2)
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	//req.ReqHeader().Set("host", "test.host.com")
	err = req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("req code:", req.ResCode())
	hdr, err := req.ResHeader()
	if err == nil {
		cookie, _ := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("req body:", string(req.ResBodyBytes()))
}

type testFuns struct {
	tstb *map[string]string
}

func (*testFuns) AuthFun() AuthFun {
	return func(c *Context) bool {
		//println("call testFuns.AuthFun")
		hdr, err := c.ReqHeader()
		if err != nil {
			c.ResString(ResStatusErr, "head err")
			return false
		}
		if hdr.Token != "123456" {
			c.ResString(ResStatusErr, "token err:"+hdr.Token)
			return false
		}
		return true
	}
}
func (e *testFuns) GetName1(c *Context, hdr *Header, body string) {
	mp := make(map[string]string)
	mp["aaa"] = "bbb"
	e.tstb = &mp
	println("call testFuns.GetName1:", e.tstb)
	c.ResHeader().Set("cookie", body)
	c.ResString(ResStatusOk, "ok")
}
func (e *testFuns) GetName2(c *Context, hdr *Header) {
	println("call testFuns.GetName2:", e.tstb)
	println("call testFuns.GetName2 aa:", (*e.tstb)["aaa"])
	c.ResHeader().Set("cookie", "1234567")
	c.ResString(ResStatusOk, "ok")
}
func (e *testFuns) Runs(c *Context, hdr *Header, body string) {
	println(fmt.Sprintf("Runs(%s):(%s)", body, time.Now().String()))
	c.ResString(ResStatusOk, "ok")
}
func TestRPCReq(t *testing.T) {
	req, err := NewRPCReq("localhost:7030", 2, "GetName1")
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	req.ReqHeader().Token = "123456"
	err = req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("GetName1 req code:", req.ResCode())
	hdr, err := req.ResHeader()
	if err == nil {
		cookie, _ := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("GetName1 req body:", string(req.ResBodyBytes()))

	req, err = NewRPCReq("localhost:7030", 2, "GetName2")
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	req.ReqHeader().Token = "123456"
	err = req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("GetName2 req code:", req.ResCode())
	hdr, err = req.ResHeader()
	if err == nil {
		cookie, _ := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("GetName2 req body:", string(req.ResBodyBytes()))
}

func testRPCs(in int) {
	req, err := NewRPCReq("localhost:7030", 2, "Runs")
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	req.ReqHeader().Token = "123456"
	err = req.Do(nil, fmt.Sprintf("%d", in))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	//println(fmt.Sprintf("GetName1 req(%d) body:%s",req.ResCode(), string(req.ResBodyBytes())))
}
func TestRPCReqs(t *testing.T) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(in int) {
			tms := time.Now()
			testRPCs(in)
			println(fmt.Sprintf("TestRPCReq(%d) end times:%0.5fs", in, time.Since(tms).Seconds()))
			wg.Done()
		}(i)
	}
	wg.Wait()
}
