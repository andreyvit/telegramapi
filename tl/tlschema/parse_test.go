package tlschema

import (
	"testing"
)

func TestF(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization", "ctor auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization"},
		{"", ""},
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
