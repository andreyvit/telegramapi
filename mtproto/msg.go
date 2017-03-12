package mtproto

type MsgType int

const (
	ContentMsg MsgType = 0
	ServiceMsg         = 1
	KeyExMsg           = 2
)

func (t MsgType) String() string {
	switch t {
	case ContentMsg:
		return "Content"
	case ServiceMsg:
		return "Service"
	case KeyExMsg:
		return "KeyEx"
	default:
		panic("invalid value")
	}
}

type Msg struct {
	Payload []byte
	Type    MsgType
	MsgID   uint64
}

func MakeMsg(b []byte, t MsgType) Msg {
	return Msg{b, t, 0}
}
