package hbtp

import "time"

var conf = newConf()

type config struct {
	tmsHead time.Duration
	tmsBody time.Duration
	maxHead uint
	maxBody uint
}

func newConf() config {
	return config{
		tmsHead: time.Second * 10,   //10s
		tmsBody: time.Second * 20,   //20s
		maxHead: 1024 * 1024 * 100,  //100M
		maxBody: 1024 * 1024 * 1024, //1G
	}
}

func SetMaxHeadLen(n uint) {
	conf.maxHead = n
}
func SetMaxBodyLen(n uint) {
	conf.maxBody = n
}

func ReadHeadTimeout(n time.Duration) {
	conf.tmsHead = n
}
func ReadBodyTimeout(n time.Duration) {
	conf.tmsBody = n
}
