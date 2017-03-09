package mtproto

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/andreyvit/telegramapi/binints"
)

type OutgoingMsg struct {
	Encrypted bool
	Payload   []byte
}

func NormalMsg(b []byte) OutgoingMsg {
	return OutgoingMsg{true, b}
}
func UnencryptedMsg(b []byte) OutgoingMsg {
	return OutgoingMsg{false, b}
}

type Framer struct {
	MsgIDOverride uint64

	gen MsgIDGen
}

func (fr *Framer) Format(msg OutgoingMsg) ([]byte, error) {
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

func (fr *Framer) Parse(msg []byte) ([]byte, error) {
	r := bytes.NewReader(msg)
	var payload []byte
	var a Accum

	authKeyID, err := binints.ReadUint64LE(r)
	a.Push(err)

	if authKeyID == 0 {
		msgID, err := binints.ReadUint64LE(r)
		a.Push(err)

		msgLen, err := binints.ReadUint32LE(r)
		a.Push(err)

		payload, err = ReadN(r, int(msgLen))
		a.Push(err)

		log.Printf("Received unencrypted: msgID=%x msgLen=%d err=%v, payload: %s", msgID, msgLen, a.Error(), hex.EncodeToString(payload))
	} else {
		// log.Printf("Received encrypted: authKeyID=%x msgID=%x msgLen=%d cmd = %08x", authKeyID, msgID, msgLen, cmd)
		panic("authKeyID != 0")
	}

	a.Push(binints.ExpectEOF(r))
	return payload, a.Error()
}
