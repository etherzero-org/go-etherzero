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
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"encoding/json"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"gopkg.in/fatih/set.v0"
)

const (
	MNPAYMENTS_SIGNATURES_REQUIRED = 6
	MNPaymentsSignaturesTotal      = 10
)

var (
	errInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	errECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

// is a callback type for vote when new block arrives
type masternodeRanksFn func(height *big.Int) map[int64]*masternode.Masternode

// Masternode Payments Class
// Keeps track of who should get paid for which blocks
type MasternodePayments struct {
	cachedBlockHeight *big.Int // keep track of current block height
	minBlocksToStore  *big.Int // ... but at least nMinBlocksToStore (payments blocks) dash default value:5000
	storageCoeff      *big.Int //masternode count times nStorageCoeff payments blocks should be stored ... default value:1.25

	votes      map[common.Hash]*masternode.MasternodePaymentVote // vote hash -> vote
	blocks     map[uint64]*MasternodeBlockPayees
	didNotVote map[string]int64
	scope      event.SubscriptionScope
	active     *masternode.ActiveMasternode
	lastVoted  map[common.Address]*big.Int // masternodeID <- height

	ranksFn    masternodeRanksFn //The callback function used to get the current position of the Masternodes
	winnerFeed event.Feed
	mu         sync.Mutex
}

// SignerFn is a signer function callback when a contract requires a method to
// sign the transaction before submission.
type SignerFn func(types.Signer, common.Address, *types.Transaction) (*types.Transaction, error)

// TransactOpts is the collection of authorization data required to create a
// valid Ethereum transaction.
type TransactOpts struct {
	From   common.Address // Ethereum account to send the transaction from
	Nonce  *big.Int       // Nonce to use for the transaction execution (nil = use pending state)
	Signer SignerFn       // Method to use for signing the transaction (mandatory)

	Value    *big.Int // Funds to transfer along along the transaction (nil = 0 = no funds)
	GasPrice *big.Int // Gas price to use for the transaction execution (nil = gas price oracle)
	GasLimit uint64   // Gas limit to set for the transaction execution (0 = estimate)

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

func (self *MasternodePayments) winners() string {
	self.mu.Lock()
	height := self.cachedBlockHeight.Uint64()
	maxnHeight := height + 21
	str := ``
	for i := height; i < maxnHeight; i++ {
		v, ok := self.blocks[i]
		if ok {
			strJson, _ := json.Marshal(v)
			str = fmt.Sprintf("%v %v \n", str, string(strJson))
		}
	}
	self.mu.Unlock()
	return str
}

func NewMasternodePayments(height *big.Int, fn masternodeRanksFn) *MasternodePayments {
	return &MasternodePayments{
		cachedBlockHeight: height,
		minBlocksToStore:  big.NewInt(1),
		storageCoeff:      big.NewInt(1),
		votes:             make(map[common.Hash]*masternode.MasternodePaymentVote),
		blocks:            make(map[uint64]*MasternodeBlockPayees),
		lastVoted:         make(map[common.Address]*big.Int),
		didNotVote:        make(map[string]int64),
		ranksFn:           fn,
	}
}

//hash is blockHash,(!GetBlockHash(blockHash, vote.nBlockHeight - 101))
func (self *MasternodePayments) Add(vote *masternode.MasternodePaymentVote) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	hash := vote.Hash()
	log.Info("MasternodePayments Add vote ", " Height ",vote.Number," masternode", vote.MasternodeId," Hash", hash)
	if temp := self.votes[hash]; temp != nil {
		return false
	}
	self.votes[hash] = vote
	var len int
	if payee := self.blocks[vote.Number.Uint64()]; payee == nil {
		blockPayees := NewMasternodeBlockPayees(vote.Number)
		blockPayees.Add(vote)
		self.blocks[vote.Number.Uint64()] = blockPayees
		len = self.blocks[vote.Number.Uint64()].Count()
	} else {
		self.blocks[vote.Number.Uint64()].Add(vote)
		len = self.blocks[vote.Number.Uint64()].Count()
	}
	log.Info("masternode block vote successed", "count", len)
	return true
}

func (self *MasternodePayments) VoteCount() int { return len(self.votes) }

func (self *MasternodePayments) BlockCount() int { return len(self.blocks) }

//vote hash
func (self *MasternodePayments) HasVerifiedVote(hash common.Hash) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	if vote := self.votes[hash]; vote != nil {
		return true
	}
	return false
}

