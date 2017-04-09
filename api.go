package telegramapi

import (
	"crypto/rsa"
	"errors"

	"github.com/andreyvit/telegramapi/mtproto"
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

	pubKey *rsa.PublicKey

	Session *mtproto.Session
}

// const msgUseLayer18 uint32 = 0x1c900537
// const msgUseLayer2 uint32 = 0x289dd1f6
// const helpGetConfig = 0xc4f9186b

func Connect(options Options) (*Conn, error) {
	if options.Endpoint == "" {
		return nil, errors.New("configuration error: missing endpoint")
	}
	if options.PublicKey == "" {
		return nil, errors.New("configuration error: missing public key")
	}
	pubKey, err := mtproto.ParsePublicKey(options.PublicKey)
	if err != nil {
		return nil, err
	}

	tr, err := mtproto.DialTCP(options.Endpoint, mtproto.TCPTransportOptions{})
	if err != nil {
		return nil, err
	}

	session := mtproto.NewSession(tr, mtproto.SessionOptions{
		PubKey:  pubKey,
		Verbose: options.Verbose,
	})

	return &Conn{
		Options: options,
		pubKey:  pubKey,

		Session: session,
	}, nil
}

func (c *Conn) Close() {
	c.Session.Close()
}

func (c *Conn) Run() {
	c.Session.Run()
}

func (c *Conn) Err() error {
	return c.Session.Err()
}
