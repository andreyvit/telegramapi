package mtproto

import (
	"errors"
	"github.com/andreyvit/telegramapi/tl"
	"math/big"
	"time"
)

const (
	TagInt                     uint32 = 0x00000000
	TagLong                           = 0x00000000
	TagDouble                         = 0x00000000
	TagString                         = 0x00000000
	TagResPQ                          = 0x05162463
	TagPQInnerData                    = 0x83c95aec
	TagServerDHParamsFail             = 0x79cb045d
	TagServerDHParamsOk               = 0xd0e8075c
	TagServerDHInnerData              = 0xb5890dba
	TagClientDHInnerData              = 0x6643b654
	TagDhGenOk                        = 0x3bcbf734
	TagDhGenRetry                     = 0x46dc1fb9
	TagDhGenFail                      = 0xa69dae02
	TagRpcResult                      = 0xf35c6d01
	TagRpcError                       = 0x2144ca19
	TagRpcAnswerUnknown               = 0x5e2ad36e
	TagRpcAnswerDroppedRunning        = 0xcd78e586
	TagRpcAnswerDropped               = 0xa43ad8b7
	TagFutureSalt                     = 0x0949d9dc
	TagFutureSalts                    = 0xae500895
	TagPong                           = 0x347773c5
	TagDestroySessionOk               = 0xe22045fc
	TagDestroySessionNone             = 0x62d350c9
	TagNewSessionCreated              = 0x9ec20908
	TagMsgContainer                   = 0x73f1f8dc
	TagMessage                        = 0x00000000
	TagMsgCopy                        = 0xe06046b2
	TagGzipPacked                     = 0x3072cfa1
	TagMsgsAck                        = 0x62d6b459
	TagBadMsgNotification             = 0xa7eff811
	TagBadServerSalt                  = 0xedab447b
	TagMsgResendReq                   = 0x7d861a08
	TagMsgsStateReq                   = 0xda69fb52
	TagMsgsStateInfo                  = 0x04deb57d
	TagMsgsAllInfo                    = 0x8cc0d131
	TagMsgDetailedInfo                = 0x276d3ec6
	TagMsgNewDetailedInfo             = 0x809db6df
	TagReqPq                          = 0x60469778
	TagReqDHParams                    = 0xd712e4be
	TagSetClientDHParams              = 0xf5045f1f
	TagRpcDropAnswer                  = 0x58e4a740
	TagGetFutureSalts                 = 0xb921bd04
	TagPing                           = 0x7abe77ec
	TagPingDelayDisconnect            = 0xf3427b8c
	TagDestroySession                 = 0xe7512126
	TagHttpWait                       = 0x9299359f
	TagInt128                         = 0x00000000
	TagInt256                         = 0x00000000
	TagBytes                          = 0x00000000
	TagBigint                         = 0x00000000
	TagUnixtime                       = 0x00000000
	TagObject                         = 0x00000000
	TagVector                         = 0x1cb5c415
)

// TLResPQ represents resPQ from TL schema
type TLResPQ struct {
	Nonce                       [16]byte // nonce: int128
	ServerNonce                 [16]byte // server_nonce: int128
	Pq                          *big.Int // pq: bytes
	ServerPublicKeyFingerprints []uint64 // server_public_key_fingerprints: Vector<long>
}

func (s *TLResPQ) Cmd() uint32 {
	return TagResPQ
}

func (s *TLResPQ) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	s.Pq = r.ReadBigInt()
	if cmd := r.ReadCmd(); cmd != TagVector {
		r.Fail(errors.New("expected: vector"))
	}
	s.ServerPublicKeyFingerprints = make([]uint64, r.ReadInt())
	for i := 0; i < len(s.ServerPublicKeyFingerprints); i++ {
		s.ServerPublicKeyFingerprints[i] = r.ReadUint64()
	}
}

func (s *TLResPQ) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.WriteBigInt(s.Pq)
	w.WriteCmd(TagVector)
	w.WriteInt(len(s.ServerPublicKeyFingerprints))
	for i := 0; i < len(s.ServerPublicKeyFingerprints); i++ {
		w.WriteUint64(s.ServerPublicKeyFingerprints[i])
	}
}

// TLPQInnerData represents p_q_inner_data from TL schema
type TLPQInnerData struct {
	Pq          *big.Int // pq: bytes
	P           *big.Int // p: bytes
	Q           *big.Int // q: bytes
	Nonce       [16]byte // nonce: int128
	ServerNonce [16]byte // server_nonce: int128
	NewNonce    [32]byte // new_nonce: int256
}

