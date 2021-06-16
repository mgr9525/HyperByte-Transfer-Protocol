package hbtp

import (
	"testing"
)

type testCS struct {
	i int16
	k int32
	a [2]byte
	b [4]byte
	c [2]byte
}

func TestS1(t *testing.T) {
	cs := &testCS{i: 123, k: 4567}
	copy(cs.a[:], LitIntToByte(123, 2))
	copy(cs.b[:], LitIntToByte(123124, 4))
	copy(cs.c[:], LitIntToByte(123, 2))
	ln2 := SizeOf(cs)
	println("size ln:", ln2)
}
func TestU1(t *testing.T) {
	Debug = true
	cs := &testCS{i: 123, k: 4567}
	copy(cs.a[:], LitIntToByte(123, 2))
	copy(cs.b[:], LitIntToByte(123124, 4))
	copy(cs.c[:], LitIntToByte(110, 2))
	bts, err := Struct2Byte(cs)
	//bts,err:=Struct2Bytes(unsafe.Pointer(cs),unsafe.Sizeof(*cs))
	if err != nil {
		println("Struct2Bytes err:" + err.Error())
		return
	}
	Debugf("Struct2Bytes hex(ln:%d):%x", len(bts), bts)
	cs = &testCS{}
	err = Byte2Struct(bts, cs)
	//err=Bytes2Struct(bts,unsafe.Pointer(cs))
	if err != nil {
		println("Bytes2Struct err:" + err.Error())
		return
	}
	Debugf("Bytes2Struct i:%d,k:%d,a:%x(%d),b:%x(%d),c:%x(%d)", cs.i, cs.k, cs.a, LitByteToInt(cs.a[:]), cs.b, LitByteToInt(cs.b[:]), cs.c, LitByteToInt(cs.c[:]))
}
