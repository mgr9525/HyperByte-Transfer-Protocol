package hbtp

import (
	"testing"
)

type testCS struct {
	i int16
	a [2]byte
	b [4]byte
	c [2]byte
}

func TestU1(t *testing.T) {
	cs := &testCS{i: 123}
	copy(cs.a[:], LitIntToByte(123, 2))
	copy(cs.b[:], LitIntToByte(123124, 4))
	copy(cs.c[:], LitIntToByte(110, 2))
	bts, err := Struct2Byte(cs)
	//bts,err:=Struct2Bytes(unsafe.Pointer(cs),unsafe.Sizeof(*cs))
	if err != nil {
		println("Struct2Bytes err:" + err.Error())
		return
	}
	Debugf("Struct2Bytes hex:%x", bts)
	cs = &testCS{}
	err = Byte2Struct(bts, cs)
	//err=Bytes2Struct(bts,unsafe.Pointer(cs))
	if err != nil {
		println("Bytes2Struct err:" + err.Error())
		return
	}
	Debugf("Bytes2Struct i:%d,a:%x(%d),b:%x(%d),c:%x(%d)", cs.i, cs.a, BigByteToInt(cs.a[:]), cs.b, BigByteToInt(cs.b[:]), cs.c, BigByteToInt(cs.c[:]))
}
