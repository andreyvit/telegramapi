package tlschema

import (
	"bufio"
	"fmt"
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

type lexer struct {
	str     string
	scanner *bufio.Scanner
	offs    int

	token      string
	tokenStart int

	prev      string
	prevStart int

	next      string
	nextStart int

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

func (lex *lexer) Reset(s string) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(lex.lexsplit)
	*lex = lexer{s, scanner, 0, "", 0, "", 0, "", 0, nil}
	lex.Advance()
}

func (lex *lexer) AtEOF() bool {
	return lex.token == ""
}

func (lex *lexer) Advance() {
	if lex.next != "" {
		lex.token, lex.tokenStart = lex.next, lex.nextStart
		lex.next, lex.nextStart = "", 0
		return
	}
	lex.prev, lex.prevStart = lex.token, lex.tokenStart
	if lex.scanner.Scan() {
		lex.token = lex.scanner.Text()
	} else {
		lex.token = ""
		lex.tokenStart = len(lex.str)
	}
}

func (lex *lexer) Unadvance() {
	if lex.prev != "" {
		lex.next, lex.nextStart = lex.token, lex.tokenStart
		lex.token, lex.tokenStart = lex.prev, lex.prevStart
	} else {
		panic("no previous token")
	}
}

func (lex *lexer) Current() string {
	return lex.token
}

func (lex *lexer) Op(token string) bool {
	if lex.token == token {
		lex.Advance()
		return true
	} else {
		return false
	}
}

func (lex *lexer) Ident() string {
	if !lex.AtEOF() && isIdentRune(rune(lex.token[0])) {
		s := lex.token
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
		lex.FailErr(lexerError{msg, lex.str, lex.tokenStart, lex.token})
	}
}

func (lex *lexer) FailPrev(msg string) {
	if lex.err == nil {
		lex.FailErr(lexerError{msg, lex.str, lex.prevStart, lex.prev})
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
		lex.tokenStart = lex.offs + start
		lex.offs += end
		return end, data[start:end], nil
	}

	// Scan until space, marking end of word.
	for i := start; i < ln; i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if !isIdentRune(r) {
			lex.tokenStart = lex.offs + start
			lex.offs += i
			return i, data[start:i], nil
		}
	}

	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && start < ln {
		lex.tokenStart = lex.offs + start
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
