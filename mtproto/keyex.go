package mtproto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/andreyvit/telegramapi/binints"
	"github.com/andreyvit/telegramapi/tl"
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
	SessionID  [8]byte
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

func (kex *KeyEx) Start() tl.Object {
	if kex.RandomReader == nil {
		kex.RandomReader = rand.Reader
	}

	_, err := io.ReadFull(kex.RandomReader, kex.nonce[:])
	if err != nil {
		panic(err)
	}

	msg := &TLReqPQ{}
	copy(msg.Nonce[:], kex.nonce[:])

	kex.state = KeyExReqPQ
	return msg
}

func (kex *KeyEx) Handle(o tl.Object) (tl.Object, error) {
	omsg, err := kex.handle(o)
	if err != nil {
		kex.state = KeyExFailed
		kex.err = err
	}
	return omsg, err
}

func (kex *KeyEx) handle(o tl.Object) (tl.Object, error) {
	switch kex.state {
	case KeyExInit:
		return nil, ErrUnexpectedCommand
	case KeyExFailed:
		return nil, ErrUnexpectedCommand

	case KeyExReqPQ:
		switch o := o.(type) {
		case *TLResPQ:
			return kex.handleResPQ(o)
		default:
			return nil, ErrUnexpectedCommand
		}

	case KeyExReqDHParams:
		switch o := o.(type) {
		case *TLServerDHParamsOK:
			return kex.handleServerDHParamsOK(o)
		case *TLServerDHParamsFail:
			return nil, errors.New("got server_DH_params_fail")
		default:
			return nil, ErrUnexpectedCommand
		}

	case KeyExSetClientDHParams:
		switch o := o.(type) {
		case *TLDHGenOK:
			return kex.handleDHGenOK(o)
		case *TLDHGenFail:
			return nil, errors.New("got dh_gen_fail")
		case *TLDHGenRetry:
			return nil, errors.New("got dh_gen_retry")
		default:
			return nil, ErrUnexpectedCommand
		}

	default:
		return nil, ErrUnexpectedCommand
	}
}

