// Copyright 2018 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-etherzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.

// Package devote implements the proof-of-stake consensus engine.

package types

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto"
)

// Vote represents an entire vote in the Etherzero blockchain.
type Vote struct {
	Cycle      uint64         `json:"cycle"      gencodec:"required"`
	Account    common.Address `json:"account"    gencodec:"required"`
	Masternode string         `json:"masternode" gencodec:"required"`
	Sign       []byte         `json:"sign"       gencodec:"required"`
}

// NewVote return a no has sign vote
func NewVote(cycle uint64, account common.Address, masternode string) *Vote {
	return &Vote{
		Cycle:      cycle,
		Account:    account,
		Masternode: masternode,
	}
}

// Hash returns the vote hash , which is simply the keccak256 hash of its
// RLP encoding.
func (v *Vote) Hash() (h common.Hash) {
	return rlpHash(v)
}

func (v *Vote) NosigHash()(h common.Hash){
	vote:=&Vote{
		Cycle:      v.Cycle,
		Account:    v.Account,
		Masternode: v.Masternode,
	}
	return rlpHash(vote)
}
// SignVote signs the transaction using the given signer and private key
func (vote *Vote) SignVote(prv *ecdsa.PrivateKey) (*Vote, error) {
	h := vote.Hash() //not sign
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	vote.Sign = sig

	return vote, nil
}

func (v *Vote) Verify(hash, sig []byte, pub *ecdsa.PublicKey) bool {
	recoveredPub1, _ := crypto.Ecrecover(hash, sig)
	recoveredPubBytes := crypto.FromECDSAPub(pub)
	return bytes.Equal(recoveredPub1, recoveredPubBytes)
}
