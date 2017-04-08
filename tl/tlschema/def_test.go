package tlschema

import (
	"testing"
)

func TestAlterDef(t *testing.T) {
	tests := []struct {
		input     string
		renamings map[string]string
		expected  string
	}{
		{
			"message msg_id:long seqno:int bytes:int body:Object = Message",
			map[string]string{"message": "proto_message", "Message": "ProtoMessage"},
			"ctor proto_message msg_id:long seqno:int bytes:int body:Object = ProtoMessage",
		},
		{
			"msg_container#73f1f8dc messages:vector<%Message> = MessageContainer",
			map[string]string{"message": "proto_message", "Message": "ProtoMessage"},
			"ctor msg_container#73f1f8dc messages:vector<%ProtoMessage> = MessageContainer",
		},
	}

	for _, tt := range tests {
		actual, _, err := ParseLine(tt.input, ParseState{})
		if err != nil {
			t.Error(err)
			continue
		}

		actual.Alter(&Alterations{Renamings: tt.renamings})
		actualStr := actual.String()

		if actualStr != tt.expected {
			t.Errorf("def(%q).Alter() == %q, expected %q", tt.input, actualStr, tt.expected)
		} else {
			t.Logf("✓ def(%q).Alter() == %q", tt.input, actualStr)
		}
	}
}

func TestFixTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"rpc_result req_msg_id:long result:Object = RpcResult",
			"ctor rpc_result#f35c6d01 req_msg_id:long result:Object = RpcResult",
		},
		{
			"message msg_id:long seqno:int bytes:int body:Object = Message",
			"ctor message#5bb8e511 msg_id:long seqno:int bytes:int body:Object = Message",
		},
	}

	for _, tt := range tests {
		actual, _, err := ParseLine(tt.input, ParseState{})
		if err != nil {
			t.Error(err)
			continue
		}

		err = actual.FixTag()
		if err != nil {
			t.Error(err)
			continue
		}
		actualStr := actual.String()

		if actualStr != tt.expected {
			t.Errorf("def(%q).FixTag() == %q, expected %q", tt.input, actualStr, tt.expected)
		} else {
			t.Logf("✓ def(%q).FixTag() == %q", tt.input, actualStr)
		}
	}
}

func TestSimplify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X",
			"ctor invokeWithLayer#da9b0d0d {X:Type} layer:int query:Object = Object",
		},
	}

	for _, tt := range tests {
		actual, _, err := ParseLine(tt.input, ParseState{})
		if err != nil {
			t.Error(err)
			continue
		}

		err = actual.Simplify()
		if err != nil {
			t.Error(err)
			continue
		}
		actualStr := actual.String()

		if actualStr != tt.expected {
			t.Errorf("def(%q).Alter() == %q, expected %q", tt.input, actualStr, tt.expected)
		} else {
			t.Logf("✓ def(%q).Alter() == %q", tt.input, actualStr)
		}
	}
}
