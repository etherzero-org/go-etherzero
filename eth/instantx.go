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
	"fmt"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/rlp"

	"hash"
	"math/big"
	"sync"
)

const (
	/*
		    At 15 signatures, 1/2 of the masternode network can be owned by
		    one party without comprimising the security of InstantSend
			在15个签名中，masternode网络的1/2可以由一方拥有，而不会包含InstantSend的安全性
		    (1000/2150.0)**10 = 0.00047382219560689856
		    (1000/2900.0)**10 = 2.3769498616783657e-05
		    ### getting 5 of 10 signatures w/ 1000 nodes of 2900
			获得10个签名中的5个/ 1000个2900的节点
		    (1000/2900.0)**5 = 0.004875397277841433
	*/
	INSTANTSEND_CONFIRMATIONS_REQUIRED = 6

	DEFAULT_INSTANTSEND_DEPTH = 5

	MIN_INSTANTSEND_PROTO_VERSION = 70208
	// For how long we are going to accept votes/locks
	// after we saw the first one for a specific transaction
	INSTANTSEND_LOCK_TIMEOUT_SECONDS = 15

	// For how long we are going to keep invalid votes and votes for failed lock attempts,
	// must be greater than INSTANTSEND_LOCK_TIMEOUT_SECONDS
	INSTANTSEND_FAILED_TIMEOUT_SECONDS = 60

)

type InstantSend struct {

	// maps for AlreadyHave
	accepted          map[common.Hash]*types.Transaction     // tx hash - tx
	rejected          map[common.Hash]*types.Transaction     // tx hash - tx
	txLockedVotes     map[common.Hash]*masternode.TxLockVote // vote hash - vote
	txLockVotesOrphan map[common.Hash]*masternode.TxLockVote // vote hash - vote

	Candidates map[common.Hash]*masternode.TxLockCondidate // tx hash - lock candidate

	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	all       map[common.Hash]int                // All votes to allow lookups
	lockedTxs map[common.Hash]*types.Transaction //
	mu        sync.Mutex

	//track masternodes who voted with no txreq (for DOS protection)
	//追踪没有txreq投票的masternodes（用于DOS保护）
	masternodeOrphanVotes map[common.Hash]int

	/*
	   At 15 signatures, 1/2 of the masternode network can be owned by
	   one party without comprimising the security of InstantSend
	   (1000/2150.0)**10 = 0.00047382219560689856
	   (1000/2900.0)**10 = 2.3769498616783657e-05

	   ### getting 5 of 10 signatures w/ 1000 nodes of 2900
	   (1000/2900.0)**5 = 0.004875397277841433
	*/

	//std::map<COutPoint, int64_t> mapMasternodeOrphanVotes; // mn outpoint - time
	cachedHeight *big.Int
	voteFeed     event.Feed

	scope event.SubscriptionScope

	Active *masternode.ActiveMasternode
}

func NewInstantx() *InstantSend {

	is := &InstantSend{
		accepted:              make(map[common.Hash]*types.Transaction),
		rejected:              make(map[common.Hash]*types.Transaction),
		txLockedVotes:         make(map[common.Hash]*masternode.TxLockVote),
		txLockVotesOrphan:     make(map[common.Hash]*masternode.TxLockVote),
		Candidates:            make(map[common.Hash]*masternode.TxLockCondidate),
		all:                   make(map[common.Hash]int),
		lockedTxs:             make(map[common.Hash]*types.Transaction),
		masternodeOrphanVotes: make(map[common.Hash]int),
	}

	return is
}

//received a consensus TxLockRequest
func (is *InstantSend) ProcessTxLockRequest(request *types.Transaction) bool {

	txHash := request.Hash()

	//check to see if we conflict with existing completed lock

	if _, ok := is.lockedTxs[txHash]; !ok {
		// Conflicting with complete lock, proceed to see if we should cancel them both
		log.Info("WARNING: Found conflicting completed Transaction Lock", "InstantSend  txid=", txHash, "completed lock txid=", is.lockedTxs[txHash])
	}

	// Check to see if there are votes for conflicting request,
	// if so - do not fail, just warn user
	if _, ok := is.all[txHash]; !ok {
		log.Info("WARNING:Double spend attempt!", "InstantSend txid=", txHash, "Voted txid count :", is.all[txHash])
	}

	if !is.CreateTxLockCandidate(request) {
		log.Info("CreateTxLockCandidate failed, txid=", txHash)
		return false
	}
	// Masternodes will sometimes propagate votes before the transaction is known to the client.
	// If this just happened - lock inputs, resolve conflicting locks, update transaction status
	// forcing external script notification.
	is.TryToFinalizeLockCandidate(is.Candidates[txHash])

	return true
}

func (is *InstantSend) vote(condidate *masternode.TxLockCondidate) {

	txHash := condidate.Hash()
	if _, ok := is.accepted[txHash]; !ok {
		return
	}

	txlockRequest := condidate.TxLockRequest()
	nonce := txlockRequest.Nonce()
	if nonce < 1 {
		log.Info("nonce error")
		return
	}

	var alreadyVoted bool = false
	if _, ok := is.all[txHash]; !ok {
		txLockCondidate := is.Candidates[txHash] //找到当前交易的侯选人
		if txLockCondidate.HasMasternodeVoted(is.Active.ID) {
			alreadyVoted = true
			log.Info("CInstantSend::Vote -- WARNING: We already voted for this outpoint, skipping: txHash=", txHash, ", masternodeid=", is.Active.ID)
			return
		}
	}

	vote := masternode.NewTxLockVote(txHash, is.Active.ID) //构建一个投票对象
	if alreadyVoted {
		return
	}
	signByte, err := vote.Sign(vote.Hash(), is.Active.PrivateKey)

	if err != nil {
		return
	}
	sigErr := vote.Verify(vote.Hash().Bytes(), signByte, is.Active.PrivateKey.Public())

	if sigErr != nil {
		return
	}

	// vote constructed sucessfully, let's store and relay it
	tvHash := vote.Hash()
	is.voteFeed.Send(vote)

	is.txLockedVotes[tvHash] = vote
	txLock := is.Candidates[txHash]

	if txLock.AddVote(vote) {
		log.Info("Vote created successfully, relaying: txHash=", txHash.String(), ", vote=", tvHash.String())
		is.all[txHash] = 1
	}

}

