package mtproto

import (
	"bytes"

	"github.com/andreyvit/telegramapi/binints"
)

type Msg struct {
	Encrypted bool
	MsgID     uint64
	Payload   []byte
}

func NormalMsg(b []byte) Msg {
	return Msg{true, 0, b}
}
func UnencryptedMsg(b []byte) Msg {
	return Msg{false, 0, b}
}

type Framer struct {
	MsgIDOverride uint64

	gen MsgIDGen
}

func (fr *Framer) Format(msg Msg) ([]byte, error) {
	var buf bytes.Buffer
	binints.WriteUint64LE(&buf, 0)

	var msgID uint64
	if fr.MsgIDOverride != 0 {
		msgID = fr.MsgIDOverride
		fr.MsgIDOverride = 0
	} else {
		msgID = fr.gen.Generate()
	}
	binints.WriteUint64LE(&buf, msgID)

	binints.WriteUint32LE(&buf, uint32(len(msg.Payload)))
	buf.Write(msg.Payload)

	return buf.Bytes(), nil
}

func (fr *Framer) Parse(raw []byte) (Msg, error) {
	r := bytes.NewReader(raw)

	authKeyID, err := binints.ReadUint64LE(r)
	if err != nil {
		return Msg{}, err
	}

	if authKeyID == 0 {
		var a Accum

		msgID, err := binints.ReadUint64LE(r)
		a.Push(err)

		msgLen, err := binints.ReadUint32LE(r)
		a.Push(err)

		payload, err := ReadN(r, int(msgLen))
		a.Push(err)

		a.Push(binints.ExpectEOF(r))

		return Msg{false, msgID, payload}, a.Error()
	} else {
		// log.Printf("Received encrypted: authKeyID=%x msgID=%x msgLen=%d cmd = %08x", authKeyID, msgID, msgLen, cmd)
		panic("authKeyID != 0")
		// a.Push(binints.ExpectEOF(r))
	}
}
