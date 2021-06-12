package hbtp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

type AuthFun = func(c *Context) bool

var hdrtyp = reflect.TypeOf(url.Values{})

type IRPCRoute interface {
	AuthFun() AuthFun
}

func appendParams(self *reflect.Value, c *Context, fnt reflect.Type) ([]reflect.Value, error) {
	nmIn := fnt.NumIn()
	nmOut := fnt.NumOut()
	if nmOut > 0 {
		return nil, errors.New("method err")
	}
	inls := make([]reflect.Value, nmIn)
	ind := 1
	inls[0] = reflect.ValueOf(c)
	if self != nil {
		inls[1] = reflect.ValueOf(c)
		ind = 2
	}
	for i := ind; i < nmIn; i++ {
		argt := fnt.In(i)
		argtr := argt
		if argt.Kind() == reflect.Ptr {
			argtr = argt.Elem()
		}
		/*if argtr == hdrtyp {
			inls[i] = reflect.ValueOf(c.args)
			continue
		}*/
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
	if self != nil {
		return inls[1:], nil
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

func GrpcFunHandle(t IRPCRoute) ConnFun {
	tv := reflect.ValueOf(t)
	ty := tv.Type()
	if ty.Kind() != reflect.Ptr {
		panic("route must struct pointer")
	}
	tyr := ty.Elem()
	if tyr.Kind() != reflect.Struct {
		return nil
	}

	mln1 := ty.NumMethod()
	//mln2 := tyr.NumMethod()
	mfn := t.AuthFun()
	return func(c *Context) {
		if mfn != nil && !mfn(c) {
			return
		}

		for i := 0; i < mln1; i++ {
			mty := ty.Method(i)
			if c.Command() == mty.Name {
				inls, err := appendParams(&tv, c, mty.Type)
				if err != nil {
					c.ResString(ResStatusErr, fmt.Sprintf("appendParams err:%+v", err))
					return
				}
				tv.Method(i).Call(inls)
				return
			}
		}

		c.ResString(ResStatusNotFound, "not found method:"+c.Command())
		return
	}
}
