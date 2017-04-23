package telegramapi

import (
	"errors"
	"log"
	"sync"

	"github.com/kr/pretty"

	"github.com/andreyvit/telegramapi/mtproto"
	"github.com/andreyvit/telegramapi/tl"
)

type Options struct {
	Endpoint  string
	PublicKey string
	Verbose   int

	APIID   int
	APIHash string
}

type Conn struct {
	Options

	delegate      Delegate
	delegateQueue chan func()

	state    *State
	stateMut sync.Mutex

	session *mtproto.Session
}

type Delegate interface {
	HandleConnectionReady()
	HandleStateChanged(newState State)
}

func New(options Options, state *State, delegate Delegate) *Conn {
	if options.Endpoint == "" {
		panic("configuration error: missing endpoint")
	}
	if options.PublicKey == "" {
		panic("configuration error: missing public key")
	}

	return &Conn{
		Options:  options,
		delegate: delegate,
		state:    state,
		session:  nil,

		delegateQueue: make(chan func(), 1),
	}
}

func (c *Conn) Send(o tl.Object) (tl.Object, error) {
	return c.session.Send(o)
}

func (c *Conn) Shutdown() {
	c.session.Shutdown()
}

func (c *Conn) dispatchDelegateCalls() {
	for f := range c.delegateQueue {
		f()
	}
}

func (c *Conn) runProcessing() {
	c.session.WaitReady()

	err := c.runProcessingErr()
	if err == nil {
		c.delegateQueue <- func() {
			c.delegate.HandleConnectionReady()
		}
	} else {
		c.session.Fail(err)
	}
}

func (c *Conn) runProcessingErr() error {
	r, err := c.session.Send(&mtproto.TLHelpGetConfig{})
	if err != nil {
		return err
	}

	switch r := r.(type) {
	case *mtproto.TLConfig:
		c.updateState(func(state *State) {
			state.KnownDCs = processDCs(r)
		})
	default:
		return c.HandleUnknownReply(r)
	}

	return nil
}

func (c *Conn) updateState(f func(state *State)) {
	c.stateMut.Lock()
	f(c.state)
	newState := *c.state
	c.stateMut.Unlock()

	c.delegateQueue <- func() {
		c.delegate.HandleStateChanged(newState)
	}
}

func (c *Conn) SwitchToDC(dc int) {
	c.updateState(func(state *State) {
		state.PreferredDC = dc
	})
}

func (c *Conn) HandleUnknownReply(r tl.Object) error {
	log.Printf("Unknown reploy: %v", r)
	return errors.New("unknown reply")
}

func (c *Conn) saveSessionState() {
	auth, fs := c.session.AuthState()
	c.updateState(func(state *State) {
		state.Auth = *auth
		state.FramerState = fs
		log.Printf("saveSessionState: %v", pretty.Sprint(c.state))
	})
}

func (c *Conn) Run() error {
	for {
		err := c.runInternal()
		if err != mtproto.ErrReconnectRequired {
			return err
		}
	}
}

func (c *Conn) runInternal() error {
	log.Printf("Running with state: %v", pretty.Sprint(c.state))

	dc := c.state.findPreferredDC()

	endpoint := c.Endpoint
	if dc != nil {
		endpoint = dc.PrimaryAddr.Endpoint()
		log.Printf("Will connect to DC %v at %v", dc.ID, endpoint)
	} else if c.state.PreferredDC != 0 {
		log.Printf("** WARNING: preferred DC %v not found, will connect to default DC at %v", c.state.PreferredDC, endpoint)
	} else {
		log.Printf("Will connect to default DC at %v", endpoint)
	}

	pubKey, err := mtproto.ParsePublicKey(c.PublicKey)
	if err != nil {
		return err
	}

	tr, err := mtproto.DialTCP(endpoint, mtproto.TCPTransportOptions{})
	if err != nil {
		return err
	}

	c.session = mtproto.NewSession(tr, mtproto.SessionOptions{
		PubKey:  pubKey,
		Verbose: c.Verbose,
	})

	c.session.OnStateChanged(c.saveSessionState)

	if c.state.Auth.KeyID != 0 {
		c.session.RestoreAuthState(&c.state.Auth, c.state.FramerState)
	}

	go c.dispatchDelegateCalls()
	go c.runProcessing()

	c.session.Run()
	c.saveSessionState()
	return c.session.Err()
}
