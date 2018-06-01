// Copyright 2017 The go-ethereum Authors
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
package eth

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
)

// TestMasternodeManager_BestMasternode
// Test function for choose BestMasternode
func TestMasternodeManager_BestMasternode(t *testing.T) {
	//// initial the parameter may needed during this test function
	manager := &MasternodeManager{
		networkId:   uint64(0),
		eventMux:    nil,
		txpool:      nil,
		blockchain:  nil,
		chainconfig: nil,
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		masternodes: &masternode.MasternodeSet{},
	}
	manager.is = NewInstantx()
	manager.winner = NewMasternodePayments(manager, big.NewInt(10))

	// init new hash
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	// init new account
	var account common.Address
	for i := range account {
		account[i] = byte(i)
	}
	// generate a new block data
	tx1 := types.NewTransaction(1, common.BytesToAddress([]byte{0x11}), big.NewInt(111), 1111, big.NewInt(111111), []byte{0x11, 0x11, 0x11})
	tx2 := types.NewTransaction(2, common.BytesToAddress([]byte{0x22}), big.NewInt(222), 2222, big.NewInt(222222), []byte{0x22, 0x22, 0x22})
	tx3 := types.NewTransaction(3, common.BytesToAddress([]byte{0x33}), big.NewInt(333), 333, big.NewInt(33333), []byte{0x33, 0x33, 0x33})
	txs := []*types.Transaction{tx1, tx2, tx3}
	block := types.NewBlock(&types.Header{Number: big.NewInt(31415926)}, txs, nil, nil)

	// generate a new masternode set and deploy the new nodeset
	ms := newMasternodeSet(true)

	//// begin to test
	testsbody := []struct {
		ms                  *masternode.MasternodeSet // input ms, MasternodeSet
		voteNum             *big.Int                  // 投票的区块
		err                 error                     // return type, error or nil
		numberofmasternodes uint32                    //  is numberofmasternodes==0 ,return the error
	}{
		{
			nil,
			big.NewInt(31415926),
			nil,
			1,
		},

		{
			ms,
			nil,
			errors.New("no masternode detected"),
			20,
		},
		{
			ms,
			nil,
			errors.New("The number of local masternodes is too less to obtain the best Masternode"),
			0,
		},
		{
			ms,
			nil,
			nil,
			20,
		},
	}
	number := big.NewInt(31415926)
	// show the test process
	for _, v := range testsbody {
		// first two line test
		if v.voteNum != nil {
			manager.winner.blocks[31415926] = NewMasternodeBlockPayees(number)
			masternodePayee := NewMasternodePayee(*(new(common.Address)), nil)
			manager.winner.blocks[31415926].payees = append(manager.winner.blocks[31415926].payees, masternodePayee)
		}
		if v.numberofmasternodes == 0 {
			v.ms = newMasternodeSet(false)
		}
		manager.masternodes = v.ms
		addr, err := manager.BestMasternode(block)
		if err != nil {
			if !strings.EqualFold(err.Error(), v.err.Error()) {
				t.Errorf("test failed %v", err)
			}
		}

		if addr != *(new(common.Address)) {
			t.Logf("winnerid is %v", addr.String())
		}
	}
}

func TestMasternodeManager_GetMasternodeScores(t *testing.T) {

}

func TestMasternodeManager_GetMasternodeRank(t *testing.T) {

}
