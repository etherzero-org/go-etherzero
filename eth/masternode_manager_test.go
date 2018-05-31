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
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"

	"github.com/ethzero/go-ethzero/accounts/abi/bind"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/p2p/discover"
)

// uintID auxiliary function
// generate new discover.NodeID
func uintID(i uint32) discover.NodeID {
	var id discover.NodeID
	binary.BigEndian.PutUint32(id[:], i)
	return id
}

func newMasterNodeSet() *masternode.MasternodeSet {
	backend := newTestBackend()

	addr0, err := deploy(key0, big.NewInt(0), backend)
	if err != nil {
		fmt.Errorf("deploy contract: expected no error, got %v", err)
	}

	contract, err1 := contract.NewContract(addr0, backend)
	if err1 != nil {
		fmt.Errorf("expected no error, got %v", err1)
	}

	var (
		id1  [32]byte
		id2  [32]byte
		misc [32]byte
	)

	addr := net.TCPAddr{net.ParseIP("127.0.0.88"), 21212, ""}

	misc[0] = 1
	copy(misc[1:17], addr.IP)
	binary.BigEndian.PutUint16(misc[17:19], uint16(addr.Port))

	nodeID, _ := discover.HexID("0x2cb5063f3fe98370023ecbf05a5f61534ac724e8bfc52e72e2f33dc57e6328a15bb6c09ce296c546a35c1469b6d2a013d6fc1f2a123ee867764e8c5e184e46ce")

	copy(id1[:], nodeID[:32])
	copy(id2[:], nodeID[32:64])

	transactOpts := bind.NewKeyedTransactor(key0)
	val, _ := new(big.Int).SetString("20000000000000000000", 10)
	transactOpts.Value = val

	tx, err := contract.Register(transactOpts, id1, id2, misc)
	fmt.Println("Register", tx, err)
	backend.Commit()

	masternodes, _ := masternode.NewMasternodeSet(contract)
	return masternodes
}

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

	// generate a new masternode set and deploy the new nodeset
	ms := newMasterNodeSet()
	nodenum := int64(10)
	for i := int64(0); i < nodenum; i++ {
		node := discover.NewNode(discover.MustHexID("0x84d9d65c4552b5eb43d5ad55a2ee3f56c6cbc1c64a5c8d659f51fcd51bace24351232b8d7821617d2b29b54b81cdefb9b3e9c37d7fd5f63270bcc9e1a6f6a439"), net.IP{127, 0, 55, byte(234 + i)}, uint16(3333+i), uint16(4444+i))
		singleNode := &masternode.Masternode{
			ID:                         fmt.Sprintf("%v", i+10),
			Height:                     big.NewInt(10),
			Node:                       node,
			Account:                    account,
			OriginBlock:                0,
			State:                      masternode.MasternodeEnable,
			ProtocolVersion:            64,
			CollateralMinConfBlockHash: hash,
		}
		// register new node in the register code
		ms.Register(singleNode)
	}
	//// begin to test
	testsbody := []struct {
		ms  *masternode.MasternodeSet // input ms, MasternodeSet
		err error                     // return type, error or nil
	}{
		{
			nil,
			errors.New("no masternode detected"),
		}, {
			ms,
			nil,
		},
	}

	// show the test process
	for _, v := range testsbody {
		manager.masternodes = v.ms
		node, err := manager.BestMasternode(hash)
		if err != nil {
			if !strings.EqualFold(err.Error(), v.err.Error()) {
				t.Errorf("test failed %v", err)
			}
		}

		if node != nil {
			t.Logf("winnerid is %v", node.ID)
		}
	}
}

func TestMasternodeManager_GetMasternodeScores(t *testing.T) {

}

func TestMasternodeManager_GetMasternodeRank(t *testing.T) {

}
