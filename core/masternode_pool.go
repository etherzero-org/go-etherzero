// Copyright 2016 The go-ethereum Authors
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

package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/state"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/masternode"
)

// txPermanent is the number of mined blocks after a mined transaction is
// considered permanent and no rollback is expected
var txPermanent = uint64(500)

// TxPool implements the transaction pool for light clients, which keeps track
// of the status of locally created transactions, detecting if they are included
// in a block (mined) or rolled back. There are no queued transactions since we
// always receive all locally signed transactions in the same order as they are
// created.
type MasternodePool struct {
	config       *params.ChainConfig
	signer       types.Signer
	quit         chan bool
	txFeed       event.Feed
	voteFeed     event.Feed
	scope        event.SubscriptionScope
	chainHeadCh  chan ChainHeadEvent
	chainHeadSub event.Subscription
	mu           sync.RWMutex
	currentState *state.StateDB // Current state in the blockchain head
	//odr          OdrBackend
	relay     TxRelayBackend
	voteRelay VoteRelayBackend
	active *masternode.Masternode
	head      common.Hash
	votes     map[common.Hash]*types.TxLockVote    //votes by vote hash
	nonce     map[common.Address]uint64            // "pending" nonce
	pending   map[common.Hash]*types.Transaction   // pending transactions by tx hash
	mined     map[common.Hash][]*types.Transaction // mined transactions by block hash

	homestead bool
}

// TxRelayBackend provides an interface to the mechanism that forwards transacions
// to the ETH network. The implementations of the functions should be non-blocking.
//
// Send instructs backend to forward new transactions
// NewHead notifies backend about a new head after processed by the tx pool,
//  including  mined and rolled back transactions since the last event
// Discard notifies backend about transactions that should be discarded either
//  because they have been replaced by a re-send or because they have been mined
//  long ago and no rollback is expected
type TxRelayBackend interface {
	Send(txs types.Transactions)
	NewHead(head common.Hash, mined []common.Hash, rollback []common.Hash)
	Discard(hashes []common.Hash)
}

// VoteRelayBackend provides an interface to the mechanism that forwards Votes
// to the Masternode network. The implementations of the functions should be non-blocking.
//
// Send instructs backend to forward new Votes
// NewHead notifies backend about a new head after processed by the tx pool,
//  including  mined and rolled back transactions since the last event
// Discard notifies backend about transactions that should be discarded either
//  because they have been replaced by a re-send or because they have been mined
//  long ago and no rollback is expected
type VoteRelayBackend interface {
	Send(vote types.TxLockVote)
	NewHead(head common.Hash, mined []common.Hash, rollback []common.Hash)
	Discard(hashes []common.Hash)
}

// NewTxPool creates a new Masternode transaction & vote pool
func NewMasternodePool(config *params.ChainConfig, relay TxRelayBackend, vr VoteRelayBackend,active *masternode.Masternode) *MasternodePool {
	pool := &MasternodePool{
		config:      config,
		signer:      types.NewEIP155Signer(config.ChainId),
		nonce:       make(map[common.Address]uint64),
		pending:     make(map[common.Hash]*types.Transaction),
		mined:       make(map[common.Hash][]*types.Transaction),
		quit:        make(chan bool),
		chainHeadCh: make(chan ChainHeadEvent, chainHeadChanSize),
		relay:       relay,
		voteRelay:   vr,
		active:active,
	}
	// Subscribe events from blockchain
	//pool.chainHeadSub = pool.chain.SubscribeChainHeadEvent(pool.chainHeadCh)
	go pool.eventLoop()

	return pool
}

// GetNonce returns the "pending" nonce of a given address. It always queries
// the nonce belonging to the latest header too in order to detect if another
// client using the same key sent a transaction.
func (pool *MasternodePool) GetNonce(ctx context.Context, addr common.Address) (uint64, error) {

	nonce := pool.currentState.GetNonce(addr)
	if pool.currentState.Error() != nil {
		return 0, pool.currentState.Error()
	}
	sn, ok := pool.nonce[addr]
	if ok && sn > nonce {
		nonce = sn
	}
	if !ok || sn < nonce {
		pool.nonce[addr] = nonce
	}
	return nonce, nil
}

// txStateChanges stores the recent changes between pending/mined states of
// transactions. True means mined, false means rolled back, no entry means no change
type txStateChanges map[common.Hash]bool