func (s *TLPQInnerData) Cmd() uint32 {
	return TagPQInnerData
}

func (s *TLPQInnerData) ReadBareFrom(r *tl.Reader) {
	s.Pq = r.ReadBigInt()
	s.P = r.ReadBigInt()
	s.Q = r.ReadBigInt()
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	r.ReadFull(s.NewNonce[:])
}

func (s *TLPQInnerData) WriteBareTo(w *tl.Writer) {
	w.WriteBigInt(s.Pq)
	w.WriteBigInt(s.P)
	w.WriteBigInt(s.Q)
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.Write(s.NewNonce[:])
}

// TLServerDHInnerData represents server_DH_inner_data from TL schema
type TLServerDHInnerData struct {
	Nonce       [16]byte  // nonce: int128
	ServerNonce [16]byte  // server_nonce: int128
	G           int       // g: int
	DhPrime     *big.Int  // dh_prime: bytes
	GA          *big.Int  // g_a: bytes
	ServerTime  time.Time // server_time: int
}

func (s *TLServerDHInnerData) Cmd() uint32 {
	return TagServerDHInnerData
}

func (s *TLServerDHInnerData) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	s.G = r.ReadInt()
	s.DhPrime = r.ReadBigInt()
	s.GA = r.ReadBigInt()
	s.ServerTime = r.ReadTimeSec32()
}

func (s *TLServerDHInnerData) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.WriteInt(s.G)
	w.WriteBigInt(s.DhPrime)
	w.WriteBigInt(s.GA)
	w.WriteTimeSec32(s.ServerTime)
}

// TLClientDHInnerData represents client_DH_inner_data from TL schema
type TLClientDHInnerData struct {
	Nonce       [16]byte // nonce: int128
	ServerNonce [16]byte // server_nonce: int128
	RetryId     uint64   // retry_id: long
	GB          *big.Int // g_b: bytes
}

func (s *TLClientDHInnerData) Cmd() uint32 {
	return TagClientDHInnerData
}

func (s *TLClientDHInnerData) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	s.RetryId = r.ReadUint64()
	s.GB = r.ReadBigInt()
}

func (s *TLClientDHInnerData) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.WriteUint64(s.RetryId)
	w.WriteBigInt(s.GB)
}

// TLRpcResult represents rpc_result from TL schema
type TLRpcResult struct {
	ReqMsgId uint64    // req_msg_id: long
	Result   tl.Object // result: Object
}

func (s *TLRpcResult) Cmd() uint32 {
	return TagRpcResult
}

func (s *TLRpcResult) ReadBareFrom(r *tl.Reader) {
	s.ReqMsgId = r.ReadUint64()
	s.Result = ReadBoxedObjectFrom(r)
}

func (s *TLRpcResult) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.ReqMsgId)
	w.WriteCmd(s.Result.Cmd())
	s.Result.WriteBareTo(w)
}

// TLRpcError represents rpc_error from TL schema
type TLRpcError struct {
	ErrorCode    int    // error_code: int
	ErrorMessage string // error_message: string
}

func (s *TLRpcError) Cmd() uint32 {
	return TagRpcError
}

func (s *TLRpcError) ReadBareFrom(r *tl.Reader) {
	s.ErrorCode = r.ReadInt()
	s.ErrorMessage = r.ReadString()
}

func (s *TLRpcError) WriteBareTo(w *tl.Writer) {
	w.WriteInt(s.ErrorCode)
	w.WriteString(s.ErrorMessage)
}

// TLFutureSalt represents future_salt from TL schema
type TLFutureSalt struct {
	ValidSince int    // valid_since: int
	ValidUntil int    // valid_until: int
	Salt       uint64 // salt: long
}

func (s *TLFutureSalt) Cmd() uint32 {
	return TagFutureSalt
}

func (s *TLFutureSalt) ReadBareFrom(r *tl.Reader) {
	s.ValidSince = r.ReadInt()
	s.ValidUntil = r.ReadInt()
	s.Salt = r.ReadUint64()
}

func (s *TLFutureSalt) WriteBareTo(w *tl.Writer) {
	w.WriteInt(s.ValidSince)
	w.WriteInt(s.ValidUntil)
	w.WriteUint64(s.Salt)
}

