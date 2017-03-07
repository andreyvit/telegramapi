package telegramapi

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"log"
	"net"
	"time"

	"github.com/andreyvit/telegramapi/mtproto"
)

const maxMsgLen = 1024 * 1024 * 10

type Conn struct {
	netconn net.Conn

	efSent bool

	seq uint32
}

const testEndpoint = "149.154.167.40:443"
const productionEndpoint = "149.154.167.50:443"

const msgUseLayer18 uint32 = 0x1c900537
const msgUseLayer2 uint32 = 0x289dd1f6
const helpGetConfig = 0xc4f9186b
const msgReqPQ = 0x60469778

func Connect() (*Conn, error) {
	netconn, err := net.Dial("tcp", testEndpoint)
	if err != nil {
		return nil, err
	}

	return &Conn{netconn: netconn}, nil
}

func (c *Conn) Close() {
	c.netconn.Close()
}

func (c *Conn) SayHello() error {
	var buf bytes.Buffer

	// writeUint32(&buf, helpGetConfig)
	// writeUint32(&buf, msgUseLayer2)
	writeUint32(&buf, msgReqPQ)
	writeUint128(&buf, 0x60469778, 0xc4f9186b)

	err := c.SendUnencrypted(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) SendUnencrypted(payload []byte) error {
	var buf bytes.Buffer
	writeUint64(&buf, 0)

	msgID := c.generateMsgID()
	writeUint64(&buf, msgID)

	writeUint32(&buf, uint32(len(payload)))
	buf.Write(payload)

	return c.sendRaw(buf.Bytes())
}

func (c *Conn) sendRaw(data []byte) error {
	msg := c.formatMessage(data)
	log.Printf("Sending %v bytes: %v", len(msg), hex.EncodeToString(msg))
	n, err := c.netconn.Write(msg)
	if err != nil {
		return err
	}
	if n < len(msg) {
		return errors.New("sent partly failed")
	}
	return nil
}

func (c *Conn) ReadMessage(timeout time.Duration) ([]byte, error) {
	raw, err := mtproto.ReadAbridgedTCPMessage(c.netconn, maxMsgLen, timeout, 1*time.Minute)
	if raw == nil || err != nil {
		return nil, err
	}

	log.Printf("Received %v raw bytes: %v", len(raw), hex.EncodeToString(raw))

	data, err := c.decrypt(raw)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Conn) decrypt(msg []byte) ([]byte, error) {
	var payload []byte
	var a mtproto.Accum
	r := bytes.NewReader(msg)

	authKeyID, err := binints.ReadUint64LE(r)
	a.Push(err)

	if authKeyID == 0 {
		msgID, err := binints.ReadUint64LE(r)
		a.Push(err)

		msgLen, err := binints.ReadUint32LE(r)
		a.Push(err)

		payload, err = mtproto.ReadN(r, int(msgLen))
		a.Push(err)

		log.Printf("Received unencrypted: msgID=%x msgLen=%d err=%v, payload: %s", msgID, msgLen, a.Error(), hex.EncodeToString(payload))
	} else {
		// log.Printf("Received encrypted: authKeyID=%x msgID=%x msgLen=%d cmd = %08x", authKeyID, msgID, msgLen, cmd)
		panic("authKeyID != 0")
	}

	a.Push(binints.ExpectEOF(r))
	return payload, a.Error()
}

func (c *Conn) PrintMessage(msg []byte) {
	r := bytes.NewReader(msg)
	var a mtproto.Accum

	cmd, err := binints.ReadUint32LE(r)
	a.Push(err)

	if cmd == mtproto.IDResPQ {
		var res mtproto.ResPQ
		err = mtproto.ReadResPQ(r, &res)
		a.Push(err)

		log.Printf("res_pq#%08x: %+#v", cmd, res)
	} else {
		log.Printf("Unknown cmd: %08x", cmd)
	}

	log.Printf("Err: %v", a.Error())
}

func (c *Conn) formatMessage(data []byte) []byte {
	var buf bytes.Buffer

	if !c.efSent {
		buf.WriteByte(0xEF)
		c.efSent = true
	}

	l := len(data)
	if l%4 != 0 {
		panic("Message length not divisible by 4")
	}
	if l == 0 {
		panic("Cannot send empty message")
	}
	l /= 4

	if l <= 0x7e {
		buf.WriteByte(byte(l))
	} else {
		buf.WriteByte(0x7F)
		writeUint24(&buf, uint32(l))
	}

	buf.Write(data)

	return buf.Bytes()
}
