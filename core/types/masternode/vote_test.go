// Copyright 2014 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-eth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.
package masternode

import (
	"fmt"
	"testing"

	"github.com/ethzero/go-ethzero/common"
)

func TestTxLockVote_IsValid(t *testing.T) {
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	txLockVote := NewTxLockVote(
		hash,
		fmt.Sprintf("0x65bc97ef01b35f86a45c319675a05699e0947743c59ed53d0a918fb215c5ee5f"),
	)
	fmt.Printf("timeout ret is %v", txLockVote.IsTimeOut())
}
