package tlschema

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

var ErrWeirdDef = errors.New("this line is best ignored")

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

	CondArgName string
	CondBit     int
}

func (a Arg) String() string {
	return fmt.Sprintf("%s:%s", a.Name, a.Type.String())
}

type TypeExpr struct {
	IsBang      bool
	IsPercent   bool
	Name        string
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
	buf.WriteString(t.Name)
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

type ParseState struct {
	InsideFuncs bool
}

func Parse(text string) ([]*Def, error) {
	var state ParseState
	var defs []*Def

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()

		def, newState, err := ParseLine(line, state)
		state = newState
		if err != nil && err != ErrWeirdDef {
			return nil, err
		}
		if def != nil {
			defs = append(defs, def)
		}
	}

	return defs, nil
}

func ParseLine(line string, state ParseState) (*Def, ParseState, error) {
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return nil, state, nil
	}

	if strings.HasPrefix(line, "---") {
		if !strings.HasSuffix(line, "---") {
			return nil, state, fmt.Errorf("expected trailing triple dash in %q", line)
		}
		line = strings.TrimSpace(line[3 : len(line)-3])
		switch line {
		case "functions":
			state.InsideFuncs = true
			return nil, state, nil
		case "types":
			state.InsideFuncs = false
			return nil, state, nil
		default:
			return nil, state, fmt.Errorf("unexpected section %q", line)
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

	if lex.Op("?") {
		lex.FailErr(ErrWeirdDef)
		return nil, false
	}

	for lex.Op("{") {
		arg, ok := scanArg(lex)
		if !ok {
			return nil, false
		}

		def.GenericArgs = append(def.GenericArgs, arg)

		lex.NeedOp("}")
	}

	for !lex.Op("=") {
		if lex.Op("[") {
			lex.FailErr(ErrWeirdDef)
			return nil, false
		}

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

	if _, ok := lex.Int(); ok {
		if lex.Op("*") {
			lex.FailErr(ErrWeirdDef)
			return arg, false
		} else {
			lex.Unadvance(1)
		}
	}

	if s := lex.Ident(); s != "" {
		if lex.Op(":") {
			arg.Name = s
		} else {
			lex.Unadvance(1)
		}
	}

	verbose := false //strings.Contains(lex.str, "profile_photo:Photo")

	if verbose {
		log.Printf("lexer (before): %s", lex.DebugString())
	}
	if argname := lex.Ident(); argname != "" {
		if verbose {
			log.Printf("lexer (argname): %s", lex.DebugString())
		}
		if lex.Op(".") {
			if verbose {
				log.Printf("lexer (op): %s", lex.DebugString())
			}
			if num, ok := lex.Int(); ok {
				if verbose {
					log.Printf("lexer (num): %s", lex.DebugString())
				}
				if lex.Op("?") {
					arg.CondArgName = argname
					arg.CondBit = num
				} else {
					lex.Unadvance(3)
					if verbose {
						log.Printf("lexer (num undo): %s", lex.DebugString())
					}
				}
			} else {
				if verbose {
					log.Printf("lexer (int failed): %s", lex.DebugString())
				}
				lex.Unadvance(2)
				if verbose {
					log.Printf("lexer (op undo): %s", lex.DebugString())
				}
			}
		} else {
			lex.Unadvance(1)
			if verbose {
				log.Printf("lexer (argname undo): %s", lex.DebugString())
			}
		}
	}

	if verbose {
		log.Printf("lexer (after): %s", lex.DebugString())
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

	if lex.Op("!") {
		typ.IsBang = true
	}
	if lex.Op("%") {
		typ.IsPercent = true
	}

	if lex.Op("#") {
		typ.Name = Natural
	} else {
		typ.Name, _, ok = scanScopedName(lex)
		if !ok {
			return typ, false
		}
	}

	if lex.Op("<") {
		for {
			subtyp, ok := scanTypeExpr(lex)
			if !ok {
				return typ, false
			}
			typ.GenericArgs = append(typ.GenericArgs, subtyp)
			if !lex.Op(",") {
				break
			}
		}
		lex.NeedOp(">")
	}

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
