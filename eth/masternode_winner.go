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

package eth

import (
	"errors"
	"math/big"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/core"
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

// Masternode Payments Class
// Keeps track of who should get paid for which blocks
type MasternodePayments struct {
	cachedBlockNumber *big.Int // Keep track of current block height
	minBlocksToStore  *big.Int
	storageCoeff      *big.Int //masternode count times nStorageCoeff payments blocks should be stored ...
	manager           *MasternodeManager

	votes      map[common.Hash]*masternode.MasternodePaymentVote
	blocks     map[uint64]*MasternodeBlockPayees
	lastVote   map[common.Hash]*big.Int
	didNotVote map[common.Hash]*big.Int
}

func NewMasternodePayments(manager *MasternodeManager, number *big.Int) *MasternodePayments {
	payments := &MasternodePayments{
		cachedBlockNumber: number,
		minBlocksToStore:  big.NewInt(1),
		storageCoeff:      big.NewInt(1),
		manager:           manager,
		votes:             make(map[common.Hash]*masternode.MasternodePaymentVote),
		blocks:            make(map[uint64]*MasternodeBlockPayees),
		lastVote:          make(map[common.Hash]*big.Int),
		didNotVote:        make(map[common.Hash]*big.Int),
	}
	return payments
}

//hash is blockHash,(!GetBlockHash(blockHash, vote.nBlockHeight - 101))
func (mp *MasternodePayments) Add(hash common.Hash, vote *masternode.MasternodePaymentVote) bool {

	if mp.Has(hash) {
		return false
	}
	mp.votes[hash] = vote

	if payee := mp.blocks[vote.Number.Uint64()]; payee == nil {
		blockPayees := NewMasternodeBlockPayees(vote.Number)
		blockPayees.Add(vote)
	} else {
		mp.blocks[vote.Number.Uint64()].Add(vote)
	}

	return true
}

func (mp *MasternodePayments) VoteCount() int {
	return len(mp.votes)
}

func (mp *MasternodePayments) BlockCount() int {
	return len(mp.blocks)
}

func (mp *MasternodePayments) Has(hash common.Hash) bool {

	if vote := mp.votes[hash]; vote != nil {
		return vote.IsVerified()
	}
	return false
}

func (mp *MasternodePayments) Clear() {
	mp.blocks = make(map[uint64]*MasternodeBlockPayees)
	mp.votes = make(map[common.Hash]*masternode.MasternodePaymentVote)

}

func (mp *MasternodePayments) ProcessBlock(block *types.Block) bool {

	rank, ok := mp.manager.GetMasternodeRank(mp.manager.active.ID)

	if ok {
		log.Info("ProcessBlock -- Unknown Masternode")
		return false
	}
	if rank > MNPAYMENTS_SIGNATURES_TOTAL {
		log.Info("Masternode not in the top ", MNPAYMENTS_SIGNATURES_TOTAL, "( ", rank, ")")
		return false
	}
	// LOCATE THE NEXT MASTERNODE WHICH SHOULD BE PAID
	log.Info("ProcessBlock -- Start: nBlockHeight=", block.String(), " masternode=", mp.manager.active.ID)

	info, err := mp.manager.GetNextMasternodeInQueueForPayment(block.Hash())
	if err != nil {
		log.Info("ERROR: Failed to find masternode to pay", err)
		return false
	}

	vote := masternode.NewMasternodePaymentVote(block.Number(), info)
	mp.Add(block.Hash(), vote)

	return true
}

//Handle the voting of other masternodes
func (m *MasternodePayments) Vote(vote *masternode.MasternodePaymentVote) bool {

	if m.votes[vote.Hash()] != nil {
		log.Trace("ERROR:Avoid processing same vote multiple times", "hash=", vote.Hash().String(), " , Height:", vote.Number.String())
		return false
	}

	m.votes[vote.Hash()] = vote
	// but first mark vote as non-verified,
	// AddPaymentVote() below should take care of it if vote is actually ok

	//vote out of range
	firstBlock := m.cachedBlockNumber.Sub(m.cachedBlockNumber, m.StorageLimit())
	if vote.Number.Cmp(firstBlock) > 0 || vote.Number.Cmp(m.cachedBlockNumber.Add(m.cachedBlockNumber, big.NewInt(20))) > 0 {
		log.Trace("ERROR:vote out of range: ", "FirstBlock=", firstBlock.String(), ", BlockHeight=", vote.Number, " CacheHeight=", m.cachedBlockNumber.String())
		return false
	}

	if !vote.IsVerified() {
		log.Trace("ERROR: invalid message, error:")
		return false
	}

	//canvote

	//checkSignature
	//if vote.CheckSignature(vote.masternode.MasternodeInfo().ID)

	log.Info("masternode_winner vote: ", "address:", vote.Masternode.Account.String(), "blockHeight:", vote.Number, "cacheHeight:", m.cachedBlockNumber.String(), "Hash:", vote.Hash().String())

	if m.Add(vote.Hash(), vote) {
		//Relay

	}

	return true

}

func (m *MasternodePayments) StorageLimit() *big.Int {

	count := m.manager.masternodes.Len()
	size := big.NewInt(1).Mul(m.storageCoeff, big.NewInt(int64(count)))

	if size.Cmp(m.minBlocksToStore) > 0 {
		return size
	}
	return m.minBlocksToStore
}

type MasternodePayee struct {
	account common.Address
	votes   []*masternode.MasternodePaymentVote
}

func NewMasternodePayee(address common.Address, vote *masternode.MasternodePaymentVote) *MasternodePayee {

	mp := &MasternodePayee{
		account: address,
	}
	mp.votes = append(mp.votes, vote)
	return mp
}

func (mp *MasternodePayee) Add(vote *masternode.MasternodePaymentVote) {

	mp.votes = append(mp.votes, vote)
}

func (mp *MasternodePayee) Count() int {
	return len(mp.votes)
}

func (mp *MasternodePayee) Votes() []*masternode.MasternodePaymentVote {
	return mp.votes
}

type MasternodeBlockPayees struct {
	number *big.Int //blockHeight
	payees []*MasternodePayee
	//payees *set.Set
}

func NewMasternodeBlockPayees(number *big.Int) *MasternodeBlockPayees {

	payee := &MasternodeBlockPayees{
		number: number,
	}
	return payee
}

//vote
func (mbp *MasternodeBlockPayees) Add(vote *masternode.MasternodePaymentVote) {

	//When the masternode has been voted
	//info := vote.masternode.MasternodeInfo()
	for _, mp := range mbp.payees {
		if mp.account == vote.Masternode.Account {
			mp.Add(vote)
			return
		}
	}
	payee := NewMasternodePayee(vote.Masternode.Account, vote)
	mbp.payees = append(mbp.payees, payee)

}

//select the Masternode that has been voted the most
func (mbp *MasternodeBlockPayees) Best() (common.Address, bool) {

	if len(mbp.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	votes := -1
	hash := common.Address{}

	for _, payee := range mbp.payees {
		if votes < payee.Count() {
			hash = payee.account
			votes = payee.Count()
		}
	}
	return hash, votes > -1
}

//Used to record the last winning block of the masternode. At least 2 votes need to be satisfied
// Has(2,masternode.account)
func (mbp *MasternodeBlockPayees) Has(num int, address common.Address) bool {
	if len(mbp.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	for _, payee := range mbp.payees {
		if payee.Count() >= num && payee.account == address {
			return true
		}
	}
	return false
}
