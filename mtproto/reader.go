package mtproto

import (
	"errors"
	"math/big"
)

var ErrMessageTooShort = errors.New("message too short")
var ErrTrailingData = errors.New("unexpected trailing data")

type Reader struct {
	cmd uint32
	rem []byte
	err error
}

func NewReader(data []byte) *Reader {
	r := &Reader{0, data, nil}
	r.cmd = r.ReadUint32()
	return r
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
func (r *Reader) ReadCmd() uint32 {
	return r.ReadUint32()
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

func (r *Reader) ReadN(n int) []byte {
	if !r.need(n) {
		return nil
	}
	v := r.rem[:n]
	r.rem = r.rem[n:]
	return v
}

func (r *Reader) Skip(n int) {
	if r.need(n) {
		r.rem = r.rem[n:]
	}
}

func (r *Reader) ReadStringLen() (int, int) {
	b, ok := r.TryReadByte()
	if !ok {
		return -1, 0
	}
	if b == 254 {
		size, ok := r.TryReadUint24()
		if !ok {
			return -1, 0
		}
		return int(size), PaddingOf(4 + int(size))
	} else if b > 254 {
		r.Fail(errors.New("unexpected string size byte FF"))
		return -1, 0
	} else {
		return int(b), PaddingOf(1 + int(b))
	}
}

func (r *Reader) ReadString() []byte {
	len, pad := r.ReadStringLen()
	if len < 0 {
		return nil
	}

	buf := r.ReadN(len)
	r.Skip(pad)
	return buf
}

func (r *Reader) ReadBigInt() *big.Int {
	b := r.ReadString()
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
