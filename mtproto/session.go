package mtproto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"io"
	"log"

	"github.com/andreyvit/telegramapi/tl"
)

//go:generate go install ../tl/cmd/tlc
//go:generate tlc -o generated.go mtproto telegram

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

type Handler func(cmd uint32, o tl.Object) ([]tl.Object, error)

var ErrCmdNotHandled = errors.New("not handled")

type Session struct {
	options   SessionOptions
	transport Transport
	framer    *Framer
	keyex     *KeyEx
	handlers  []Handler

	connKeyExDone bool
	connInitSent  bool

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

func (sess *Session) AddHandler(handler func(cmd uint32, o tl.Object) ([]tl.Object, error)) {
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

func (sess *Session) sendInternal(o tl.Object) {
	if sess.err != nil {
		return
	}

	if sess.connKeyExDone && !sess.connInitSent {
		o = &TLInvokeWithLayer{
			Layer: knownschemas.TelegramLayer,
			Query: &TLInitConnection{
				APIID:         88766,
				DeviceModel:   "Mac",
				SystemVersion: "10.11",
				AppVersion:    "0.1",
				LangCode:      "en",
				Query:         o,
			},
		}
		sess.connInitSent = true
	}

	msg := MsgFromObj(o)

	raw, err := sess.framer.Format(msg)
	if err != nil {
		sess.failInternal(err)
		return
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v): %v", o, len(msg.Payload), msg.Type, hex.EncodeToString(msg.Payload))
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v)", tl.Name(o), len(msg.Payload), msg.Type)
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

	o, err := Schema.ReadBoxedObject(msg.Payload)
	if err != nil {
		if sess.options.Verbose >= 2 {
			log.Printf("mtproto.Session received %s (%v bytes, %v): %v", Schema.DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type, hex.EncodeToString(msg.Payload))
		} else if sess.options.Verbose >= 1 {
			log.Printf("mtproto.Session received %s (%v bytes, %v)", Schema.DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type)
		}
		return err
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session received %v (%v bytes, %v)", o, len(msg.Payload), msg.Type)
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session received %s (%v bytes, %v)", tl.Name(o), len(msg.Payload), msg.Type)
	}

	sess.invokeHandlersInternal(o)

	return nil
}

func (sess *Session) invokeHandlersInternal(o tl.Object) {
	msgs, err := sess.invokeHandlersInternalReturnCmds(o)
	if err == ErrCmdNotHandled {
		sess.logDroppedIncomingMsg(o)
	} else if err != nil {
		sess.failInternal(err)
	} else {
		for _, msg := range msgs {
			sess.sendInternal(msg)
		}
	}
}

func (sess *Session) logDroppedIncomingMsg(o tl.Object) {
	if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session: dropping unhandled message %v", o)
	}
}

func (sess *Session) invokeHandlersInternalReturnCmds(o tl.Object) ([]tl.Object, error) {
	for _, h := range sess.handlers {
		msgs, err := h(o.Cmd(), o)
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
		log.Printf("mtproto.Session broadcasting %08x", cmd)
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

func (sess *Session) handleKeyEx(cmd uint32, o tl.Object) ([]tl.Object, error) {
	if sess.connKeyExDone {
		return nil, ErrCmdNotHandled
	}

	if cmd == PseudoIDKeyExStart {
		omsg := sess.keyex.Start()
		return []tl.Object{omsg}, nil
	} else if o != nil {
		omsg, err := sess.keyex.Handle(o)
		if err != nil {
			return nil, err
		}
		if omsg != nil {
			return []tl.Object{omsg}, nil
		} else {
			auth, err := sess.keyex.Result()
			if err != nil {
				return nil, err
			}
			sess.ApplyAuth(auth)
			return []tl.Object{}, nil
		}
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) handleConfig(cmd uint32, o tl.Object) ([]tl.Object, error) {
	if cmd == PseudoIDHandshakeDone {
		msg := &TLHelpGetNearestDC{}
		return []tl.Object{msg}, nil
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) handleRPCResult(cmd uint32, o tl.Object) ([]tl.Object, error) {
	switch o := o.(type) {
	case *TLRPCResult:
		msgs, err := sess.invokeHandlersInternalReturnCmds(o.Result)
		if err == ErrCmdNotHandled {
			sess.logDroppedIncomingMsg(o.Result)
			return nil, nil
		} else {
			return msgs, err
		}
	case *TLMsgContainer:
		var replies []tl.Object
		for _, msg := range o.Messages {
			// TODO: verify msg.MsgId
			// TODO: verify msg.Seqno
			r, err := sess.invokeHandlersInternalReturnCmds(msg.Body)
			if err == ErrCmdNotHandled {
				sess.logDroppedIncomingMsg(msg.Body)
			} else if err != nil {
				return nil, err
			}
			replies = append(replies, r...)
		}
		return replies, nil
	case *TLNewSessionCreated:
		log.Printf("NOTICE: %v", o)
		return nil, nil
	case *TLMsgsAck:
		for _, msgID := range o.MsgIDs {
			// TODO: ack
			log.Printf("TODO: ack'ed %08x", msgID)
		}
		return nil, nil
	default:
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
	sess.connKeyExDone = true
	sess.Notify(PseudoIDHandshakeDone)
}

func (sess *Session) Close() {
	sess.transport.Close()
}

func (sess *Session) Wait() {
	// TODO
}
