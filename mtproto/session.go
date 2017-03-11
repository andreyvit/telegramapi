package mtproto

import (
	"crypto/rsa"
	"io"
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
	}
}

func (sess *Session) Send(msg Msg) {
	sess.sendc <- msg
}

func (sess *Session) Run() {
	incomingc := make(chan []byte, 1)

	go sess.listen(incomingc)

loop:
	for {
		select {
		case raw, ok := <-incomingc:
			if ok {
				sess.handle(raw)
			} else {
				break loop
			}
		case err := <-sess.failc:
			sess.failInternal(err)
		}
	}
}

func (sess *Session) listen(incomingc chan<- []byte) {
	for {
		raw, err := sess.transport.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			sess.Fail(err)
			break
		}

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

	// TODO
	_ = msg

	return nil
}

func (sess *Session) Close() {
	sess.transport.Close()
}

func (sess *Session) Wait() {
	// TODO
}
