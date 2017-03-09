package mtproto

import (
	"math"
)

var knownPs = []uint64{1229739323}

func factorize(pq uint64) (uint64, uint64) {
	for _, p := range knownPs {
		if pq%p == 0 {
			return p, pq / p
		}
	}

	var p uint64
	for p = 3; p < math.MaxUint32; p += 2 {
		if pq%p == 0 {
			return p, pq / p
		}
	}
	return 0, 0
}
