package telegramapi

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

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
	c.netconn.SetReadDeadline(time.Now().Add(timeout))

	var sizebuf [3]byte
	n, err := c.netconn.Read(sizebuf[0:1])
	if _, ok := err.(net.Error); ok {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	// if n != 1 {
	// 	return nil, errors.New("partial read")
	// }

	var msglen int
	if sizebuf[0] == 0x7F {
		panic("cannot handle 7F yet")
	} else if sizebuf[0] >= 0x7F {
		panic("WAT?!")
	}
	msglen = int(sizebuf[0]) * 4

	data := make([]byte, msglen)
	c.netconn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	n, err = c.netconn.Read(data)
	if _, ok := err.(net.Error); ok {
		// timeout
	} else if err != nil {
		return nil, err
	}
	if n < msglen {
		return nil, errors.New("partial read")
	}

	log.Printf("Received %v bytes: %v", len(data), hex.EncodeToString(data))

	return data, nil
}

func (c *Conn) PrintMessage(msg []byte) {
	var authKeyID uint64
	reader := bytes.NewReader(msg)
	err := binary.Read(reader, binary.LittleEndian, &authKeyID)
	if err != nil {
		panic(err)
	}

	if authKeyID != 0 {
		panic("authKeyID != 0")
	}

	var msgID uint64
	err = binary.Read(reader, binary.LittleEndian, &msgID)
	if err != nil {
		panic(err)
	}

	var msgLen uint32
	err = binary.Read(reader, binary.LittleEndian, &msgLen)
	if err != nil {
		panic(err)
	}

	var cmd uint32
	err = binary.Read(reader, binary.LittleEndian, &cmd)
	if err != nil {
		panic(err)
	}

	off, _ := reader.Seek(0, io.SeekCurrent)

	log.Printf("Received: authKeyID=%x msgID=%x msgLen=%d cmd = %08x", authKeyID, msgID, msgLen, cmd)
	msgLen -= 4

	payload := msg[int(off) : int(off)+int(msgLen)]
	log.Printf("Received payload: %v", hex.EncodeToString(payload))
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
