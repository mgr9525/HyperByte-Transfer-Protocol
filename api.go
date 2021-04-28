package hbtp

import (
	"errors"
	"fmt"
	"time"
)

var (
	doHost     = ""
	doTokens   = ""
	doTokenFun DoGenTokenHandle
)

type DoGenTokenHandle func(request *Request) string

func InitDo(host, tks string, tkfs ...DoGenTokenHandle) {
	doHost = host
	doTokens = tks
	if len(tkfs) > 0 {
		doTokenFun = tkfs[0]
	}
}

func NewDoReq(code int, method string, ipr ...string) (*Request, error) {
	if doHost == "" {
		return nil, errors.New("hbtp do is not init")
	}
	req, err := NewRPCReq(doHost, code, method)
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
func DoJson(code int, method string, in, out interface{}, hd ...map[string]interface{}) error {
	req, err := NewDoReq(code, method)
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
func DoString(code int, method string, in interface{}, hd ...Mp) (int, []byte, error) {
	req, err := NewDoReq(code, method)
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
