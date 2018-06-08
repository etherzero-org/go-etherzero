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
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/consensus/ethash"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/core/vm"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/ethdb"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/params"
)

func genNewgspec() core.Genesis {
	return core.Genesis{
		Config: params.TestChainConfig,
		Alloc:  core.GenesisAlloc{addr1: {Balance: big.NewInt(10000000000000)}},
	}
}

func newBlockChain() *core.BlockChain {
	db, _ := ethdb.NewMemDatabase()
	fmt.Println("etherzero", genNewgspec().Config)
	//bc, _ := core.NewBlockChain(db, nil, oldcustomg.Config, ethash.NewFullFaker(), vm.Config{})
	cacheConfig := &core.CacheConfig{
		Disabled:      true, // Whether to disable trie write caching (archive node)
		TrieNodeLimit: 1,    // Memory limit (MB) at which to flush the current in-memory trie to disk
		TrieTimeLimit: time.Duration(10),
	}
	vmConfig := vm.Config{
		Debug:     true,
		EnableJit: true,
		ForceJit:  true,

		NoRecursion:             true,
		EnablePreimageRecording: true,
	}
	a := genNewgspec()
	core.SetupGenesisBlock(db, &a)
	chainman, _ := core.NewBlockChain(db, cacheConfig, a.Config, ethash.NewFaker(), vmConfig)
	hash, number := common.Hash{0: 0xff}, uint64(314)
	core.WriteCanonicalHash(db, hash, number)
	return chainman
}

// 获得交易的确认数,当一笔交易完成了投票锁定,则一次性返回五个确认
// TestInstantSend_GetConfirmations
// when a transaction is voted_locked ,return five confirmations once
func TestInstantSend_GetConfirmations(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}
	is := NewInstantx(newChainConfig(), newBlockChain())
	is.Active = returnNewActinveNode()

	is.GetConfirmations(txHash)
}

func transaction(nonce uint64, gaslimit uint64, key *ecdsa.PrivateKey) *types.Transaction {
	return pricedTransaction(nonce, gaslimit, big.NewInt(1), key)
}

func pricedTransaction(nonce uint64, gaslimit uint64, gasprice *big.Int, key *ecdsa.PrivateKey) *types.Transaction {
	tx, _ := types.SignTx(types.NewTransaction(nonce, common.Address{}, big.NewInt(100), gaslimit, gasprice, nil), types.HomesteadSigner{}, key)
	return tx
}

// TestInstantSend_ProcessTxLockRequest
// test for ProcessTxLockRequest
// lock the transaction then creare an instance when start an payable
func TestInstantSend_ProcessTxLockRequest(t *testing.T) {
	is := NewInstantx(newChainConfig(), newBlockChain())
	key, _ := crypto.GenerateKey()
	tx0 := transaction(0, 100000, key)
	fmt.Printf("key %v\n,tx0 %v", key, tx0)
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}

	hashTmp := tx0.Hash()
	//is.lockedTxs[txHash] = txlockcondidate.TxLockRequest()
	//is.lockedTxs[hashTmp] = request
	//is.all[hashTmp] = 1
	is.Candidates[hashTmp] = masternode.NewTxLockCondidate(newTestTransaction(testAccount, 0, 0))

	is.ProcessTxLockRequest(tx0)
}

