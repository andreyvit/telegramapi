package tlschema

import (
	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization", "ctor auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization"},
		{"auth.sendInvites#771c1d97 phone_numbers:Vector<string> message:string = Bool;", "ctor auth.sendInvites#771c1d97 phone_numbers:Vector<string> message:string = Bool"},
		{"invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X;", "ctor invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X"},
	}

	for _, tt := range tests {
		actual, _, err := ParseLine(tt.input, ParseState{false})
		var actualStr string
		if actual != nil {
			actualStr = actual.String()
		} else {
			actualStr = ""
		}

		if err != nil {
			t.Errorf("ParseLine(%q) failed: %v", tt.input, err)
		} else if actualStr != tt.expected {
			t.Errorf("ParseLine(%q) == %q, expected %q", tt.input, actualStr, tt.expected)
		} else {
			t.Logf("ParseLine(%q) == %q", tt.input, actualStr)
		}
	}
}

func TestParseDef(t *testing.T) {
	tests := []string{knownschemas.MTProtoSchema, knownschemas.TelegramSchema}

	for _, str := range tests {
		defs, err := Parse(str)
		if err != nil {
			t.Errorf("Parse failed: %v", err)
		}
		if len(defs) == 0 {
			t.Error("Parse returned zero defs")
		}
	}
}
