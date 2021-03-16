package utils

import (
	"context"
	"errors"
	"net"
	"os"
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

func CheckContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func TcpRead(ctx context.Context, conn net.Conn, ln uint) ([]byte, error) {
	if conn == nil || ln <= 0 {
		return nil, errors.New("handleRead ln<0")
	}
	rn := uint(0)
	rt := make([]byte, ln)
	for {
		if CheckContext(ctx) {
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
