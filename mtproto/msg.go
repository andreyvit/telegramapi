package mtproto

import (
	"github.com/andreyvit/telegramapi/tl"
)

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

func MsgFromObj(o tl.Object) Msg {
	if o == nil {
		panic("MsgFromObj for nil object")
	} else {
		var t MsgType
		if IsContentMsg(o) {
			t = ContentMsg
		} else {
			t = KeyExMsg
		}
		return Msg{tl.Bytes(o), t, 0}
	}
}

func IsContentMsg(o tl.Object) bool {
	return combOrigins[o.Cmd()] == SchemaOriginTelegram
}

func RequiresAck(o tl.Object) bool {
	return IsContentMsg(o)
}
