package mtproto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"io"
	"log"
	"math/big"
)

type keyExState int

var ErrAfterKeyExchangeFailed = errors.New("no commands can be processed after failed key exchange")
var ErrKeyExchangeNotFinished = errors.New("key exchange not yet finished")

const (
	KeyExInit keyExState = iota
	KeyExDone
	KeyExFailed
	KeyExReqPQ
	KeyExReqDHParams
	KeyExSetClientDHParams
)

type AuthResult struct {
	Key        []byte
	KeyID      uint64
	ServerSalt [8]byte
	TimeOffset int
}

type KeyEx struct {
	RandomReader io.Reader
	PubKey       *rsa.PublicKey

	state keyExState
	err   error

	nonce       [16]byte
	newNonce    [32]byte
	serverNonce [16]byte
	p, q        uint64

	tmpAESKey [32]byte
	tmpAESIV  [32]byte

	g       int
	b       *big.Int
	dhPrime *big.Int

	auth           AuthResult
	authKeyAuxHash uint64
}

func (kex *KeyEx) Result() (*AuthResult, error) {
	switch kex.state {
	case KeyExDone:
		return &kex.auth, nil
	case KeyExFailed:
		return nil, kex.err
	default:
		return nil, ErrKeyExchangeNotFinished
	}
}

func (kex *KeyEx) IsFinished() bool {
	switch kex.state {
	case KeyExDone, KeyExFailed:
		return true
	default:
		return false
	}
}

func (kex *KeyEx) Start() Msg {
	if kex.RandomReader == nil {
		kex.RandomReader = rand.Reader
	}

	_, err := io.ReadFull(kex.RandomReader, kex.nonce[:])
	if err != nil {
		panic(err)
	}

	w := NewWriterCmd(IDReqPQ)
	w.WriteUint128(kex.nonce[:])
	msg := MakeMsg(w.Bytes(), KeyExMsg)

	kex.state = KeyExReqPQ
	return msg
}

func (kex *KeyEx) Handle(r *Reader) (*Msg, error) {
	msg, err := kex.handle(r)
	if err != nil {
		kex.state = KeyExFailed
		kex.err = err
	}
	return msg, err
}

func (kex *KeyEx) handle(r *Reader) (*Msg, error) {
	cmd := r.Cmd()

	switch kex.state {
	case KeyExInit:
		return nil, ErrUnexpectedCommand
	case KeyExFailed:
		return nil, ErrUnexpectedCommand

	case KeyExReqPQ:
		if cmd != IDResPQ {
			return nil, ErrUnexpectedCommand
		}

		return kex.handleResPQ(r)

	case KeyExReqDHParams:
		if cmd == IDServerDHParamsOK {
			return kex.handleServerDHParamsOK(r)
		} else if cmd == IDServerDHParamsFail {
			return nil, errors.New("got server_DH_params_fail")
		} else {
			return nil, ErrUnexpectedCommand
		}

	case KeyExSetClientDHParams:
		if cmd == IDDHGenOK {
			return kex.handleDHGenOK(r)
		} else if cmd == IDDHGenRetry {
			return nil, errors.New("got dh_gen_retry")
		} else if cmd == IDDHGenFail {
			return nil, errors.New("got dh_gen_fail")
		} else {
			return nil, ErrUnexpectedCommand
		}

	default:
		return nil, ErrUnexpectedCommand
	}
}

func (kex *KeyEx) handleNoncePair(r *Reader) error {
	var nonce [16]byte
	if r.ReadUint128(nonce[:]) {
		if 1 != subtle.ConstantTimeCompare(nonce[:], kex.nonce[:]) {
			return errors.New("bad nonce")
		}
	}
	if r.ReadUint128(nonce[:]) {
		if 1 != subtle.ConstantTimeCompare(nonce[:], kex.nonce[:]) {
			return errors.New("bad server nonce")
		}
	}
	return nil
}

