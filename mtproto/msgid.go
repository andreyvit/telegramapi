package mtproto

import (
	"time"
)

type MsgIDGen struct {
	min uint64
}

func (g *MsgIDGen) GenerateAt(tm time.Time) uint64 {
	nano := uint64(tm.UnixNano())

	// we need to divide by 10^9 (nanoseconds in a second) and multiply by 2^32
	nano = nano / 1000
	nano = nano << 16
	nano = nano / 1000
	nano = nano << 16
	nano = nano / 10

	// clear the last two bits because it must be divisible by 4
	nano = nano & ^uint64(0x3)

	if nano < g.min {
		nano = g.min
	}
	g.min = nano + 4

	return nano
}

func (g *MsgIDGen) Generate() uint64 {
	return g.GenerateAt(time.Now())
}
