package binints

import (
	"testing"
)

func TestDecodeUint32LE(t *testing.T) {
	i := []byte{0x63, 0x24, 0x16, 0x05}
	e := uint32(0x05162463)
	a := DecodeUint32LE(i)
	if a != e {
		t.Errorf("DecodeUint32LE(%08x) == %08x, expected %08x", i, a, e)
	}
}
