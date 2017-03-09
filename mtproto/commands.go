package mtproto

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"io"
	"log"
	"math/big"

	"github.com/andreyvit/telegramapi/binints"
)

var ErrMalformedCommand = errors.New("malformed command")
var ErrUnexpectedCommand = errors.New("unexpected command")

const (
	IDVectorLong       uint32 = 0x1cb5c415
	IDReqPQ                   = 0x60469778
	IDResPQ                   = 0x05162463
	IDPQInnerData             = 0x83c95aec
	IDReqDHParams             = 0xd712e4be
	IDServerDHParamsOK        = 0xd0e8075c
)

type ResPQ struct {
	Nonce       [16]byte
	ServerNonce [16]byte
	PQ          *big.Int

	ServerPubKeyFingerprints []uint64
}

type ServerDHParamsOK struct {
	Nonce       [16]byte
	ServerNonce [16]byte
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

	pq, err := ReadString(r)
	if err != nil {
		return err
	}

	res.PQ = big.NewInt(0)
	res.PQ.SetBytes(pq)

	res.ServerPubKeyFingerprints, err = ReadVectorLong(r)
	if err != nil {
		return err
	}

	return nil
}

type PQInnerData struct {
	PQ *big.Int
	P  *big.Int
	Q  *big.Int

	Nonce       [16]byte
	ServerNonce [16]byte
	NewNonce    [32]byte
}

func (data *PQInnerData) WriteTo(w io.Writer) error {
	err := binints.WriteUint32LE(w, IDPQInnerData)
	if err != nil {
		return err
	}

	err = WriteString(w, data.PQ.Bytes())
	if err != nil {
		return err
	}

	err = WriteString(w, data.P.Bytes())
	if err != nil {
		return err
	}

	err = WriteString(w, data.Q.Bytes())
	if err != nil {
		return err
	}

	_, err = w.Write(data.Nonce[:])
	if err != nil {
		return err
	}

	_, err = w.Write(data.ServerNonce[:])
	if err != nil {
		return err
	}

	_, err = w.Write(data.NewNonce[:])
	if err != nil {
		return err
	}

	return nil
}

type ReqDHParams struct {
	PQInnerData
	ServerPubKeyFingerprint uint64
	PubKey                  *rsa.PublicKey
	Random255               [255]byte
}

func (data *ReqDHParams) WriteTo(w io.Writer) error {
	var unencrypted bytes.Buffer
	err := data.PQInnerData.WriteTo(&unencrypted)
	if err != nil {
		return err
	}

	hash := sha1.Sum(unencrypted.Bytes())

	var dataWithHash bytes.Buffer
	_, err = dataWithHash.Write(hash[:])
	if err != nil {
		return err
	}
	_, err = dataWithHash.Write(unencrypted.Bytes())
	if err != nil {
		return err
	}

	// pad to 255
	const tlen = 255
	olen := dataWithHash.Len()
	if olen > tlen {
		panic("dataWithHash.Len() > 255")
	}
	_, err = dataWithHash.Write(data.Random255[:tlen-olen])
	if err != nil {
		return err
	}

	encrypted := EncryptRSA(dataWithHash.Bytes(), data.PubKey)
	log.Printf("ReqDHParams: encrypted %v bytes (%v + padding) into %v bytes", dataWithHash.Len(), olen, len(encrypted))
	if len(encrypted) != 256 {
		panic("len(encrypted) != 256")
	}

	err = binints.WriteUint32LE(w, IDReqDHParams)
	if err != nil {
		return err
	}

	err = binints.WriteUint128LE(w, data.PQInnerData.Nonce[:])
	if err != nil {
		return err
	}
	err = binints.WriteUint128LE(w, data.PQInnerData.ServerNonce[:])
	if err != nil {
		return err
	}

	err = WriteString(w, data.P.Bytes())
	if err != nil {
		return err
	}
	err = WriteString(w, data.Q.Bytes())
	if err != nil {
		return err
	}

	err = binints.WriteUint64LE(w, data.ServerPubKeyFingerprint)
	if err != nil {
		return err
	}

	err = WriteString(w, encrypted)
	if err != nil {
		return err
	}

	return nil
}