func (self *MasternodePayments) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.blocks = make(map[uint64]*MasternodeBlockPayees)
	self.votes = make(map[common.Hash]*masternode.MasternodePaymentVote)

}

func (self *MasternodePayments) ProcessBlock(block *types.Block, rank int) bool {

	var height *big.Int

	if self.active.State() != masternode.ACTIVE_MASTERNODE_STARTED {
		log.Info("Masternode not active")
		return false
	}
	if rank > MNPaymentsSignaturesTotal {
		log.Info("Masternode not in the top ", "Total", MNPaymentsSignaturesTotal, "rank ", rank)
		return false
	}
	height = new(big.Int).Add(block.Number(), big.NewInt(1))

	// LOCATE THE NEXT MASTERNODE WHICH SHOULD BE PAID
	log.Info("ProcessBlock -- Start: ", "vote blockheight", height, " masternodeId", self.active.ID)

	vote := masternode.NewMasternodePaymentVote(height, self.active.ID, self.active.Account)
	log.Info("MasternodePayments::ProcessBlock -- Signing vote ")
	hash := vote.Hash()
	sig, err := vote.Sign(hash[:], self.active.PrivateKey)
	vote.Sig = sig
	if err == nil {
		if vote.Verify(hash[:], sig, &self.active.PrivateKey.PublicKey) {
			// vote constructed sucessfully, let's store and relay it
			self.Add(vote)
			go self.PostVoteEvent(vote)
			log.Info("MasternodePayments:: ProcessBlock vote successed!")
			return true
		}
	}
	log.Error("MasternodePayments::processBlock -- Failed to sign consensus vote")
	return false

}

