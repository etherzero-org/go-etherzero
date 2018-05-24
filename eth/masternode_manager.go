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
	"sort"
	"sync"
	"sync/atomic"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/eth/downloader"
	"github.com/ethzero/go-ethzero/eth/fetcher"
	"github.com/ethzero/go-ethzero/ethdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/masternode"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/params"
	"github.com/pkg/errors"
)

const (
	SIGNATURES_TOTAL = 10
)

type MasternodeManager struct {
	networkId uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool      txPool
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	maxPeers    int

	fetcher *fetcher.Fetcher
	peers   *peerSet

	masternodes map[string]*masternode.Masternode //id -> masternode

	enableds map[string]*masternode.Masternode //id -> masternode

	is *InstantSend

	winner *MasternodePayments

	active *masternode.Masternode

	SubProtocols []p2p.Protocol

	eventMux      *event.TypeMux
	txCh          chan core.TxPreEvent
	txSub         event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan *peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup

	log log.Logger
}

func (m *MasternodeManager) List() map[string]*masternode.Masternode {

	return m.masternodes

}

func (m *MasternodeManager) Add(node *masternode.Masternode) {

	info := node.MasternodeInfo()

	if m.masternodes[info.ID] == nil {
		m.masternodes[info.ID] = node
	}
	log.Warn(" The Masternode already exists ", "Masternode ID", info.ID)
}

// NewProtocolManager returns a new ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the ethereum network.
func NewMasternodeManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, txpool txPool, engine consensus.Engine, blockchain *core.BlockChain, chaindb ethdb.Database) (*MasternodeManager, error) {
	// Create the protocol manager with the base fields
	manager := &MasternodeManager{
		networkId:   networkId,
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chainconfig: config,
		peers:       newPeerSet(),
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}

	//if len(manager.SubProtocols) == 0 {
	//	return nil, errIncompatibleConfig
	//}
	validator := func(header *types.Header) error {
		return engine.VerifyHeader(blockchain, header, true)
	}
	heighter := func() uint64 {
		return blockchain.CurrentBlock().NumberU64()
	}

	inserter := func(blocks types.Blocks) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			log.Warn("Discarded bad propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		return manager.blockchain.InsertChain(blocks)
	}

	vote := func(block *types.Block) bool {
		return manager.winner.ProcessBlock(block)
	}

	manager.fetcher = fetcher.New(blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, inserter, manager.removePeer, vote)

	return manager, nil
}

func (mm *MasternodeManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := mm.peers.Peer(id)
	if peer == nil {
		return
	}
	log.Debug("Removing Etherzero masternode peer", "peer", id)

	if err := mm.peers.Unregister(id); err != nil {
		log.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.Peer.Disconnect(p2p.DiscUselessPeer)
	}
}

func (mm *MasternodeManager) Start(maxPeers int) {
	mm.maxPeers = maxPeers

	// broadcast transactions
	mm.txCh = make(chan core.TxPreEvent, txChanSize)
	mm.txSub = mm.txpool.SubscribeTxPreEvent(mm.txCh)
	go mm.txBroadcastLoop()

	// broadcast mined blocks
	mm.minedBlockSub = mm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	go mm.minedBroadcastLoop()

	// start sync handlers
	go mm.syncer()
	go mm.txsyncLoop()
}

func (mm *MasternodeManager) Stop() {
	log.Info("Stopping Etherzero masternode protocol")

	mm.txSub.Unsubscribe()         // quits txBroadcastLoop
	mm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	mm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(mm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to mm.peers yet
	// will exit when they try to register.
	mm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	mm.wg.Wait()

	log.Info("Etherzero masternode protocol stopped")
}

func (mm *MasternodeManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// Deterministically select the oldest/best masternode to pay on the network
// Pass in the hash value of the block that participates in the calculation.
// Dash is the Hash passed to the first 100 blocks.
// If use the current block Hash, there is a risk that the current block will be discarded.
func (mm *MasternodeManager) GetNextMasternodeInQueueForPayment(block common.Hash) (*masternode.Masternode, error) {

	var (
		paids        []int
		tenthNetWork = len(mm.masternodes) / 10
		countTenth   = 0
		highest      *big.Int
		winner       *masternode.Masternode
		sortMap      map[int]*masternode.Masternode
	)
	if mm.masternodes == nil {
		return nil, errors.New("no masternode detected")
	}
	for _, node := range mm.masternodes {
		i := int(node.Height.Int64())
		paids = append(paids, i)
		sortMap[i] = node
	}

	sort.Ints(paids)

	for _, i := range paids {
		fmt.Printf("%s\t%d\n", i, sortMap[i].CalculateScore(block))
		score := sortMap[i].CalculateScore(block)
		if score.Cmp(highest) > 0 {
			highest = score
			winner = sortMap[i]
		}
		countTenth++
		if countTenth >= tenthNetWork {
			break
		}
	}

	return winner, nil
}

func (mm *MasternodeManager) GetMasternodeRank(id string) (int, bool) {

	var rank int = 0
	mm.syncer()
	block := mm.blockchain.CurrentBlock()

	if block == nil {
		mm.log.Info("ERROR: GetBlockHash() failed at BlockHeight:%d ", block.Number())
		return rank, false
	}
	masternodeScores := mm.GetMasternodeScores(block.Hash(), 1)

	tRank := 0
	for _, masternode := range masternodeScores {
		info := masternode.MasternodeInfo()
		tRank++
		if id == info.ID {
			rank = tRank
			break
		}
	}
	return rank, true
}

func (mm *MasternodeManager) GetMasternodeScores(blockHash common.Hash, minProtocol int) map[*big.Int]*masternode.Masternode {

	masternodeScores := make(map[*big.Int]*masternode.Masternode)

	for _, m := range mm.masternodes {
		masternodeScores[m.CalculateScore(blockHash)] = m
	}
	return masternodeScores
}


func (mm *MasternodeManager) ProcessTxLockVotes(votes []*types.TxLockVote) bool {

	info := mm.active.MasternodeInfo()
	rank, ok := mm.GetMasternodeRank(info.ID)
	if !ok {
		log.Info("InstantSend::Vote -- Can't calculate rank for masternode ", info.ID, " rank: ", rank)
		return false
	} else if rank > SIGNATURES_TOTAL {
		log.Info("InstantSend::Vote -- Masternode not in the top ", SIGNATURES_TOTAL, " (", rank, ")")
		return false
	}
	log.Info("InstantSend::Vote -- In the top ", SIGNATURES_TOTAL, " (", rank, ")")
	return mm.is.ProcessTxLockVotes(votes)
}

func (mm *MasternodeManager) ProcessPaymentVotes(vote *MasternodePaymentVote) bool {

	return mm.winner.Vote(vote)
}



func(mn *MasternodeManager) ProcessTxVote(tx *types.Transaction) bool{


	mn.is.ProcessTxLockRequest(tx)
	log.Info("Transaction Lock Request accepted,","txHash:",tx.Hash().String(),"MasternodeId",mn.active.MasternodeInfo().ID)
	mn.is.Accept(tx)
	mn.is.Vote(tx.Hash())

	return true
}