func (is *InstantSend) Vote(hash common.Hash) {

	txLockCondidate, ok := is.Candidates[hash]
	if !ok {
		return
	}
	is.vote(txLockCondidate)
	is.TryToFinalizeLockCandidate(txLockCondidate)
}

func (is *InstantSend) CreateTxLockCandidate(request *types.Transaction) bool {

	if !request.CheckNonce() {
		return false
	}
	txhash := request.Hash()

	if is.Candidates == nil {
		log.Info("CreateTxLockCandidate -- new,txid=", txhash.String())
		txlockcondidate := masternode.NewTxLockCondidate(request)
		is.Candidates[txhash] = txlockcondidate
	} else {
		log.Info("CreateTxLockCandidate -- seen, txid", txhash.String())
	}

	return true
}

func (is *InstantSend) ProcessTxLockVote(vote *masternode.TxLockVote) bool {

	txhash := vote.Hash()
	txLockCondidate := is.Candidates[txhash]

	log.Info("ProcessTxLockVote -- Transaction Lock Vote, txid=", txhash.String())
	if _, ok := is.all[txhash]; !ok {
		is.all[txhash]++
	}
	if txLockCondidate.AddVote(vote) {
		return false
	}

	signatures := txLockCondidate.CountVotes()
	signaturesMax := txLockCondidate.MaxSignatures()
	log.Info("ProcessTxLockVote Transaction Lock signatures count:", signatures, "/", signaturesMax, ",vote Hash:", vote.Hash().String())

	is.TryToFinalizeLockCandidate(txLockCondidate)

	return true
}

func (is *InstantSend) ProcessTxLockVotes(votes []*masternode.TxLockVote) bool {
	for i := range votes {
		if !is.ProcessTxLockVote(votes[i]) {
			log.Info("processTxLockVotes vote failed vote Hash:", votes[i].Hash())
		}
	}
	return true
}

func (is *InstantSend) Accept(tx *types.Transaction) {
	if is.accepted[tx.Hash()] != nil {
		is.accepted[tx.Hash()] = tx
	} else {
		log.Info("transaction already exists in the Accept Map", "tx hash:", tx.Hash().String())
	}
}

func (is *InstantSend) Reject(tx *types.Transaction) {
	if is.rejected[tx.Hash()] != nil {
		is.rejected[tx.Hash()] = tx
	} else {
		log.Info("transaction already exists in the Reject Map", "tx hash:", tx.Hash().String())
	}
}

func (is *InstantSend) IsLockedInstantSendTransaction(hash common.Hash) bool {

	_, ok := is.Candidates[hash]
	if !ok {
		return false
	}
	return is.lockedTxs[hash] != nil

}

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *masternode.TxLockCondidate) {
	txLockRequest := condidate.TxLockRequest()

	txHash := txLockRequest.Hash()
	if condidate.IsReady() {
		is.lockedTxs[txHash] = txLockRequest
	}
}

func (is *InstantSend) PostVoteEvent(vote *masternode.TxLockVote) {

	is.voteFeed.Send(core.VoteEvent{vote})
}

// SubscribeTxPreEvent registers a subscription of VoteEvent and
// starts sending event to the given channel.
func (self *InstantSend) SubscribeVoteEvent(ch chan<- core.VoteEvent) event.Subscription {
	return self.scope.Track(self.voteFeed.Subscribe(ch))
}

func (self *InstantSend) CheckAndRemove() {

	self.mu.Lock()
	defer self.mu.Unlock()

	for txHash, lockCondidate := range self.Candidates {

		if lockCondidate.IsExpired(self.cachedHeight) {
			log.Info("InstantSend::CheckAndRemove -- Removing expired Transaction Lock Candidate: txid= \n", txHash.String())
			delete(self.rejected, txHash)
			delete(self.accepted, txHash)
			delete(self.Candidates, txHash)
		}
	}

	for txHash, lockVote := range self.txLockedVotes {

		if lockVote.IsExpired(self.cachedHeight) {
			log.Info("InstantSend::CheckAndRemove -- Removing expired vote: txid=", txHash.String(), "  masternode= ", lockVote.MasternodeId())
			delete(self.txLockedVotes, txHash)
		}
	}

	for txHash, lockVote := range self.txLockedVotes {

		if lockVote.IsFailed() {
			log.Info("InstantSend::CheckAndRemove -- Removing Failed vote: txid=", txHash.String(), "Masternode= ", lockVote.MasternodeId())
		}
	}

}

func (is *InstantSend) GetConfirmations(hash common.Hash) int{

	if is.IsLockedInstantSendTransaction(hash) {
		return DEFAULT_INSTANTSEND_DEPTH
	}
	return 0
}

func (is *InstantSend) Have(hash common.Hash) bool {
	return is.lockedTxs[hash] != nil
}

func (is *InstantSend) String() string {

	str := fmt.Sprintf("InstantSend Lock Candidates :", len(is.Candidates), ", Votes :", len(is.all))
	return str
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
