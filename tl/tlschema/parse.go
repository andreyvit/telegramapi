package tlschema

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
)

var ErrWeirdDef = errors.New("this line is best ignored")

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
	def.OriginalStr = line
	return def, state, nil
}

func scanDef(lex *lexer) (*Def, bool) {
	var def Def
	var ok bool

	def.CombName, ok = scanScopedName(lex)
	if !ok {
		return nil, false
	}

	if lex.Op("#") {
		def.Tag, ok = lex.NeedHex32()
		if !ok {
			return nil, false
		}
	}

	if lex.Op("?") {
		def.IsWeird = true
		def.IsQuestionMark = true
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

	def.ResultType, ok = scanTypeExpr(lex)
	if !ok {
		return nil, false
	}

	return &def, true
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
		typ.Name = MakeScopedName(Natural)
	} else {
		typ.Name, ok = scanScopedName(lex)
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

func scanScopedName(lex *lexer) (ScopedName, bool) {
	n := lex.NeedIdent()
	if n == "" {
		return ScopedName{}, false
	}

	if lex.Op(".") {
		n2 := lex.NeedIdent()
		if n2 == "" {
			return ScopedName{}, false
		}
		return ScopedName{n + "." + n2, len(n) + 1}, true
	} else {
		return ScopedName{n, 0}, true
	}
}
