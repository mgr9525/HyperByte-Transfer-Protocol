package hbtp

import "time"

const MaxOther uint64 = 1024 * 1024 * 20   //20M
const MaxHeads uint64 = 1024 * 1024 * 100  //100M
const MaxBodys uint64 = 1024 * 1024 * 1024 //1G

type LmtTmConfig struct {
	TmOhther time.Duration
	TmHeads  time.Duration
	TmBodys  time.Duration
}
type LmtMaxConfig struct {
	MaxOhther uint64
	MaxHeads  uint64
	MaxBodys  uint64
}

func MakeLmtTmCfg() *LmtTmConfig {
	return &LmtTmConfig{
		TmOhther: time.Second * 10, //10s
		TmHeads:  time.Second * 30, //10s
		TmBodys:  time.Second * 50, //20s
	}
}
func MakeLmtMaxCfg() *LmtMaxConfig {
	return &LmtMaxConfig{
		MaxOhther: 1024 * 1024 * 2,  //2M
		MaxHeads:  1024 * 1024 * 10, //10M
		MaxBodys:  1024 * 1024 * 50, //50M
	}
}
