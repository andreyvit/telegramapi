package tlschema

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"strconv"
)

const Natural = "#"

type Def struct {
	Tag      uint32
	CombName ScopedName

	IsFunc     bool
	IsInternal bool

	IsQuestionMark bool
	IsWeird        bool

	OriginalStr string

	GenericArgs []Arg
	Args        []Arg

	ResultType TypeExpr
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
	if d.IsQuestionMark {
		buf.WriteString(" ?")
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
	buf.WriteString(d.ResultType.String())
	return buf.String()
}

func (d Def) CanonicalString() string {
	var buf bytes.Buffer
	buf.WriteString(d.CombName.String())
	if d.IsQuestionMark {
		buf.WriteString(" ?")
	}
	for _, arg := range d.GenericArgs {
		buf.WriteString(" {")
		buf.WriteString(arg.CanonicalString())
		buf.WriteString("}")
	}
	for _, arg := range d.Args {
		buf.WriteString(" ")
		buf.WriteString(arg.CanonicalString())
	}
	buf.WriteString(" = ")
	buf.WriteString(d.ResultType.CanonicalString())
	return buf.String()
}

func (d *Def) Alter(alter *Alterations) {
	d.CombName.Alter(alter)
	for _, arg := range d.GenericArgs {
		arg.Alter(alter)
	}
	for _, arg := range d.Args {
		arg.Alter(alter)
	}
	d.ResultType.Alter(alter)
}

func (d *Def) FixTag() error {
	if d.Tag != 0 {
		return nil
	}

	str := d.CanonicalString()
	if str != d.OriginalStr {
		return fmt.Errorf("refusing to compute tag because canonical string %q differs from original string %q", str, d.OriginalStr)
	}

	d.Tag = crc32.ChecksumIEEE([]byte(str))
	return nil
}

func (d *Def) Simplify() error {
	if d.GenericArgs == nil {
		return nil
	}

	var typeArgNames []string
	for _, arg := range d.GenericArgs {
		if arg.Type.String() == "Type" {
			typeArgNames = append(typeArgNames, arg.Name)
		}
	}

	for i := range d.Args {
		d.Args[i].Type.Simplify(typeArgNames)
	}

	d.ResultType.Simplify(typeArgNames)

	return nil
}

type Arg struct {
	Name string
	Type TypeExpr

	CondArgName string
	CondBit     int
}

func (a Arg) String() string {
	var buf bytes.Buffer
	if a.CondArgName != "" {
		buf.WriteString(a.CondArgName)
		buf.WriteString(".")
		buf.WriteString(strconv.Itoa(a.CondBit))
		buf.WriteString("?")
	}
	if a.Name != "" {
		buf.WriteString(a.Name)
		buf.WriteString(":")
	}
	buf.WriteString(a.Type.String())
	return buf.String()
}

func (a Arg) CanonicalString() string {
	var buf bytes.Buffer
	if a.CondArgName != "" {
		buf.WriteString(a.CondArgName)
		buf.WriteString(".")
		buf.WriteString(strconv.Itoa(a.CondBit))
		buf.WriteString("?")
	}
	if a.Name != "" {
		buf.WriteString(a.Name)
		buf.WriteString(":")
	}
	buf.WriteString(a.Type.CanonicalString())
	return buf.String()
}

func (a *Arg) Alter(alter *Alterations) {
	a.Type.Alter(alter)
}

type TypeExpr struct {
	IsBang      bool
	IsPercent   bool
	Name        ScopedName
	GenericArgs []TypeExpr
}

func (t TypeExpr) IsBare() bool {
	return t.IsPercent || t.Name.IsBare()
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

func (t TypeExpr) CanonicalString() string {
	var buf bytes.Buffer
	if t.IsBang {
		buf.WriteString("!")
	}
	if t.IsPercent {
		buf.WriteString("%")
	}
	buf.WriteString(t.Name.String())
	for _, arg := range t.GenericArgs {
		buf.WriteString(" ")
		buf.WriteString(arg.CanonicalString())
	}
	return buf.String()
}

func (t TypeExpr) IsJustTypeName() bool {
	return !t.IsBang && !t.IsPercent && len(t.GenericArgs) == 0
}

func (t *TypeExpr) Alter(alter *Alterations) {
	t.Name.Alter(alter)
	for i := range t.GenericArgs {
		t.GenericArgs[i].Alter(alter)
	}
}

func (t *TypeExpr) Simplify(genericTypeNames []string) {
	if containsStr(genericTypeNames, t.Name.Full()) {
		t.Name = MakeScopedNameComponents("", "Object")
		t.IsBang = false
	}
}
