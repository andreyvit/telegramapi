package tlschema

import (
	"bytes"
	"fmt"
	"strings"
)

const Natural = "#"

type Def struct {
	Tag      uint32
	CombName ScopedName
	IsFunc   bool
	IsWeird  bool

	GenericArgs []Arg
	Args        []Arg

	Type TypeExpr
}

func (d Def) String() string {
	var buf bytes.Buffer
	if d.IsFunc {
		buf.WriteString("func")
	} else {
		buf.WriteString("ctor")
	}
	buf.WriteString(" ")
	buf.WriteString(d.CombName.String())
	if d.Tag != 0 {
		buf.WriteString(fmt.Sprintf("#%08x", d.Tag))
	}
	for _, arg := range d.GenericArgs {
		buf.WriteString(" {")
		buf.WriteString(arg.String())
		buf.WriteString("}")
	}
	for _, arg := range d.Args {
		buf.WriteString(" ")
		buf.WriteString(arg.String())
	}
	buf.WriteString(" = ")
	buf.WriteString(d.Type.String())
	return buf.String()
}

type Arg struct {
	Name string
	Type TypeExpr

	CondArgName string
	CondBit     int
}

func (a Arg) String() string {
	return fmt.Sprintf("%s:%s", a.Name, a.Type.String())
}

type TypeExpr struct {
	IsBang      bool
	IsPercent   bool
	Name        ScopedName
	GenericArgs []TypeExpr
}

func (t TypeExpr) String() string {
	var buf bytes.Buffer
	if t.IsBang {
		buf.WriteString("!")
	}
	if t.IsPercent {
		buf.WriteString("%")
	}
	buf.WriteString(t.Name.String())
	if len(t.GenericArgs) > 0 {
		buf.WriteString("<")
		for idx, arg := range t.GenericArgs {
			if idx > 0 {
				buf.WriteString(",")
			}
			buf.WriteString(arg.String())
		}
		buf.WriteString(">")
	}
	return buf.String()
}

func (t TypeExpr) IsJustTypeName() bool {
	return !t.IsBang && !t.IsPercent && len(t.GenericArgs) == 0
}

type ScopedName struct {
	full     string
	shortidx int
}

func (n ScopedName) String() string {
	return n.full
}

func (n ScopedName) Full() string {
	return n.full
}

func (n ScopedName) Short() string {
	return n.full[n.shortidx:]
}

func MakeScopedName(s string) ScopedName {
	i := strings.IndexRune(s, '.')
	if i < 0 {
		return ScopedName{s, 0}
	} else {
		return ScopedName{s, i + 1}
	}
}
