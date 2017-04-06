package tlschema

import (
	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"testing"
)

func TestMTProto(t *testing.T) {
	sch := MustParse(knownschemas.MTProtoSchema)
	c := sch.ByName("set_client_DH_params")
	if c == nil {
		t.Fatal("cannot find combinator set_client_DH_params")
	}
	if a, e := c.String(), "func set_client_DH_params#f5045f1f nonce:int128 server_nonce:int128 encrypted_data:bytes = Set_client_DH_params_answer"; a != e {
		t.Fatalf("got %s\nwanted %s", a, e)
	}
}

func TestTelegram(t *testing.T) {
	sch := MustParse(knownschemas.TelegramSchema)

	c := sch.ByName("messageEmpty")
	if c == nil {
		t.Fatal("cannot find combinator messageEmpty")
	}
	if a, e := c.String(), "ctor messageEmpty#83e5de54 id:int = Message"; a != e {
		t.Fatalf("messageEmpty: got %s\nwanted %s", a, e)
	}

	typ := sch.Type("Message")
	if typ == nil {
		t.Fatal("cannot find combinator messageEmpty")
	}
	if a, e := typ.String(), "Message => messageEmpty | message | messageService"; a != e {
		t.Fatalf("Message: got %s\nwanted %s", a, e)
	}
}
