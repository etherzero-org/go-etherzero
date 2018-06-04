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
	"math/big"
	"sync"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/rlp"

	"time"
)

const (
	/*
	   At 15 signatures, 1/2 of the masternode network can be owned by
	   one party without comprimising the security of InstantSend
	   (1000/2150.0)**10 = 0.00047382219560689856
	   (1000/2900.0)**10 = 2.3769498616783657e-05
	   ### getting 5 of 10 signatures w/ 1000 nodes of 2900
	   (1000/2900.0)**5 = 0.004875397277841433
	*/
	INSTANTSEND_CONFIRMATIONS_REQUIRED = 6

	DEFAULT_INSTANTSEND_DEPTH = 5

	MIN_INSTANTSEND_PROTO_VERSION = 70208

	SIGNATURES_REQUIRED = 6
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
	masternodeOrphanVotes map[string]uint64 //masternodeID - Orphan time
	votesOrphan           map[common.Hash]*masternode.TxLockVote

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
	scope        event.SubscriptionScope

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
		masternodeOrphanVotes: make(map[string]uint64),
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

	txlockRequest := condidate.TxLockRequest

	nonce := txlockRequest.Nonce()
	if nonce < 1 {
		log.Info("nonce error")
		return
	}

	var alreadyVoted bool = false
	if _, ok := is.all[txHash]; !ok {
		txLockCondidate := is.Candidates[txHash]
		if txLockCondidate.HasMasternodeVoted(is.Active.ID) {
			alreadyVoted = true
			log.Info("CInstantSend::Vote -- WARNING: We already voted for this outpoint, skipping: txHash=", txHash, ", masternodeid=", is.Active.ID)
			return
		}
	}

	vote := masternode.NewTxLockVote(txHash, is.Active.ID)
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
	var txlockcondidate *masternode.TxLockCondidate

	if is.Candidates == nil {
		log.Info("CreateTxLockCandidate -- new,txid=", txhash.String())
		txlockcondidate = masternode.NewTxLockCondidate(request)
		is.Candidates[txhash] = txlockcondidate
	} else if is.Candidates[request.Hash()] == nil {
		txlockcondidate.TxLockRequest = request
		log.Info("CreateTxLockCandidate -- seen, txid", txhash.String())
		if txlockcondidate.IsTimeout() {
			log.Info("InstantSend::CreateTxLockCandidate -- timed out, txid=%s\n", txhash.String())
			return false
		}
		log.Info("InstantSend::CreateTxLockCandidate -- update empty, txid=%s\n", txhash.String())
	}
	return true
}

func (self *InstantSend) ProcessTxLockVote(vote *masternode.TxLockVote) bool {

	txHash := vote.Hash()
	// TODO:Verification work is handled in the MasternodeManager
	//if !vote.IsValid() {
	//	log.Error("CInstantSend::ProcessTxLockVote -- Vote is invalid, txid=", txHash.String())
	//	return false
	//}

	self.voteFeed.Send(vote)
	txLockCondidate := self.Candidates[txHash]

	// Masternodes will sometimes propagate votes before the transaction is known to the client,
	// will actually process only after the lock request itself has arrived
	if txLockCondidate == nil {
		if self.votesOrphan[txHash] == nil {
			//createEmptyCondidate
			self.votesOrphan[txHash] = vote
			reProcess := true
			log.Info("CInstantSend::ProcessTxLockVote -- Orphan vote: txid=", txHash.String(), " masternodeId=", vote.MasternodeId())

			var tx *types.Transaction

			if tx = self.accepted[txHash]; tx != nil {
				if tx = self.rejected[txHash]; tx != nil {
					reProcess = false
				}
			}
			// We have enough votes for corresponding lock to complete,
			// tx lock request should already be received at this stage.
			if reProcess && self.IsEnoughOrphanVotesForTx(txHash) {
				log.Info("InstantSend::ProcessTxLockVote -- Found enough orphan votes, reprocessing Transaction Lock Request: txid=", txHash.String())
				self.ProcessTxLockRequest(tx)
				return true
			}
		} else {
			log.Info("InstantSend::ProcessTxLockVote -- Orphan vote: txid= ", txHash.String(), "  masternode= ", vote.MasternodeId())
		}
		// This tracks those messages and allows only the same rate as of the rest of the network
		// TODO: make sure this works good enough for multi-quorum
		MasternodeOrphanExpireTime := 60 * uint64(time.Second) * 10 // keep time data for 10 minutes
		if self.masternodeOrphanVotes[vote.MasternodeId()] == 0 {
			self.masternodeOrphanVotes[vote.MasternodeId()] = MasternodeOrphanExpireTime
		} else {
			preOrphanVote := self.masternodeOrphanVotes[vote.MasternodeId()]
			if preOrphanVote > uint64(time.Now().Unix()) && preOrphanVote > self.GetAverageMasternodeOrphanVoteTime() {
				log.Info("InstantSend::ProcessTxLockVote -- masternode is spamming orphan Transaction Lock Votes: txid=",
					txHash.String(), "masternode= \n", vote.MasternodeId())
				return false
			}
			// not spamming, refresh
			self.masternodeOrphanVotes[vote.MasternodeId()] = MasternodeOrphanExpireTime
		}
		return true
	}

	log.Info("ProcessTxLockVote -- Transaction Lock Vote, txid=", txHash.String())
	if _, ok := self.all[txHash]; !ok {
		self.all[txHash]++
	}
	if txLockCondidate.AddVote(vote) {
		return false
	}

	signatures := txLockCondidate.CountVotes()
	signaturesMax := txLockCondidate.MaxSignatures()
	log.Info("ProcessTxLockVote Transaction Lock signatures count:", signatures, "/", signaturesMax, ",vote Hash:", vote.Hash().String())

	self.TryToFinalizeLockCandidate(txLockCondidate)

	return true
}

func (self *InstantSend) GetAverageMasternodeOrphanVoteTime() uint64 {
	self.mu.Lock()
	defer self.mu.Unlock()
	// NOTE: should never actually call this function when masternodeOrphanVotes is empty
	if len(self.masternodeOrphanVotes) < 1 {
		return 0
	}
	var total uint64 = 0
	for moVote := range self.masternodeOrphanVotes {
		total += self.masternodeOrphanVotes[moVote]
	}
	return total / uint64(len(self.masternodeOrphanVotes))
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

func (self *InstantSend) IsEnoughOrphanVotesForTx(hash common.Hash) bool {

	var countVotes int = 0
	for txHash := range self.votesOrphan {
		if txHash == hash {
			countVotes++
			if countVotes >= SIGNATURES_REQUIRED {
				return true
			}
		}
	}
	return false
}

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *masternode.TxLockCondidate) {

	txLockRequest := condidate.TxLockRequest

	txHash := txLockRequest.Hash()
	if condidate.IsReady() {
		is.lockedTxs[txHash] = txLockRequest
	}
}

//we have enough votes now
func (is *InstantSend) ResolveConflicts(condidate masternode.TxLockCondidate) bool {

	return true
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

func (is *InstantSend) GetConfirmations(hash common.Hash) int {

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
