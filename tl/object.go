package tl

import (
	"fmt"
)

type Object interface {
	Cmd() uint32
	ReadBareFrom(r *Reader)
	WriteBareTo(w *Writer)
}

type Schema struct {
	Factory func(uint32) Object
}

func (schema *Schema) ReadBoxedObjectFrom(r *Reader) Object {
	cmd := r.PeekCmd()
	o := schema.Factory(cmd)
	if o != nil {
		r.ReadCmd()
		o.ReadBareFrom(r)
		return o
	} else {
		r.Fail(fmt.Errorf("unknown object %08x", cmd))
		return nil
	}
}

func (schema *Schema) ReadLimitedBoxedObjectFrom(r *Reader, cmds ...uint32) Object {
	if r.ExpectCmd(cmds...) {
		return schema.ReadBoxedObjectFrom(r)
	} else {
		return nil
	}
}
