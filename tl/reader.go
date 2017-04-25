package tl

import (
	"errors"
	"math"
	"math/big"
	"time"
)

var ErrMessageTooShort = errors.New("message too short")
var ErrTrailingData = errors.New("unexpected trailing data")
var ErrUnexpectedCommand = errors.New("unexpected command")

type Reader struct {
	rem []byte
	cmd uint32
	err error
}

func CmdOfPayload(b []byte) uint32 {
	if len(b) < 4 {
		return 0
	} else {
		v := uint32(b[0])
		v |= uint32(b[1]) << (8 * 1)
		v |= uint32(b[2]) << (8 * 2)
		v |= uint32(b[3]) << (8 * 3)
		return v
	}
}

func NewReader(data []byte) *Reader {
	r := &Reader{data, 0, nil}
	r.StartInnerCmd()
	return r
}

func (r *Reader) Reset(data []byte) {
	*r = Reader{data, 0, nil}
}

func (r *Reader) Cmd() uint32 {
	return r.cmd
}

func (r *Reader) Fail(err error) {
	if r.err == nil {
		r.err = err
	}
}

func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) need(cb int) bool {
	if r.err != nil {
		return false
	}
	if len(r.rem) < cb {
		r.Fail(ErrMessageTooShort)
		return false
	} else {
		return true
	}
}

func (r *Reader) ReadByte() byte {
	if !r.need(1) {
		return 0
	}
	v := r.rem[0]
	r.rem = r.rem[1:]
	return v
}

func (r *Reader) TryReadByte() (byte, bool) {
	if !r.need(1) {
		return 0, false
	}
	v := r.rem[0]
	r.rem = r.rem[1:]
	return v, true
}

func (r *Reader) TryReadUint24() (uint32, bool) {
	if !r.need(3) {
		return 0, false
	}

	b := r.rem
	v := uint32(b[0])
	v |= uint32(b[1]) << (8 * 1)
	v |= uint32(b[2]) << (8 * 2)
	r.rem = b[3:]
	return v, true
}

func (r *Reader) PeekUint32() uint32 {
	if len(r.rem) < 4 {
		return 0
	}

	b := r.rem
	v := uint32(b[0])
	v |= uint32(b[1]) << (8 * 1)
	v |= uint32(b[2]) << (8 * 2)
	v |= uint32(b[3]) << (8 * 3)
	return v
}

func (r *Reader) TryReadUint32() (uint32, bool) {
	if !r.need(4) {
		return 0, false
	}

	b := r.rem
	v := uint32(b[0])
	v |= uint32(b[1]) << (8 * 1)
	v |= uint32(b[2]) << (8 * 2)
	v |= uint32(b[3]) << (8 * 3)
	r.rem = b[4:]
	return v, true
}
func (r *Reader) ReadUint32() uint32 {
	v, _ := r.TryReadUint32()
	return v
}
func (r *Reader) ReadInt() int {
	return int(r.ReadUint32())
}
func (r *Reader) ReadBool() bool {
	return r.ReadUint32() != 0
}
func (r *Reader) ReadCmd() uint32 {
	return r.ReadUint32()
}
func (r *Reader) PeekCmd() uint32 {
	return r.PeekUint32()
}
func (r *Reader) StartInnerCmd() uint32 {
	r.cmd = r.PeekUint32()
	return r.cmd
}
func (r *Reader) ReadTimeSec32() time.Time {
	u, ok := r.TryReadUint32()
	if !ok {
		return time.Time{}
	}
	return time.Unix(int64(u), 0)
}
func (r *Reader) ReadFloat64() float64 {
	u, ok := r.TryReadUint64()
	if !ok {
		return 0
	}
	return math.Float64frombits(u)
}

func (r *Reader) TryReadUint64() (uint64, bool) {
	if !r.need(8) {
		return 0, false
	}

	b := r.rem
	v := uint64(b[0])
	v |= uint64(b[1]) << (8 * 1)
	v |= uint64(b[2]) << (8 * 2)
	v |= uint64(b[3]) << (8 * 3)
	v |= uint64(b[4]) << (8 * 4)
	v |= uint64(b[5]) << (8 * 5)
	v |= uint64(b[6]) << (8 * 6)
	v |= uint64(b[7]) << (8 * 7)
	r.rem = b[8:]
	return v, true
}
func (r *Reader) ReadUint64() uint64 {
	v, _ := r.TryReadUint64()
	return v
}

func (r *Reader) ReadUint128(buf []byte) bool {
	if !r.need(16) {
		return false
	}
	_ = buf[15]
	copy(buf[:16], r.rem)
	r.rem = r.rem[16:]
	return true
}

func (r *Reader) ExpectEOF() {
	if len(r.rem) > 0 {
		r.Fail(ErrTrailingData)
	}
}

func (r *Reader) ExpectCmd(cmds ...uint32) bool {
	cmd := r.PeekUint32()
	for _, valid := range cmds {
		if valid == cmd {
			return true
		}
	}
	r.Fail(ErrUnexpectedCommand)
	return false
}

func (r *Reader) ReadN(n int) []byte {
	if !r.need(n) {
		return nil
	}
	v := r.rem[:n]
	r.rem = r.rem[n:]
	return v
}

func (r *Reader) ReadFull(b []byte) {
	n := len(b)
	if !r.need(n) {
		return
	}
	copy(b, r.rem[:n])
	r.rem = r.rem[n:]
}

func (r *Reader) Skip(n int) {
	if r.need(n) {
		r.rem = r.rem[n:]
	}
}

func (r *Reader) ReadToEnd() []byte {
	return r.rem
}

func (r *Reader) ReadBlobLen() (int, int) {
	b, ok := r.TryReadByte()
	if !ok {
		return -1, 0
	}
	if b == 254 {
		size, ok := r.TryReadUint24()
		if !ok {
			return -1, 0
		}
		return int(size), paddingOf(4 + int(size))
	} else if b > 254 {
		r.Fail(errors.New("unexpected string size byte FF"))
		return -1, 0
	} else {
		return int(b), paddingOf(1 + int(b))
	}
}

func (r *Reader) ReadBlob() []byte {
	len, pad := r.ReadBlobLen()
	if len < 0 {
		return nil
	}

	buf := r.ReadN(len)
	r.Skip(pad)
	return buf
}

func (r *Reader) ReadString() string {
	blob := r.ReadBlob()
	if blob != nil {
		return string(blob)
	} else {
		return ""
	}
}

func (r *Reader) ReadBigInt() *big.Int {
	b := r.ReadBlob()
	if b == nil {
		return nil
	}

	n := new(big.Int)
	n.SetBytes(b)
	return n
}

func (r *Reader) ReadVectorLong() []uint64 {
	cmd, ok := r.TryReadUint32()
	if !ok {
		return nil
	}
	if cmd != IDVectorLong {
		r.Fail(errors.New("expected %(Vector long)"))
		return nil
	}

	len, ok := r.TryReadUint32()
	if !ok {
		return nil
	}

	res := make([]uint64, len)
	for i := 0; i < int(len); i++ {
		res[i], ok = r.TryReadUint64()
		if !ok {
			return nil
		}
	}
	return res
}
