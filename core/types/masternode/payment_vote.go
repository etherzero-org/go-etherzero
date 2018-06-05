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
	"github.com/ethzero/go-ethzero/crypto"
)

const (
	MNPAYMENTS_SIGNATURES_REQUIRED = 6
	MNPAYMENTS_SIGNATURES_TOTAL    = 10

	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_1 = 70206
	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_2 = 70208
)

var (
	errInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	errECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

// vote for the winning payment
type MasternodePaymentVote struct {
	Number            *big.Int //blockHeight
	MasternodeId      string
	MasternodeAccount common.Address
	createTime        time.Time
	Sig               []byte
}

//Voted block number,activeMasternode
func NewMasternodePaymentVote(blockHeight *big.Int, id string, account common.Address) *MasternodePaymentVote {

	vote := MasternodePaymentVote{
		Number:            blockHeight,
		MasternodeId:      id,
		MasternodeAccount: account,
		createTime:        time.Now(),
	}

	return &vote
}

func (pv *MasternodePaymentVote) Hash() common.Hash {

	return rlpHash([]interface{}{
		pv.Number,
		pv.MasternodeId,
		pv.MasternodeAccount,
		pv.createTime,
	})
}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *MasternodePaymentVote) Verify(pubkey, hash, signature []byte) bool {

	return crypto.VerifySignature(pubkey, hash, signature)
}

// Implements the Sign method from SigningMethod
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *MasternodePaymentVote) Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {

	sig, err = crypto.Sign(hash[:], prv)
	if err != nil {
		return nil, err
	}
	return sig, nil
}
