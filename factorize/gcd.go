package factorize

func gcd(a, b uint64) uint64 {
	for a != 0 && b != 0 {
		for (b & 1) == 0 {
			b = b >> 1
		}
		for (a & 1) == 0 {
			a = a >> 1
		}
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}
	if b == 0 {
		return a
	} else {
		return b
	}
}
