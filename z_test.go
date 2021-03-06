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
	eg.RegParamFun(1, helloWorldFun)
	eg.RegGrpcFun(2, &testFuns{})
	/*go func() {
		time.Sleep(time.Second * 10)
		eg.Stop()
	}()*/
	if err := eg.Run(":7030"); err != nil {
		println("engine run err:", err.Error())
	}
}

func helloWorldFun(c *Context) {
	host := c.ReqHeader().GetString("host")
	println("helloWorldFun header host:" + host)
	println("helloWorldFun body:" + string(c.BodyBytes()))
	c.ResHeader().Set("cookie", "1234567")
	c.ResString(ResStatusOk, "ok")
}

func TestRequest(t *testing.T) {
	req := NewRequest("localhost:7030", 1).Command("/hello/test")
	defer req.Close()
	//req.ReqHeader().Set("host", "test.host.com")
	res, err := req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("req code:", res.Code())
	hdr, err := res.Header()
	if err == nil {
		cookie := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("req body:", string(res.BodyBytes()))
}

type testFuns struct {
}

func (*testFuns) AuthFun() AuthFun {
	return func(c *Context) bool {
		tks := c.Args().Get("token")
		if tks != "123456" {
			c.ResString(ResStatusErr, "token err:"+tks)
			return false
		}
		return true
	}
}
func (e *testFuns) GetName1(c *Context, body string) {
	println("call testFuns.GetName1:", body)
	c.ResString(ResStatusOk, "ok")
}
func (e *testFuns) GetName2(c *Context) {
	println("call testFuns.GetName2:")
	c.ResString(ResStatusOk, "ok")
}
func (e *testFuns) Runs(c *Context, body string) {
	Debugf("Runs(%s):(%s)", body, time.Now().String())
	c.ResString(ResStatusOk, "ok")
}
func TestRPCReq(t *testing.T) {
	req := NewRequest("localhost:7030", 2).
		Command("GetName1").SetArg("token", "123456")
	defer req.Close()
	res, err := req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("GetName1 res code:", res.Code())
	hdr, err := res.Header()
	if err == nil {
		cookie := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("GetName1 req body:", string(res.BodyBytes()))

	req = NewRequest("localhost:7030", 2).
		Command("GetName2").SetArg("token", "1234567")
	defer req.Close()
	res, err = req.Do(nil, []byte("hello world"))
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	println("GetName2 res code:", res.Code())
	hdr, err = res.Header()
	if err == nil {
		cookie := hdr.GetString("cookie")
		println("req cookie:", cookie)
	}
	println("GetName2 req body:", string(res.BodyBytes()))
}

func testRPCs(in int) {
	req := NewRequest("localhost:7030", 2).
		Command("Runs").SetArg("token", "123456")
	res, err := req.Do(nil, fmt.Sprintf("%d", in))
	defer req.Close()
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	fmt.Println(fmt.Sprintf("Runs res(%d) body:%s", res.Code(), string(res.BodyBytes())))
}
func TestRPCReqs(t *testing.T) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(in int) {
			tms := time.Now()
			testRPCs(in)
			fmt.Println(fmt.Sprintf("TestRPCReq(%d) end times:%0.5fs", in, time.Since(tms).Seconds()))
			wg.Done()
		}(i)
	}
	wg.Wait()
}
