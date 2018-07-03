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
	"io"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/rlp"
)

// Vote represents an entire vote in the Etherzero blockchain.
type Vote struct {
	cycle      int64          `json:"cycle"      gencodec:"required"`
	account    common.Address `json:"account"    gencodec:"required"`
	masternode common.Address `json:"masternode" gencodec:"required"`
	sign       []byte         `json:"sign"       gencodec:"required"`
}

// NewVote return a no has sign vote
func NewVote(cycle int64, account common.Address, masternode common.Address) *Vote {
	vote := &Vote{
		cycle:      cycle,
		account:    account,
		masternode: masternode,
	}
	return vote
}

// Hash returns the vote hash , which is simply the keccak256 hash of its
// RLP encoding.
func (v *Vote) Hash() (h common.Hash) {
	return rlpHash(v)
}

// DecodeRLP decodes the Vote
func (v *Vote) DecodeRLP(s *rlp.Stream) error {
	var vt Vote
	if err := s.Decode(&vt); err != nil {
		return err
	}
	v.account, v.cycle, v.masternode, v.sign = vt.account, vt.cycle, vt.masternode, vt.sign

	return nil
}

// EncodeRLP serializes v into the Ethereum RLP vote format.
func (v *Vote) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, Vote{
		account:    v.account,
		cycle:      v.cycle,
		masternode: v.masternode,
		sign:       v.sign,
	})
}

func (v *Vote) Account() common.Address {
	return v.account
}

func (v *Vote) Masternode() common.Address {
	return v.masternode
}

func (v *Vote) Cycle() int64 {
	return v.cycle
}

func (v *Vote) Sign() []byte {
	return v.sign
}
func (v *Vote) SetSign(sign []byte) {
	v.sign = sign
}

// SignVote signs the transaction using the given signer and private key
func (vote *Vote) SignVote(prv *ecdsa.PrivateKey) (*Vote, error) {
	h := vote.Hash() //not sign
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	vote.sign = sig

	return vote, nil
}

func (v *Vote) Verify(hash, sig []byte, pub *ecdsa.PublicKey) bool {
	recoveredPub1, _ := crypto.Ecrecover(hash, sig)
	recoveredPubBytes := crypto.FromECDSAPub(pub)
	return bytes.Equal(recoveredPub1, recoveredPubBytes)
}
