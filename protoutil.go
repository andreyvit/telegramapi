package telegramapi

import (
	"time"
)

// const nanoSecInSec uint64 = 1000000000
const seqNonce uint64 = 0xDEADBEEF

func (c *Conn) generateMsgID() uint64 {
	tm := time.Now()
	sec := uint64(tm.Unix())

	c.seq++
	seq := seqNonce + uint64(c.seq)

	return (sec << 16) | (seq << 2)
}
