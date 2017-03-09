package mtproto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
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

	clientNonce    [16]byte
	newClientNonce [16]byte
	serverNonce    [16]byte
	p, q           uint64
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

	err := binints.ReadFull(kex.RandomReader, kex.clientNonce[:])
	if err != nil {
		panic(err)
	}
	err = binints.WriteUint128LE(&buf, kex.clientNonce[:])
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
		if cmd == IDResPQ {
			var res ResPQ
			err := ReadResPQ(r, &res)
			if err != nil {
				return nil, err
			}
			return kex.handleResPQ(&res)
		} else {
			return nil, ErrUnexpectedCommand
		}

	default:
		return nil, ErrUnexpectedCommand
	}
}

func (kex *KeyEx) handleResPQ(res *ResPQ) (*OutgoingMsg, error) {
	log.Printf("res_pq: %+#v", *res)

	if 1 != subtle.ConstantTimeCompare(res.Nonce[:], kex.clientNonce[:]) {
		log.Printf("res_pq client nonce = %v, wanted %v", res.Nonce[:], kex.clientNonce[:])
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

	copy(msgdata.PQInnerData.Nonce[:], kex.clientNonce[:])
	copy(msgdata.PQInnerData.ServerNonce[:], kex.serverNonce[:])
	err := binints.ReadFull(kex.RandomReader, msgdata.PQInnerData.NewNonce[:])
	if err != nil {
		panic(err)
	}
	err = binints.ReadFull(kex.RandomReader, msgdata.Random255[:])
	if err != nil {
		panic(err)
	}

	var msgbuf bytes.Buffer
	err = msgdata.WriteTo(&msgbuf)
	if err != nil {
		return nil, err
	}
	msg := UnencryptedMsg(msgbuf.Bytes())
	kex.state = KeyExReqDHParams
	return &msg, nil
}
