package binints

func EncodeUint24LE(v uint32, b []byte) {
	if (v & 0xFF000000) != 0 {
		panic("does not fit into uint24")
	}
	b[0] = byte(v)
	b[1] = byte(v >> (8 * 1))
	b[2] = byte(v >> (8 * 2))
}

func EncodeUint32LE(v uint32, b []byte) {
	b[0] = byte(v)
	b[1] = byte(v >> (8 * 1))
	b[2] = byte(v >> (8 * 2))
	b[3] = byte(v >> (8 * 3))
}

func EncodeUint64LE(v uint64, b []byte) {
	b[0] = byte(v)
	b[1] = byte(v >> (8 * 1))
	b[2] = byte(v >> (8 * 2))
	b[3] = byte(v >> (8 * 3))
	b[4] = byte(v >> (8 * 4))
	b[5] = byte(v >> (8 * 5))
	b[6] = byte(v >> (8 * 6))
	b[7] = byte(v >> (8 * 7))
}

func EncodeUint128LE(u, v uint64, b []byte) {
	EncodeUint64LE(u, b[0:8])
	EncodeUint64LE(v, b[0:8])
}
