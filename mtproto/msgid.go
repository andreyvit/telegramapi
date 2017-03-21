package mtproto

import (
	"log"
	"time"
)

type MsgIDGen struct {
	min uint64
}

func (g *MsgIDGen) GenerateAt(tm time.Time) uint64 {
	fnano := float64(tm.UnixNano())
	fnano /= 1000000000
	fnano *= 4294967296

	nano := uint64(fnano)

	// clear the last two bits because it must be divisible by 4
	nano = nano & ^uint64(0x3)

	if nano < g.min {
		nano = g.min
	}
	g.min = nano + 4

	u := tm.Unix()
	a := int64(nano >> 32)
	d := u - a
	if d > 1 {
		log.Printf("nano = %d (0x%x), unix = %d (0x%x), expected = %d (0x%x), diff = %d", nano, nano, a, a, u, u, d)
		panic("invalid msg id generated")
	}

	return nano
}

func (g *MsgIDGen) Generate() uint64 {
	return g.GenerateAt(time.Now())
}
