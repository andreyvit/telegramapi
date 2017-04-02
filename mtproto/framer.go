package mtproto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"github.com/andreyvit/telegramapi/binints"
	"io"
	"log"
)

type Framer struct {
	MsgIDOverride uint64
	RandomReader  io.Reader

	gen  MsgIDGen
	auth *AuthResult

	seqNo uint32
}

func (fr *Framer) SetAuth(auth *AuthResult) {
	fr.auth = auth
	if fr.RandomReader == nil {
		fr.RandomReader = rand.Reader
	}
}

func (fr *Framer) Format(msg Msg) ([]byte, error) {
	var msgID uint64
	if fr.MsgIDOverride != 0 {
		msgID = fr.MsgIDOverride
		fr.MsgIDOverride = 0
	} else {
		msgID = fr.gen.Generate()
	}

	w := NewWriter()
	if fr.auth == nil {
		if msg.Type != KeyExMsg {
			panic("cannot send encrypted messages before key exchange is finished")
		}

		w.WriteUint64(0)
		w.WriteUint64(msgID)
		w.WriteInt(len(msg.Payload))
		w.Write(msg.Payload)
	} else {
		var seqNo uint32
		if msg.Type == ContentMsg {
			seqNo = fr.seqNo + 1
			fr.seqNo += 2
		} else {
			seqNo = fr.seqNo
		}

		w.Write(fr.auth.ServerSalt[:])
		w.Write(fr.auth.SessionID[:])
		w.WriteUint64(msgID)
		w.WriteUint32(seqNo)
		w.WriteInt(len(msg.Payload))
		w.Write(msg.Payload)
		hash := sha1.Sum(w.Bytes())

		var msgKey [16]byte
		copy(msgKey[:], hash[4:20])

		pad := w.PaddingTo(16)
		if pad > 0 {
			var padding [16]byte
			_, err := io.ReadFull(fr.RandomReader, padding[:pad])
			if err != nil {
				log.Printf("failed to read padding (%d): %v", pad, err)
				return nil, err
			}
			w.Write(padding[:pad])
		}
		data := w.Bytes()

		var key, iv [32]byte
		deriveAESKey(fr.auth.Key, msgKey[:], key[:], iv[:], true)

		log.Printf("AES key: %x", key)
		log.Printf("AES iv: %x", key)

		encrypted, err := AESIGEPadEncrypt(nil, data, key[:], iv[:], nil)
		if err != nil {
			log.Printf("encryption failed: %v", pad, err)
			return nil, err
		}

		w.Clear()
		w.WriteUint64(fr.auth.KeyID)
		w.Write(msgKey[:])
		w.Write(encrypted)
	}

	return w.Bytes(), nil
}

func (fr *Framer) Parse(raw []byte) (Msg, error) {
	r := bytes.NewReader(raw)

	authKeyID, err := binints.ReadUint64LE(r)
	if err != nil {
		return Msg{}, err
	}

	if authKeyID == 0 {
		var a Accum

		msgID, err := binints.ReadUint64LE(r)
		a.Push(err)

		msgLen, err := binints.ReadUint32LE(r)
		a.Push(err)

		payload, err := ReadN(r, int(msgLen))
		a.Push(err)

		a.Push(binints.ExpectEOF(r))

		return Msg{payload, KeyExMsg, msgID}, a.Error()
	} else {
		// log.Printf("Received encrypted: authKeyID=%x msgID=%x msgLen=%d cmd = %08x", authKeyID, msgID, msgLen, cmd)
		panic("authKeyID != 0")
		// a.Push(binints.ExpectEOF(r))
	}
}

func deriveAESKey(authKey, msgKey []byte, key, iv []byte, isClient bool) {
	if len(authKey) != 256 {
		panic("invalid auth key len")
	}
	if len(msgKey) != 16 {
		panic("invalid msg key len")
	}

	var x int
	if isClient {
		x = 0
	} else {
		x = 8
	}
	var src [48]byte

	// sha1_a = SHA1(msg_key + substr(auth_key, x, 32))
	copy(src[0:16], msgKey)
	copy(src[16:48], authKey[x:x+32])
	a := sha1.Sum(src[:48])

	// sha1_b = SHA1(substr(auth_key, 32+x, 16) + msg_key + substr(auth_key, 48+x, 16))
	copy(src[0:16], authKey[32+x:32+x+16])
	copy(src[16:32], msgKey)
	copy(src[32:48], authKey[48+x:48+x+16])
	b := sha1.Sum(src[:48])

	// sha1_с = SHA1(substr(auth_key, 64+x, 32) + msg_key)
	copy(src[0:32], authKey[64+x:64+x+32])
	copy(src[32:48], msgKey)
	c := sha1.Sum(src[:48])

	// sha1_d = SHA1(msg_key + substr(auth_key, 96+x, 32))
	copy(src[0:16], msgKey)
	copy(src[16:48], authKey[96+x:96+x+32])
	d := sha1.Sum(src[:48])

	// aes_key = substr(sha1_a, 0, 8) + substr(sha1_b, 8, 12) + substr(sha1_c, 4, 12)
	copy(key[0:8], a[0:8])
	copy(key[8:20], b[8:8+12])
	copy(key[20:32], c[4:4+12])

	// aes_iv = substr(sha1_a, 8, 12) + substr(sha1_b, 0, 8) + substr(sha1_c, 16, 4) + substr(sha1_d, 0, 8)
	copy(iv[0:12], a[8:8+12])
	copy(iv[12:20], b[0:8])
	copy(iv[20:24], c[16:16+4])
	copy(iv[24:32], d[0:8])
}
