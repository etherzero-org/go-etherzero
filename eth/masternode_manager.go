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
	"github.com/ethzero/go-ethzero/p2p/discover"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/rlp"
	"sort"
)

type MasternodeManager struct {
	networkId uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool      txPool
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	maxPeers    int

	fetcher    *fetcher.Fetcher
	peers      *peerSet

	masternodes map[string]*masternode.Masternode //id -> masternode

	is *InstantSend

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

type MasterNodelist struct {
	url string
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

// local MasterNodelists.
type MasterNodelists []*MasterNodelist


type ltrInfo struct {
	tx     *types.Transaction
	sentTo map[*peer]struct{}
}
type LesTxRelay struct {
	txSent       map[common.Hash]*ltrInfo
	txPending    map[common.Hash]struct{}
	ps           *peerSet
	peerList     []*peer
	peerStartPos int
	lock         sync.RWMutex
}


func MasterNodelistManager(ma *masternode.Masternode) {
	//Managme the masterNodelist
	//srv := ma.Server()
	//masterpeers := srv.PeersInfo()
	//fmt.Println("Master Node list :")
	//for _,ma:=range masterpeers{
	//	fmt.Println( ma.Name,ma.ID,ma.Network.LocalAddress,ma.MasterState)
	//}

}

func MasterNodeAdd(ma *masternode.Masternode) {
	srv := ma.Server()
	node, _ := discover.ParseNode("enode://d79eee6402b5e61d846cdbd068a4db9d4a392c2c1929a205bb91abfea1723b63c02595156b11f340585e7fdf1918f1143067287e0efb45e8029b82e9a9abe6c0@127.0.0.1:31211")
	srv.AddPeer(node)

}

// send sends a list of transactions to at most a given number of peers at
// once, never resending any particular transaction to the same peer twice
func (self *LesTxRelay)sendMasterNodelist(txs types.Transactions, mlist *MasterNodelists, count int) {
	sendTo := make(map[*peer]types.Transactions)
	self.peerStartPos++ // rotate the starting position of the peer list
	if self.peerStartPos >= len(self.peerList) {
		self.peerStartPos = 0
	}

	for _, tx := range txs {
		hash := tx.Hash()
		ltr, ok := self.txSent[hash]
		if !ok {
			ltr = &ltrInfo{
				sentTo: make(map[*peer]struct{}),
			}
			self.txSent[hash] = ltr
			self.txPending[hash] = struct{}{}
		}

		if len(self.peerList) > 0 {
			cnt := count
			pos := self.peerStartPos
			for {
				peer := self.peerList[pos]
				if _, ok := ltr.sentTo[peer]; !ok {
					sendTo[peer] = append(sendTo[peer], tx)
					ltr.sentTo[peer] = struct{}{}
					cnt--
				}
				if cnt == 0 {
					break // sent it to the desired number of peers
				}
				pos++
				if pos == len(self.peerList) {
					pos = 0
				}
				if pos == self.peerStartPos {
					break // tried all available peers
				}
			}
		}
	}
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

	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}
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
	manager.fetcher = fetcher.New(blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, inserter, manager.removePeer)

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
func (mm *MasternodeManager) GetNextMasternodeInQueueForPayment(hash common.Hash) *masternode.Masternode {

	var paidMasternodes map[int]*masternode.Masternode

	for _, masternode := range mm.masternodes {
		paidMasternodes[masternode.Paid()] = masternode
	}

	var paids []int
	sort.Ints(paids)

	var tenthNetWork = len(mm.masternodes) / 10
	var countTenth = 0
	var highest *big.Int

	var winnerMasternode *masternode.Masternode

	for _, paid := range paids {
		fmt.Printf("%s\t%d\n", paid, paidMasternodes[paid].CalculateScore(hash))
		score := paidMasternodes[paid].CalculateScore(hash)
		if score.Cmp(highest) > 0 {
			highest = score
			winnerMasternode = paidMasternodes[paid]
		}
		countTenth++
		if countTenth >= tenthNetWork {
			break
		}
	}
	return winnerMasternode
}

func (mm *MasternodeManager) GetMasternodeRank(	id discover.NodeID) (int, bool) {

	var rank int = 0
	mm.syncer()
	block := mm.blockchain.CurrentBlock()

	if block == nil {
		mm.log.Info("ERROR: GetBlockHash() failed at nBlockHeight:%d ", block.Number())
	}
	masternodeScores:=mm.GetMasternodeScores(block.Hash(),1)

	tRank:=0
	for _,masternode := range masternodeScores{
		tRank++
		if id == masternode.ID{
			rank=tRank
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

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (mm *MasternodeManager) handleMsg(p *peer) error {
	// Read the next message from the remote peer, and ensure it's fully consumed
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	defer msg.Discard()

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")
	case p.version >= etz64 && msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			if entry, err := mm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}
		}
		return p.SendNodeData(data)

	case p.version >= etz64 && msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
	case p.version >= etz64 && msg.Code == GetReceiptsMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash     common.Hash
			bytes    int
			receipts []rlp.RawValue
		)
		for bytes < softResponseLimit && len(receipts) < downloader.MaxReceiptFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := mm.blockchain.GetReceiptsByHash(hash)
			if results == nil {
				if header := mm.blockchain.GetHeaderByHash(hash); header == nil || header.ReceiptHash != types.EmptyRootHash {
					continue
				}
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				log.Error("Failed to encode receipt", "err", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendReceiptsRLP(receipts)

	case p.version >= etz64 && msg.Code == NewVoteMsg:
		// A batch of vote arrived to one of our previous requests
		var votes []*types.TxLockVote
		if err := msg.Decode(&votes); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		rank, ok := mm.GetMasternodeRank(mm.active.ID)
		if !ok {
			mm.log.Info("InstantSend::Vote -- Can't calculate rank for masternode ", mm.active.ID.String(), " rank: ", rank)
			return errResp(ErrCalculateRankForMasternode,"msg %v: %v", msg, ok)
		} else if rank > SIGNATURES_TOTAL {
			mm.log.Info("InstantSend::Vote -- Masternode not in the top ", SIGNATURES_TOTAL, " (", rank, ")")
			return errResp(ErrMasternodeNotInTheTop,"msg %v: %v", msg, ok)
		}
		mm.log.Info("InstantSend::Vote -- In the top ", SIGNATURES_TOTAL, " (", rank, ")")
		mm.is.ProcessTxLockVotes(votes)

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&mm.acceptTxs) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}
			p.MarkTransaction(tx.Hash())
		}
		mm.txpool.AddRemotes(txs)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}
