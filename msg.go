package hbtp

import (
	"context"
	"errors"
	"net"
	"time"
)

type MessageRecv interface {
	OnCheck()
	OnMsg(msg *Message) error
}

var ErrInterrupt = errors.New("on msg err will interrupt")

type Messager struct {
	ctx    context.Context
	cncl   context.CancelFunc
	conn   net.Conn
	shuted bool
	servs  bool

	ctms   time.Time
	ctmout time.Time
	pipes  chan *Messages
	rcver  MessageRecv
}

func NewMessager(ctx context.Context, conn net.Conn, rcver MessageRecv, chanln int) *Messager {
	c := &Messager{
		conn:   conn,
		shuted: false,
		rcver:  rcver,
	}
	c.ctx, c.cncl = context.WithCancel(ctx)
	if chanln > 0 {
		c.pipes = make(chan *Messages, chanln)
	} else {
		c.pipes = make(chan *Messages)
	}
	return c
}
func (c *Messager) Stop() {
	if c.shuted {
		return
	}
	println("msger conn will stop")
	c.shuted = true
	c.cncl()
	close(c.pipes)
	c.conn.Close()
}
func (c *Messager) Run(servs bool) {
	c.servs = servs
	c.ctmout = time.Now()
	go func() {
		for !EndContext(c.ctx) {
			c.runSend()
		}
	}()
	go func() {
		for !EndContext(c.ctx) {
			c.runRecv()
		}
	}()
	println("Messager start run check")
	for !EndContext(c.ctx) {
		c.runCheck()
		time.Sleep(time.Millisecond * 100)
	}
	c.Stop()
	println("Messager end run check")
}

func (c *Messager) runRecv() {
	defer func() {
		if err := recover(); err != nil {
			println("Messager.runRecv recover:%v", err)
		}
	}()
	msg, err := parseMessage(c.ctx, c.conn)
	if err != nil {
		println("run_send sendMessage err:", err.Error())
		c.cncl()
		time.Sleep(time.Millisecond * 100)
		return
	}
	if msg.Control == 0 {
		c.ctmout = time.Now()
		if c.servs {
			msgs := &Messages{
				Control: 0,
				Cmds:    "heart",
				Heads:   nil,
				Bodys:   nil,
			}
			c.Send(msgs)
		} else if c.rcver != nil {
			err = c.rcver.OnMsg(msg)
			if err != nil {
				println("rcver on_msg err:", err.Error())
				if err == ErrInterrupt {
					c.cncl()
				}
			}
		}
	}
}
func (c *Messager) runSend() {
	defer func() {
		if err := recover(); err != nil {
			println("Messager.runSend recover:%v", err)
		}
	}()

	data := <-c.pipes
	err := sendMessage(c.ctx, c.conn, data)
	if err != nil {
		println("run_send sendMessage err:", err.Error())
		time.Sleep(time.Millisecond * 10)
	}
}
func (c *Messager) runCheck() {
	defer func() {
		if err := recover(); err != nil {
			println("Messager.runCheck recover:%v", err)
		}
	}()

	if time.Since(c.ctmout).Seconds() > 30 {
		println("msger heart timeout!!")
		c.cncl()
		return
	}
	if !c.servs && time.Since(c.ctms).Seconds() > 20 {
		c.ctms = time.Now()
		msgs := &Messages{
			Control: 0,
			Cmds:    "heart",
			Heads:   nil,
			Bodys:   nil,
		}
		c.Send(msgs)
	}

	c.rcver.OnCheck()
}
func (c *Messager) Send(m *Messages) {
	if !c.shuted {
		c.pipes <- m
	}
}
