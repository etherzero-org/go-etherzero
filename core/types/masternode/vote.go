// Copyright 2015 The go-ethereum Authors
// Copyright 2018 The go-etherzero Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package masternode

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/params"
)

const (
	SIGNATURES_REQUIRED = 6
	SIGNATURES_TOTAL    = 10
)

var (
	ErrInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	ErrECDSAVerification = errors.New("crypto/ecdsa: verification error")

	InstantSendKeepLock = big.NewInt(24)
)

type TxLockVote struct {
	txHash          common.Hash
	masternodeId    string
	sig             []byte
	ConfirmedHeight *big.Int
	createdTime     time.Time
	KeySize         int
}

func (tlv *TxLockVote) MasternodeId() string {
	return tlv.masternodeId
}

func NewTxLockVote(hash common.Hash, id string) *TxLockVote {
	return &TxLockVote{
		txHash:          hash,
		masternodeId:    id,
		createdTime:     time.Now(),
		ConfirmedHeight: big.NewInt(-1),
		KeySize:         32, //TODO
	}
}

func (v *TxLockVote) Hash() common.Hash {
	h := rlpHash(v)
	return h
}

func (tlv *TxLockVote) CheckSignature(pubkey, signature []byte) bool {
	return crypto.VerifySignature(pubkey, tlv.Hash().Bytes(), signature)

}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *TxLockVote) Verify(pubkey, hash, signature []byte) bool {
	return crypto.VerifySignature(pubkey,hash,signature)

}

// Implements the Sign method from SigningMethod
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *TxLockVote) Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {

	sig, err = crypto.Sign(hash[:], prv)
	if err != nil {
		return nil, err
	}
	return sig,nil

}

// Locks and votes expire nInstantSendKeepLock blocks after the block
// corresponding tx was included into.
func (self *TxLockVote) IsExpired(height *big.Int) bool {

	return (self.ConfirmedHeight.Cmp(big.NewInt(-1)) > 0) && (new(big.Int).Sub(height, self.ConfirmedHeight).Cmp(InstantSendKeepLock) > 0)
}

func (self *TxLockVote) IsFailed() bool {

	return uint64(time.Now().Sub(self.createdTime)) > params.InstantSendFailedTimeoutSeconds
}

func (self *TxLockVote) IsTimeOut() bool {
	return uint64(time.Now().Sub(self.createdTime)) > params.InstantSendLockTimeoutSeconds
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

// TODO:Verification work is handled in the MasternodeManager
func (tq *TxLockRequest) IsVeified() bool {

	return tq.tx.CheckNonce()
}

func (tq *TxLockRequest) Tx() *types.Transaction {
	return tq.tx
}

type TxLockCondidate struct {
	confirmedHeight *big.Int
	createdTime     time.Time
	TxLockRequest   *types.Transaction
	masternodeVotes map[string]*TxLockVote
	attacked        bool
}

func NewTxLockCondidate(request *types.Transaction) *TxLockCondidate {
	txLockCondidate := &TxLockCondidate{
		confirmedHeight: big.NewInt(-1),
		createdTime:     time.Now(),
		TxLockRequest:   request,
		masternodeVotes: make(map[string]*TxLockVote),
		attacked:        false,
	}
	return txLockCondidate
}

func (tc *TxLockCondidate) Hash() common.Hash {

	return tc.TxLockRequest.Hash()
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

// Locks and votes expire nInstantSendKeepLock blocks after the block
// corresponding tx was included into.
func (self *TxLockCondidate) IsExpired(height *big.Int) bool {

	return (self.confirmedHeight.Cmp(big.NewInt(-1)) > 0) && (new(big.Int).Sub(height, self.confirmedHeight).Cmp(InstantSendKeepLock) > 0)
}

func (self *TxLockCondidate) IsTimeout() bool {

	if uint64(time.Now().Sub(self.createdTime)) > params.InstantSendLockTimeoutSeconds {
		return false
	}
	return true
}

func (tc *TxLockCondidate) HasMasternodeVoted(id string) bool {

	return tc.masternodeVotes[id] != nil
}

func (tc *TxLockCondidate) MaxSignatures() int {

	return int(tc.TxLockRequest.Size()) * SIGNATURES_TOTAL

}
