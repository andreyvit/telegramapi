package mtproto

import (
	"crypto/rsa"
	"crypto/sha1"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"

	"github.com/andreyvit/telegramapi/binints"
	"github.com/andreyvit/telegramapi/tl"
)

const rsaBlockLen = 256

func ParsePublicKey(s string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	key := new(rsa.PublicKey)
	_, err := asn1.Unmarshal(block.Bytes, key)
	if err != nil {
		return nil, err
	}

	if key.E == 0 || key.N == nil {
		return nil, errors.New("failed to correctly parse the public key")
	}

	return key, nil
}

func ComputePubKeyFingerprint(key *rsa.PublicKey) uint64 {
	var w tl.Writer
	w.WriteBigInt(key.N)
	w.WriteBigInt(big.NewInt(int64(key.E)))
	sha1 := sha1.Sum(w.Bytes())
	return binints.DecodeUint64LE(sha1[12:20])
}

func EncryptRSA(data []byte, key *rsa.PublicKey) []byte {
	e := big.NewInt(int64(key.E))

	var m big.Int
	m.SetBytes(data)

	var result big.Int
	result.Exp(&m, e, key.N)

	return leftZeroPad(result.Bytes(), rsaBlockLen)
}