// TLFutureSalts represents future_salts from TL schema
type TLFutureSalts struct {
	ReqMsgId uint64          // req_msg_id: long
	Now      int             // now: int
	Salts    []*TLFutureSalt // salts: vector<future_salt>
}

func (s *TLFutureSalts) Cmd() uint32 {
	return TagFutureSalts
}

func (s *TLFutureSalts) ReadBareFrom(r *tl.Reader) {
	s.ReqMsgId = r.ReadUint64()
	s.Now = r.ReadInt()
	s.Salts = make([]*TLFutureSalt, r.ReadInt())
	for i := 0; i < len(s.Salts); i++ {
		s.Salts[i] = new(TLFutureSalt)
		s.Salts[i].ReadBareFrom(r)
	}
}

func (s *TLFutureSalts) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.ReqMsgId)
	w.WriteInt(s.Now)
	w.WriteInt(len(s.Salts))
	for i := 0; i < len(s.Salts); i++ {
		s.Salts[i].WriteBareTo(w)
	}
}

// TLPong represents pong from TL schema
type TLPong struct {
	MsgId  uint64 // msg_id: long
	PingId uint64 // ping_id: long
}

func (s *TLPong) Cmd() uint32 {
	return TagPong
}

func (s *TLPong) ReadBareFrom(r *tl.Reader) {
	s.MsgId = r.ReadUint64()
	s.PingId = r.ReadUint64()
}

func (s *TLPong) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.MsgId)
	w.WriteUint64(s.PingId)
}

// TLNewSession represents new_session_created from TL schema
type TLNewSession struct {
	FirstMsgId uint64 // first_msg_id: long
	UniqueId   uint64 // unique_id: long
	ServerSalt uint64 // server_salt: long
}

func (s *TLNewSession) Cmd() uint32 {
	return TagNewSessionCreated
}

func (s *TLNewSession) ReadBareFrom(r *tl.Reader) {
	s.FirstMsgId = r.ReadUint64()
	s.UniqueId = r.ReadUint64()
	s.ServerSalt = r.ReadUint64()
}

func (s *TLNewSession) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.FirstMsgId)
	w.WriteUint64(s.UniqueId)
	w.WriteUint64(s.ServerSalt)
}

// TLMessageContainer represents msg_container from TL schema
type TLMessageContainer struct {
	Messages []*TLMessage // messages: vector<%Message>
}

func (s *TLMessageContainer) Cmd() uint32 {
	return TagMsgContainer
}

func (s *TLMessageContainer) ReadBareFrom(r *tl.Reader) {
	s.Messages = make([]*TLMessage, r.ReadInt())
	for i := 0; i < len(s.Messages); i++ {
		s.Messages[i] = new(TLMessage)
		s.Messages[i].ReadBareFrom(r)
	}
}

func (s *TLMessageContainer) WriteBareTo(w *tl.Writer) {
	w.WriteInt(len(s.Messages))
	for i := 0; i < len(s.Messages); i++ {
		s.Messages[i].WriteBareTo(w)
	}
}

// TLMessage represents message from TL schema
type TLMessage struct {
	MsgId uint64    // msg_id: long
	Seqno int       // seqno: int
	Bytes int       // bytes: int
	Body  tl.Object // body: Object
}

func (s *TLMessage) Cmd() uint32 {
	return TagMessage
}

func (s *TLMessage) ReadBareFrom(r *tl.Reader) {
	s.MsgId = r.ReadUint64()
	s.Seqno = r.ReadInt()
	s.Bytes = r.ReadInt()
	s.Body = ReadBoxedObjectFrom(r)
}

func (s *TLMessage) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.MsgId)
	w.WriteInt(s.Seqno)
	w.WriteInt(s.Bytes)
	w.WriteCmd(s.Body.Cmd())
	s.Body.WriteBareTo(w)
}

// TLMessageCopy represents msg_copy from TL schema
type TLMessageCopy struct {
	OrigMessage *TLMessage // orig_message: Message
}

func (s *TLMessageCopy) Cmd() uint32 {
	return TagMsgCopy
}

func (s *TLMessageCopy) ReadBareFrom(r *tl.Reader) {
	if cmd := r.ReadCmd(); cmd != TagMessage {
		r.Fail(errors.New("expected: message"))
	}
	s.OrigMessage = new(TLMessage)
	s.OrigMessage.ReadBareFrom(r)
}

func (s *TLMessageCopy) WriteBareTo(w *tl.Writer) {
	w.WriteCmd(TagMessage)
	s.OrigMessage.WriteBareTo(w)
}