// TestInstantSend_Vote
// test for Vote
// vote for the transaction in masternode
// ranking the top 10 masternodes from the masternodes
func TestInstantSend_Vote(t *testing.T) {

	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}

	can1 := masternode.NewTxLockCondidate(newTestTransaction(testAccount, 1, 0))
	hash1 := can1.Hash()
	can0 := masternode.NewTxLockCondidate(newTestTransaction(testAccount, 0, 0))
	hash0 := can0.Hash()
	is := NewInstantx(newChainConfig(), newBlockChain())
	tests := []struct {
		is         *InstantSend
		hash       common.Hash
		can        *masternode.TxLockCondidate
		acceptHash bool
		hasCan     bool
		isVoted    bool
	}{
		{is,
			txHash,
			nil,
			false,
			false,
			false,
		},
		{is,
			txHash,
			can0,
			false,
			false,
			false,
		},
		{is,
			hash0,
			can0,
			true,
			false,
			false,
		},
		{
			is,
			hash1,
			can1,
			true,
			true,
			true,
		},
	}

	for _, v := range tests {
		if v.can != nil {
			v.is.Candidates[v.hash] = v.can
		}
		if v.acceptHash && v.can != nil {
			v.is.accepted[v.hash] = newTestTransaction(testAccount, 0, 0)
		}
		if v.hasCan && v.isVoted {
			v.is.Active = returnNewActinveNode()
			v.is.Active.PrivateKey = testAccount
			v.is.Candidates[v.hash] = v.can
			v.is.Active.ID = fmt.Sprintf("%v", 0xc5d24601)
		}
		v.is.Vote(v.hash)
	}
}

func returnNewActinveNode() *masternode.ActiveMasternode {
	srvr := &p2p.Server{}
	mns := newMasternodeSet(true)
	return masternode.NewActiveMasternode(srvr, mns)
}

// 当收到一笔交易投票时,对该笔投票进行处理,会出现当投票先于交易到达主节点时需要进行Orphan处理
// process the vote when reciving a transaction for vote
// if the vote is earlier reached the masternode than its transaction ,
// Orphan processing is needed
func TestInstantSend_ProcessTxLockVote(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}

	is := NewInstantx(newChainConfig(), newBlockChain())
	is.Active = returnNewActinveNode()

	vote := masternode.NewTxLockVote(txHash, is.Active.ID)

	is.ProcessTxLockVote(vote)
}

// TestInstantSend_CreateTxLockCandidate
// test for CreateTxLockCandidate
// create candidate instance for vote
func TestInstantSend_CreateTxLockCandidate(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}
	is := NewInstantx(newChainConfig(), newBlockChain())
	request := newTestTransaction(testAccount, 1, 0)
	is.CreateTxLockCandidate(request)
}

// 投票转发,当新建一笔投票,收到一笔有效投票,都需要对该笔投票进行转发
// masternode need  to retransfer the voting when generating
// a new vote or receiving a valid vote
func TestInstantSend_PostVoteEvent(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}
	is := NewInstantx(newChainConfig(), newBlockChain())
	is.Active = returnNewActinveNode()

	vote := masternode.NewTxLockVote(txHash, is.Active.ID)
	is.PostVoteEvent(vote)

}

func newChainConfig() *params.ChainConfig {
	return &params.ChainConfig{
		ChainId: big.NewInt(0),
	}
}

// 当处理一笔交易投票时,需要判断该笔交易的投票是否有满足六个投票,如果满足则要进行该方法的调用,结束交易锁定
// TestInstantSend_TryToFinalizeLockCandidate
// when processing a transaction,if it's satisfied to have 6 votes,
// process the TryToFinalizeLockCandidate and finish the vote_locked
func TestInstantSend_TryToFinalizeLockCandidate(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}
	is := NewInstantx(newChainConfig(), newBlockChain())
	is.Active = returnNewActinveNode()
	can1 := masternode.NewTxLockCondidate(newTestTransaction(testAccount, 1, 0))

	is.TryToFinalizeLockCandidate(can1)
}

// 当一笔交易已经获得了足够的投票时,需要对发生冲突的投票进行处理,主要就是进行清理工作.
// when a transaction has got enough voting tickets
// process the conflict,which means CheckAndRemove
func TestInstantSend_CheckAndRemove(t *testing.T) {
	var txHash common.Hash
	for i := range txHash {
		txHash[i] = byte(i)
	}
	is := NewInstantx(newChainConfig(), newBlockChain())
	is.Active = returnNewActinveNode()

	is.CheckAndRemove()
}
