package mtproto

import (
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"io"
)

func ReadUint128(r io.Reader, buf []byte) error {
	_ = buf[15]
	return binints.ReadFull(r, buf)
}

func ReadN(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	err := binints.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func SkipN(r io.Reader, n int) error {
	buf := make([]byte, n)
	return binints.ReadFull(r, buf)
}

func ReadStringLen(r io.Reader) (int, int, error) {
	b, err := binints.ReadByte(r)
	if err != nil {
		return -1, 0, err
	}

	if b == 254 {
		size, err := binints.ReadUint24LE(r)
		if err != nil {
			return -1, 0, err
		}
		return int(size), PaddingOf(4 + int(size)), nil
	} else if b == 254 {
		return -1, 0, errors.New("unexpected string size byte FF")
	} else {
		return int(b), PaddingOf(1 + int(b)), nil
	}
}

func PaddingOf(len int) int {
	rem := len % 4
	if rem == 0 {
		return 0
	} else {
		return 4 - rem
	}
}

func ReadString(r io.Reader) ([]byte, error) {
	len, pad, err := ReadStringLen(r)
	if err != nil {
		return nil, err
	}

	buf, err := ReadN(r, len)
	if err != nil {
		return nil, err
	}

	SkipN(r, pad)

	return buf, nil
}

func ReadVectorLong(r io.Reader) ([]uint64, error) {
	cmd, err := binints.ReadUint32LE(r)
	if err != nil {
		return nil, err
	}
	if cmd != IDVectorLong {
		return nil, errors.New("expected %(Vector long)")
	}

	len, err := binints.ReadUint32LE(r)
	if err != nil {
		return nil, err
	}

	res := make([]uint64, len)
	for i := 0; i < int(len); i++ {
		res[i], err = binints.ReadUint64LE(r)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