// TLMsgsAck represents msgs_ack from TL schema
type TLMsgsAck struct {
	MsgIds []uint64 // msg_ids: Vector<long>
}

func (s *TLMsgsAck) Cmd() uint32 {
	return TagMsgsAck
}

func (s *TLMsgsAck) ReadBareFrom(r *tl.Reader) {
	if cmd := r.ReadCmd(); cmd != TagVector {
		r.Fail(errors.New("expected: vector"))
	}
	s.MsgIds = make([]uint64, r.ReadInt())
	for i := 0; i < len(s.MsgIds); i++ {
		s.MsgIds[i] = r.ReadUint64()
	}
}

func (s *TLMsgsAck) WriteBareTo(w *tl.Writer) {
	w.WriteCmd(TagVector)
	w.WriteInt(len(s.MsgIds))
	for i := 0; i < len(s.MsgIds); i++ {
		w.WriteUint64(s.MsgIds[i])
	}
}

// TLMsgResendReq represents msg_resend_req from TL schema
type TLMsgResendReq struct {
	MsgIds []uint64 // msg_ids: Vector<long>
}

func (s *TLMsgResendReq) Cmd() uint32 {
	return TagMsgResendReq
}

func (s *TLMsgResendReq) ReadBareFrom(r *tl.Reader) {
	if cmd := r.ReadCmd(); cmd != TagVector {
		r.Fail(errors.New("expected: vector"))
	}
	s.MsgIds = make([]uint64, r.ReadInt())
	for i := 0; i < len(s.MsgIds); i++ {
		s.MsgIds[i] = r.ReadUint64()
	}
}

func (s *TLMsgResendReq) WriteBareTo(w *tl.Writer) {
	w.WriteCmd(TagVector)
	w.WriteInt(len(s.MsgIds))
	for i := 0; i < len(s.MsgIds); i++ {
		w.WriteUint64(s.MsgIds[i])
	}
}

// TLMsgsStateReq represents msgs_state_req from TL schema
type TLMsgsStateReq struct {
	MsgIds []uint64 // msg_ids: Vector<long>
}

func (s *TLMsgsStateReq) Cmd() uint32 {
	return TagMsgsStateReq
}

func (s *TLMsgsStateReq) ReadBareFrom(r *tl.Reader) {
	if cmd := r.ReadCmd(); cmd != TagVector {
		r.Fail(errors.New("expected: vector"))
	}
	s.MsgIds = make([]uint64, r.ReadInt())
	for i := 0; i < len(s.MsgIds); i++ {
		s.MsgIds[i] = r.ReadUint64()
	}
}

func (s *TLMsgsStateReq) WriteBareTo(w *tl.Writer) {
	w.WriteCmd(TagVector)
	w.WriteInt(len(s.MsgIds))
	for i := 0; i < len(s.MsgIds); i++ {
		w.WriteUint64(s.MsgIds[i])
	}
}

// TLMsgsStateInfo represents msgs_state_info from TL schema
type TLMsgsStateInfo struct {
	ReqMsgId uint64 // req_msg_id: long
	Info     []byte // info: bytes
}

func (s *TLMsgsStateInfo) Cmd() uint32 {
	return TagMsgsStateInfo
}

func (s *TLMsgsStateInfo) ReadBareFrom(r *tl.Reader) {
	s.ReqMsgId = r.ReadUint64()
	s.Info = r.ReadBlob()
}

func (s *TLMsgsStateInfo) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.ReqMsgId)
	w.WriteBlob(s.Info)
}

// TLMsgsAllInfo represents msgs_all_info from TL schema
type TLMsgsAllInfo struct {
	MsgIds []uint64 // msg_ids: Vector<long>
	Info   []byte   // info: bytes
}

func (s *TLMsgsAllInfo) Cmd() uint32 {
	return TagMsgsAllInfo
}

func (s *TLMsgsAllInfo) ReadBareFrom(r *tl.Reader) {
	if cmd := r.ReadCmd(); cmd != TagVector {
		r.Fail(errors.New("expected: vector"))
	}
	s.MsgIds = make([]uint64, r.ReadInt())
	for i := 0; i < len(s.MsgIds); i++ {
		s.MsgIds[i] = r.ReadUint64()
	}
	s.Info = r.ReadBlob()
}

