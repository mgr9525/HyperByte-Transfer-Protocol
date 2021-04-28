package hbtp

import (
	"errors"
	"fmt"
	"time"
)

type DoGenTokenHandle func(request *Request) string

func NewDoReq(host string, code int, pth, ipr string, tkfn DoGenTokenHandle, tmots ...time.Duration) (*Request, error) {
	req, err := NewRequest(host, code, tmots...)
	if err != nil {
		return nil, err
	}
	hd := req.ReqHeader()
	hd.RelIp = ipr
	hd.Path = pth
	hd.Times = time.Now().Format(time.RFC3339Nano)
	if tkfn != nil {
		hd.Token = tkfn(req)
	}
	return req, err
}
func NewDoRPCReq(host string, code int, method, ipr string, tkfn DoGenTokenHandle, tmots ...time.Duration) (*Request, error) {
	req, err := NewRPCReq(host, code, method, tmots...)
	if err != nil {
		return nil, err
	}
	hd := req.ReqHeader()
	hd.RelIp = ipr
	hd.Times = time.Now().Format(time.RFC3339Nano)
	if tkfn != nil {
		hd.Token = tkfn(req)
	}
	return req, err
}
func DoJson(req *Request, in, out interface{}, hd ...map[string]interface{}) error {
	if req == nil {
		return errors.New("req is nil")
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err := req.Do(nil, in)
	if err != nil {
		return err
	}
	if req.ResCode() != ResStatusOk {
		return fmt.Errorf("res err(%d):%s", req.ResCode(), string(req.ResBodyBytes()))
	}
	return req.ResBodyJson(out)
}
func DoString(req *Request, in interface{}, hd ...Mp) (int, []byte, error) {
	if req == nil {
		return 0, nil, errors.New("req is nil")
	}
	defer req.Close()
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.ReqHeader().Set(k, v)
		}
	}
	err := req.Do(nil, in)
	if err != nil {
		return 0, nil, err
	}
	return req.ResCode(), req.ResBodyBytes(), nil
}
