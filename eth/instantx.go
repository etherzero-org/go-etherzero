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
	"fmt"
	"math/big"
	"sync"
	"time"

	"sync/atomic"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/state"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/rlp"
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
	InstantSendConfirmationsRequired = 6

	DefaultInstantSendDepth = 5

	SignaturesRequired = 6
)

var (
	// ErrInvalidSender is returned if the transaction contains an invalid signature.
	ErrInvalidSender = errors.New("invalid sender")

	// ErrInsufficientFunds is returned if the total cost of executing a transaction
	// is higher than the balance of the user's account.
	ErrInsufficientFundsMin = errors.New("keeping 0.01 etz at least on your wallet")

	//ErrCreateCandidate is returned if the transaction Create TxLockCandidate Failed
	ErrCreateCandidate = errors.New("create Tx Candidate failed")
)

// blockChain provides the state of blockchain and current gas limit to do
// some pre checks in tx pool and event subscribers.
type blockChain interface {
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
	StateAt(root common.Hash) (*state.StateDB, error)
}

type InstantSend struct {
	// maps for AlreadyHave
	accepted          map[common.Hash]*types.Transaction     // tx hash - tx
	rejected          map[common.Hash]*types.Transaction     // tx hash - tx
	txLockedVotes     map[common.Hash]*masternode.TxLockVote // vote hash - vote
	txLockVotesOrphan map[common.Hash]*masternode.TxLockVote // vote hash - vote

	Candidates map[common.Hash]*masternode.TxLockCondidate // tx hash - lock candidate

	all       map[common.Hash]int                // All votes to allow lookups
	lockedTxs map[common.Hash]*types.Transaction //Store all transactions that have completed voting
	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	mu sync.Mutex
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
	cachedBlockHeight *big.Int // Keep track of current block height
	voteFeed          event.Feed
	scope             event.SubscriptionScope
	atWork            int32 // atWork indicates wether the InstantSend is currently working

	Active       *masternode.ActiveMasternode
	chain        blockChain
	txCh         chan core.TxPreEvent
	txSub        event.Subscription
	currentState *state.StateDB // Current state in the blockchain head
	signer       types.Signer
	eth          Backend
}

// NewInstantx new an InstantSend
func NewInstantx(chainconfig *params.ChainConfig, eth Backend) *InstantSend {
	instantSend := &InstantSend{
		accepted:              make(map[common.Hash]*types.Transaction),
		rejected:              make(map[common.Hash]*types.Transaction),
		txLockedVotes:         make(map[common.Hash]*masternode.TxLockVote),
		txLockVotesOrphan:     make(map[common.Hash]*masternode.TxLockVote),
		Candidates:            make(map[common.Hash]*masternode.TxLockCondidate),
		all:                   make(map[common.Hash]int),
		eth:                   eth,
		chain:                 eth.BlockChain(),
		cachedBlockHeight:     eth.BlockChain().CurrentBlock().Number(),
		signer:                types.NewEIP155Signer(chainconfig.ChainId),
		lockedTxs:             make(map[common.Hash]*types.Transaction),
		masternodeOrphanVotes: make(map[string]uint64),
		votesOrphan:           make(map[common.Hash]*masternode.TxLockVote),
	}
	instantSend.txSub = eth.TxPool().SubscribeTxPreEvent(instantSend.txCh)
	instantSend.reset()

	go instantSend.update()
	instantSend.commitNewWork()

	return instantSend
}

//received a consensus TxLockRequest
func (is *InstantSend) ProcessTxLockRequest(request *types.Transaction) error {

	txHash := request.Hash()

	// check to see if we conflict with existing completed lock
	if _, ok := is.lockedTxs[txHash]; !ok {
		// Conflicting with complete lock, proceed to see if we should cancel them both
		log.Info("WARNING: Found conflicting completed Transaction Lock", "InstantSend  txid=", txHash, "completed lock txid=", is.lockedTxs[txHash])
	}
	// Check to see if there are votes for conflicting request,
	// if so - do not fail, just warn user
	if _, ok := is.all[txHash]; !ok {
		log.Info("WARNING:Double spend attempt!", "InstantSend txid=", txHash, "Voted txid count :", is.all[txHash])
	}

	sender, err := is.signer.Sender(request)
	if err == nil {
		return core.ErrInvalidSender
	}
	nonce := is.currentState.GetNonce(sender)

	if nonce < request.Nonce() {
		return core.ErrNonceTooHigh
	}
	if nonce > request.Nonce() {
		return core.ErrNonceTooLow
	}
	if !is.CreateTxLockCandidate(request) {
		log.Info("CreateTxLockCandidate failed, txid=", txHash)
		return ErrCreateCandidate
	}
	// Masternodes will sometimes propagate votes before the transaction is known to the client.
	// If this just happened - lock inputs, resolve conflicting locks, update transaction status
	// forcing external script notification.
	is.TryToFinalizeLockCandidate(is.Candidates[txHash])

	return nil
}