func (s *TLMsgsAllInfo) WriteBareTo(w *tl.Writer) {
	w.WriteCmd(TagVector)
	w.WriteInt(len(s.MsgIds))
	for i := 0; i < len(s.MsgIds); i++ {
		w.WriteUint64(s.MsgIds[i])
	}
	w.WriteBlob(s.Info)
}

// TLReqPq represents req_pq from TL schema
type TLReqPq struct {
	Nonce [16]byte // nonce: int128
}

func (s *TLReqPq) Cmd() uint32 {
	return TagReqPq
}

func (s *TLReqPq) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
}

func (s *TLReqPq) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
}

// TLReqDHParams represents req_DH_params from TL schema
type TLReqDHParams struct {
	Nonce                [16]byte // nonce: int128
	ServerNonce          [16]byte // server_nonce: int128
	P                    []byte   // p: bytes
	Q                    []byte   // q: bytes
	PublicKeyFingerprint uint64   // public_key_fingerprint: long
	EncryptedData        []byte   // encrypted_data: bytes
}

func (s *TLReqDHParams) Cmd() uint32 {
	return TagReqDHParams
}

func (s *TLReqDHParams) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	s.P = r.ReadBlob()
	s.Q = r.ReadBlob()
	s.PublicKeyFingerprint = r.ReadUint64()
	s.EncryptedData = r.ReadBlob()
}

func (s *TLReqDHParams) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.WriteBlob(s.P)
	w.WriteBlob(s.Q)
	w.WriteUint64(s.PublicKeyFingerprint)
	w.WriteBlob(s.EncryptedData)
}

// TLSetClientDHParams represents set_client_DH_params from TL schema
type TLSetClientDHParams struct {
	Nonce         [16]byte // nonce: int128
	ServerNonce   [16]byte // server_nonce: int128
	EncryptedData []byte   // encrypted_data: bytes
}

func (s *TLSetClientDHParams) Cmd() uint32 {
	return TagSetClientDHParams
}

func (s *TLSetClientDHParams) ReadBareFrom(r *tl.Reader) {
	r.ReadUint128(s.Nonce[:])
	r.ReadUint128(s.ServerNonce[:])
	s.EncryptedData = r.ReadBlob()
}

func (s *TLSetClientDHParams) WriteBareTo(w *tl.Writer) {
	w.WriteUint128(s.Nonce[:])
	w.WriteUint128(s.ServerNonce[:])
	w.WriteBlob(s.EncryptedData)
}

// TLRpcDropAnswer represents rpc_drop_answer from TL schema
type TLRpcDropAnswer struct {
	ReqMsgId uint64 // req_msg_id: long
}

func (s *TLRpcDropAnswer) Cmd() uint32 {
	return TagRpcDropAnswer
}

func (s *TLRpcDropAnswer) ReadBareFrom(r *tl.Reader) {
	s.ReqMsgId = r.ReadUint64()
}

func (s *TLRpcDropAnswer) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.ReqMsgId)
}

// TLGetFutureSalts represents get_future_salts from TL schema
type TLGetFutureSalts struct {
	Num int // num: int
}

func (s *TLGetFutureSalts) Cmd() uint32 {
	return TagGetFutureSalts
}

func (s *TLGetFutureSalts) ReadBareFrom(r *tl.Reader) {
	s.Num = r.ReadInt()
}

func (s *TLGetFutureSalts) WriteBareTo(w *tl.Writer) {
	w.WriteInt(s.Num)
}

// TLPing represents ping from TL schema
type TLPing struct {
	PingId uint64 // ping_id: long
}

func (s *TLPing) Cmd() uint32 {
	return TagPing
}

func (s *TLPing) ReadBareFrom(r *tl.Reader) {
	s.PingId = r.ReadUint64()
}

func (s *TLPing) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.PingId)
}

// TLPingDelayDisconnect represents ping_delay_disconnect from TL schema
type TLPingDelayDisconnect struct {
	PingId          uint64 // ping_id: long
	DisconnectDelay int    // disconnect_delay: int
}

func (s *TLPingDelayDisconnect) Cmd() uint32 {
	return TagPingDelayDisconnect
}

func (s *TLPingDelayDisconnect) ReadBareFrom(r *tl.Reader) {
	s.PingId = r.ReadUint64()
	s.DisconnectDelay = r.ReadInt()
}

func (s *TLPingDelayDisconnect) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.PingId)
	w.WriteInt(s.DisconnectDelay)
}

