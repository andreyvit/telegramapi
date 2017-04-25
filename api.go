package telegramapi

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/kr/pretty"

	"github.com/andreyvit/telegramapi/mtproto"
	"github.com/andreyvit/telegramapi/tl"
)

type Options struct {
	SeedAddr  Addr
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
	if options.SeedAddr.IP == "" {
		panic("configuration error: missing SeedAddr")
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
		c.session.SetDC(r.ThisDC)
		c.updateState(func(state *State) {
			updateDCs(state.DCs, r)
		})
	default:
		return c.HandleUnknownReply(r)
	}

	return nil
}

func (c *Conn) Fail(err error) {
	c.session.Fail(err)
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
	switch r := r.(type) {
	case *mtproto.TLRPCError:
		if r.ErrorCode == 303 {
			if nstr := stripPrefix(r.ErrorMessage, "PHONE_MIGRATE_"); nstr != "" {
				n, err := strconv.Atoi(nstr)
				if err != nil {
					return errors.New("X not numeric in PHONE_MIGRATE_X")
				}
				c.SwitchToDC(n)
				return mtproto.ErrReconnectRequired
			}
		}
		log.Printf("RPC error: %v", r)
		return fmt.Errorf("telegram error: %s", r.ErrorMessage)
	default:
		log.Printf("Unknown reply: %v", r)
		return errors.New("unknown reply")
	}
}

func (c *Conn) saveSessionState() {
	auth, fs := c.session.AuthState()
	c.updateState(func(state *State) {
		dc := state.DCs[c.session.DC()]
		if dc != nil {
			dc.Auth = *auth
			dc.FramerState = fs
		}
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
	c.state.initialize()
	log.Printf("Running with state: %v", pretty.Sprint(c.state))

	dc := c.state.findPreferredDC()

	if dc != nil {
		log.Printf("Will connect to DC %v at %v", dc.ID, dc.PrimaryAddr.Endpoint())
	} else {
		dc = &DCState{
			ID:          0,
			PrimaryAddr: c.SeedAddr,
		}
		if c.state.PreferredDC != 0 {
			log.Printf("** WARNING: preferred DC %v not found, will connect to default DC at %v", c.state.PreferredDC, dc.PrimaryAddr.Endpoint())
		} else {
			log.Printf("Will connect to default DC at %v", dc.PrimaryAddr.Endpoint())
		}
	}

	pubKey, err := mtproto.ParsePublicKey(c.PublicKey)
	if err != nil {
		return err
	}

	tr, err := mtproto.DialTCP(dc.PrimaryAddr.Endpoint(), mtproto.TCPTransportOptions{})
	if err != nil {
		return err
	}

	c.session = mtproto.NewSession(tr, mtproto.SessionOptions{
		PubKey:  pubKey,
		Verbose: c.Verbose,
	})

	c.session.OnStateChanged(c.saveSessionState)

	if dc.Auth.KeyID != 0 {
		c.session.RestoreAuthState(&dc.Auth, dc.FramerState)
	} else {
		c.state.LoginState = LoggedOut
	}

	go c.dispatchDelegateCalls()
	go c.runProcessing()

	c.session.Run()
	c.saveSessionState()
	return c.session.Err()
}
