package telegramapi

import (
	"errors"
	"fmt"

	"github.com/andreyvit/telegramapi/mtproto"
	"github.com/andreyvit/telegramapi/tl"
)

// Key        []byte
// KeyID      uint64
// ServerSalt [8]byte
// TimeOffset int
// SessionID  [8]byte

func readAuth(o *mtproto.AuthResult, fs *mtproto.FramerState, r *tl.Reader, ver int) {
	o.KeyID = r.ReadUint64()
	if o.KeyID != 0 {
		o.Key = r.ReadBlob()
		r.ReadFull(o.ServerSalt[:])
		r.ReadFull(o.SessionID[:])
		o.TimeOffset = r.ReadInt()
	}

	fs.SeqNo = r.ReadUint32()
}

func writeAuth(o *mtproto.AuthResult, fs *mtproto.FramerState, w *tl.Writer) {
	w.WriteUint64(o.KeyID)
	if o.KeyID != 0 {
		w.WriteBlob(o.Key)
		w.Write(o.ServerSalt[:])
		w.Write(o.SessionID[:])
		w.WriteInt(o.TimeOffset)
	}

	w.WriteUint32(fs.SeqNo)
}

type Addr struct {
	IP   string
	Port int
}

func (o *Addr) Endpoint() string {
	return fmt.Sprintf("%s:%d", o.IP, o.Port)
}

func (o *Addr) Read(r *tl.Reader, ver int) {
	o.IP = r.ReadString()
	o.Port = r.ReadInt()
}

func (o *Addr) Write(w *tl.Writer) {
	w.WriteString(o.IP)
	w.WriteInt(o.Port)
}

type DCState struct {
	ID int

	PrimaryAddr Addr

	Auth        mtproto.AuthResult
	FramerState mtproto.FramerState
}

func (o *DCState) Read(r *tl.Reader, ver int) {
	o.ID = r.ReadInt()
	o.PrimaryAddr.Read(r, 1)
	readAuth(&o.Auth, &o.FramerState, r, 1)
}

func (o *DCState) Write(w *tl.Writer) {
	w.WriteInt(o.ID)
	o.PrimaryAddr.Write(w)
	writeAuth(&o.Auth, &o.FramerState, w)
}

type State struct {
	PreferredDC int

	DCs map[int]*DCState

	LoginState    LoginState
	PhoneNumber   string
	PhoneCodeHash string

	UserID    int
	FirstName string
	LastName  string
	Username  string
}

func (o *State) Cmd() uint32 {
	return 0
}

func (o *State) initialize() {
	if o.DCs == nil {
		o.DCs = make(map[int]*DCState)
	}
}

func (o *State) findPreferredDC() *DCState {
	id := o.PreferredDC
	if id == 0 {
		return nil
	}

	for _, dc := range o.DCs {
		if dc.ID == id {
			return dc
		}
	}

	return nil
}

func (o *State) WriteBareTo(w *tl.Writer) {
	w.WriteInt(4)
	w.WriteInt(o.PreferredDC)

	w.WriteInt(len(o.DCs))
	for _, dc := range o.DCs {
		dc.Write(w)
	}

	w.WriteUint32(uint32(o.LoginState))
	w.WriteString(o.PhoneNumber)
	w.WriteString(o.PhoneCodeHash)
	w.WriteInt(o.UserID)
	w.WriteString(o.FirstName)
	w.WriteString(o.LastName)
	w.WriteString(o.Username)
}

func (o *State) ReadBareFrom(r *tl.Reader) {
	ver := r.ReadInt()
	if ver < 1 || ver > 4 {
		r.Fail(errors.New("Unsupported version"))
	}

	o.PreferredDC = r.ReadInt()

	o.DCs = make(map[int]*DCState)
	n := r.ReadInt()
	for i := 0; i < n; i++ {
		dc := new(DCState)
		dc.Read(r, 1)
		o.DCs[dc.ID] = dc
	}

	if ver >= 2 {
		o.LoginState = LoginState(r.ReadUint32())
		o.PhoneNumber = r.ReadString()
		o.PhoneCodeHash = r.ReadString()
	}
	if ver >= 4 {
		o.UserID = r.ReadInt()
	}
	if ver >= 3 {
		o.FirstName = r.ReadString()
		o.LastName = r.ReadString()
		o.Username = r.ReadString()
	}
}