// setState sets the status of a tx to either recently mined or recently rolled back
func (txc txStateChanges) setState(txHash common.Hash, mined bool) {
	val, ent := txc[txHash]
	if ent && (val != mined) {
		delete(txc, txHash)
	} else {
		txc[txHash] = mined
	}
}

// getLists creates lists of mined and rolled back tx hashes
func (txc txStateChanges) getLists() (mined []common.Hash, rollback []common.Hash) {
	for hash, val := range txc {
		if val {
			mined = append(mined, hash)
		} else {
			rollback = append(rollback, hash)
		}
	}
	return
}

// rollbackTxs marks the transactions contained in recently rolled back blocks
// as rolled back. It also removes any positional lookup entries.
func (pool *MasternodePool) rollbackTxs(hash common.Hash, txc txStateChanges) {
	if list, ok := pool.mined[hash]; ok {
		for _, tx := range list {
			txHash := tx.Hash()
			pool.pending[txHash] = tx
			txc.setState(txHash, false)
		}
		delete(pool.mined, hash)
	}
}

// blockCheckTimeout is the time limit for checking new blocks for mined
// transactions. Checking resumes at the next chain head event if timed out.
const blockCheckTimeout = time.Second * 3

// eventLoop processes chain head events and also notifies the tx relay backend
// about the new head hash and tx state changes
func (pool *MasternodePool) eventLoop() {
	for {
		select {
		//case ev := <-pool.chainHeadCh:
		//case _ := <-pool.chainHeadCh:
		//pool.setNewHead(ev.Block.Header())
		// hack in order to avoid hogging the lock; this part will
		// be replaced by a subsequent PR.
		//	time.Sleep(time.Millisecond)

		// System stopped
		case <-pool.chainHeadSub.Err():
			return
		}
	}
}

// Stop stops the light transaction pool
func (pool *MasternodePool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	pool.scope.Close()
	// Unsubscribe subscriptions registered from blockchain
	pool.chainHeadSub.Unsubscribe()
	close(pool.quit)
	log.Info("Transaction pool stopped")
}

// SubscribeTxPreEvent registers a subscription of core.TxPreEvent and
// starts sending event to the given channel.
func (pool *MasternodePool) SubscribeTxPreEvent(ch chan<- TxPreEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

// Stats returns the number of currently pending (locally created) transactions
func (pool *MasternodePool) Stats() (pending int) {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	pending = len(pool.pending)
	return
}

// validateTx checks whether a transaction is valid according to the consensus rules.
func (pool *MasternodePool) validateTx(ctx context.Context, tx *types.Transaction) error {
	// Validate sender
	var (
		from common.Address
		err  error
	)

	// Validate the transaction sender and it's sig. Throw
	// if the from fields is invalid.
	if from, err = types.Sender(pool.signer, tx); err != nil {
		return ErrInvalidSender
	}
	// Last but not least check for nonce errors

	if n := pool.currentState.GetNonce(from); n > tx.Nonce() {
		return ErrNonceTooLow
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if b := pool.currentState.GetBalance(from); b.Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}

	// Should supply enough intrinsic gas
	gas, err := IntrinsicGas(tx.Data(), tx.To() == nil, pool.homestead)
	if err != nil {
		return err
	}
	if tx.Gas() < gas {
		return ErrIntrinsicGas
	}
	return pool.currentState.Error()
}

// add validates a new transaction and sets its state pending if processable.
// It also updates the locally stored nonce if necessary.
func (self *MasternodePool) add(ctx context.Context, tx *types.Transaction) error {
	hash := tx.Hash()

	if self.pending[hash] != nil {
		return fmt.Errorf("Known transaction (%x)", hash[:4])
	}
	err := self.validateTx(ctx, tx)
	if err != nil {
		return err
	}

	if _, ok := self.pending[hash]; !ok {
		self.pending[hash] = tx
		nonce := tx.Nonce() + 1
		addr, _ := types.Sender(self.signer, tx)
		if nonce > self.nonce[addr] {
			self.nonce[addr] = nonce
		}
		info:=self.active.MasternodeInfo()
		vote:=types.NewTxLockVote(hash,info.ID)
		self.addVote(ctx,vote)

		// Notify the subscribers. This event is posted in a goroutine
		// because it's possible that somewhere during the post "Remove transaction"
		// gets called which will then wait for the global tx pool lock and deadlock.
		go self.txFeed.Send(TxPreEvent{Tx: tx})
	}

	// Print a log message if low enough level is set
	log.Debug("Pooled new transaction", "hash", hash, "from", log.Lazy{Fn: func() common.Address { from, _ := types.Sender(self.signer, tx); return from }}, "to", tx.To())
	return nil
}

func (self *MasternodePool) AddVote(ctx context.Context, vote *types.TxLockVote) error {

	self.mu.Lock()
	defer self.mu.Unlock()

	if err := self.addVote(ctx, vote); err != nil {
		return err
	}
	//fmt.Println("Send", vote.Hash())
	self.voteRelay.Send(*vote)
	return nil
}

func (self *MasternodePool) addVote(ctx context.Context, vote *types.TxLockVote) error {

	hash := vote.Hash()
	if self.votes[hash] != nil {
		return fmt.Errorf("Known vote (%x)", hash[:4])
	}
	if _, ok := self.votes[hash]; !ok {
		self.votes[hash] = vote

		// Notify the subscribers. This event is posted in a goroutine
		// because it's possible that somewhere during the post "Remove transaction"
		// gets called which will then wait for the global tx pool lock and deadlock.
		go self.voteFeed.Send(VoteEvent{Vote: vote})
	}
	// Print a log message if low enough level is set
	//log.Debug("Pooled new transaction", "hash", hash, "from", log.Lazy{Fn: func() common.Address { from, _ := types.Sender(self.signer, tx); return from }}, "to", tx.To())
	return nil

}

// RemoveVote removes the vote with the given hash from the pool.
func (self *MasternodePool) RemoveVote(hash common.Hash) {
	self.mu.Lock()
	defer self.mu.Unlock()
	//delete from votes map
	delete(self.votes, hash)
	self.relay.Discard([]common.Hash{hash})
}

// Add adds a transaction to the pool if valid and passes it to the tx relay
// backend
func (self *MasternodePool) Add(ctx context.Context, tx *types.Transaction) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	if err := self.add(ctx, tx); err != nil {
		return err
	}
	//fmt.Println("Send", tx.Hash())
	self.relay.Send(types.Transactions{tx})
	return nil
}

