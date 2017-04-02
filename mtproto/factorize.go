package mtproto

import (
	"math/rand"
	// "math"
	"sync"

	fact "github.com/andreyvit/telegramapi/factorize"
)

// var knownPs = []uint64{1229739323}

var factorizeRand *rand.Rand
var factorizeMut sync.Mutex

func factorize(pq uint64) (uint64, uint64) {
	factorizeMut.Lock()
	defer factorizeMut.Unlock()

	p, q := fact.Lopatin(pq, factorizeRand)
	if p*q != pq {
		panic("p*q != pq")
	}
	return p, q

	// for _, p := range knownPs {
	// 	if pq%p == 0 {
	// 		return p, pq / p
	// 	}
	// }

	// var p uint64
	// for p = 3; p < math.MaxUint32; p += 2 {
	// 	if pq%p == 0 {
	// 		return p, pq / p
	// 	}
	// }
	// return 0, 0
}

func init() {
	factorizeRand = rand.New(rand.NewSource(1))
}
