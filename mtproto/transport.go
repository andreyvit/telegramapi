package mtproto

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/andreyvit/telegramapi/binints"
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
	r.SetReadDeadline(time.Now().Add(firstByteTimeout))

	msglen, err := ReadAbridgedTCPMessageLen(r)
	if _, ok := err.(net.Error); ok {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if msglen > maxMsgLen {
		return nil, errors.New("message too large")
	}

	data := make([]byte, msglen)
	r.SetReadDeadline(time.Now().Add(msgTimeout))
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type TCPTransport struct {
}
