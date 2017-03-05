package telegramapi

import (
	"io"
)

func writeUint24(w io.ByteWriter, v uint32) error {
	if (v & 0xFF000000) != 0 {
		panic("does not fit into Uint24")
	}

	err := w.WriteByte(byte(v))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 8))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 16))
	if err != nil {
		return err
	}

	return nil
}

func writeUint32(w io.ByteWriter, v uint32) error {
	err := w.WriteByte(byte(v))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 8))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 16))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 24))
	if err != nil {
		return err
	}

	return nil
}

func writeUint64(w io.ByteWriter, v uint64) error {
	err := w.WriteByte(byte(v))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 8))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 16))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 24))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 32))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 40))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 48))
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(v >> 56))
	if err != nil {
		return err
	}

	return nil
}

func writeUint128(w io.ByteWriter, a, b uint64) error {
	err := writeUint64(w, a)
	if err != nil {
		return err
	}
	err = writeUint64(w, b)
	if err != nil {
		return err
	}

	return nil
}
