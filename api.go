package telegramapi

import (
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"log"
	"net"
	"time"

	"github.com/andreyvit/telegramapi/mtproto"
)

const maxMsgLen = 1024 * 1024 * 10

type Options struct {
	Endpoint  string
	PublicKey string
}

type Conn struct {
	Options

	pubKey *rsa.PublicKey

	netconn net.Conn
	framer  *mtproto.Framer
	keyex   *mtproto.KeyEx

	efSent bool
}

const msgUseLayer18 uint32 = 0x1c900537
const msgUseLayer2 uint32 = 0x289dd1f6
const helpGetConfig = 0xc4f9186b

var ErrAuthTimeout = errors.New("authentication timeout")

func Connect(options Options) (*Conn, error) {
	if options.Endpoint == "" {
		return nil, errors.New("configuration error: missing endpoint")
	}
	if options.PublicKey == "" {
		return nil, errors.New("configuration error: missing public key")
	}
	pubKey, err := mtproto.ParsePublicKey(options.PublicKey)
	if err != nil {
		return nil, err
	}

	netconn, err := net.Dial("tcp", options.Endpoint)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Options: options,
		pubKey:  pubKey,
		netconn: netconn,
		framer:  &mtproto.Framer{},
		keyex: &mtproto.KeyEx{
			PubKey: pubKey,
		},
	}, nil
}

func (c *Conn) Close() {
	c.netconn.Close()
}

func (c *Conn) SayHello() error {
	msg := c.keyex.Start()
	err := c.send(msg)
	if err != nil {
		return err
	}

	for !c.keyex.IsFinished() {
		payload, err := c.ReadMessage(10 * time.Second)
		if err != nil {
			return err
		}
		if payload == nil {
			return ErrAuthTimeout
		}

		msg, err := c.keyex.Handle(payload)
		if err != nil {
			return err
		}
		if msg != nil {
			err := c.send(*msg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) send(msg mtproto.OutgoingMsg) error {
	data, err := c.framer.Format(msg)
	if err != nil {
		return err
	}

	raw := c.formatTCPMessage(data)
	log.Printf("Sending %v bytes: %v", len(raw), hex.EncodeToString(raw))
	n, err := c.netconn.Write(raw)
	if err != nil {
		return err
	}
	if n < len(raw) {
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

	data, err := c.framer.Parse(raw)
	if err != nil {
		return nil, err
	}

	return data, nil
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

func (c *Conn) formatTCPMessage(data []byte) []byte {
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
		binints.WriteUint24LE(&buf, uint32(l))
	}

	buf.Write(data)

	return buf.Bytes()
}
