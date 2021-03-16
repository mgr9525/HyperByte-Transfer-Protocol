package iotconn

import (
	"context"
	"testing"
)

func TestServer(t *testing.T) {
	eg := NewEngine(context.Background())
	eg.RegFun(1, ParamFunHandle(helloWorldFun))
	eg.RegFun(2, RPCFunHandle(&testFuns{}))
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
	req, err := NewRequest("localhost:7030")
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	//req.ReqHeader().Set("host", "test.host.com")
	err = req.Do(1, []byte("hello world"))
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
}

func (testFuns) GetName(c *Context) {
	c.ResHeader().Set("cookie", "1234567")
	c.ResString(ResStatusOk, "ok")
}
func TestRPCReq(t *testing.T) {
	req, err := NewRPCReq("localhost:7030", "GetName")
	if err != nil {
		println("NewRequest err:", err.Error())
		return
	}
	defer req.Close()
	err = req.Do(2, []byte("hello world"))
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
