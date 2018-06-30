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
	"io"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/rlp"
)

// Vote represents an entire vote in the Etherzero blockchain.
type Vote struct {
	voteId     common.Hash    `json:"voteId"     gencodec:"required`
	cycle      uint64         `json:"cycle"      gencodec:"required"`
	masternode common.Address `json:"masternode" gencodec:"required"`
	value      common.Address `json:"value"      gencodec:"required"`
}

// Hash returns the vote hash , which is simply the keccak256 hash of its
// RLP encoding.
func (v *Vote) Hash() (h common.Hash) {
	return rlpHash(v)
}

// DecodeRLP decodes the Ethereum
func (v *Vote) DecodeRLP(s *rlp.Stream) error {
	var vt Vote
	if err := s.Decode(&vt); err != nil {
		return err
	}
	v.masternode, v.voteId, v.cycle, v.value = vt.masternode, vt.voteId, vt.cycle, vt.value

	return nil
}

// EncodeRLP serializes v into the Ethereum RLP vote format.
func (v *Vote) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, Vote{
		masternode: v.masternode,
		voteId:     v.voteId,
		cycle:      v.cycle,
		value:      v.value,
	})
}
