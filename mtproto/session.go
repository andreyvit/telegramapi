package mtproto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/andreyvit/telegramapi/binints"
	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"io"
	"log"
	"sync"

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

var ErrReconnectRequired = errors.New("reconnect required")

var ErrInvalidMsg = errors.New("invalid message")

type Session struct {
	options   SessionOptions
	transport Transport
	framer    *Framer
	keyex     *KeyEx
	handlers  []Handler

	connKeyExDone bool
	connInitSent  bool
	inFlight      map[uint64]*rpcInFlight

	failc  chan error
	sendc  chan outgoingMsg
	closec chan struct{}
	eventc chan uint32

	stateMut  sync.Mutex
	stateCond *sync.Cond
	isReady   bool

	dc int

	onstatechanged func()

	err error
}

type outgoingMsg struct {
	Obj   tl.Object
	Reply chan<- reply
}

type reply struct {
	Obj tl.Object
	Err error
}

type rpcInFlight struct {
	MsgID uint64
	Reply chan<- reply
}

func NewSession(transport Transport, options SessionOptions) *Session {
	s := &Session{
		options:   options,
		transport: transport,
		framer:    &Framer{},
		keyex: &KeyEx{
			PubKey: options.PubKey,
		},

		inFlight: make(map[uint64]*rpcInFlight),

		failc:  make(chan error, 1),
		sendc:  make(chan outgoingMsg, 1),
		closec: make(chan struct{}),
		eventc: make(chan uint32, 10),
	}
	s.stateCond = sync.NewCond(&s.stateMut)
	s.AddHandler(s.handleKeyEx)
	s.AddHandler(s.handleRPCResult)
	return s
}

func (sess *Session) OnStateChanged(f func()) {
	sess.stateMut.Lock()
	defer sess.stateMut.Unlock()
	sess.onstatechanged = f
}

func (sess *Session) DC() int {
	sess.stateMut.Lock()
	defer sess.stateMut.Unlock()

	return sess.dc
}
func (sess *Session) SetDC(dc int) {
	sess.stateMut.Lock()
	defer sess.stateMut.Unlock()

	if sess.dc != 0 && sess.dc != dc {
		panic("cannot change session DC once it has been set")
	}
	sess.dc = dc
}

func (sess *Session) AddHandler(handler func(cmd uint32, o tl.Object) ([]tl.Object, error)) {
	h := Handler(handler)
	sess.handlers = append(sess.handlers, h)
}

func (sess *Session) WaitReady() {
	sess.stateMut.Lock()
	defer sess.stateMut.Unlock()

	for !sess.isReady {
		sess.stateCond.Wait()
	}
}

func (sess *Session) RunJob(f func() error) {
	go func() {
		sess.WaitReady()
		err := f()
		if err != nil {
			sess.failInternal(err)
		}
	}()
}

func (sess *Session) Send(o tl.Object) (tl.Object, error) {
	replyc := make(chan reply)
	sess.sendc <- outgoingMsg{o, replyc}
	reply := <-replyc
	return reply.Obj, reply.Err
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

	if !sess.connKeyExDone {
		sess.startKeyEx()
	}

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
		case msg := <-sess.sendc:
			sess.sendInternal(msg.Obj, msg.Reply)
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
		// panic("failed")
	}
}

func (sess *Session) sendInternal(o tl.Object, replyc chan<- reply) {
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

	// if sess.options.Verbose >= 2 {
	// 	log.Printf("mtproto.Session sending %s (%v bytes, %v)", o, len(msg.Payload), msg.Type)
	// } else if sess.options.Verbose >= 1 {
	// 	log.Printf("mtproto.Session sending %s (%v bytes, %v)", tl.Name(o), len(msg.Payload), msg.Type)
	// }

	sess.stateMut.Lock()
	raw, msgID, err := sess.framer.Format(msg)
	sess.stateMut.Unlock()
	if err != nil {
		sess.failInternal(err)
		return
	}

	if sess.options.Verbose >= 2 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v, msgID %08x)", o, len(msg.Payload), msg.Type, msgID)
	} else if sess.options.Verbose >= 1 {
		log.Printf("mtproto.Session sending %s (%v bytes, %v, msgID %08x)", tl.Name(o), len(msg.Payload), msg.Type, msgID)
	}

	if replyc != nil {
		sess.startPendingRPC(msgID, replyc)
	}

	err = sess.transport.Send(raw)
	if err != nil {
		sess.failInternal(err)
		return
	}
}

