package hbtp

var (
	lenMsgInfo   = SizeOf(new(msgInfo))
	lenResInfoV1 = SizeOf(new(resInfoV1))
)

type msgInfo struct {
	Version uint16
	Control int32
	LenCmd  uint16
	LenArg  uint16
	LenHead uint32
	LenBody uint32
}
type resInfoV1 struct {
	Code    int32
	LenHead uint32
	LenBody uint32
}
