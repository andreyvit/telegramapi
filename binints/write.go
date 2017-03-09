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

func WriteUint32LE(w io.Writer, v uint32) error {
	var b [4]byte
	EncodeUint32LE(v, b[:])
	_, err := w.Write(b[:])
	return err
}

func WriteUint64LE(w io.Writer, v uint64) error {
	var b [8]byte
	EncodeUint64LE(v, b[:])
	_, err := w.Write(b[:])
	return err
}

func WriteUint128LE(w io.Writer, b []byte) error {
	if len(b) != 16 {
		panic("16-byte buffer required")
	}
	_, err := w.Write(b)
	return err
}
