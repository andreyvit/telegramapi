package mtproto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
)

type Transport interface {
	Send(data []byte) error
	Recv() ([]byte, int, error)
	Close()
}

type SessionOptions struct {
	PubKey  *rsa.PublicKey
	AppID   string
	APIHash string
	Verbose int
}

type Handler func(cmd uint32, r *Reader) ([]Msg, error)

var ErrCmdNotHandled = errors.New("not handled")

type Session struct {
	options   SessionOptions
	transport Transport
	framer    *Framer
	keyex     *KeyEx
	handlers  []Handler

	failc  chan error
	sendc  chan Msg
	closec chan struct{}
	eventc chan uint32

	err error
}

const (
	PseudoIDInvalidCommand uint32 = iota
	PseudoIDKeyExStart
	PseudoIDHandshakeDone
)

func NewSession(transport Transport, options SessionOptions) *Session {
	s := &Session{
		options:   options,
		transport: transport,
		framer:    &Framer{},
		keyex: &KeyEx{
			PubKey: options.PubKey,
		},

		failc:  make(chan error, 1),
		sendc:  make(chan Msg, 1),
		closec: make(chan struct{}),
		eventc: make(chan uint32, 10),
	}
	s.AddHandler(s.handleKeyEx)
	s.AddHandler(s.handleConfig)
	s.AddHandler(s.handleRPCResult)
	return s
}

func (sess *Session) AddHandler(handler func(cmd uint32, r *Reader) ([]Msg, error)) {
	h := Handler(handler)
	sess.handlers = append(sess.handlers, h)
}

func (sess *Session) Send(msg Msg) {
	sess.sendc <- msg
}

func (sess *Session) Err() error {
	return sess.err
}

func (sess *Session) Notify(pseudocmd uint32) {
	sess.eventc <- pseudocmd
}

func (sess *Session) Run() {
	incomingc := make(chan []byte, 1)

	go sess.listen(incomingc)

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session running...")
	}

	sess.eventc <- PseudoIDKeyExStart

loop:
	for sess.err == nil {
		select {
		case raw, ok := <-incomingc:
			if ok {
				sess.handle(raw)
			} else {
				if sess.options.Verbose >= 2 {
					log.Printf("mtproto.Session incoming closed")
				}
				break loop
			}
		case err := <-sess.failc:
			sess.failInternal(err)
		case pseudocmd := <-sess.eventc:
			sess.broadcastInternal(pseudocmd)
		}
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session quitting, err: %v", sess.err)
	}
}

func (sess *Session) listen(incomingc chan<- []byte) {
	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session listening...")
	}
	for {
		raw, errcode, err := sess.transport.Recv()
		if err == io.EOF {
			if sess.options.Verbose >= 2 {
				log.Printf("mtproto.Session Recv'd EOF")
			}
			break
		} else if err != nil {
			if sess.options.Verbose >= 2 {
				log.Printf("mtproto.Session Recv failed: %v", err)
			}
			sess.failc <- err
			break
		} else if raw == nil && errcode != 0 {
			if sess.options.Verbose >= 1 {
				log.Printf("mtproto.Session Recv returned error code %v", errcode)
			}
			sess.failc <- fmt.Errorf("error code %v", errcode)
			break
		}
		// if sess.options.Verbose >= 2 {
		// 	log.Printf("mtproto.Session Recv'ed %d bytes", len(raw))
		// }

		incomingc <- raw
	}
	close(incomingc)
}

func (sess *Session) Fail(err error) {
	if err == nil {
		panic("Fail(nil)")
	}
	sess.failc <- err
}

func (sess *Session) failInternal(err error) {
	if err == nil {
		panic("fail(nil)")
	}
	if sess.err == nil {
		sess.err = err
		if sess.options.Verbose >= 1 {
			log.Printf("mtproto.Session failed: %v", err)
		}
		panic("failed")
	}
}

func (sess *Session) sendInternal(msg Msg) {
	if sess.err != nil {
		return
	}

	raw, err := sess.framer.Format(msg)
	if err != nil {
		sess.failInternal(err)
		return
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v): %v", DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type, hex.EncodeToString(msg.Payload))
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v)", DescribeCmdOfPayload(msg.Payload), msg.Type, len(msg.Payload))
	}

	err = sess.transport.Send(raw)
	if err != nil {
		sess.failInternal(err)
		return
	}
}

func (sess *Session) handle(msg []byte) {
	err := sess.doHandle(msg)
	if err != nil {
		sess.failInternal(err)
	}
}

