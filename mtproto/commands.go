package mtproto

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"log"
	"math/big"
)

var ErrMalformedCommand = errors.New("malformed command")
var ErrUnexpectedCommand = errors.New("unexpected command")

const (
	IDVectorLong         uint32 = 0x1cb5c415
	IDReqPQ                     = 0x60469778
	IDResPQ                     = 0x05162463
	IDPQInnerData               = 0x83c95aec
	IDReqDHParams               = 0xd712e4be
	IDServerDHParamsOK          = 0xd0e8075c
	IDServerDHParamsFail        = 0x79cb045d
	IDServerDHInnerData         = 0xb5890dba
	IDClientDHInnerData         = 0x6643b654
	IDSetClientDHParams         = 0xf5045f1f
	IDDHGenOK                   = 0x3bcbf734
	IDDHGenRetry                = 0x46dc1fb9
	IDDHGenFail                 = 0xa69dae02
)

type ResPQ struct {
	Nonce       [16]byte
	ServerNonce [16]byte
	PQ          *big.Int

	ServerPubKeyFingerprints []uint64
}

func (data *ResPQ) ReadFrom(r *Reader) {
	r.ReadUint128(data.Nonce[:])
	r.ReadUint128(data.ServerNonce[:])
	data.PQ = r.ReadBigInt()
	data.ServerPubKeyFingerprints = r.ReadVectorLong()
}

type PQInnerData struct {
	PQ *big.Int
	P  *big.Int
	Q  *big.Int

	Nonce       [16]byte
	ServerNonce [16]byte
	NewNonce    [32]byte
}

type DHGenOK struct {
	Nonce        [16]byte
	ServerNonce  [16]byte
	NewNonceHash [16]byte
}

func (data *PQInnerData) WriteTo(w *Writer) {
	w.WriteCmd(IDPQInnerData)
	w.WriteBigInt(data.PQ)
	w.WriteBigInt(data.P)
	w.WriteBigInt(data.Q)
	w.Write(data.Nonce[:])
	w.Write(data.ServerNonce[:])
	w.Write(data.NewNonce[:])
}

type ReqDHParams struct {
	PQInnerData
	ServerPubKeyFingerprint uint64
	PubKey                  *rsa.PublicKey
	Random255               [255]byte
}

func (data *ReqDHParams) WriteTo(w *Writer) {
	unencrypted := BytesOf(&data.PQInnerData)
	hash := sha1.Sum(unencrypted)

	var dataWithHash bytes.Buffer
	_, _ = dataWithHash.Write(hash[:])
	_, _ = dataWithHash.Write(unencrypted)

	// pad to 255
	const tlen = 255
	olen := dataWithHash.Len()
	if olen > tlen {
		panic("dataWithHash.Len() > 255")
	}
	_, _ = dataWithHash.Write(data.Random255[:tlen-olen])

	encrypted := EncryptRSA(dataWithHash.Bytes(), data.PubKey)
	log.Printf("ReqDHParams: encrypted %v bytes (%v + padding) into %v bytes", dataWithHash.Len(), olen, len(encrypted))
	if len(encrypted) != 256 {
		panic("len(encrypted) != 256")
	}

	w.WriteCmd(IDReqDHParams)
	w.WriteUint128(data.PQInnerData.Nonce[:])
	w.WriteUint128(data.PQInnerData.ServerNonce[:])
	w.WriteBigInt(data.P)
	w.WriteBigInt(data.Q)
	w.WriteUint64(data.ServerPubKeyFingerprint)
	w.WriteBlob(encrypted)
}

type ServerDHParamsOK struct {
	Nonce       [16]byte
	ServerNonce [16]byte
	// G           int
	// DHPrime    *big.Int
	GA         *big.Int
	ServerTime int
}