// TLDestroySession represents destroy_session from TL schema
type TLDestroySession struct {
	SessionId uint64 // session_id: long
}

func (s *TLDestroySession) Cmd() uint32 {
	return TagDestroySession
}

func (s *TLDestroySession) ReadBareFrom(r *tl.Reader) {
	s.SessionId = r.ReadUint64()
}

func (s *TLDestroySession) WriteBareTo(w *tl.Writer) {
	w.WriteUint64(s.SessionId)
}

// TLHttpWait represents http_wait from TL schema
type TLHttpWait struct {
	MaxDelay  int // max_delay: int
	WaitAfter int // wait_after: int
	MaxWait   int // max_wait: int
}

func (s *TLHttpWait) Cmd() uint32 {
	return TagHttpWait
}

func (s *TLHttpWait) ReadBareFrom(r *tl.Reader) {
	s.MaxDelay = r.ReadInt()
	s.WaitAfter = r.ReadInt()
	s.MaxWait = r.ReadInt()
}

func (s *TLHttpWait) WriteBareTo(w *tl.Writer) {
	w.WriteInt(s.MaxDelay)
	w.WriteInt(s.WaitAfter)
	w.WriteInt(s.MaxWait)
}

func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
	cmd := r.ReadCmd()
	switch cmd {
	case TagResPQ:
		s := new(TLResPQ)
		s.ReadBareFrom(r)
		return s
	case TagPQInnerData:
		s := new(TLPQInnerData)
		s.ReadBareFrom(r)
		return s
	case TagServerDHInnerData:
		s := new(TLServerDHInnerData)
		s.ReadBareFrom(r)
		return s
	case TagClientDHInnerData:
		s := new(TLClientDHInnerData)
		s.ReadBareFrom(r)
		return s
	case TagRpcResult:
		s := new(TLRpcResult)
		s.ReadBareFrom(r)
		return s
	case TagRpcError:
		s := new(TLRpcError)
		s.ReadBareFrom(r)
		return s
	case TagFutureSalt:
		s := new(TLFutureSalt)
		s.ReadBareFrom(r)
		return s
	case TagFutureSalts:
		s := new(TLFutureSalts)
		s.ReadBareFrom(r)
		return s
	case TagPong:
		s := new(TLPong)
		s.ReadBareFrom(r)
		return s
	case TagNewSessionCreated:
		s := new(TLNewSession)
		s.ReadBareFrom(r)
		return s
	case TagMsgContainer:
		s := new(TLMessageContainer)
		s.ReadBareFrom(r)
		return s
	case TagMessage:
		s := new(TLMessage)
		s.ReadBareFrom(r)
		return s
	case TagMsgCopy:
		s := new(TLMessageCopy)
		s.ReadBareFrom(r)
		return s
	case TagMsgsAck:
		s := new(TLMsgsAck)
		s.ReadBareFrom(r)
		return s
	case TagMsgResendReq:
		s := new(TLMsgResendReq)
		s.ReadBareFrom(r)
		return s
	case TagMsgsStateReq:
		s := new(TLMsgsStateReq)
		s.ReadBareFrom(r)
		return s
	case TagMsgsStateInfo:
		s := new(TLMsgsStateInfo)
		s.ReadBareFrom(r)
		return s
	case TagMsgsAllInfo:
		s := new(TLMsgsAllInfo)
		s.ReadBareFrom(r)
		return s
	case TagReqPq:
		s := new(TLReqPq)
		s.ReadBareFrom(r)
		return s
	case TagReqDHParams:
		s := new(TLReqDHParams)
		s.ReadBareFrom(r)
		return s
	case TagSetClientDHParams:
		s := new(TLSetClientDHParams)
		s.ReadBareFrom(r)
		return s
	case TagRpcDropAnswer:
		s := new(TLRpcDropAnswer)
		s.ReadBareFrom(r)
		return s
	case TagGetFutureSalts:
		s := new(TLGetFutureSalts)
		s.ReadBareFrom(r)
		return s
	case TagPing:
		s := new(TLPing)
		s.ReadBareFrom(r)
		return s
	case TagPingDelayDisconnect:
		s := new(TLPingDelayDisconnect)
		s.ReadBareFrom(r)
		return s
	case TagDestroySession:
		s := new(TLDestroySession)
		s.ReadBareFrom(r)
		return s
	case TagHttpWait:
		s := new(TLHttpWait)
		s.ReadBareFrom(r)
		return s
	default:
		return nil
	}
}
