package hbtp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"reflect"
	"time"
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
	if ctx == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var Debug = false

func Infof(s string, args ...interface{}) {
	tms := time.Now().Format("2006-01-02 15:04:05")
	if len(args) > 0 {
		fmt.Println(tms + " [info] " + fmt.Sprintf(s, args...))
	} else {
		fmt.Println(tms + " [info] " + s)
	}
}
func Errorf(s string, args ...interface{}) {
	tms := time.Now().Format("2006-01-02 15:04:05")
	if len(args) > 0 {
		println(tms + " [err] " + fmt.Sprintf(s, args...))
	} else {
		println(tms + " [err] " + s)
	}
}
func Debugf(s string, args ...interface{}) {
	if !Debug {
		return
	}
	tms := time.Now().Format("2006-01-02 15:04:05")
	if len(args) > 0 {
		println(tms + " [debug] " + fmt.Sprintf(s, args...))
	} else {
		println(tms + " [debug] " + s)
	}
}

func TcpRead(ctx context.Context, conn net.Conn, ln uint) ([]byte, error) {
	if conn == nil || ln <= 0 {
		return nil, errors.New("handleRead ln<0")
	}
	var buf *bytes.Buffer
	rn := uint(0)
	tn := ln
	if ln > 10240 {
		tn = 10240
		buf = &bytes.Buffer{}
	}
	bts := make([]byte, tn)
	for {
		if EndContext(ctx) {
			return nil, errors.New("context dead")
		}
		n, err := conn.Read(bts)
		if n > 0 {
			rn += uint(n)
			if buf != nil {
				buf.Write(bts[:n])
			}
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
	if buf != nil {
		return buf.Bytes(), nil
	}
	return bts, nil
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

func FlcStructSizeof(pt interface{}) int {
	if pt == nil {
		return 0
	}
	ptv := reflect.ValueOf(pt)
	if ptv.Kind() == reflect.Ptr {
		if ptv.IsZero() {
			return 0
		}
		ptv = ptv.Elem()
	}

	ln := ptv.NumField()
	num := 0
	for i := 0; i < ln; i++ {
		fd := ptv.Field(i)
		num += int(fd.Type().Size())
	}
	return num
}
func FlcStruct2Byte(pt interface{}) ([]byte, error) {
	if pt == nil {
		return nil, errors.New("param is nil")
	}
	ptv := reflect.ValueOf(pt)
	if ptv.Kind() == reflect.Ptr {
		if ptv.IsZero() {
			return nil, errors.New("pt is null")
		}
		ptv = ptv.Elem()
	}
	ln := ptv.NumField()
	buf := &bytes.Buffer{}
	for i := 0; i < ln; i++ {
		fdt := ptv.Type().Field(i)
		fd := ptv.Field(i)
		sz := int(fd.Type().Size())
		val := fd.Interface()
		var bts []byte
		switch fd.Kind() {
		case reflect.Int8:
			bts = LitIntToByte(int64(val.(int8)), sz)
		case reflect.Uint8:
			bts = LitIntToByte(int64(val.(uint8)), sz)
		case reflect.Int16:
			bts = LitIntToByte(int64(val.(int16)), sz)
		case reflect.Uint16:
			bts = LitIntToByte(int64(val.(uint16)), sz)
		case reflect.Int32:
			bts = LitIntToByte(int64(val.(int32)), sz)
		case reflect.Uint32:
			bts = LitIntToByte(int64(val.(uint32)), sz)
		case reflect.Int64:
			bts = LitIntToByte(val.(int64), sz)
		case reflect.Uint64:
			bts = LitIntToByte(int64(val.(uint64)), sz)
		case reflect.Int:
			bts = LitIntToByte(int64(val.(int)), sz)
		case reflect.Uint:
			bts = LitIntToByte(int64(val.(uint)), sz)
		case reflect.Float32:
			bts = LitIntToByte(int64(val.(float32)), sz)
		case reflect.Float64:
			bts = LitIntToByte(int64(val.(float64)), sz)
		default:
			println("field not found type:%s", fdt.Name)
			continue
		}
		buf.Write(bts)
	}

	return buf.Bytes(), nil
}
func FlcByte2Struct(bts []byte, pt interface{}) error {
	if pt == nil {
		return errors.New("param is nil")
	}
	ptv := reflect.ValueOf(pt)
	if ptv.Kind() != reflect.Ptr || ptv.IsZero() {
		return errors.New("pt is not ptr")
	}
	ptv = ptv.Elem()
	ln := ptv.NumField()
	buf := bytes.NewBuffer(bts)
	for i := 0; i < ln; i++ {
		fdt := ptv.Type().Field(i)
		fd := ptv.Field(i)
		sz := int(fd.Type().Size())
		bts := make([]byte, sz)
		n, _ := buf.Read(bts)
		if n != sz {
			println("field set err:%s", fdt.Name)
			continue
		}
		val := LitByteToInt(bts)
		switch fd.Kind() {
		case reflect.Int8:
			fd.Set(reflect.ValueOf(int8(val)))
		case reflect.Uint8:
			fd.Set(reflect.ValueOf(uint8(val)))
		case reflect.Int16:
			fd.Set(reflect.ValueOf(int16(val)))
		case reflect.Uint16:
			fd.Set(reflect.ValueOf(uint16(val)))
		case reflect.Int32:
			fd.Set(reflect.ValueOf(int32(val)))
		case reflect.Uint32:
			fd.Set(reflect.ValueOf(uint32(val)))
		case reflect.Int64:
			fd.Set(reflect.ValueOf(uint64(val)))
		case reflect.Uint64:
			fd.Set(reflect.ValueOf(uint64(val)))
		case reflect.Int:
			fd.Set(reflect.ValueOf(int(val)))
		case reflect.Uint:
			fd.Set(reflect.ValueOf(uint(val)))
		case reflect.Float32:
			fd.Set(reflect.ValueOf(float32(val)))
		case reflect.Float64:
			fd.Set(reflect.ValueOf(float64(val)))
		default:
			println("field not found type:%s", fdt.Name)
			continue
		}
	}

	return nil
}

/*func Struct2Byte(pt interface{}) ([]byte, error) {
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
}*/
func Ptr2Bytes(pt unsafe.Pointer, ln int) ([]byte, error) {
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
func Bytes2Ptr(data []byte, pt unsafe.Pointer) error {
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
