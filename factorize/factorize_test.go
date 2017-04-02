package factorize

import (
	"math/rand"
	"testing"
)

var factorizeTests = []struct {
	p, q uint64
}{
	{3, 5},
	{13, 17},
	{1246778989, 1441161677},
}

func TestLopatin(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	for _, tt := range factorizeTests {
		p, q := Lopatin(tt.p*tt.q, r)
		if p != tt.p || q != tt.q {
			t.Errorf("FindTwoMultipliers(%v) == %v * %v, expected %v * %v", tt.p*tt.q, p, q, tt.p, tt.q)
		} else {
			t.Logf("FindTwoMultipliers(%v) == %v * %v", tt.p*tt.q, p, q)
		}
	}
}