func (sess *Session) startPendingRPC(msgID uint64, replyc chan<- reply) {
	if sess.inFlight[msgID] != nil {
		panic("duplicate msgID")
	}

	infl := &rpcInFlight{msgID, replyc}
	sess.inFlight[msgID] = infl
}

func (sess *Session) finishPendingRPC(msgID uint64, obj tl.Object, err error) {
	infl := sess.inFlight[msgID]
	if infl == nil {
		log.Printf("WARNING: dropping reply to unknown msgID %08x: %v, %v", msgID, obj, err)
		return
	}
	delete(sess.inFlight, msgID)

	infl.Reply <- reply{obj, err}
}

func (sess *Session) handle(msg []byte) {
	err := sess.doHandle(msg)
	if err != nil {
		sess.failInternal(err)
	}
}

func (sess *Session) doHandle(raw []byte) error {
	sess.stateMut.Lock()
	msg, err := sess.framer.Parse(raw)
	sess.stateMut.Unlock()
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
			sess.sendInternal(msg, nil)
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
		sess.processResult(msgs, err)
	}
}

func (sess *Session) processResult(msgs []tl.Object, err error) {
	if err != nil && err != ErrCmdNotHandled {
		sess.failInternal(err)
		return
	}
	for _, msg := range msgs {
		sess.sendInternal(msg, nil)
	}
}

func (sess *Session) startKeyEx() {
	omsg := sess.keyex.Start()
	sess.processResult([]tl.Object{omsg}, nil)
}

func (sess *Session) handleKeyEx(cmd uint32, o tl.Object) ([]tl.Object, error) {
	if sess.connKeyExDone {
		return nil, ErrCmdNotHandled
	}

	if o != nil {
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
			sess.applyAuth(auth)
			return []tl.Object{}, nil
		}
	} else {
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) handleRPCResult(cmd uint32, o tl.Object) ([]tl.Object, error) {
	switch o := o.(type) {
	case *TLRPCResult:
		sess.finishPendingRPC(o.ReqMsgID, o.Result, nil)
		return nil, nil
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
	case *TLBadServerSalt:
		sess.stateMut.Lock()
		auth, _ := sess.framer.State()
		sess.stateMut.Unlock()
		binints.EncodeUint64LE(o.NewServerSalt, auth.ServerSalt[:])
		sess.applyAuth(auth)
		return nil, ErrReconnectRequired
	case *TLBadMsgNotification:
		log.Printf("WARNING: bad msg %08x: err code %d, seq no %d", o.BadMsgID, o.ErrorCode, o.BadMsgSeqno)
		sess.finishPendingRPC(o.BadMsgID, nil, ErrInvalidMsg)
		return nil, nil
	default:
		return nil, ErrCmdNotHandled
	}
}

func (sess *Session) AuthState() (*AuthResult, FramerState) {
	sess.stateMut.Lock()
	defer sess.stateMut.Unlock()

	return sess.framer.State()
}

func (sess *Session) RestoreAuthState(auth *AuthResult, fs FramerState) {
	sess.stateMut.Lock()
	sess.framer.Restore(fs)
	sess.stateMut.Unlock()

	sess.applyAuth(auth)
}

func (sess *Session) applyAuth(auth *AuthResult) {
	var zero [8]byte
	if bytes.Equal(zero[:], auth.SessionID[:]) {
		_, err := io.ReadFull(rand.Reader, auth.SessionID[:])
		if err != nil {
			panic(err)
		}
	}

	sess.stateMut.Lock()

	sess.framer.SetAuth(auth)

	sess.connKeyExDone = true
	sess.isReady = true
	sess.stateCond.Broadcast()
	f := sess.onstatechanged
	sess.stateMut.Unlock()

	if f != nil {
		f()
	}
}

func (sess *Session) Shutdown() {
	sess.transport.Close()
}

func (sess *Session) Wait() {
	// TODO
}
