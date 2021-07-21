package hbtp

import (
	"encoding/json"
	"errors"
)

type Response struct {
	code   int32
	heads  []byte
	bodys  []byte
	header *Map
}

func (c *Response) Header() (*Map, error) {
	if c.heads == nil {
		return nil, errors.New("is do?")
	}
	if c.header == nil {
		c.header = NewMap()
		err := json.Unmarshal(c.heads, &c.header)
		if err != nil {
			return nil, err
		}
	}
	return c.header, nil
}
func (c *Response) Code() int32 {
	return c.code
}
func (c *Response) HeadBytes() []byte {
	return c.heads
}
func (c *Response) BodyBytes() []byte {
	return c.bodys
}
func (c *Response) BodyJson(bd interface{}) error {
	if c.bodys == nil {
		return errors.New("is do?")
	}
	return json.Unmarshal(c.bodys, bd)
}
