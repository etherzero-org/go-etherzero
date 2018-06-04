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
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"sync"
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
	minBlocksToStore  *big.Int // ... but at least nMinBlocksToStore (payments blocks) dash default value:5000
	storageCoeff      *big.Int //masternode count times nStorageCoeff payments blocks should be stored ... default value:1.25

	votes      map[common.Hash]*masternode.MasternodePaymentVote
	blocks     map[uint64]*MasternodeBlockPayees
	didNotVote map[common.Hash]*big.Int
	scope      event.SubscriptionScope
	active     *masternode.ActiveMasternode
	lastVoted  map[common.Address]*big.Int // masternodeID <- height

	winnerFeed event.Feed
	mu sync.Mutex
}

func NewMasternodePayments(manager *MasternodeManager, number *big.Int) *MasternodePayments {

	payments := &MasternodePayments{
		cachedBlockNumber: number,
		minBlocksToStore:  big.NewInt(1),
		storageCoeff:      big.NewInt(1),
		votes:             make(map[common.Hash]*masternode.MasternodePaymentVote),
		blocks:            make(map[uint64]*MasternodeBlockPayees),
		lastVoted:         make(map[common.Address]*big.Int),
		didNotVote:        make(map[common.Hash]*big.Int),
	}
	return payments
}

//hash is blockHash,(!GetBlockHash(blockHash, vote.nBlockHeight - 101))
func (mp *MasternodePayments) Add(hash common.Hash, vote *masternode.MasternodePaymentVote) bool {

	if mp.HasVerifiedVote(hash) {
		return false
	}
	mp.votes[hash] = vote

	if payee := mp.blocks[vote.Number.Uint64()]; payee == nil {
		blockPayees := NewMasternodeBlockPayees(vote.Number)
		blockPayees.Add(vote)
		mp.blocks[vote.Number.Uint64()] = blockPayees
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

//vote hash
func (self *MasternodePayments) HasVerifiedVote(hash common.Hash) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	if vote := self.votes[hash]; vote != nil {
		return vote.IsVerified()
	}
	return false
}

func (self *MasternodePayments) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.blocks = make(map[uint64]*MasternodeBlockPayees)
	self.votes = make(map[common.Hash]*masternode.MasternodePaymentVote)

}

func (mp *MasternodePayments) ProcessBlock(block *types.Block, rank int) bool {

	if rank > MNPAYMENTS_SIGNATURES_TOTAL {
		log.Info("Masternode not in the top ", MNPAYMENTS_SIGNATURES_TOTAL, "( ", rank, ")")
		return false
	}
	// LOCATE THE NEXT MASTERNODE WHICH SHOULD BE PAID
	log.Info("ProcessBlock -- Start: nBlockHeight=", block.String(), " masternodeId=", mp.active.ID)

	vote := masternode.NewMasternodePaymentVote(block.Number(), mp.active.Account)
	// vote constructed sucessfully, let's store and relay it
	mp.Add(block.Hash(), vote)

	return true
}

//Handle the voting of other masternodes
func (self *MasternodePayments) Vote(vote *masternode.MasternodePaymentVote, storageLimit *big.Int) bool {

	if self.votes[vote.Hash()] != nil {
		log.Trace("ERROR:Avoid processing same vote multiple times", "hash=", vote.Hash().String(), " , Height:", vote.Number.String())
		return false
	}
	self.mu.Lock()
	defer self.mu.Unlock()

	self.votes[vote.Hash()] = vote
	// but first mark vote as non-verified,
	// AddPaymentVote() below should take care of it if vote is actually ok

	//vote out of range
	firstBlock := self.cachedBlockNumber.Sub(self.cachedBlockNumber, storageLimit)
	if vote.Number.Cmp(firstBlock) > 0 || vote.Number.Cmp(self.cachedBlockNumber.Add(self.cachedBlockNumber, big.NewInt(20))) > 0 {
		log.Trace("ERROR:vote out of range: ", "FirstBlock=", firstBlock.String(), ", BlockHeight=", vote.Number, " CacheHeight=", self.cachedBlockNumber.String())
		return false
	}

	if !vote.IsVerified() {
		log.Trace("ERROR: invalid message, error: Verified false")
		return false
	}

	//canvote
	if !self.CanVote(vote.Number, vote.MasternodeAccount) {
		log.Info("masternode already voted, masternode account:", vote.MasternodeAccount.String())
		return false
	}
	//checkSignature
	//if vote.CheckSignature(vote.masternode.MasternodeInfo().ID)

	log.Info("masternode_winner vote: ", "blockHeight:", vote.Number, "cacheHeight:", self.cachedBlockNumber.String(), "Hash:", vote.Hash().String())

	if self.Add(vote.Hash(), vote) {
		//Relay
		self.winnerFeed.Send(vote)
	}
	return true
}

