package mtproto

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"

	"github.com/andreyvit/telegramapi/binints"
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
	var buf bytes.Buffer
	err := WriteString(&buf, key.N.Bytes())
	if err != nil {
		panic(err)
	}

	e := big.NewInt(int64(key.E))
	err = WriteString(&buf, e.Bytes())
	if err != nil {
		panic(err)
	}

	sha1 := sha1.Sum(buf.Bytes())
	return binints.DecodeUint64LE(sha1[12:20])
}

func EncryptRSA(data []byte, key *rsa.PublicKey) []byte {
	e := big.NewInt(int64(key.E))

	var m big.Int
	m.SetBytes(data)

	var result big.Int
	result.Exp(&m, e, key.N)

	bytes := result.Bytes()
	if len(bytes) >= rsaBlockLen {
		return bytes
	} else {
		result := make([]byte, rsaBlockLen)
		copy(result[rsaBlockLen-len(bytes):rsaBlockLen], bytes)
		return result
	}
}
