package tl

import (
	"fmt"

	"github.com/kr/pretty"
)

type Object interface {
	Cmd() uint32
	ReadBareFrom(r *Reader)
	WriteBareTo(w *Writer)
}

type Schema struct {
	Factory func(uint32) Object
}

func (schema *Schema) ReadBoxedObject(raw []byte) (Object, error) {
	var r Reader
	r.Reset(raw)
	o := schema.ReadBoxedObjectFrom(&r)
	r.ExpectEOF()

	return o, r.Err()
}

func (schema *Schema) MustReadBoxedObject(raw []byte) Object {
	o, err := schema.ReadBoxedObject(raw)
	if err != nil {
		panic(err)
	}
	return o
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

func Pretty(o Object) string {
	return pretty.Sprint(o)
}
