package main

import (
	"errors"

	"github.com/andreyvit/telegramapi/tl"
)

type State struct {
}

func (o *State) ReadBareFrom(r *tl.Reader) {
	ver := r.ReadInt()
	if ver < 1 || ver > 1 {
		r.Fail(errors.New("Unsupported version"))
	}
}

func (o *State) WriteBareTo(w *tl.Writer) {
	w.WriteInt(1)
}
