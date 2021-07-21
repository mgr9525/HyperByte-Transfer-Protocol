package hbtp

import (
	"fmt"
	"testing"
	"unsafe"
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
	ln2 := FlcStructSizeof(cs)
	println("size ln:", ln2)
}

/*func TestU1(t *testing.T) {
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
}*/
func TestS3(t *testing.T) {
	info := &msgInfo{
		Version: 1,
		Control: 2,
		LenCmd:  1000,
		LenBody: 50,
	}
	bts, err := FlcStruct2Byte(info)
	if err != nil {
		println("err1:", err.Error())
		return
	}
	println(fmt.Sprintf("bts(%d/%d):%v", len(bts), unsafe.Sizeof(info.Control), bts))
	infos := &msgInfo{}
	err = FlcByte2Struct(bts, infos)
	if err != nil {
		println("err2:", err.Error())
		return
	}
	println(fmt.Sprintf("infos(%d/%d):%d,%d", infos.Version, infos.Control, infos.LenCmd, infos.LenBody))
}
