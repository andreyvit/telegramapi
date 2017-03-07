package mtproto

import (
	// "github.com/andreyvit/telegramapi/binints"
	"io"
)

const IDVectorLong = 0x1cb5c415
const IDResPQ = 0x05162463

type ResPQ struct {
	Nonce       [16]byte
	ServerNonce [16]byte
	PQ          []byte

	ServerPubKeyFingerprints []uint64
}

func ReadResPQ(r io.Reader, res *ResPQ) error {
	err := ReadUint128(r, res.Nonce[:])
	if err != nil {
		return err
	}

	err = ReadUint128(r, res.ServerNonce[:])
	if err != nil {
		return err
	}

	res.PQ, err = ReadString(r)
	if err != nil {
		return err
	}

	res.ServerPubKeyFingerprints, err = ReadVectorLong(r)
	if err != nil {
		return err
	}

	return nil
}
