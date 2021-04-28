package hbtp

import (
	"fmt"
	"time"
)

var (
	doTokens   = ""
	doTokenFun DoGenTokenHandle
)

type DoGenTokenHandle func(request *Request) string

func InitDoToken(tks string, tkfs ...DoGenTokenHandle) {
	doTokens = tks
	if len(tkfs) > 0 {
		doTokenFun = tkfs[0]
	}
}

func NewDoReq(host string, code int, pth string, ipr ...string) (*Request, error) {
	req, err := NewRequest(host, code)
	if err != nil {
		return nil, err
	}
	hd := req.ReqHeader()
	if len(ipr) > 0 {
		hd.RelIp = ipr[1]
	}
	hd.Path = pth
	hd.Times = time.Now().Format(time.RFC3339Nano)
	hd.Token = doTokens
	if doTokens == "" && doTokenFun != nil {
		hd.Token = doTokenFun(req)
	}
	return req, err
}
func DoJson(host string, code int, pth string, in, out interface{}, hd ...map[string]interface{}) error {
	req, err := NewDoReq(host, code, pth)
	if err != nil {
		return err
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err = req.Do(in)
	if err != nil {
		return err
	}
	if req.ResCode() != ResStatusOk {
		return fmt.Errorf("res err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return req.ResBodyJson(out)
}
func DoString(host string, code int, pth string, in interface{}, hd ...Mp) (int, []byte, error) {
	req, err := NewDoReq(host, code, pth)
	if err != nil {
		return 0, nil, err
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err = req.Do(in)
	if err != nil {
		return 0, nil, err
	}
	return req.ResCode(), req.ResBodyBytes(), nil
}

func NewDoRPCReq(host string, code int, method string, ipr ...string) (*Request, error) {
	req, err := NewRPCReq(host, code, method)
	if err != nil {
		return nil, err
	}
	hd := req.ReqHeader()
	if len(ipr) > 0 {
		hd.RelIp = ipr[1]
	}
	hd.Times = time.Now().Format(time.RFC3339Nano)
	hd.Token = doTokens
	if doTokens == "" && doTokenFun != nil {
		hd.Token = doTokenFun(req)
	}
	return req, err
}
func DoRPCJson(host string, code int, method string, in, out interface{}, hd ...map[string]interface{}) error {
	req, err := NewDoRPCReq(host, code, method)
	if err != nil {
		return err
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err = req.Do(in)
	if err != nil {
		return err
	}
	if req.ResCode() != ResStatusOk {
		return fmt.Errorf("res err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return req.ResBodyJson(out)
}
func DoRPCString(host string, code int, method string, in interface{}, hd ...Mp) (int, []byte, error) {
	req, err := NewDoRPCReq(host, code, method)
	if err != nil {
		return 0, nil, err
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err = req.Do(in)
	if err != nil {
		return 0, nil, err
	}
	return req.ResCode(), req.ResBodyBytes(), nil
}
