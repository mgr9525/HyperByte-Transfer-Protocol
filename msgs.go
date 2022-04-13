package hbtp

import (
	"context"
	"errors"
	"fmt"
	"net"
)

type Messages struct {
	Control int32
	Cmds    string
	Heads   []byte
	Bodys   []byte
}
type Message struct {
	Version uint16
	Control int32
	Cmds    string
	Heads   []byte
	Bodys   []byte
}

func parseMessage(ctx context.Context, conn net.Conn) (*Message, error) {
	bts, err := TcpRead(ctx, conn, 1)
	if err != nil {
		return nil, err
	}
	if len(bts) < 1 || bts[0] != 0x8d {
		return nil, fmt.Errorf("first byte err:%v", bts)
	}
	bts, err = TcpRead(ctx, conn, 1)
	if err != nil {
		return nil, err
	}
	if len(bts) < 1 || bts[0] != 0x8f {
		return nil, fmt.Errorf("second byte err:%v", bts)
	}
	info := &msgsInfo{}
	infoln := FlcStructSizeof(info)
	bts, err = TcpRead(ctx, conn, uint(infoln))
	if err != nil {
		return nil, err
	}
	err = FlcByte2Struct(bts, info)
	if err != nil {
		return nil, err
	}
	if info.Version != 1 {
		return nil, errors.New("not found version")
	}
	rt := &Message{
		Version: info.Version,
		Control: info.Control,
	}
	if info.LenCmd > 0 {
		bts, err = TcpRead(ctx, conn, uint(info.LenCmd))
		if err != nil {
			return nil, err
		}
		rt.Cmds = string(bts)
	}
	if info.LenHead > 0 {
		rt.Heads, err = TcpRead(ctx, conn, uint(info.LenHead))
		if err != nil {
			return nil, err
		}
	}

	if info.LenBody > 0 {
		rt.Bodys, err = TcpRead(ctx, conn, uint(info.LenBody))
		if err != nil {
			return nil, err
		}
	}
	bts, err = TcpRead(ctx, conn, 2)
	if err != nil {
		return nil, err
	}
	if len(bts) < 2 || bts[0] != 0x8e || bts[1] != 0x8f {
		return nil, fmt.Errorf("end byte err:%v", bts)
	}
	return rt, nil
}

func sendMessage(ctx context.Context, conn net.Conn, msg *Messages) error {
	info := &msgsInfo{
		Version: 1,
		Control: msg.Control,
		LenCmd:  uint16(len(msg.Cmds)),
		LenHead: uint32(len(msg.Heads)),
		LenBody: uint32(len(msg.Bodys)),
	}
	bts, err := FlcStruct2Byte(info)
	if err != nil {
		return err
	}
	err = TcpWrite(ctx, conn, []byte{0x8d, 0x8f})
	if err != nil {
		return err
	}
	err = TcpWrite(ctx, conn, bts)
	if err != nil {
		return err
	}
	if info.LenCmd > 0 {
		err = TcpWrite(ctx, conn, []byte(msg.Cmds))
		if err != nil {
			return err
		}
	}
	if info.LenHead > 0 {
		err = TcpWrite(ctx, conn, msg.Heads)
		if err != nil {
			return err
		}
	}
	if info.LenBody > 0 {
		err = TcpWrite(ctx, conn, msg.Bodys)
		if err != nil {
			return err
		}
	}
	err = TcpWrite(ctx, conn, []byte{0x8e, 0x8f})
	if err != nil {
		return err
	}
	return nil
}
