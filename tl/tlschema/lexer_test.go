package tlschema

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization", "auth . signUp # 1b067634 phone_number : string phone_code_hash : string phone_code : string first_name : string last_name : string = auth . Authorization"},
		{"invokeAfterMsg#cb9f372d {X:Type} msg_id:long query:!X = X", "invokeAfterMsg # cb9f372d { X : Type } msg_id : long query : ! X = X"},
	}

	for _, tt := range tests {
		actual := strings.Join(tokenize(tt.input), " ")
		if actual != tt.expected {
			t.Errorf("tokenize(%q) == %q, expected %q", tt.input, actual, tt.expected)
		} else {
			t.Logf("tokenize(%q) == %q", tt.input, actual)
		}
	}
}

func TestLexerMethods(t *testing.T) {
	var lex lexer
	lex.Reset("auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization")

	if a, e := lex.Ident(), "auth"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}
	if !lex.Op(".") {
		t.Fatal(`lex.Op(".") didn't match`)
	}
	if a, e := lex.Ident(), "signUp"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}
	if !lex.Op("#") {
		t.Fatal(`lex.Op("#") didn't match`)
	}
	if a, e := lex.Ident(), "1b067634"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}
	if a, e := lex.Ident(), "phone_number"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}
	if !lex.Op(":") {
		t.Fatal(`lex.Op(":") didn't match`)
	}
	if a, e := lex.Ident(), "string"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}

	lex.Unadvance(1)
	if a, e := lex.Ident(), "string"; a != e {
		t.Fatalf("lex.Ident() == %q, expected %q", a, e)
	}
}
