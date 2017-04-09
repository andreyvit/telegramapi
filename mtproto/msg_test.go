package mtproto

import (
	"testing"

	"github.com/andreyvit/telegramapi/tl"
)

func TestMsgType(t *testing.T) {
	tests := []struct {
		input    tl.Object
		expected MsgType
	}{
		{&TLReqPQ{}, KeyExMsg},
		{&TLNearestDC{}, ContentMsg},
	}
	for _, tt := range tests {
		actual := MsgFromObj(tt.input).Type
		if actual != tt.expected {
			t.Errorf("MsgFromObj(%v) == %v, expected %v", tt.input, actual, tt.expected)
		} else {
			t.Logf("MsgFromObj(%v) == %v", tt.input, actual)
		}
	}
}
