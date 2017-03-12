package mtproto

type MsgType int

const (
	ContentMsg MsgType = 0
	ServiceMsg         = 1
	KeyExMsg           = 2
)

type Msg struct {
	Payload []byte
	Type    MsgType
	MsgID   uint64
}

func MakeMsg(b []byte, t MsgType) Msg {
	return Msg{b, t, 0}
}
