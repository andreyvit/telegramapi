package tlschema

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexerError struct {
	msg   string
	str   string
	offs  int
	token string
}

func (e lexerError) Error() string {
	p := e.str[:e.offs]
	s := e.str[e.offs:]

	msg := strings.Replace(e.msg, "<TOKEN>", e.token, -1)
	return fmt.Sprintf("%s: %s →ERR→ %s", msg, p, s)
}

type internalToken struct {
	str   string
	start int
}

const ageLim = 4
const unadvanceLim = ageLim - 1

type lexer struct {
	str     string
	scanner *bufio.Scanner
	offs    int

	tokens [ageLim + 1]internalToken
	age    int

	err error
}

func tokenize(s string) []string {
	var lex lexer
	lex.Reset(s)

	var tokens []string
	for !lex.AtEOF() {
		tokens = append(tokens, lex.Current())
		lex.Advance()
	}

	return tokens
}

func (lex *lexer) DebugString() string {
	var b bytes.Buffer
	b.WriteString("Lexer(")
	for age, tok := range lex.tokens[:] {
		if age != 0 && tok.str == "" {
			continue
		}
		if age > 0 {
			b.WriteString(" ")
		}
		if age == lex.age {
			b.WriteString("→")
		}
		b.WriteString(strconv.Itoa(age))
		b.WriteString(":")
		b.WriteString(strconv.Quote(tok.str))
	}
	b.WriteString(")")
	return b.String()
}

func (lex *lexer) Reset(s string) {
	var newlex lexer
	newlex.str = s
	newlex.scanner = bufio.NewScanner(strings.NewReader(s))
	newlex.scanner.Split(lex.lexsplit)
	*lex = newlex
	lex.Advance()
}

func (lex *lexer) AtEOF() bool {
	return lex.Current() == ""
}

func (lex *lexer) Current() string {
	return lex.tokens[lex.age].str
}

func (lex *lexer) Advance() {
	if lex.age > 0 {
		lex.age--
		return
	}

	for i := ageLim; i >= 1; i-- {
		lex.tokens[i] = lex.tokens[i-1]
	}

	if lex.scanner.Scan() {
		lex.tokens[0].str = lex.scanner.Text()
	} else {
		lex.tokens[0] = internalToken{"", len(lex.str)}
	}
}

func (lex *lexer) Unadvance(n int) {
	if lex.age+n > unadvanceLim {
		panic("cannot unadvance beyond ageLim")
	}
	lex.age += n
	if lex.tokens[lex.age].str == "" {
		panic("no tokens this far back yet")
	}
}

func (lex *lexer) Op(token string) bool {
	if lex.Current() == token {
		lex.Advance()
		return true
	} else {
		return false
	}
}

func (lex *lexer) Ident() string {
	s := lex.Current()
	if s != "" && isIdentRune(rune(s[0])) {
		lex.Advance()
		return s
	} else {
		return ""
	}
}

func (lex *lexer) NeedIdent() string {
	s := lex.Ident()
	if s == "" {
		lex.FailNext("expected identifier, got <TOKEN>")
	}
	return s
}

func (lex *lexer) Int() (int, bool) {
	s := lex.Ident()
	if s == "" {
		return -1, false
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		lex.Unadvance(1)
		return -1, false
	}
	return int(n), true
}

func (lex *lexer) NeedOp(op string) bool {
	if !lex.Op(op) {
		lex.FailNext("expected " + op + ", got <TOKEN>")
		return false
	}
	return true
}

func (lex *lexer) Err() error {
	return lex.err
}

func (lex *lexer) FailNext(msg string) {
	if lex.err == nil {
		lex.FailErr(lexerError{msg, lex.str, lex.tokens[lex.age].start, lex.tokens[lex.age].str})
	}
}

func (lex *lexer) FailPrev(msg string) {
	if lex.err == nil {
		lex.FailErr(lexerError{msg, lex.str, lex.tokens[lex.age+1].start, lex.tokens[lex.age+1].str})
	}
}

func (lex *lexer) FailErr(err error) {
	if lex.err == nil {
		lex.err = err
	}
}

func (lex *lexer) lexsplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var r rune
	var width int
	ln := len(data)

	start := 0
	for width = 0; start < ln; start += width {
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}

	if start >= ln {
		lex.offs += start
		return start, nil, nil
	}

	// If not an ident char, return it
	r, width = utf8.DecodeRune(data[start:])
	if !isIdentRune(r) {
		end := start + width
		lex.tokens[0].start = lex.offs + start
		lex.offs += end
		return end, data[start:end], nil
	}

	// Scan until space, marking end of word.
	for i := start; i < ln; i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if !isIdentRune(r) {
			lex.tokens[0].start = lex.offs + start
			lex.offs += i
			return i, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && start < ln {
		lex.tokens[0].start = lex.offs + start
		lex.offs += ln
		return ln, data[start:], nil
	}

	// Request more data.
	lex.offs += start
	return start, nil, nil
}

func isIdentRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}