// SubscribeWinnerVoteEvent registers a subscription of PaymentVoteEvent and
// starts sending event to the given channel.
func (self *MasternodePayments) SubscribeWinnerVoteEvent(ch chan<- core.PaymentVoteEvent) event.Subscription {
	return self.scope.Track(self.winnerFeed.Subscribe(ch))
}

func (is *MasternodePayments) PostVoteEvent(vote *masternode.MasternodePaymentVote) {

	is.winnerFeed.Send(core.PaymentVoteEvent{vote})
}

// Find an entry in the masternode list that is next to be paid
// height is blockNumber , address is masternode account
func (mp *MasternodePayments) BlockWinner(height *big.Int) (common.Address, bool) {

	if mp.blocks[height.Uint64()] != nil {
		return mp.blocks[height.Uint64()].Best()
	}
	return common.Address{}, false
}

//Detect whether the current masternode can vote
func (self *MasternodePayments) CanVote(height *big.Int, address common.Address) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.lastVoted[address] != nil && self.lastVoted[address].Cmp(height) == 0 {
		return false
	}
	self.lastVoted[address] = height
	return true
}

// Must be called at startup or init
func (self *MasternodePayments) CheckAndRemove(limit *big.Int){

	for hash,vote := range self.votes {
		if new(big.Int).Sub(self.cachedBlockNumber,vote.Number).Cmp(limit)>0 {
			log.Info("CheckAndRemove -- Removing old Masternode payment: BlockHeight=",vote.Number.String())
			delete(self.votes, hash)
			delete(self.blocks,vote.Number.Uint64())
		}
	}

}

func (self *MasternodePayments) CheckPreviousBlockVotes(preBlockHash common.Hash) bool{

	return true
}

type MasternodePayee struct {
	masternodeAccount common.Address
	votes             []*masternode.MasternodePaymentVote

	mu sync.Mutex
}

func NewMasternodePayee(account common.Address, vote *masternode.MasternodePaymentVote) *MasternodePayee {

	mp := &MasternodePayee{
		masternodeAccount: account,
	}
	mp.votes = append(mp.votes, vote)
	return mp
}

func (self *MasternodePayee) Add(vote *masternode.MasternodePaymentVote) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.votes = append(self.votes, vote)
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
	mu sync.Mutex
	//payees *set.Set
}

func NewMasternodeBlockPayees(number *big.Int) *MasternodeBlockPayees {

	payee := &MasternodeBlockPayees{
		number: number,
	}
	return payee
}

//vote
func (self *MasternodeBlockPayees) Add(vote *masternode.MasternodePaymentVote) {

	self.mu.Lock()
	defer self.mu.Unlock()
	//When the masternode has been voted
	//info := vote.masternode.MasternodeInfo()
	for _, mp := range self.payees {
		if mp.masternodeAccount == vote.MasternodeAccount {
			mp.Add(vote)
			return
		}
	}
	payee := NewMasternodePayee(vote.MasternodeAccount, vote)
	self.payees = append(self.payees, payee)

}

//select the Masternode that has been voted the most
func (self *MasternodeBlockPayees) Best() (common.Address, bool) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if len(self.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	var (
		votes             = -1
		masternodeAccount common.Address
	)
	for _, payee := range self.payees {
		if votes < payee.Count() {
			masternodeAccount = payee.masternodeAccount
			votes = payee.Count()
		}
	}
	return masternodeAccount, votes > -1
}

//Used to record the last winning block of the masternode. At least 2 votes need to be satisfied
// Has(2,masternode.account)
func (self *MasternodeBlockPayees) Has(num int, account common.Address) bool {

	self.mu.Lock()
	defer self.mu.Unlock()

	if len(self.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	for _, payee := range self.payees {
		if payee.Count() >= num && payee.masternodeAccount == account {
			return true
		}
	}
	return false
}
