package hbtp

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type AuthFun = func(c *Context) bool

var hdrtyp = reflect.TypeOf(Header{})

func appendParams(self *reflect.Value, c *Context, fnt reflect.Type) ([]reflect.Value, error) {
	hdr, _ := c.ReqHeader()
	nmIn := fnt.NumIn()
	var inls []reflect.Value
	ind := 1
	if self == nil {
		inls = make([]reflect.Value, nmIn)
	} else {
		inls = make([]reflect.Value, nmIn-1)
		ind = 2
	}
	inls[0] = reflect.ValueOf(c)
	for i := ind; i < nmIn; i++ {
		argt := fnt.In(i)
		argtr := argt
		if argt.Kind() == reflect.Ptr {
			argtr = argt.Elem()
		}
		if argtr == hdrtyp {
			inls[i] = reflect.ValueOf(hdr)
			continue
		}
		if argtr.Kind() == reflect.Struct || argtr.Kind() == reflect.Map {
			argv := reflect.New(argtr)
			if err := json.Unmarshal(c.BodyBytes(), argv.Interface()); err != nil {
				c.ResString(ResStatusErr, fmt.Sprintf("params err[%d]:%+v", i, err))
				return nil, err
			}
			if argt.Kind() == reflect.Ptr {
				inls[i] = argv
			} else {
				inls[i] = argv.Elem()
			}
		} else if argtr.Kind() == reflect.String {
			inls[i] = reflect.ValueOf(string(c.BodyBytes()))
		}
	}
	return inls, nil
}
func ParamFunHandle(fn interface{}, authfn ...AuthFun) ConnFun {
	fnv := reflect.ValueOf(fn)
	if fnv.Kind() != reflect.Func {
		return nil
	}
	fnt := fnv.Type()
	return func(c *Context) {
		if len(authfn) > 0 {
			if !authfn[0](c) {
				return
			}
		}
		inls, err := appendParams(nil, c, fnt)
		if err != nil {
			c.ResString(ResStatusErr, fmt.Sprintf("appendParams err:%+v", err))
			return
		}
		fnv.Call(inls)
	}
}

func RPCFunHandle(t interface{}, authfn ...AuthFun) ConnFun {
	tv := reflect.ValueOf(t)
	ty := tv.Type()
	tyr := ty
	if ty.Kind() == reflect.Ptr {
		tyr = ty.Elem()
	}
	if tyr.Kind() != reflect.Struct {
		return nil
	}

	mln := tyr.NumMethod()
	return func(c *Context) {
		hdr, err := c.ReqHeader()
		if err != nil {
			c.ResString(ResStatusErr, fmt.Sprintf("ReqHeader err:%+v", err))
			return
		}
		if len(authfn) > 0 {
			if !authfn[0](c) {
				return
			}
		}

		for i := 0; i < mln; i++ {
			mty := tyr.Method(i)
			if hdr.Path == mty.Name {
				inls, err := appendParams(&tv, c, mty.Type)
				if err != nil {
					c.ResString(ResStatusErr, fmt.Sprintf("appendParams err:%+v", err))
					return
				}
				tv.Method(i).Call(inls)
				return
			}
		}

		c.ResString(ResStatusNotFound, "not found method:"+hdr.Path)
		return
	}
}
