// PQ factorization derived from org.telegram.mtproto.secure.pq.PQLopatin of github.com/ex3ndr/telegram-mt
package factorize

import (
	"math/rand"
)

func Lopatin(what uint64, r *rand.Rand) (uint64, uint64) {
	var it int = 0
	for i := uint(0); i < 3; i++ {
		q := uint64((r.Intn(128) & 15) + 17)
		x := uint64(r.Intn(1000000000) + 1)
		y := x
		lim := 1 << (i + 18)
		for j := 1; j < lim; j++ {
			it++
			a, b, c := x, x, q
			for b != 0 {
				if (b & 1) != 0 {
					c += a
					if c >= what {
						c -= what
					}
				}
				a += a
				if a >= what {
					a -= what
				}
				b = b >> 1
			}

			x = c
			var z uint64
			if x < y {
				z = y - x
			} else {
				z = x - y
			}
			g := gcd(z, what)
			if g != 1 {
				p := what / g
				if p <= g {
					return p, g
				} else {
					return g, p
				}
			}
			if (j & (j - 1)) == 0 {
				y = x
			}
		}
	}

	return 1, what
}
