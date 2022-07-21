package hbtp

import (
	"errors"
	"fmt"
)

func DoJson(req *Request, in, out interface{}, hd ...Map) error {
	if req == nil {
		return errors.New("req is nil")
	}
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.Header().Set(k, v)
		}
	}
	res, err := req.Do(nil, in)
	if err != nil {
		return err
	}
	defer res.Close()
	if res.Code() != ResStatusOk {
		return fmt.Errorf("res err(%d):%s", res.Code(), string(res.BodyBytes()))
	}
	return res.BodyJson(out)
}
func DoString(req *Request, in interface{}, hd ...Map) (int32, []byte, error) {
	if req == nil {
		return 0, nil, errors.New("req is nil")
	}
	if len(hd) > 0 && hd[0] != nil {
		for k, v := range hd[0] {
			req.Header().Set(k, v)
		}
	}
	res, err := req.Do(nil, in)
	if err != nil {
		return 0, nil, err
	}
	defer res.Close()
	return res.Code(), res.BodyBytes(), nil
}
