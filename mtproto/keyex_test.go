package mtproto

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

var randomness = `
3E0549828CCA27E966B301A48FECE2FC
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D

311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC024D
311C85DB234AA2640AFC4A76A735CF5B1F0FD68BD17FA181E1229AD867CC02
`

var req1 = `
00 00 00 00 00 00 00 00 4A 96 70 27 C4 7A E5 51
14 00 00 00 78 97 46 60 3E 05 49 82 8C CA 27 E9
66 B3 01 A4 8F EC E2 FC
`

var res1 = `
00 00 00 00 00 00 00 00 01 C8 83 1E C9 7A E5 51
40 00 00 00 63 24 16 05 3E 05 49 82 8C CA 27 E9
66 B3 01 A4 8F EC E2 FC A5 CF 4D 33 F4 A1 1E A8
77 BA 4A A5 73 90 73 30 08 17 ED 48 94 1A 08 F9
81 00 00 00 15 C4 B5 1C 01 00 00 00 21 6B E8 6C
02 2B B4 C3`

var req2 = `
00 00 00 00 00 00 00 00 27 7A 71 17 C9 7A E5 51
40 01 00 00 BE E4 12 D7 3E 05 49 82 8C CA 27 E9
66 B3 01 A4 8F EC E2 FC A5 CF 4D 33 F4 A1 1E A8
77 BA 4A A5 73 90 73 30 04 49 4C 55 3B 00 00 00
04 53 91 10 73 00 00 00 21 6B E8 6C 02 2B B4 C3
FE 00 01 00 7B B0 10 0A 52 31 61 90 4D 9C 69 FA
04 BC 60 DE CF C5 DD 74 B9 99 95 C7 68 EB 60 D8
71 6E 21 09 BA F2 D4 60 1D AB 6B 09 61 0D C1 10
67 BB 89 02 1E 09 47 1F CF A5 2D BD 0F 23 20 4A
D8 CA 8B 01 2B F4 0A 11 2F 44 69 5A B6 C2 66 95
53 86 11 4E F5 21 1E 63 72 22 7A DB D3 49 95 D3
E0 E5 FF 02 EC 63 A4 3F 99 26 87 89 62 F7 C5 70
E6 A6 E7 8B F8 36 6A F9 17 A5 27 26 75 C4 60 64
BE 62 E3 E2 02 EF A8 B1 AD FB 1C 32 A8 98 C2 98
7B E2 7B 5F 31 D5 7C 9B B9 63 AB CB 73 4B 16 F6
52 CE DB 42 93 CB B7 C8 78 A3 A3 FF AC 9D BE A9
DF 7C 67 BC 9E 95 08 E1 11 C7 8F C4 6E 05 7F 5C
65 AD E3 81 D9 1F EE 43 0A 6B 57 6A 99 BD F8 55
1F DB 1B E2 B5 70 69 B1 A4 57 30 61 8F 27 42 7E
8A 04 72 0B 49 71 EF 4A 92 15 98 3D 68 F2 83 0C
3E AA 6E 40 38 55 62 F9 70 D3 8A 05 C9 F1 24 6D
C3 34 38 E6
`

func TestKeyExchange(t *testing.T) {
	var keyex KeyEx
	var framer Framer
	var err error

	keyex.RandomReader = bytes.NewReader(fromHex(randomness))
	keyex.PubKey, err = ParsePublicKey(publicKey)
	if err != nil {
		t.Fatal(err)
	}

	framer.MsgIDOverride = 0x51e57ac42770964a
	bytes, err := framer.Format(keyex.Start())
	if err != nil {
		t.Fatal(err)
	}

	a, e := hex.EncodeToString(bytes), hex.EncodeToString(fromHex(req1))
	if a != e {
		t.Errorf("req_pq is %q, expected %q", a, e)
	}

	payload, err := framer.Parse(fromHex(res1))
	if err != nil {
		t.Fatal(err)
	}
	msg, err := keyex.Handle(payload)
	if err != nil {
		t.Fatal(err)
	}

	if msg == nil {
		t.Fatal("no reply to res_pq")
	}
	framer.MsgIDOverride = 0x51e57ac917717a27
	bytes, err = framer.Format(*msg)
	if err != nil {
		t.Fatal(err)
	}
	ebytes := fromHex(req2)
	a, e = hex.EncodeToString(bytes), hex.EncodeToString(ebytes)
	if len(bytes) != len(ebytes) {
		t.Errorf("req_DH_params is %v, expected %v (len mismatch: got %v, wanted %v)", a, e, len(bytes), len(ebytes))
	}
	// if a != e {
	// 	t.Errorf("req_DH_params is %v, expected %v", a, e)
	// }
}

func fromHex(s string) []byte {
	data, err := hex.DecodeString(strings.Map(dropSpace, s))
	if err != nil {
		panic(err)
	}
	return data
}

func dropSpace(r rune) rune {
	if r == ' ' || r == '\t' || r == '\n' {
		return -1
	} else {
		return r
	}
}
