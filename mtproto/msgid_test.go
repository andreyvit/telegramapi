package mtproto

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	g := &MsgIDGen{}
	a, e := g.GenerateAt(time.Date(2017, 1, 1, 0, 0, 0, 12345678, time.UTC)), uint64(0x5868468003291400)
	if a != e {
		t.Errorf("round 1 generated %016x, expected %016x", a, e)
	}

	a, e = g.GenerateAt(time.Date(2017, 1, 1, 0, 0, 0, 12345678, time.UTC)), uint64(0x5868468003291404)
	if a != e {
		t.Errorf("round 2 generated %016x, expected %016x", a, e)
	}

	a = g.GenerateAt(time.Now())
	if (a % 4) != 0 {
		t.Errorf("generated value is not divisible by 4: %016x", a)
	}
}