func (kex *KeyEx) handleResPQ(r *Reader) (*Msg, error) {
	var res ResPQ
	res.ReadFrom(r)
	if r.Err() != nil {
		return nil, r.Err()
	}

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		return nil, errors.New("bad nonce")
	}

	copy(kex.serverNonce[:], res.ServerNonce[:])

	//log.Printf("res_pq: %+#v", *res)

	expectedFingerprint := ComputePubKeyFingerprint(kex.PubKey)
	var keyOK bool
	for _, fingerprint := range res.ServerPubKeyFingerprints {
		if fingerprint == expectedFingerprint {
			keyOK = true
		}
	}
	if !keyOK {
		log.Printf("res_pq public key fingerprints = %v, wanted %v", res.ServerPubKeyFingerprints, expectedFingerprint)
		return nil, errors.New("public key fingerprint mismatch")
	}

	if res.PQ.BitLen() > 64 {
		log.Printf("mtproto/keyex: PQ number does not fit into uint64: %v", res.PQ)
		return nil, errors.New("PQ too large")
	}
	pqn := res.PQ.Uint64()
	p, q := factorize(pqn)
	log.Printf("mtproto/keyex: %v = %v (p) * %v (q)", pqn, p, q)

	msgdata := &ReqDHParams{
		PQInnerData: PQInnerData{
			PQ: res.PQ,
			P:  big.NewInt(int64(p)),
			Q:  big.NewInt(int64(q)),
		},
		ServerPubKeyFingerprint: expectedFingerprint,
		PubKey:                  kex.PubKey,
	}

	_, err := io.ReadFull(kex.RandomReader, kex.newNonce[:])
	if err != nil {
		panic(err)
	}
	_, err = io.ReadFull(kex.RandomReader, msgdata.Random255[:])
	if err != nil {
		panic(err)
	}

	copy(msgdata.PQInnerData.Nonce[:], kex.nonce[:])
	copy(msgdata.PQInnerData.ServerNonce[:], kex.serverNonce[:])
	copy(msgdata.PQInnerData.NewNonce[:], kex.newNonce[:])

	msg := MakeMsg(BytesOf(msgdata), KeyExMsg)
	kex.state = KeyExReqDHParams
	return &msg, nil
}

func (kex *KeyEx) handleServerDHParamsOK(r *Reader) (*Msg, error) {
	var res ServerDHParamsOK

	r.ReadUint128(res.Nonce[:])
	r.ReadUint128(res.ServerNonce[:])
	encrypted := r.ReadString()
	if r.Err() != nil {
		return nil, r.Err()
	}

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		// log.Printf("server_dh_params_ok nonce = %v, wanted %v", res.Nonce[:], kex.nonce[:])
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(res.ServerNonce[:], kex.serverNonce[:]) {
		// log.Printf("server_dh_params_ok server nonce = %v, wanted %v", res.Nonce[:], kex.serverNonce[:])
		return nil, errors.New("bad server nonce")
	}

	deriveTempAESKey(kex.serverNonce[:], kex.newNonce[:], kex.tmpAESKey[:], kex.tmpAESIV[:])

	answer, answerHash, err := AESIGEDecryptWithHash(nil, encrypted, kex.tmpAESKey[:], kex.tmpAESIV[:])
	if err != nil {
		return nil, err
	}

	// DECRYPTION

	if true {
		if false {
			log.Printf("Server nonce: %v", hex.EncodeToString(kex.serverNonce[:]))
			log.Printf("New nonce: %v", hex.EncodeToString(kex.newNonce[:]))
			log.Printf("Decryption temp AES key: %v", hex.EncodeToString(kex.tmpAESKey[:]))
			log.Printf("Decryption temp AES IV: %v", hex.EncodeToString(kex.tmpAESIV[:]))
		}
		log.Printf("Decrypted: %v", hex.EncodeToString(answer))
	}

	// TODO: check hash here (need to determine the reader offset here)
	_ = answerHash

	r = NewReader(answer)
	if r.Cmd() != IDServerDHInnerData {
		return nil, errors.New("expected server_DH_inner_data")
	}

	r.ReadUint128(res.Nonce[:])
	r.ReadUint128(res.ServerNonce[:])
	kex.g = r.ReadInt()
	kex.dhPrime = r.ReadBigInt()
	res.GA = r.ReadBigInt()
	res.ServerTime = r.ReadInt()
	if r.Err() != nil {
		return nil, r.Err()
	}

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(res.ServerNonce[:], kex.serverNonce[:]) {
		return nil, errors.New("bad server nonce")
	}

	// VERIFICATION

	if !kex.dhPrime.ProbablyPrime(20) {
		return nil, errors.New("DHPrime not prime")
	}
	// TODO: more checks required by MTProto protocol

	// PROCESSING

	var bbytes [256]byte
	_, err = io.ReadFull(kex.RandomReader, bbytes[:])
	if err != nil {
		return nil, err
	}
	kex.b = new(big.Int)
	kex.b.SetBytes(bbytes[:])

	retryID := kex.authKeyAuxHash // zero on first attempt

	gb := new(big.Int)
	gb.Exp(big.NewInt(int64(kex.g)), kex.b, kex.dhPrime)

	gab := new(big.Int)
	gab.Exp(res.GA, kex.b, kex.dhPrime)
	kex.auth.Key = leftZeroPad(gab.Bytes(), 256)

	authKeyHash := sha1.Sum(kex.auth.Key)
	kex.auth.KeyID = binints.DecodeUint64LE(authKeyHash[12:])
	kex.authKeyAuxHash = binints.DecodeUint64LE(authKeyHash[:8])
	for i := 0; i < 8; i++ {
		kex.auth.ServerSalt[i] = kex.newNonce[i] ^ kex.serverNonce[i]
	}

	log.Printf("Auth key: %v (key ID: %x, server salt: %x)", hex.EncodeToString(kex.auth.Key), kex.auth.KeyID, kex.auth.ServerSalt)

	// RESPONSE

	w := NewWriterCmd(IDClientDHInnerData)
	w.WriteUint128(kex.nonce[:])
	w.WriteUint128(kex.serverNonce[:])
	w.WriteUint64(retryID)
	w.WriteBigInt(gb)

	encrypted, err = AESIGEPadEncryptWithHash(nil, w.Bytes(), kex.tmpAESKey[:], kex.tmpAESIV[:], kex.RandomReader)
	if err != nil {
		return nil, err
	}

	w = NewWriterCmd(IDSetClientDHParams)
	w.WriteUint128(kex.nonce[:])
	w.WriteUint128(kex.serverNonce[:])
	w.WriteString(encrypted)

	msg := MakeMsg(w.Bytes(), KeyExMsg)
	kex.state = KeyExSetClientDHParams
	return &msg, nil
}

