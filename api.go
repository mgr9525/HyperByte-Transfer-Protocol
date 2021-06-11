package hbtp

import (
	"errors"
	"fmt"
)

func DoJson(req *Request, in, out interface{}, hd ...Map) error {
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
func DoString(req *Request, in interface{}, hd ...Map) (int32, []byte, error) {
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
