package binints

func DecodeUint24LE(b []byte) uint32 {
	v := uint32(b[0])
	v |= uint32(b[1]) << (8 * 1)
	v |= uint32(b[2]) << (8 * 2)
	return v
}

func DecodeUint32LE(b []byte) uint32 {
	v := uint32(b[0])
	v |= uint32(b[1]) << (8 * 1)
	v |= uint32(b[2]) << (8 * 2)
	v |= uint32(b[3]) << (8 * 3)
	return v
}

func DecodeUint64LE(b []byte) uint64 {
	v := uint64(b[0])
	v |= uint64(b[1]) << (8 * 1)
	v |= uint64(b[2]) << (8 * 2)
	v |= uint64(b[3]) << (8 * 3)
	v |= uint64(b[4]) << (8 * 4)
	v |= uint64(b[5]) << (8 * 5)
	v |= uint64(b[6]) << (8 * 6)
	v |= uint64(b[7]) << (8 * 7)
	return v
}
