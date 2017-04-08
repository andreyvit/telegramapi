package tlschema

import (
	"bytes"
)

type Comb struct {
	*Def
	TypeStr string

	Origin   string
	Priority Priority
}

func (c *Comb) FullName() string {
	return c.CombName.Full()
}

type Type struct {
	Name  ScopedName
	Ctors []*Comb

	Origin   string
	Priority Priority
}

func (t *Type) String() string {
	var buf bytes.Buffer
	buf.WriteString(t.Name.Full())
	buf.WriteString(" => ")
	for i, ctor := range t.Ctors {
		if i > 0 {
			buf.WriteString(" | ")
		}
		buf.WriteString(ctor.CombName.Full())
	}
	return buf.String()
}