func (sess *Session) doHandle(raw []byte) error {
	msg, err := sess.framer.Parse(raw)
	if err != nil {
		if sess.options.Verbose >= 2 {
			log.Printf("mtproto.Session failed to parse incoming data (%v bytes): %v - error: %v", len(raw), hex.EncodeToString(raw), err)
		} else if sess.options.Verbose >= 1 {
			log.Printf("mtproto.Session failed to parse incoming data (%v bytes) - error: %v", len(raw), err)
		}
		return err
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session received %s (%v bytes, %v): %v", DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type, hex.EncodeToString(msg.Payload))
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session received %s (%v bytes, %v)", DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type)
	}

	r := NewReader(msg.Payload)
	sess.invokeHandlersInternal(r.Cmd(), r)

	return nil
}

func (sess *Session) invokeHandlersInternal(cmd uint32, r *Reader) {
	msgs, err := sess.invokeHandlersInternalReturnCmds(cmd, r)
	if err == ErrCmdNotHandled {
		if sess.options.Verbose >= 1 {
			log.Printf("mtproto.Session: dropping unhandled message %s (within %s)", DescribeCmd(r.Cmd()), DescribeCmd(cmd))
		}
	} else if err != nil {
		sess.failInternal(err)
	} else {
		for _, msg := range msgs {
			sess.sendInternal(msg)
		}
	}
}

func (sess *Session) invokeHandlersInternalReturnCmds(cmd uint32, r *Reader) ([]Msg, error) {
	for _, h := range sess.handlers {
		msgs, err := h(cmd, r)
		if err == ErrCmdNotHandled {
			continue
		} else if err != nil {
			return nil, err
		} else {
			return msgs, nil
		}
	}

	return nil, ErrCmdNotHandled
}

func (sess *Session) broadcastInternal(cmd uint32) {
	if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session broadcasting %s", CmdName(cmd))
	}
	for _, h := range sess.handlers {
		msgs, err := h(cmd, nil)
		if err != nil && err != ErrCmdNotHandled {
			sess.failInternal(err)
			return
		}
		for _, msg := range msgs {
			sess.sendInternal(msg)
		}
	}
}

func (sess *Session) handleKeyEx(cmd uint32, r *Reader) ([]Msg, error) {
	if sess.keyex.IsFinished() {
		return nil, ErrCmdNotHandled
	}

	if cmd == PseudoIDKeyExStart {
		omsg := sess.keyex.Start()
		return []Msg{omsg}, nil
	} else if r != nil {
		omsg, err := sess.keyex.Handle(r)
		if err != nil {
			return nil, err
		}
		if omsg != nil {
			return []Msg{*omsg}, nil
		} else {
			auth, err := sess.keyex.Result()
			if err != nil {
				return nil, err
			}
			sess.ApplyAuth(auth)
			return []Msg{}, nil
		}
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) handleConfig(cmd uint32, r *Reader) ([]Msg, error) {
	if cmd == PseudoIDHandshakeDone {
		w := NewWriterCmd(Cmd("invokeWithLayer"))
		w.WriteInt(apiLayer)
		w.WriteCmd(Cmd("initConnection"))
		w.WriteInt(88766)
		w.WriteBlobStr("Mac")
		w.WriteBlobStr("10.11")
		w.WriteBlobStr("0.1")
		w.WriteBlobStr("en")
		w.WriteCmd(Cmd("help.getNearestDc"))
		return []Msg{MakeMsg(w.Bytes(), ContentMsg)}, nil
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) handleRPCResult(cmd uint32, r *Reader) ([]Msg, error) {
	if cmd == Cmd("rpc_result") {
		reqMsgID := r.ReadUint64()
		_ = reqMsgID
		subcmd := r.StartInnerCmd()
		return sess.invokeHandlersInternalReturnCmds(subcmd, r)
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) ApplyAuth(auth *AuthResult) {
	var zero [8]byte
	if bytes.Equal(zero[:], auth.SessionID[:]) {
		_, err := io.ReadFull(rand.Reader, auth.SessionID[:])
		if err != nil {
			panic(err)
		}
	}

	sess.framer.SetAuth(auth)
	sess.Notify(PseudoIDHandshakeDone)
}

func (sess *Session) Close() {
	sess.transport.Close()
}

func (sess *Session) Wait() {
	// TODO
}

func init() {
	RegisterCmd(PseudoIDKeyExStart, "KeyExStart", "")
	RegisterCmd(PseudoIDHandshakeDone, "HandshakeDone", "")
}
