package mtproto

import (
	"crypto/rsa"
	"encoding/hex"
	"io"
	"log"
)

type Transport interface {
	Send(data []byte) error
	Recv() ([]byte, error)
	Close()
}

type SessionOptions struct {
	PubKey  *rsa.PublicKey
	AppID   string
	APIHash string
	Verbose int
}

type Session struct {
	options   SessionOptions
	transport Transport
	framer    *Framer
	keyex     *KeyEx

	failc  chan error
	sendc  chan Msg
	closec chan struct{}

	err error
}

func NewSession(transport Transport, options SessionOptions) *Session {
	return &Session{
		options:   options,
		transport: transport,
		framer:    &Framer{},
		keyex: &KeyEx{
			PubKey: options.PubKey,
		},

		failc:  make(chan error, 1),
		sendc:  make(chan Msg, 1),
		closec: make(chan struct{}),
	}
}

func (sess *Session) Send(msg Msg) {
	sess.sendc <- msg
}

func (sess *Session) Err() error {
	return sess.err
}

func (sess *Session) Run() {
	incomingc := make(chan []byte, 1)

	go sess.listen(incomingc)

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session running...")
	}

	sess.sendInternal(sess.keyex.Start())

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
		raw, err := sess.transport.Recv()
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
		return err
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session received %s (%v bytes, %v): %v", DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type, hex.EncodeToString(msg.Payload))
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session received %s (%v bytes, %v)", DescribeCmdOfPayload(msg.Payload), len(msg.Payload), msg.Type)
	}

	if !sess.keyex.IsFinished() {
		omsg, err := sess.keyex.Handle(msg)
		if err != nil {
			return err
		}
		if omsg != nil {
			sess.sendInternal(*omsg)
		}
	} else {
		log.Printf("mtproto.Session TODO: handle normal msg")
	}

	return nil
}

func (sess *Session) Close() {
	sess.transport.Close()
}

func (sess *Session) Wait() {
	// TODO
}
