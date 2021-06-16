package hbtp

type msgInfo struct {
	Version int16
	Control int32
	LenCmd  uint16
	LenArg  uint16
	LenHead uint32
	LenBody uint32
}
type resInfo struct {
	Code    int32
	LenHead uint32
	LenBody uint32
}