func (is *InstantSend) vote(condidate *masternode.TxLockCondidate) bool {

	txHash := condidate.Hash()
	if _, ok := is.accepted[txHash]; !ok {
		return false
	}

	txlockRequest := condidate.TxLockRequest
	nonce := txlockRequest.Nonce()
	if nonce < 1 {
		log.Info("nonce error")
		return false
	}

	var alreadyVoted bool = false
	if _, ok := is.all[txHash]; !ok {
		txLockCondidate := is.Candidates[txHash]
		if txLockCondidate != nil {
			if txLockCondidate.HasMasternodeVoted(is.Active.ID) {
				alreadyVoted = true
				log.Info("CInstantSend::Vote -- WARNING: We already voted for this outpoint, skipping: txHash=", txHash, ", masternodeid=", is.Active.ID)
				return false
			}
		}
	}
	if alreadyVoted {
		return false
	}
	vote := masternode.NewTxLockVote(txHash, is.Active.ID)
	hash := vote.Hash()
	signByte, err := vote.Sign(hash[:], is.Active.PrivateKey)
	//signByte, err := vote.Sign(is.Active.PrivateKey)
	vote.Sig = signByte
	if err != nil {
		return false
	}
	//publicKey:=crypto.FromECDSAPub(&is.Active.PrivateKey.PublicKey)
	if vote.Verify(&is.Active.PrivateKey.PublicKey) {
		//if vote.CheckSignature(publicKey,vote.Sig){
		log.Info("InstantSend sign Verify valid")
		// vote constructed sucessfully, let's store and relay it
		is.voteFeed.Send(vote)
		// add to txLockedVotes
		_, ok1 := is.txLockedVotes[hash]
		if !ok1 {
			is.txLockedVotes[hash] = vote
		} else {
			return false
		}

		txLock := is.Candidates[txHash]
		if txLock.AddVote(vote) {
			log.Info("Vote created successfully, relaying: txHash=", txHash.String(), ", vote=", hash.String())
			is.all[txHash] = 1
			return true
		}
	} else {
		log.Info("vote Sign verify failed vote hash:", vote.Hash().String())
	}
	return false
}

func (is *InstantSend) Vote(hash common.Hash) bool {

	txLockCondidate, ok := is.Candidates[hash]
	if !ok {
		return false
	}
	if is.vote(txLockCondidate) {
		return is.TryToFinalizeLockCandidate(txLockCondidate)
	}
	return false
}

func (is *InstantSend) CreateTxLockCandidate(request *types.Transaction) bool {

	if !request.CheckNonce() {
		return false
	}
	txhash := request.Hash()
	txlockcondidate := masternode.NewTxLockCondidate(request)
	if is.Candidates == nil {
		log.Info("CreateTxLockCandidate -- new,txid=", txhash.String())
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
	if txLockCondidate.TxLockRequest.CheckNonce() {
		txLockCondidate.MarkAsAttacked()
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
	// there must be a lock candidate
	if _, ok := is.Candidates[hash]; !ok {
		return false
	}
	// and all of these outputs must be included in mapLockedOutpoints with correct hash
	return is.lockedTxs[hash] != nil
}

func (self *InstantSend) IsEnoughOrphanVotesForTx(hash common.Hash) bool {
	var countVotes int = 0
	for txHash := range self.votesOrphan {
		if txHash == hash {
			countVotes++
			if countVotes >= SignaturesRequired {
				return true
			}
		}
	}
	return false
}

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *masternode.TxLockCondidate) bool {
	is.mu.Lock()
	defer is.mu.Unlock()

	txLockRequest := condidate.TxLockRequest
	txHash := txLockRequest.Hash()
	if condidate.IsReady() && !is.IsLockedInstantSendTransaction(txHash) {
		//we have enough votes now
		log.Info("InstantSend ::TryToFinalizeLockCandidate -- Transaction Lock is ready to comply ,txid =", txHash.String())
		//dash LockTransactionInputs

		if is.ResolveConflicts(condidate) {
			is.lockedTxs[txHash] = txLockRequest
			//do something
			//UpdateLockedTransaction
		}
	}
	return true
}

//we have enough votes now
func (is *InstantSend) ResolveConflicts(condidate *masternode.TxLockCondidate) bool {
	// make sure the lock is ready
	if !condidate.IsReady() {
		return false
	}
	is.mu.Lock()
	defer is.mu.Unlock()
	tx := condidate.TxLockRequest

	from, err := is.signer.Sender(tx)
	if err != nil {
		log.Error("ResolveConflicts error,", ErrInvalidSender.Error())
		return false
	}
	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if is.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		log.Error("ResolveConflicts error,", ErrInsufficientFundsMin.Error())
		return false
	}

	txs := is.GetLockedTxListByAccount(from)
	sum := big.NewInt(0)

	for _, tx := range txs {
		sum = new(big.Int).Add(sum, tx.Cost())
	}
	if is.currentState.GetBalance(from).Cmp(sum) < 0 {
		log.Error("ResolveConflicts error")
		for _, tx := range txs {
			candidate := is.Candidates[tx.Hash()]
			candidate.SetConfirmedHeight(big.NewInt(0))
			is.Reject(tx)
			log.Info("ResolveConflicts :: Found conflicting completed Transaction Lock, dropping both txid ", tx.Hash().String())
		}
		is.CheckAndRemove()
		log.Info("ResolveConflicts :: Found conflicting completed Transaction Lock, dropping both ")
		return false
	}

	log.Info("ResolveConflicts -- Done, txid=", tx.Hash().String())
	return true
}