func (self *MasternodePayments) transact(opts *TransactOpts, contract common.Address, input []byte) (*types.Transaction, error) {

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64

	nonce = opts.Nonce.Uint64()
	// Figure out the gas allowance and gas price values
	gasPrice := opts.GasPrice
	gasLimit := opts.GasLimit
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction

	rawTx = types.NewTransaction(nonce, contract, value, gasLimit, gasPrice, input)
	if opts.Signer == nil {
		return nil, errors.New("no signer to authorize the transaction with")
	}
	signedTx, err := opts.Signer(types.HomesteadSigner{}, opts.From, rawTx)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

//Handle the voting of other masternodes
func (self *MasternodePayments) Vote(vote *masternode.MasternodePaymentVote, storageLimit *big.Int) bool {

	log.Info("MasternodePayments vote begin ")
	self.mu.Lock()
	if self.votes[vote.Hash()] != nil {
		log.Trace("ERROR:Avoid processing same vote multiple times", "hash=", vote.Hash(), " , Height:", vote.Number)
		return false
	}
	self.mu.Unlock()
	//vote out of range
	//firstBlock := self.cachedBlockHeight.Sub(self.cachedBlockHeight, storageLimit)
	//if vote.Number.Cmp(firstBlock) > 0 || vote.Number.Cmp(self.cachedBlockHeight.Add(self.cachedBlockHeight, big.NewInt(20))) > 0 {
	//	log.Trace("ERROR:vote out of range: ", "FirstBlock=", firstBlock.String(), ", BlockHeight=", vote.Number, " CacheHeight=", self.cachedBlockHeight.String())
	//	fmt.Printf("ERROR:vote out of range , FirstBlock=%d, BlockHeight=%d,CacheHeight=%d",  firstBlock, vote.Number, self.cachedBlockHeight)
	//	return false
	//}
	//canvote
	if !self.CanVote(vote.Number, vote.MasternodeAccount) {
		log.Info("masternode already voted, masternode account:", "masternodeAccount", vote.MasternodeAccount)
		return false
	}
	log.Info("masternode_winner vote: ", "blockHeight:", vote.Number, "cacheHeight:", self.cachedBlockHeight, "Hash:", vote.Hash().String())
	if self.Add(vote) {
		//Relay
		go self.PostVoteEvent(vote)
	}
	return true
}

// SubscribeWinnerVoteEvent registers a subscription of PaymentVoteEvent and
// starts sending event to the given channel.
func (self *MasternodePayments) SubscribeWinnerVoteEvent(ch chan<- core.BlockVoteEvent) event.Subscription {
	return self.scope.Track(self.winnerFeed.Subscribe(ch))
}

func (is *MasternodePayments) PostVoteEvent(vote *masternode.MasternodePaymentVote) {
	is.winnerFeed.Send(core.BlockVoteEvent{vote})
}

// Find an entry in the masternode list that is next to be paid
// height is blockNumber , address is masternode account
func (self *MasternodePayments) BlockWinner(height *big.Int) (common.Address, bool) {

	if self.blocks[height.Uint64()] != nil {
		return self.blocks[height.Uint64()].Best()
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
func (self *MasternodePayments) CheckAndRemove(limit *big.Int) {
	for hash, vote := range self.votes {
		if new(big.Int).Sub(self.cachedBlockHeight, vote.Number).Cmp(limit) > 0 {
			log.Info("CheckAndRemove -- Removing old Masternode payment: BlockHeight=", vote.Number.String())
			delete(self.votes, hash)
			delete(self.blocks, vote.Number.Uint64())
		}
	}
}

func (self *MasternodePayments) CheckPreviousBlockVotes(height *big.Int) {

	ranks := self.ranksFn(height)
	var (
		debugStr = ""
		found    = false
		account  = common.Address{}
	)
	for i := 0; i < MNPaymentsSignaturesTotal && i < len(ranks); i++ {
		node := ranks[int64(i)]
		if payees := self.blocks[height.Uint64()]; payees != nil {
			hashs := payees.hashs.List()
			for i = 0; i < payees.hashs.Size(); i++ {
				voteHash := hashs[i].(common.Hash)
				var vote *masternode.MasternodePaymentVote
				if vote = self.votes[voteHash]; vote == nil {
					continue
				}
				if vote.MasternodeId == node.ID {
					found = true
					account = node.Account
					break
				}
			}
		}
		if !found {
			debugStr = fmt.Sprintf("CheckPreviousBlockVotes --   %s - no vote received", node.ID)
			self.didNotVote[node.ID]++
		}
		debugStr += fmt.Sprintf("CheckPreviousBlockVotes --   %s - voted for %s \n", node.ID, account.String())
	}
	debugStr += fmt.Sprintf("CheckPreviousBlockVotes -- Masternodes which missed a vote in the past:\n")
	for it, i := range self.didNotVote {
		debugStr += fmt.Sprintf("CheckPreviousBlockVotes --   %s: %d\n", it, i)
	}
	log.Info("MasternodePayments CheckPreviousBlockVotes", "debug info", debugStr)
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

func (self *MasternodePayee) Count() int {
	return len(self.votes)
}

func (self *MasternodePayee) Votes() []*masternode.MasternodePaymentVote {
	return self.votes
}

type MasternodeBlockPayees struct {
	number *big.Int //blockHeight
	payees []*MasternodePayee
	mu     sync.Mutex
	hashs  *set.Set
}

func NewMasternodeBlockPayees(number *big.Int) *MasternodeBlockPayees {
	payee := &MasternodeBlockPayees{
		number: number,
		payees: make([]*MasternodePayee, 0),
		hashs:  set.New(),
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

func (self *MasternodeBlockPayees) Count() int {

	return len(self.payees)
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
		masternodeAccount = common.Address{}
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

func (self *MasternodeBlockPayees) AddVoteHash(hash common.Hash) {
	self.hashs.Add(hash)
}

func (self *MasternodeBlockPayees) AllVoteHash() []interface{} {
	return self.hashs.List()
}
