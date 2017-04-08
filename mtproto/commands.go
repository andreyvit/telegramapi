package mtproto

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
)

func EncryptRSAWithHash(unencrypted []byte, paddingSource []byte, pubKey *rsa.PublicKey) []byte {
	hash := sha1.Sum(unencrypted)

	var dataWithHash bytes.Buffer
	dataWithHash.Write(hash[:])
	dataWithHash.Write(unencrypted)

	// pad to 255
	const tlen = 255
	olen := dataWithHash.Len()
	if olen > tlen {
		panic("dataWithHash.Len() > 255")
	}
	dataWithHash.Write(paddingSource[:tlen-olen])

	encrypted := EncryptRSA(dataWithHash.Bytes(), pubKey)
	// log.Printf("ReqDHParams: encrypted %v bytes (%v + padding) into %v bytes", dataWithHash.Len(), olen, len(encrypted))
	if len(encrypted) != 256 {
		panic("len(encrypted) != 256")
	}

	return encrypted
}
