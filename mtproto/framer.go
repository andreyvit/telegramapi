package mtproto

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"github.com/andreyvit/telegramapi/tl"
	"io"
	"log"
)

var ErrUnknownKeyID = errors.New("unknown auth key ID")

type FramerState struct {
	SeqNo uint32
}

type Framer struct {
	MsgIDOverride uint64
	RandomReader  io.Reader

	gen  MsgIDGen
	auth *AuthResult

	FramerState
}

func (fr *Framer) State() (*AuthResult, FramerState) {
	return fr.auth, fr.FramerState
}

func (fr *Framer) Restore(state FramerState) {
	fr.FramerState = state
}

func (fr *Framer) SetAuth(auth *AuthResult) {
	fr.auth = auth
	if fr.RandomReader == nil {
		fr.RandomReader = rand.Reader
	}
}

func (fr *Framer) Format(msg Msg) ([]byte, uint64, error) {
	var msgID uint64
	if fr.MsgIDOverride != 0 {
		msgID = fr.MsgIDOverride
		fr.MsgIDOverride = 0
	} else {
		msgID = fr.gen.Generate()
	}

	w := tl.NewWriter()
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
			seqNo = fr.SeqNo + 1
			fr.SeqNo += 2
		} else {
			seqNo = fr.SeqNo
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
				return nil, 0, err
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
			return nil, 0, err
		}

		w.Clear()
		w.WriteUint64(fr.auth.KeyID)
		w.Write(msgKey[:])
		w.Write(encrypted)
	}

	return w.Bytes(), msgID, nil
}

func (fr *Framer) Parse(raw []byte) (Msg, error) {
	var r tl.Reader
	r.Reset(raw)

	authKeyID, ok := r.TryReadUint64()
	if !ok {
		return Msg{}, r.Err()
	}

	if authKeyID == 0 {
		msgID := r.ReadUint64()
		msgLen := r.ReadInt()
		payload := r.ReadN(msgLen)
		r.ExpectEOF()

		return Msg{payload, KeyExMsg, msgID}, r.Err()
	} else {
		if authKeyID != fr.auth.KeyID {
			return Msg{}, ErrUnknownKeyID
		}

		var msgKey [16]byte
		r.ReadUint128(msgKey[:])
		enc := r.ReadToEnd()
		if err := r.Err(); err != nil {
			return Msg{}, r.Err()
		}

		log.Printf("Received encrypted: authKeyID=%x data=(%d) %x", authKeyID, len(enc), enc)

		var key, iv [32]byte
		deriveAESKey(fr.auth.Key, msgKey[:], key[:], iv[:], false)
		log.Printf("AES key: %x", key)
		log.Printf("AES iv: %x", key)

		decrypted, err := AESIGEDecrypt(nil, enc, key[:], iv[:])
		if err != nil {
			return Msg{}, err
		}

		r.Reset(decrypted)
		var salt [8]byte
		var sessid [8]byte
		r.ReadFull(salt[:])
		r.ReadFull(sessid[:])
		msgID := r.ReadUint64()
		seqNo := r.ReadInt()
		msgLen := r.ReadInt()
		payload := r.ReadN(msgLen)
		if err := r.Err(); err != nil {
			return Msg{}, r.Err()
		}

		log.Printf("Received: authKeyID=%x msgID=%v seqNo=%v payload=(%d) %x", authKeyID, msgID, seqNo, len(payload), payload)

		var typ MsgType
		if (seqNo & 1) == 0 {
			typ = ServiceMsg
		} else {
			typ = ContentMsg
		}

		return Msg{payload, typ, msgID}, nil
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
