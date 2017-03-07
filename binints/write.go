package binints

import (
	"io"
)

func WriteUint24LE(w io.Writer, v uint32) error {
	var b [3]byte
	EncodeUint24LE(v, b[:])
	_, err := w.Write(b[:])
	return err
}

func WriteUint128LE(w io.Writer, u, v uint64) error {
	var b [16]byte
	EncodeUint128LE(u, v, b[:])
	_, err := w.Write(b[:])
	return err
}
