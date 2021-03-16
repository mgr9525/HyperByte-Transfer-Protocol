package iotconn

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Header struct {
	Path  string `json:"path"`
	RelIp string `json:"ip"`    // 真实IP
	Token string `json:"token"` // 请求token
	Times string `json:"times"` // 请求时间

	Info map[string]interface{} `json:"info"`
}

func ParseHeader(bts []byte) (*Header, error) {
	if bts == nil {
		return nil, errors.New("bts is nil")
	}
	rt := &Header{}
	return rt, json.Unmarshal(bts, rt)
}

func (c *Header) Set(key string, val interface{}) {
	if c.Info == nil {
		c.Info = make(map[string]interface{})
	}
	c.Info[key] = val
}
func (c *Header) Del(key string) {
	if c.Info == nil {
		return
	}
	delete(c.Info, key)
}
func (c *Header) Bytes() []byte {
	if c == nil {
		return nil
	}
	bts, _ := json.Marshal(c)
	return bts
}
func (c *Header) Get(key string) (interface{}, bool) {
	if c.Info == nil {
		return nil, false
	}
	rt, ok := c.Info[key]
	return rt, ok
}
func (c *Header) GetString(key string) (string, bool) {
	v, ok := c.Get(key)
	if !ok {
		return "", false
	}
	switch v.(type) {
	case string:
		return v.(string), true
	}
	return fmt.Sprintf("%v", v), true
}
func (c *Header) GetInt(key string) (int64, error) {
	v, ok := c.Get(key)
	if !ok {
		return 0, errors.New("not found")
	}
	switch v.(type) {
	case int:
		return v.(int64), nil
	case string:
		return strconv.ParseInt(v.(string), 10, 64)
	case int64:
		return v.(int64), nil
	case float32:
		return int64(v.(float32)), nil
	case float64:
		return int64(v.(float64)), nil
	}
	return 0, errors.New("not found")
}
func (c *Header) GetFloat(key string) (float64, error) {
	v, ok := c.Get(key)
	if !ok {
		return 0, errors.New("not found")
	}
	switch v.(type) {
	case int:
		return float64(v.(int)), nil
	case string:
		return strconv.ParseFloat(v.(string), 64)
	case int64:
		return float64(v.(int64)), nil
	case float32:
		return float64(v.(float32)), nil
	case float64:
		return v.(float64), nil
	}
	return 0, errors.New("not found")
}
func (c *Header) GetBool(key string) bool {
	v, ok := c.Get(key)
	if ok {
		switch v.(type) {
		case bool:
			return v.(bool)
		}
	}
	return false
}
