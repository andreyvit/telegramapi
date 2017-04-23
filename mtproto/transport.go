package mtproto

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/andreyvit/telegramapi/binints"
	"github.com/andreyvit/telegramapi/tl"
)

type TCPReader interface {
	io.Reader

	// see net.Conn.SetReadDeadline
	SetReadDeadline(t time.Time) error
}

func ReadAbridgedTCPMessageLen(r io.Reader) (int, error) {
	var sizebuf [3]byte
	_, err := io.ReadFull(r, sizebuf[0:1])
	if err != nil {
		return -1, err
	}

	if sizebuf[0] == 0x7F {
		_, err = io.ReadFull(r, sizebuf[0:3])
		if err != nil {
			return -1, err
		}

		return int(binints.DecodeUint24LE(sizebuf[0:3])) * 4, nil
	} else if sizebuf[0] > 0x7F {
		return -1, errors.New("unexpected message size byte >0x7F")
	} else {
		return int(sizebuf[0]) * 4, nil
	}
}

func ReadAbridgedTCPMessage(r TCPReader, maxMsgLen int, firstByteTimeout time.Duration, msgTimeout time.Duration) ([]byte, error) {
	if firstByteTimeout > 0 {
		r.SetReadDeadline(time.Now().Add(firstByteTimeout))
	}

	msglen, err := ReadAbridgedTCPMessageLen(r)
	if _, ok := err.(net.Error); ok {
		return nil, nil
	} else if err != nil {
		log.Printf("mtproto.TCPTransport: failed to read TCP message length: %v", msglen, err)
		return nil, err
	}

	if msglen > maxMsgLen {
		return nil, errors.New("message too large")
	}

	data := make([]byte, msglen)
	if msgTimeout > 0 {
		r.SetReadDeadline(time.Now().Add(msgTimeout))
	}
	_, err = io.ReadFull(r, data)
	if err != nil {
		log.Printf("mtproto.TCPTransport: failed to read TCP message (%d): %v", msglen, err)
		return nil, err
	}
	log.Printf("mtproto.TCPTransport: received %d bytes", len(data))

	return data, nil
}

func formatTCPMessage(data []byte, isFirst bool) []byte {
	var buf bytes.Buffer

	if isFirst {
		buf.WriteByte(0xEF)
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

type TCPTransportOptions struct {
	MaxMsgLen int
}

type TCPTransport struct {
	options TCPTransportOptions
	Conn    net.Conn

	firstSent bool
}

func DialTCP(endpoint string, options TCPTransportOptions) (*TCPTransport, error) {
	c, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}

	if options.MaxMsgLen == 0 {
		options.MaxMsgLen = 1024 * 1024 * 10
	}

	return &TCPTransport{
		options: options,
		Conn:    c,
	}, nil
}

func (tr *TCPTransport) Close() {
	tr.Conn.Close()
}

func (tr *TCPTransport) Send(data []byte) error {
	data = formatTCPMessage(data, !tr.firstSent)
	tr.firstSent = true
	log.Printf("mtproto.TCPTransport: sending %d bytes", len(data))
	_, err := tr.Conn.Write(data)
	return err
}

func (tr *TCPTransport) Recv() ([]byte, int, error) {
	raw, err := ReadAbridgedTCPMessage(tr.Conn, tr.options.MaxMsgLen, 0, 0)
	if err != nil {
		return nil, 0, err
	}
	if raw == nil {
		return nil, 0, io.EOF
	}

	if len(raw) == 4 {
		errcode := int(int32(tl.NewReader(raw).Cmd()))
		return nil, errcode, nil
	}

	return raw, 0, nil
}
