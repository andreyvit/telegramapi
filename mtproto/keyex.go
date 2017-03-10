package mtproto

import (
	"bytes"
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

const (
	KeyExInit keyExState = iota
	KeyExDone
	KeyExFailed
	KeyExReqPQ
	KeyExReqDHParams
)

type KeyEx struct {
	RandomReader io.Reader
	PubKey       *rsa.PublicKey

	state keyExState

	nonce       [16]byte
	newNonce    [32]byte
	serverNonce [16]byte
	p, q        uint64

	tmpAESKey [32]byte
	tmpAESIV  [32]byte

	g       int
	b       *big.Int
	dhPrime *big.Int
}

func (kex *KeyEx) IsFinished() bool {
	switch kex.state {
	case KeyExDone, KeyExFailed:
		return true
	default:
		return false
	}
}

func (kex *KeyEx) Start() OutgoingMsg {
	if kex.RandomReader == nil {
		kex.RandomReader = rand.Reader
	}

	var buf bytes.Buffer
	binints.WriteUint32LE(&buf, IDReqPQ)

	_, err := io.ReadFull(kex.RandomReader, kex.nonce[:])
	if err != nil {
		panic(err)
	}
	err = binints.WriteUint128LE(&buf, kex.nonce[:])
	if err != nil {
		panic(err)
	}

	kex.state = KeyExReqPQ
	return UnencryptedMsg(buf.Bytes())
}

func (kex *KeyEx) Handle(payload []byte) (*OutgoingMsg, error) {
	msg, err := kex.handle(payload)
	if err != nil {
		kex.state = KeyExFailed
	}
	return msg, err
}

func (kex *KeyEx) handle(payload []byte) (*OutgoingMsg, error) {
	r := bytes.NewReader(payload)
	cmd, err := binints.ReadUint32LE(r)
	if err != nil {
		return nil, ErrMalformedCommand
	}

	switch kex.state {
	case KeyExInit:
		return nil, ErrUnexpectedCommand
	case KeyExFailed:
		return nil, ErrUnexpectedCommand

	case KeyExReqPQ:
		if cmd != IDResPQ {
			return nil, ErrUnexpectedCommand
		}

		var res ResPQ
		err := ReadResPQ(r, &res)
		if err != nil {
			return nil, err
		}
		return kex.handleResPQ(&res)

	case KeyExReqDHParams:
		if cmd == IDServerDHParamsOK {
			return kex.handleServerDHParamsOK(r)
		} else if cmd == IDServerDHParamsFail {
			return nil, errors.New("got server_DH_params_fail")
		} else {
			return nil, ErrUnexpectedCommand
		}

	default:
		return nil, ErrUnexpectedCommand
	}
}

func (kex *KeyEx) handleResPQ(res *ResPQ) (*OutgoingMsg, error) {
	log.Printf("res_pq: %+#v", *res)

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		log.Printf("res_pq nonce = %v, wanted %v", res.Nonce[:], kex.nonce[:])
		return nil, errors.New("bad client nonce")
	}

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

	copy(kex.serverNonce[:], res.ServerNonce[:])

	if res.PQ.BitLen() > 64 {
		log.Printf("mtproto/keyex: PQ number does not fit into uint64: %v", res.PQ)
		return nil, errors.New("PQ too large")
	}
	pq := res.PQ.Uint64()
	p, q := factorize(pq)
	log.Printf("mtproto/keyex: %v = %v (p) * %v (q)", pq, p, q)

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

	var msgbuf bytes.Buffer
	err = msgdata.WriteTo(&msgbuf)
	if err != nil {
		return nil, err
	}
	msg := UnencryptedMsg(msgbuf.Bytes())
	kex.state = KeyExReqDHParams
	return &msg, nil
}

func (kex *KeyEx) handleServerDHParamsOK(r io.Reader) (*OutgoingMsg, error) {
	var res ServerDHParamsOK

	err := binints.ReadUint128LE(r, res.Nonce[:])
	if err != nil {
		return nil, err
	}

	err = binints.ReadUint128LE(r, res.ServerNonce[:])
	if err != nil {
		return nil, err
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

	encrypted, err := ReadString(r)
	if err != nil {
		return nil, err
	}

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

	r = bytes.NewReader(answer)

	ansCmd, err := binints.ReadUint32LE(r)
	if err != nil {
		return nil, err
	}
	if ansCmd != IDServerDHInnerData {
		return nil, errors.New("expected server_DH_inner_data")
	}

	err = binints.ReadUint128LE(r, res.Nonce[:])
	if err != nil {
		return nil, err
	}

	err = binints.ReadUint128LE(r, res.ServerNonce[:])
	if err != nil {
		return nil, err
	}

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.nonce[:]) {
		return nil, errors.New("bad nonce")
	}
	if 1 != subtle.ConstantTimeCompare(res.ServerNonce[:], kex.serverNonce[:]) {
		return nil, errors.New("bad server nonce")
	}

	kex.g, err = binints.ReadUint32LEAsInt(r)
	if err != nil {
		return nil, err
	}

	kex.dhPrime, err = ReadBigIntBE(r)
	if err != nil {
		return nil, err
	}

	res.GA, err = ReadBigIntBE(r)
	if err != nil {
		return nil, err
	}

	res.ServerTime, err = binints.ReadUint32LEAsInt(r)
	if err != nil {
		return nil, err
	}

	// VERIFICATION

	if !kex.dhPrime.ProbablyPrime(20) {
		return nil, errors.New("DHPrime not prime")
	}
	// TODO: more checks required by MTProto protocol

	// RESPONSE

	var bbytes [256]byte
	_, err = io.ReadFull(kex.RandomReader, bbytes[:])
	if err != nil {
		panic(err)
	}
	kex.b = new(big.Int)
	kex.b.SetBytes(bbytes[:])

	gb := new(big.Int)
	gb.Exp(big.NewInt(int64(kex.g)), kex.b, kex.dhPrime)

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
