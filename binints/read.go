package binints

import (
	"errors"
	"io"
)

func ReadFull(r io.Reader, buf []byte) error {
	_, err := io.ReadFull(r, buf)
	return err
}

func ReadByte(r io.Reader) (byte, error) {
	var buf [1]byte
	err := ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func ReadUint24LE(r io.Reader) (uint32, error) {
	var buf [3]byte
	err := ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return DecodeUint24LE(buf[:]), nil
}

func ReadUint32LE(r io.Reader) (uint32, error) {
	var buf [4]byte
	err := ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return DecodeUint32LE(buf[:]), nil
}

func ReadUint32LEAsInt(r io.Reader) (int, error) {
	n, err := ReadUint32LE(r)
	return int(n), err
}

func ReadUint64LE(r io.Reader) (uint64, error) {
	var buf [8]byte
	err := ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return DecodeUint64LE(buf[:]), nil
}

func ReadUint128LE(r io.Reader, buf []byte) error {
	if len(buf) != 16 {
		panic("16-byte buffer required")
	}
	return ReadFull(r, buf)
}

var ErrTrailingData = errors.New("unexpected trailing data")

func ExpectEOF(r io.Reader) error {
	var buf [1]byte
	n, err := r.Read(buf[:])
	if n > 0 {
		return ErrTrailingData
	}
	if err != io.EOF {
		return err
	}
	return nil
}
