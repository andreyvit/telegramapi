package tlschema

import (
	"testing"
)

func TestToGoName(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"foo", "Foo"},
		{"foo_bar", "FooBar"},
		{"req_DH_params", "ReqDHParams"},
		{"inputMediaGame", "InputMediaGame"},
	}

	for _, tt := range tests {
		actual := ToGoName(tt.input)
		if actual != tt.expected {
			t.Errorf("ToGoName(%q) == %q, expected %q", tt.input, actual, tt.expected)
		} else {
			t.Logf("ToGoName(%q) == %q", tt.input, actual)
		}
	}
}

func TestScopedGoName(t *testing.T) {
	var tests = []struct {
		input    string
		expected string
	}{
		{"foo", "Foo"},
		{"req_DH_params", "ReqDHParams"},
		{"foo.bar", "FooBar"},
		{"storage.FileType", "StorageFileType"},
	}

	for _, tt := range tests {
		actual := MakeScopedName(tt.input).GoName()
		if actual != tt.expected {
			t.Errorf("ToGoName(%q) == %q, expected %q", tt.input, actual, tt.expected)
		} else {
			t.Logf("ToGoName(%q) == %q", tt.input, actual)
		}
	}
}
