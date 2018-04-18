package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"math/big"
	"time"
)

const (
	SIGNATURES_REQUIRED = 6
	SIGNATURES_TOTAL    = 10
)

var (
	ErrInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	ErrECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

//这个类是投票的辅助类，投票和创建侯选对象都需要用到
type TxLock struct {
	Txhash          common.Hash
	masternodeVotes map[discover.NodeID]*TxLockVote
	attacked        bool
}

func NewTxLock(txid common.Hash) *TxLock {

	txlock := &TxLock{
		Txhash:          txid,
		masternodeVotes: make(map[discover.NodeID]*TxLockVote),
		attacked:        false}
	return txlock
}

func (tl *TxLock) CountVotes() int {
	if tl.attacked {
		return 0
	} else {
		return len(tl.masternodeVotes)
	}
}

func (tl *TxLock) HasMasternodeVoted(id discover.NodeID) bool {

	return tl.masternodeVotes[id] != nil
}

func (tl *TxLock) IsReady() bool {

	return !tl.attacked && tl.CountVotes() >= SIGNATURES_REQUIRED
}

func (tl *TxLock) AddVote(vote *TxLockVote) bool {

	if tl.masternodeVotes[vote.MasternodeId()] == nil {
		tl.masternodeVotes[vote.MasternodeId()] = vote
		return true
	}
	return false
}

type TxLockVote struct {
	txHash          common.Hash
	masternodeId    discover.NodeID
	sig             []byte
	confirmedHeight int
	createdTime     time.Time
	txLocks         map[common.Hash]*TxLock
	KeySize         int
}

func (tlv *TxLockVote) MasternodeId() discover.NodeID {
	return tlv.masternodeId
}

func NewTxLockVote(hash common.Hash, id discover.NodeID) *TxLockVote {

	tv := &TxLockVote{
		txHash:          hash,
		masternodeId:    id,
		createdTime:     time.Now(),
		confirmedHeight: -1,
		KeySize:         256,
	}
	return tv
}

func (tlv *TxLockVote) Hash() common.Hash {

	tlvHash := rlpHash([]interface{}{
		tlv.txHash,
		tlv.masternodeId,
	})
	return tlvHash
}

func (tlv *TxLockVote) CheckSignature(pubkey, signature []byte) bool {
	return crypto.VerifySignature(pubkey, tlv.Hash().Bytes(), signature)
}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *TxLockVote) Verify(sighash []byte, signature string, key interface{}) error {

	// Get the key
	var ecdsaKey *ecdsa.PublicKey
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		ecdsaKey = k
	default:
		return ErrInvalidKeyType
	}

	r := big.NewInt(0).SetBytes(sighash[:m.KeySize])
	s := big.NewInt(0).SetBytes(sighash[m.KeySize:])

	// Verify the signature
	if verifystatus := ecdsa.Verify(ecdsaKey, sighash, r, s); verifystatus == true {
		return nil
	} else {
		return ErrECDSAVerification
	}
}

// Implements the Sign method from SigningMethod
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *TxLockVote) Sign(signingString common.Hash, key interface{}) (string, error) {
	// Get the key
	var ecdsaKey *ecdsa.PrivateKey
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		ecdsaKey = k
	default:
		return "", ErrInvalidKeyType
	}
	// Sign the string and return r, s
	if r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, signingString[:]); err == nil {
		curveBits := ecdsaKey.Curve.Params().BitSize
		keyBytes := curveBits / 8
		if curveBits%8 > 0 {
			keyBytes += 1
		}

		// We serialize the outpus (r and s) into big-endian byte arrays and pad
		// them with zeros on the left to make sure the sizes work out. Both arrays
		// must be keyBytes long, and the output must be 2*keyBytes long.
		rBytes := r.Bytes()
		rBytesPadded := make([]byte, keyBytes)
		copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

		sBytes := s.Bytes()
		sBytesPadded := make([]byte, keyBytes)
		copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

		out := append(rBytesPadded, sBytesPadded...)

		return string(out[:]), nil
	} else {
		return "", err
	}
}

//主要目的是为了获取投票对象对应的交易，需要相关的内容以及投票的相应规则参数
type TxLockRequest struct {
	tx *Transaction
}

func (tq *TxLockRequest) Hash() common.Hash {
	return tq.tx.Hash()
}

func (tq *TxLockRequest) MaxSignatures() int {
	return int(tq.tx.Size()) * SIGNATURES_TOTAL
}

func (tq *TxLockRequest) IsValid() bool {

	return tq.tx.CheckNonce()
}

func (tq *TxLockRequest) Tx() *Transaction {
	return tq.tx
}

type TxLockCondidate struct {
	confirmedHeight int
	createdTime     time.Time
	txLockRequest   *TxLockRequest
	txLock          *TxLock // TxLockRequests by tx hash

}

func (tc *TxLockCondidate) TxLock() *TxLock {
	return tc.txLock
}

func (tc *TxLockCondidate) TxLockRequest() *TxLockRequest {
	return tc.txLockRequest
}

func NewTxLockCondidata(request *TxLockRequest) TxLockCondidate {

	txlockcondidata := TxLockCondidate{confirmedHeight: -1, createdTime: time.Now(), txLockRequest: request}
	return txlockcondidata

}

func (tc *TxLockCondidate) HasMasternodeVoted(hash common.Hash, id discover.NodeID) bool {

	return tc.txLock.HasMasternodeVoted(id)
}

func (tc *TxLockCondidate) Hash() common.Hash {

	return tc.txLockRequest.Hash()
}

func (tc *TxLockCondidate) AddTxLock(txid common.Hash) {
	txlock := NewTxLock(txid)
	tc.txLock = txlock
}

func (tc *TxLockCondidate) AddVote(vote *TxLockVote) bool {
	txlock := tc.txLock
	return txlock.AddVote(vote)
}

func (tc *TxLockCondidate) IsAllTxReady() bool {
	if tc.txLock == nil {
		return false
	}
	return true
}

func (tc *TxLockCondidate) CountVotes() int {
	return tc.txLock.CountVotes()
}
