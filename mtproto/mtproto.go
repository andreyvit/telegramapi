package mtproto

type MessageHeader struct {
	MsgID uint64
	Seq   uint32
	Len   uint32
}
