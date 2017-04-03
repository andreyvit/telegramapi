package tlschema

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const Natural = "#"

type CombinatorRef struct {
}

type Def struct {
	Comb   Comb
	IsFunc bool

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
	buf.WriteString(d.Comb.String())
	for _, arg := range d.Args {
		buf.WriteString(" ")
		buf.WriteString(arg.String())
	}
	buf.WriteString(" = ")
	buf.WriteString(d.Type.String())
	return buf.String()
}

type Comb struct {
	Number    uint32
	FullName  string
	ShortName string
}

func (c Comb) String() string {
	if c.Number == 0 {
		return c.FullName
	} else {
		return fmt.Sprintf("%s#%08x", c.FullName, c.Number)
	}
}

type Arg struct {
	Name string
	Type TypeExpr

	CondFieldName string
	CondBit       int
}

func (a Arg) String() string {
	return fmt.Sprintf("%s:%s", a.Name, a.Type.String())
}

type TypeExpr struct {
	Name        string
	GenericArgs []string
}

func (t TypeExpr) String() string {
	return t.Name
}

type ParseState struct {
	InsideFuncs bool
}

func ParseLine(line string, state ParseState) (*Def, ParseState, error) {
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return nil, state, nil
	}

	if strings.HasPrefix(line, "---") {
		if !strings.HasSuffix(line, "---") {
			return nil, state, errors.New("expected trailing triple dash")
		}
		line = strings.TrimSpace(line[3 : len(line)-6])
		switch line {
		case "functions":
			state.InsideFuncs = true
			return nil, state, nil
		case "types":
			state.InsideFuncs = false
			return nil, state, nil
		default:
			return nil, state, errors.New("unexpected section")
		}
	}

	if strings.HasSuffix(line, ";") {
		line = line[:len(line)-1]
	}

	lex := new(lexer)
	lex.Reset(line)

	def, ok := scanDef(lex)
	if !ok {
		return nil, state, lex.Err()
	}
	def.IsFunc = state.InsideFuncs
	return def, state, nil
}

func scanDef(lex *lexer) (*Def, bool) {
	var def Def

	comb, ok := scanComb(lex)
	if !ok {
		return nil, false
	}
	def.Comb = comb

	for !lex.Op("=") {
		arg, ok := scanArg(lex)
		if !ok {
			return nil, false
		}

		def.Args = append(def.Args, arg)
	}

	def.Type, ok = scanTypeExpr(lex)
	if !ok {
		return nil, false
	}

	return &def, true
}

func scanComb(lex *lexer) (Comb, bool) {
	var comb Comb
	var ok bool

	comb.FullName, comb.ShortName, ok = scanScopedName(lex)
	if !ok {
		return comb, false
	}

	if lex.Op("#") {
		numstr := lex.NeedIdent()
		if numstr == "" {
			return comb, false
		}
		num, err := strconv.ParseUint(numstr, 16, 32)
		if err != nil {
			lex.FailPrev("invalid hex number")
			return comb, false
		}
		comb.Number = uint32(num)
	}
	return comb, true
}

func scanArg(lex *lexer) (Arg, bool) {
	var arg Arg

	if s := lex.Ident(); s != "" {
		if lex.Op(":") {
			arg.Name = s
		} else {
			lex.Unadvance()
		}
	}

	typ, ok := scanTypeExpr(lex)
	if !ok {
		return arg, false
	}
	arg.Type = typ

	return arg, true
}

func scanTypeExpr(lex *lexer) (TypeExpr, bool) {
	var typ TypeExpr
	var ok bool

	typ.Name, _, ok = scanScopedName(lex)
	if !ok {
		return typ, false
	}

	// TODO: < ... >

	return typ, true
}

func scanScopedName(lex *lexer) (string, string, bool) {
	n := lex.NeedIdent()
	if n == "" {
		return "", "", false
	}

	if lex.Op(".") {
		n2 := lex.NeedIdent()
		if n2 == "" {
			return "", "", false
		}
		return n + "." + n2, n2, true
	} else {
		return n, n, true
	}
}

func parseCombinatorName(s string) (string, uint32) {
	idx := strings.IndexRune(s, '#')
	if idx < 0 {
		return s, 0
	}

	name := s[:idx]
	cmdstr := s[idx+1:]
	if len(cmdstr) > 8 {
		log.Panicf("invalid schema, cmd hex code > 8 chars in %#v", s)
	}
	cmd, err := strconv.ParseUint(cmdstr, 16, 32)
	if err != nil {
		log.Panicf("invalid schema, cannot parse hex in %#v: %v", s, err)
	}
	return name, uint32(cmd)
}