// AddTransactions adds all valid transactions to the pool and passes them to
// the tx relay backend
func (self *MasternodePool) AddBatch(ctx context.Context, txs []*types.Transaction) {
	self.mu.Lock()
	defer self.mu.Unlock()
	var sendTx types.Transactions

	for _, tx := range txs {
		if err := self.add(ctx, tx); err == nil {
			sendTx = append(sendTx, tx)
		}
	}
	if len(sendTx) > 0 {
		self.relay.Send(sendTx)
	}
}

// GetTransaction returns a transaction if it is contained in the pool
// and nil otherwise.
func (tp *MasternodePool) GetTransaction(hash common.Hash) *types.Transaction {
	// check the txs first
	if tx, ok := tp.pending[hash]; ok {
		return tx
	}
	return nil
}

// GetTransactions returns all currently processable transactions.
// The returned slice may be modified by the caller.
func (self *MasternodePool) GetTransactions() (txs types.Transactions, err error) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	txs = make(types.Transactions, len(self.pending))
	i := 0
	for _, tx := range self.pending {
		txs[i] = tx
		i++
	}
	return txs, nil
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and nonce.
func (self *MasternodePool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	// Retrieve all the pending transactions and sort by account and by nonce
	pending := make(map[common.Address]types.Transactions)
	for _, tx := range self.pending {
		account, _ := types.Sender(self.signer, tx)
		pending[account] = append(pending[account], tx)
	}
	// There are no queued transactions in a light pool, just return an empty map
	queued := make(map[common.Address]types.Transactions)
	return pending, queued
}

// RemoveTransactions removes all given transactions from the pool.
func (self *MasternodePool) RemoveTransactions(txs types.Transactions) {
	self.mu.Lock()
	defer self.mu.Unlock()
	var hashes []common.Hash
	for _, tx := range txs {
		//self.RemoveTx(tx.Hash())
		hash := tx.Hash()
		delete(self.pending, hash)
		hashes = append(hashes, hash)
	}
	self.relay.Discard(hashes)
}

// RemoveTx removes the transaction with the given hash from the pool.
func (pool *MasternodePool) RemoveTx(hash common.Hash) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	// delete from pending pool
	delete(pool.pending, hash)
	pool.relay.Discard([]common.Hash{hash})
}
