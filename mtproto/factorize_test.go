package mtproto

import (
	"testing"
)

func TestFactorize(t *testing.T) {
	var tests = []struct {
		pq   uint64
		p, q uint64
	}{
		{35, 5, 7},
	}

	for _, tt := range tests {
		p, q := factorize(tt.pq)
		if p != tt.p || q != tt.q {
			t.Errorf("factorize(%v) == %v * %v, expected %v * %v", tt.pq, p, q, tt.p, tt.q)
		} else {
			t.Logf("factorize(%v) == %v * %v", tt.pq, p, q)
		}
	}
}
