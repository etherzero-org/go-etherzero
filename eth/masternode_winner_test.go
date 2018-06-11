// Copyright 2018 The go-etherzero Authors
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
package eth

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/ethdb"
)

var (
	testdb, _   = ethdb.NewMemDatabase()
	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddress = crypto.PubkeyToAddress(testKey.PublicKey)
	genesis     = core.GenesisBlockForTesting(testdb, testAddress, big.NewInt(1000000000))
)

// 当一个区块到达时,需要进行本地的主节点投票处理
// TestMasternodePayments_ProcessBlock
// a masternode need voting for transaction when it is arrived
func TestMasternodePayments_ProcessBlock(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(10), ranksFn)
	tests := []struct {
		rank int
	}{
		{11},
		{1},
	}
	for _, v := range tests {
		manager.winner.active = returnNewActinveNode()
		manager.winner.active.PrivateKey = key0

		manager.winner.ProcessBlock(genesis, v.rank)
	}
}

func TestMasternodePayments_winners(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(10), ranksFn)
	number := big.NewInt(31415926)
	a := NewMasternodeBlockPayees(number)
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	a.AddVoteHash(hash)

	manager.winner.blocks[31415926] = NewMasternodeBlockPayees(number)
	//manager.winner.cachedBlockNumber = big.NewInt(31415926)
	fmt.Println("winners is ", manager.winner.winners())
}

// 获取指定区块的获胜主节点
// TestMasternodePayments_BlockWinner
// get the winner masternode
func TestMasternodePayments_BlockWinner(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(0), ranksFn)
	number := big.NewInt(31415926)
	manager.winner.blocks[31415926] = NewMasternodeBlockPayees(number)
	manager.winner.BlockWinner(big.NewInt(31415926))
}

// 对本地主节点或者是收到其它主节点发起的一笔有效投票进行转发
// TestMasternodePayments_PostVoteEvent
// Retransport a transaction when receive it directly or got from other masternodes
// TODO verify the transaction before retransport the transaction
func TestMasternodePayments_PostVoteEvent(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(0), ranksFn)
	manager.active = returnNewActinveNode()
	vote := masternode.NewMasternodePaymentVote(genesis.Number(), "", manager.active.Account)
	manager.winner.PostVoteEvent(vote)
}

// TestMasternodePayments_Vote
// when voting , the block number is not currently but the 100th blocks previous
// for the current blocks may be abandoned
func TestMasternodePayments_Vote(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(1), ranksFn)
	manager.active = returnNewActinveNode()
	vote := masternode.NewMasternodePaymentVote(genesis.Number(), "", manager.active.Account)
	manager.winner.Vote(vote, big.NewInt(1))
}

// 当区块已经超过了一次的数量时,需要对已经过时的区块投票进行清理
// TestMasternodePayments_Vote
// need to clear the posted votting when the blocks has already been over limited
func TestMasternodePayments_Clear(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(0), ranksFn)
	manager.active = returnNewActinveNode()
	manager.winner.Clear()
}

//当收到一笔其它主节点的投票时,由该方法进行处理
// TestMasternodePayments_Add
// when receiveing a voting from other masternodes
func TestMasternodePayments_Add(t *testing.T) {
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(0), ranksFn)
	manager.active = returnNewActinveNode()
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	manager.active = returnNewActinveNode()
	vote := masternode.NewMasternodePaymentVote(genesis.Number(), "", manager.active.Account)
	manager.winner.Add(hash, vote)
}

// 验证上一个区块的投票是否正确
// TestMasternodePayments_CheckPreviousBlockVotes
// verify an vote for a certain block is valid or not
//func TestMasternodePayments_CheckPreviousBlockVotes(t *testing.T) {
//	manager := returnMasternodeManager()
//	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
//		return manager.GetMasternodeRanks(height)
//	}
//	manager.winner = NewMasternodePayments(big.NewInt(0),ranksFn)
//	manager.active = returnNewActinveNode()
//	var hash common.Hash
//	for i := range hash {
//		hash[i] = byte(i)
//	}
//	manager.winner.CheckPreviousBlockVotes(hash)
//}