func (kex *KeyEx) handleNoncePair(r *tl.Reader) error {
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

func (kex *KeyEx) handleResPQ(in *TLResPQ) (tl.Object, error) {
	if 1 != subtle.ConstantTimeCompare(in.Nonce[:], kex.nonce[:]) {
		return nil, errors.New("bad nonce")
	}

	copy(kex.serverNonce[:], in.ServerNonce[:])

	//log.Printf("res_pq: %+#v", *in)

	expectedFingerprint := ComputePubKeyFingerprint(kex.PubKey)
	var keyOK bool
	for _, fingerprint := range in.ServerPublicKeyFingerprints {
		if fingerprint == expectedFingerprint {
			keyOK = true
		}
	}
	if !keyOK {
		log.Printf("res_pq public key fingerprints = %v, wanted %v", in.ServerPublicKeyFingerprints, expectedFingerprint)
		return nil, errors.New("public key fingerprint mismatch")
	}

	if in.PQ.BitLen() > 64 {
		log.Printf("mtproto/keyex: PQ number does not fit into uint64: %v", in.PQ)
		return nil, errors.New("PQ too large")
	}
	pqn := in.PQ.Uint64()
	p, q := factorize(pqn)
	// log.Printf("mtproto/keyex: %v = %v (p) * %v (q)", pqn, p, q)

	_, err := io.ReadFull(kex.RandomReader, kex.newNonce[:])
	if err != nil {
		panic(err)
	}
	var randomPadding [255]byte
	_, err = io.ReadFull(kex.RandomReader, randomPadding[:])
	if err != nil {
		panic(err)
	}

	inner := &TLPQInnerData{
		PQ: in.PQ,
		P:  big.NewInt(int64(p)),
		Q:  big.NewInt(int64(q)),
	}
	copy(inner.Nonce[:], kex.nonce[:])
	copy(inner.ServerNonce[:], kex.serverNonce[:])
	copy(inner.NewNonce[:], kex.newNonce[:])

	// TODO: fill randomPadding

	m := &TLReqDHParams{
		P:                    inner.P,
		Q:                    inner.Q,
		PublicKeyFingerprint: expectedFingerprint,
		EncryptedData:        EncryptRSAWithHash(tl.Bytes(inner), randomPadding[:], kex.PubKey),
	}
	copy(m.Nonce[:], kex.nonce[:])
	copy(m.ServerNonce[:], kex.serverNonce[:])

	kex.state = KeyExReqDHParams
	return m, nil
}

func (kex *KeyEx) handleServerDHParamsOK(o *TLServerDHParamsOK) (tl.Object, error) {
	if 1 != subtle.ConstantTimeCompare(o.Nonce[:], kex.nonce[:]) {
		// log.Printf("server_dh_params_ok nonce = %v, wanted %v", o.Nonce[:], kex.nonce[:])
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(o.ServerNonce[:], kex.serverNonce[:]) {
		// log.Printf("server_dh_params_ok server nonce = %v, wanted %v", o.Nonce[:], kex.serverNonce[:])
		return nil, errors.New("bad server nonce")
	}

	deriveTempAESKey(kex.serverNonce[:], kex.newNonce[:], kex.tmpAESKey[:], kex.tmpAESIV[:])

	answer, answerHash, err := AESIGEDecryptWithHash(nil, o.EncryptedAnswer, kex.tmpAESKey[:], kex.tmpAESIV[:])
	if err != nil {
		return nil, err
	}

	// DECRYPTION

	if false {
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

	rawinner, err := Schema.ReadLimitedBoxedObjectNoEOFCheck(answer, TagServerDHInnerData)
	if err != nil {
		return nil, err
	}
	inner := rawinner.(*TLServerDHInnerData)

	if 1 != subtle.ConstantTimeCompare(inner.Nonce[:], kex.nonce[:]) {
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(inner.ServerNonce[:], kex.serverNonce[:]) {
		return nil, errors.New("bad server nonce")
	}

	kex.g = inner.G
	kex.dhPrime = inner.DHPrime

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
	gab.Exp(inner.GA, kex.b, kex.dhPrime)
	kex.auth.Key = leftZeroPad(gab.Bytes(), 256)

	authKeyHash := sha1.Sum(kex.auth.Key)
	kex.auth.KeyID = binints.DecodeUint64LE(authKeyHash[12:])
	kex.authKeyAuxHash = binints.DecodeUint64LE(authKeyHash[:8])
	for i := 0; i < 8; i++ {
		kex.auth.ServerSalt[i] = kex.newNonce[i] ^ kex.serverNonce[i]
	}

	// log.Printf("Auth key: %v (key ID: %x, server salt: %x)", hex.EncodeToString(kex.auth.Key), kex.auth.KeyID, kex.auth.ServerSalt)

	// RESPONSE

	replyInner := &TLClientDHInnerData{
		RetryID: retryID,
		GB:      gb,
	}
	copy(replyInner.Nonce[:], kex.nonce[:])
	copy(replyInner.ServerNonce[:], kex.serverNonce[:])

	encrypted, err := AESIGEPadEncryptWithHash(nil, tl.Bytes(replyInner), kex.tmpAESKey[:], kex.tmpAESIV[:], kex.RandomReader)
	if err != nil {
		return nil, err
	}

	reply := &TLSetClientDHParams{
		EncryptedData: encrypted,
	}
	copy(reply.Nonce[:], kex.nonce[:])
	copy(reply.ServerNonce[:], kex.serverNonce[:])

	kex.state = KeyExSetClientDHParams
	return reply, nil
}

func (kex *KeyEx) handleDHGenOK(in *TLDHGenOK) (tl.Object, error) {
	if 1 != subtle.ConstantTimeCompare(in.Nonce[:], kex.nonce[:]) {
		// log.Printf("server_dh_params_ok nonce = %v, wanted %v", in.Nonce[:], kex.nonce[:])
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(in.ServerNonce[:], kex.serverNonce[:]) {
		// log.Printf("server_dh_params_ok server nonce = %v, wanted %v", in.Nonce[:], kex.serverNonce[:])
		return nil, errors.New("bad server nonce")
	}

	// TODO: check in.NewNonceHash1

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
