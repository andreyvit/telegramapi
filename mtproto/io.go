package mtproto

import (
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"io"
	"math/big"
)

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
	} else if b > 254 {
		return -1, 0, errors.New("unexpected string size byte FF")
	} else {
		return int(b), PaddingOf(1 + int(b)), nil
	}
}

func WriteStringLen(w io.Writer, len int) (int, error) {
	var bb [4]byte
	var b []byte
	var pad int
	if len < 254 {
		bb[0] = byte(len)
		b = bb[:1]
		pad = PaddingOf(1 + len)
	} else {
		bb[0] = 254
		binints.EncodeUint24LE(uint32(len), bb[1:4])
		b = bb[:4]
		pad = PaddingOf(4 + len)
	}
	_, err := w.Write(b)
	return pad, err
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

func WriteString(w io.Writer, b []byte) error {
	pad, err := WriteStringLen(w, len(b))
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	if pad > 0 {
		var padding [4]byte
		_, err = w.Write(padding[:pad])
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadBigIntBE(r io.Reader) (*big.Int, error) {
	b, err := ReadString(r)
	if err != nil {
		return nil, err
	}

	n := new(big.Int)
	n.SetBytes(b)
	return n, nil
}

func WriteBigIntBE(w io.Writer, n *big.Int) error {
	b := n.Bytes()
	if len(b) == 0 {
		b = []byte{0}
	}

	return WriteString(w, b)
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
