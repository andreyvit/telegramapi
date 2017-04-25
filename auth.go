package telegramapi

import (
	"bytes"
	"crypto/sha256"
	"log"

	"github.com/andreyvit/telegramapi/mtproto"
)

type LoginState int

const (
	LoggedOut LoginState = iota
	LoggedIn
	WaitingForCode
	WaitingFor2FA
)

func (c *Conn) LoginState() LoginState {
	c.stateMut.Lock()
	defer c.stateMut.Unlock()
	return c.state.LoginState
}

func (c *Conn) StartLogin(phoneNumber string) error {
	r, err := c.Send(&mtproto.TLAuthSendCode{
		Flags:         1,
		PhoneNumber:   phoneNumber,
		CurrentNumber: true,
		APIID:         c.APIID,
		APIHash:       c.APIHash,
	})
	if err != nil {
		return err
	}
	switch r := r.(type) {
	case *mtproto.TLAuthSentCode:
		log.Printf("Got auth.sendCode response: %v", r)
		c.updateState(func(state *State) {
			state.LoginState = WaitingForCode
			state.PhoneNumber = phoneNumber
			state.PhoneCodeHash = r.PhoneCodeHash
		})
		return nil
	default:
		return c.HandleUnknownReply(r)
	}
}

func (c *Conn) CompleteLoginWithCode(code string) error {
	c.stateMut.Lock()
	req := &mtproto.TLAuthSignIn{
		PhoneNumber:   c.state.PhoneNumber,
		PhoneCodeHash: c.state.PhoneCodeHash,
		PhoneCode:     code,
	}
	c.stateMut.Unlock()

	r, err := c.Send(req)
	if err != nil {
		return err
	}
	if r1, ok := r.(*mtproto.TLAuthAuthorization); ok {
		log.Printf("Got auth.signIn response: %v", r1)
		c.completeLogin(r1)
		return nil
	} else if r2, ok := r.(*mtproto.TLRPCError); ok && r2.ErrorMessage == "SESSION_PASSWORD_NEEDED" {
		log.Printf("Got auth.signIn response: %v", r2)
		c.updateState(func(state *State) {
			state.LoginState = WaitingFor2FA
		})
	} else {
		return c.HandleUnknownReply(r)
	}

	return nil
}

func (c *Conn) completeLogin(auth *mtproto.TLAuthAuthorization) {
	c.updateState(func(state *State) {
		state.LoginState = LoggedIn
		if user, ok := auth.User.(*mtproto.TLUser); ok {
			state.UserID = user.ID
			state.PhoneNumber = user.Phone
			state.FirstName = user.FirstName
			state.LastName = user.LastName
			state.Username = user.Username
		}
	})
}

func (c *Conn) CompleteLoginWith2FAPassword(password []byte) error {
	var curSalt []byte
	r, err := c.Send(&mtproto.TLAccountGetPassword{})
	if err != nil {
		return err
	}
	if r, ok := r.(*mtproto.TLAccountPassword); ok {
		log.Printf("Got account.getPassword response: %v", r)
		curSalt = r.CurrentSalt
	} else {
		return c.HandleUnknownReply(r)
	}

	var data bytes.Buffer
	data.Write(curSalt)
	data.Write(password)
	data.Write(curSalt)
	hash := sha256.Sum256(data.Bytes())

	r, err = c.Send(&mtproto.TLAuthCheckPassword{PasswordHash: hash[:]})
	if err != nil {
		return err
	}
	if r, ok := r.(*mtproto.TLAuthAuthorization); ok {
		log.Printf("Got auth.checkPassword response: %v", r)
		c.completeLogin(r)
		return nil
	} else {
		return c.HandleUnknownReply(r)
	}

	return nil
}
