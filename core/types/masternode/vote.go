package masternode

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/core/types"
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

type TxLockVote struct {
	txHash          common.Hash
	masternodeId    string
	sig             []byte
	confirmedHeight int
	createdTime     time.Time
	KeySize         int
}

func (tlv *TxLockVote) MasternodeId() string {
	return tlv.masternodeId
}

func NewTxLockVote(hash common.Hash, id string) *TxLockVote {

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

type TxLockRequest struct {
	tx *types.Transaction
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

func (tq *TxLockRequest) Tx() *types.Transaction {
	return tq.tx
}

type TxLockCondidate struct {
	confirmedHeight int
	createdTime     time.Time
	txLockRequest   *types.Transaction
	masternodeVotes map[string]*TxLockVote
	attacked        bool
}

func (tc *TxLockCondidate) TxLockRequest() *types.Transaction {
	return tc.txLockRequest
}

func NewTxLockCondidate(request *types.Transaction) TxLockCondidate {

	txLockCondidate := TxLockCondidate{
		confirmedHeight: -1,
		createdTime:     time.Now(),
		txLockRequest:   request,
		masternodeVotes: make(map[string]*TxLockVote),
		attacked:        false,
	}

	return txLockCondidate
}

func (tc *TxLockCondidate) Hash() common.Hash {

	return tc.txLockRequest.Hash()
}

func (tc *TxLockCondidate) AddVote(vote *TxLockVote) bool {

	if node := tc.masternodeVotes[vote.MasternodeId()]; node == nil {
		tc.masternodeVotes[vote.MasternodeId()] = vote
		return true
	}
	return false
}

func (tc *TxLockCondidate) IsReady() bool {
	return !tc.attacked && tc.CountVotes() >= SIGNATURES_REQUIRED
}

func (tc *TxLockCondidate) CountVotes() int {
	if tc.attacked {
		return 0
	} else {
		return len(tc.masternodeVotes)
	}
}

func (tc *TxLockCondidate) HasMasternodeVoted(id string) bool {

	return tc.masternodeVotes[id] != nil
}

func (tc *TxLockCondidate) MaxSignatures() int {
	return int(tc.txLockRequest.Size()) * SIGNATURES_TOTAL
}