func (is *InstantSend) GetLockedTxListByAccount(address common.Address) []*types.Transaction {
	txs := make([]*types.Transaction, 0, len(is.Candidates))
	for _, candidate := range is.Candidates {
		tx := candidate.TxLockRequest
		if from, err := is.signer.Sender(tx); err == nil && from == address {
			txs = append(txs, tx)
		}
	}
	return txs
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
		if lockCondidate.IsExpired(self.cachedBlockHeight) {
			log.Info("InstantSend::CheckAndRemove -- Removing expired Transaction Lock Candidate: txid= \n", txHash.String())
			delete(self.rejected, txHash)
			delete(self.accepted, txHash)
			delete(self.Candidates, txHash)
		}
	}

	for txHash, lockVote := range self.txLockedVotes {
		if lockVote.IsExpired(self.cachedBlockHeight) {
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
		return DefaultInstantSendDepth
	}
	return 0
}

func (is *InstantSend) reset() {
	newHead := is.chain.CurrentBlock().Header() // Special case during testing
	statedb, err := is.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset instantSend state", "err", err)
		return
	}
	is.currentState = statedb
}

func (is *InstantSend) String() string {
	str := fmt.Sprintf("InstantSend Lock Candidates :", len(is.Candidates), ", Votes :", len(is.all))
	return str
}

func (self *InstantSend) commitNewWork() {

	pending, err := self.eth.TxPool().Pending()
	if err != nil {
		log.Error("Failed to fetch pending transactions", "err", err)
		return
	}
	txs := types.NewTransactionsByPriceAndNonce(self.signer, pending)

	for {
		tx := txs.Peek()
		if tx == nil {
			break
		}
		from, _ := self.signer.Sender(tx)
		err := self.ProcessTxLockRequest(tx)
		switch err {
		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", tx.Nonce())
			txs.Shift()

		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", tx.Nonce())
			txs.Pop()

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			log.Info("Everything ok, collect the logs and shift in the next transaction from the same account")
			txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", tx.Hash(), "err", err)
			txs.Shift()
		}
	}
}

func (self *InstantSend) commitTransactions(txs *types.TransactionsByPriceAndNonce) {

	for {
		tx := txs.Peek()
		// Retrieve the next transaction and abort if all done
		if tx == nil {
			break
		}
		// Error may be ignored here. The error has already been checked
		// during transaction acceptance is the transaction pool.
		//
		// We use the eip155 signer regardless of the current hf.
		from, _ := types.Sender(self.signer, tx)
		err := self.ProcessTxLockRequest(tx)
		switch err {
		case core.ErrNonceTooLow:
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "sender", from, "nonce", tx.Nonce())
			txs.Shift()

		case core.ErrNonceTooHigh:
			// Reorg notification data race between the transaction pool and miner, skip account =
			log.Trace("Skipping account with hight nonce", "sender", from, "nonce", tx.Nonce())
			txs.Pop()

		case nil:
			// Everything ok, collect the logs and shift in the next transaction from the same account
			log.Info("Everything ok, collect the logs and shift in the next transaction from the same account")
			txs.Shift()

		default:
			// Strange error, discard the transaction and get the next in line (note, the
			// nonce-too-high clause will prevent us from executing in vain).
			log.Debug("Transaction failed, account skipped", "hash", tx.Hash(), "err", err)
			txs.Shift()
		}
	}
}

func (self *InstantSend) update() {
	defer self.txSub.Unsubscribe()

	for {
		// A real event arrived, process interesting content
		select {
		case ev := <-self.txCh:
			// Apply transaction to the pending state if we're not mining
			acc, _ := types.Sender(self.signer, ev.Tx)
			txs := map[common.Address]types.Transactions{acc: {ev.Tx}}
			txset := types.NewTransactionsByPriceAndNonce(self.signer, txs)
			self.commitTransactions(txset)
			//system stoped
		case <-self.txSub.Err():
			return
		}
	}
}

func (self *InstantSend) Start() {

	if !atomic.CompareAndSwapInt32(&self.atWork, 0, 1) {
		return //InstantSend sever already started
	}
	atomic.StoreInt32(&self.atWork, 1)
	go self.update()
}

func (self *InstantSend) Stop() {
	self.mu.Lock()
	defer self.mu.Unlock()

	if !atomic.CompareAndSwapInt32(&self.atWork, 1, 0) {
		return //InstantSend sever already stopped
	}
	atomic.StoreInt32(&self.atWork, 0)
	self.CheckAndRemove()
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