func (kex *KeyEx) handleDHGenOK(r *Reader) (*Msg, error) {
	var res DHGenOK

	r.ReadUint128(res.Nonce[:])
	r.ReadUint128(res.ServerNonce[:])
	r.ReadUint128(res.NewNonceHash[:])
	if r.Err() != nil {
		return nil, r.Err()
	}

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		// log.Printf("server_dh_params_ok nonce = %v, wanted %v", res.Nonce[:], kex.nonce[:])
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(res.ServerNonce[:], kex.serverNonce[:]) {
		// log.Printf("server_dh_params_ok server nonce = %v, wanted %v", res.Nonce[:], kex.serverNonce[:])
		return nil, errors.New("bad server nonce")
	}

	// TODO: check NewNonceHash

	log.Printf("✓ Key exchange complete")

	kex.state = KeyExDone
	return nil, nil
}

func deriveTempAESKey(serverNonce, newNonce []byte, key, iv []byte) {
	if len(key) != 32 {
		panic("len(key) != 32")
	}
	if len(iv) != 32 {
		panic("len(iv) != 32")
	}

	var src [64]byte
	copy(src[:32], newNonce)
	copy(src[32:48], serverNonce)
	nnsn := sha1.Sum(src[:48]) // NewNonce, ServerNonce

	copy(src[:16], serverNonce)
	copy(src[16:48], newNonce)
	snnn := sha1.Sum(src[:48]) // ServerNonce, NewNonce

	copy(src[:32], newNonce)
	copy(src[32:64], newNonce)
	nnnn := sha1.Sum(src[:64]) // NewNonce, NewNonce

	// tmp_aes_key := SHA1(new_nonce + server_nonce) + substr (SHA1(server_nonce + new_nonce), 0, 12);
	copy(key[:20], nnsn[:])
	copy(key[20:], snnn[:12])

	// tmp_aes_iv := substr (SHA1(server_nonce + new_nonce), 12, 8) + SHA1(new_nonce + new_nonce) + substr (new_nonce, 0, 4);
	copy(iv[:8], snnn[12:])
	copy(iv[8:28], nnnn[:])
	copy(iv[28:], newNonce[:4])
}
