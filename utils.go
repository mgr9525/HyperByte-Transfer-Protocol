package hbtp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"reflect"
	"unsafe"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func EndContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var Debug = false

func Debugf(s string, args ...interface{}) {
	if !Debug {
		return
	}
	if len(args) > 0 {
		println(fmt.Sprintf(s, args...))
	} else {
		println(s)
	}
}

func TcpRead(ctx context.Context, conn net.Conn, ln uint) ([]byte, error) {
	if conn == nil || ln <= 0 {
		return nil, errors.New("handleRead ln<0")
	}
	rn := uint(0)
	rt := make([]byte, ln)
	for {
		if EndContext(ctx) {
			return nil, errors.New("context dead")
		}
		n, err := conn.Read(rt[rn:])
		if n > 0 {
			rn += uint(n)
		}
		if rn >= ln {
			break
		}
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, errors.New("conn abort")
		}
	}
	return rt, nil
}

// BigEndian
func BigByteToInt(data []byte) int64 {
	ln := len(data)
	rt := int64(0)
	for i := 0; i < ln; i++ {
		rt |= int64(data[ln-1-i]) << (i * 8)
	}
	return rt
}
func BigIntToByte(data int64, ln int) []byte {
	rt := make([]byte, ln)
	for i := 0; i < ln; i++ {
		rt[ln-1-i] = byte(data >> (i * 8))
	}
	return rt
}

// LittleEndian
func LitByteToInt(data []byte) int64 {
	ln := len(data)
	rt := int64(0)
	for i := 0; i < ln; i++ {
		rt |= int64(data[i]) << (i * 8)
	}
	return rt
}
func LitIntToByte(data int64, ln int) []byte {
	rt := make([]byte, ln)
	for i := 0; i < ln; i++ {
		rt[i] = byte(data >> (i * 8))
	}
	return rt
}

// BigEndian
func BigByteToFloat32(data []byte) float32 {
	v := BigByteToInt(data)
	return math.Float32frombits(uint32(v))
}
func BigFloatToByte32(data float32) []byte {
	v := math.Float32bits(data)
	return BigIntToByte(int64(v), 4)
}
func BigByteToFloat64(data []byte) float64 {
	v := BigByteToInt(data)
	return math.Float64frombits(uint64(v))
}
func BigFloatToByte64(data float64) []byte {
	v := math.Float64bits(data)
	return BigIntToByte(int64(v), 8)
}

// LittleEndian
func LitByteToFloat32(data []byte) float32 {
	v := LitByteToInt(data)
	return math.Float32frombits(uint32(v))
}
func LitFloatToByte32(data float32) []byte {
	v := math.Float32bits(data)
	return LitIntToByte(int64(v), 4)
}
func LitByteToFloat64(data []byte) float64 {
	v := LitByteToInt(data)
	return math.Float64frombits(uint64(v))
}
func LitFloatToByte64(data float64) []byte {
	v := math.Float64bits(data)
	return LitIntToByte(int64(v), 8)
}

type SliceMock struct {
	addr unsafe.Pointer
	len  uint
	cap  uint
}

func Struct2Byte(pt interface{}) ([]byte, error) {
	if pt == nil {
		return nil, errors.New("param is nil")
	}
	ln := SizeOf(pt)
	return Struct2ByteLen(pt, ln)
}
func Struct2ByteLen(pt interface{}, ln int) ([]byte, error) {
	if pt == nil {
		return nil, errors.New("param is nil")
	}
	ptv := reflect.ValueOf(pt)
	if ptv.Kind() != reflect.Ptr && !ptv.IsZero() {
		return nil, errors.New("pt is not ptr")
	}
	return Struct2Bytes(unsafe.Pointer(ptv.Pointer()), ln)
}
func Byte2Struct(data []byte, pt interface{}) error {
	if pt == nil {
		return errors.New("param is nil")
	}
	ptv := reflect.ValueOf(pt)
	if ptv.Kind() != reflect.Ptr && !ptv.IsZero() {
		return errors.New("pt is not ptr")
	}
	pte := ptv.Elem()
	if pte.Kind() != reflect.Struct {
		return fmt.Errorf("*pt is not struct:%s", pte.Kind())
	}
	return Bytes2Struct(data, unsafe.Pointer(ptv.Pointer()))
}
func Struct2Bytes(pt unsafe.Pointer, ln int) ([]byte, error) {
	if pt == nil {
		return nil, errors.New("param is nil")
	}
	mock := &SliceMock{
		addr: pt,
		cap:  uint(ln),
		len:  uint(ln),
	}
	bts := *(*[]byte)(unsafe.Pointer(mock))
	rtbts := make([]byte, mock.len)
	copy(rtbts, bts)
	return rtbts, nil
}
func Bytes2Struct(data []byte, pt unsafe.Pointer) error {
	ln := len(data)
	if pt == nil || ln <= 0 {
		return errors.New("param is nil")
	}
	mock := &SliceMock{
		addr: pt,
		cap:  uint(ln),
		len:  uint(ln),
	}
	bts := *(*[]byte)(unsafe.Pointer(mock))
	copy(bts, data)
	return nil
}
