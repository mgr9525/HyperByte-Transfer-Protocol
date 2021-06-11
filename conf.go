package hbtp

import "time"

type Config struct {
	TmsInfo time.Duration
	TmsHead time.Duration
	TmsBody time.Duration
	MaxHead uint
	MaxBody uint
}

func MakeConfig() Config {
	return Config{
		TmsInfo: time.Second * 10,   //10s
		TmsHead: time.Second * 20,   //10s
		TmsBody: time.Second * 30,   //20s
		MaxHead: 1024 * 1024 * 100,  //100M
		MaxBody: 1024 * 1024 * 1024, //1G
	}
}
