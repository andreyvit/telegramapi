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

func (o *State) Read(r *tl.Reader) {
	ver := r.ReadInt()
	if ver < 1 || ver > 1 {
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
}

func (o *State) Write(w *tl.Writer) {
	w.WriteInt(1)
	w.WriteInt(o.PreferredDC)

	w.WriteInt(len(o.DCs))
	for _, dc := range o.DCs {
		dc.Write(w)
	}
}

func (o *State) ReadBytes(data []byte) error {
	var r tl.Reader
	r.Reset(data)
	o.Read(&r)
	r.ExpectEOF()
	return r.Err()
}

func (o *State) Bytes() []byte {
	var w tl.Writer
	o.Write(&w)
	return w.Bytes()
}
