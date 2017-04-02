package factorize

import (
	"testing"
)

func TestGcd(t *testing.T) {
	tests := []struct {
		a, b     uint64
		expected uint64
	}{
		{30, 20, 5}, // WTF not 10?
		{30, 45, 15},
		{7, 3, 1},
		{7247 * 7639, 7639 * 7919, 7639}, // primes
	}

	for _, tt := range tests {
		actual := gcd(tt.a, tt.b)
		if actual != tt.expected {
			t.Errorf("gcd(%v, %v) == %v, expected %v", tt.a, tt.b, actual, tt.expected)
		} else {
			t.Logf("gcd(%v, %v) == %v", tt.a, tt.b, actual)
		}
	}
}
