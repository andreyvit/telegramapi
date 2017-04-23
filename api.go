package telegramapi

import (
	"log"

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

	state *State

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

func (c *Conn) waitForReadiness() {
	c.session.WaitReady()

	// TODO: locking
	auth, fs := c.session.AuthState()
	c.state.Auth = *auth
	c.state.FramerState = fs
	newState := *c.state

	c.delegateQueue <- func() {
		c.delegate.HandleStateChanged(newState)
		c.delegate.HandleConnectionReady()
	}
}

func (c *Conn) Run() error {
	log.Printf("Running with state: %v", pretty.Sprint(c.state))

	pubKey, err := mtproto.ParsePublicKey(c.PublicKey)
	if err != nil {
		return err
	}

	tr, err := mtproto.DialTCP(c.Endpoint, mtproto.TCPTransportOptions{})
	if err != nil {
		return err
	}

	c.session = mtproto.NewSession(tr, mtproto.SessionOptions{
		PubKey:  pubKey,
		Verbose: c.Verbose,
	})

	if c.state.Auth.KeyID != 0 {
		c.session.RestoreAuthState(&c.state.Auth, c.state.FramerState)
	}

	go c.dispatchDelegateCalls()
	go c.waitForReadiness()

	c.session.Run()
	return c.session.Err()
}
