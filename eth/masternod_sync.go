package eth

import (
	"math/big"
	"math/rand"
	"sync/atomic"
	"time"
	"math"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/eth/downloader"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/p2p/discover"
)

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (mm *MasternodeManager) BroadcastBlock(block *types.Block, propagate bool) {
	hash := block.Hash()
	peers := mm.peers.PeersWithoutBlock(hash)

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		var td *big.Int
		if parent := mm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
			td = new(big.Int).Add(block.Difficulty(), mm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		} else {
			log.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}
		// Send the block to a subset of our peers
		transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		for _, peer := range transfer {
			peer.SendNewBlock(block, td)
		}
		log.Trace("Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if mm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			peer.SendNewBlockHashes([]common.Hash{hash}, []uint64{block.NumberU64()})
		}
		log.Trace("Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

// BroadcastTx will propagate a transaction to all peers which are not known to
// already have the given transaction.
func (mm *MasternodeManager) BroadcastTx(hash common.Hash, tx *types.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it
	peers := mm.peers.PeersWithoutTx(hash)
	//FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for _, peer := range peers {
		peer.SendTransactions(types.Transactions{tx})
	}
	log.Trace("Broadcast transaction", "hash", hash, "recipients", len(peers))
}

// Mined broadcast loop
func (self *MasternodeManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case core.NewMinedBlockEvent:
			self.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			self.BroadcastBlock(ev.Block, false) // Only then announce to the rest
		}
	}
}

func (self *MasternodeManager) txBroadcastLoop() {
	for {
		select {
		case event := <-self.txCh:
			self.BroadcastTx(event.Tx.Hash(), event.Tx)
			// Err() channel will be closed when unsubscribing.
		case <-self.txSub.Err():
			return
		}
	}
}



// syncTransactions starts sending all currently pending transactions to the given peer.
func (mm *MasternodeManager) syncTransactions(p *peer) {
	var txs types.Transactions
	pending, _ := mm.txpool.Pending()
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	if len(txs) == 0 {
		return
	}
	select {
	case mm.txsyncCh <- &txsync{p, txs}:
	case <-mm.quitSync:
	}
}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (mm *MasternodeManager) txsyncLoop() {
	var (
		pending = make(map[discover.NodeID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		// Fill pack with transactions up to the target size.
		size := common.StorageSize(0)
		pack.p = s.p
		pack.txs = pack.txs[:0]
		for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
			pack.txs = append(pack.txs, s.txs[i])
			size += s.txs[i].Size()
		}
		// Remove the transactions that will be sent.
		s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
		if len(s.txs) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		s.p.Log().Trace("Sending batch of transactions", "count", len(pack.txs), "bytes", size)
		sending = true
		go func() { done <- pack.p.SendTransactions(pack.txs) }()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-mm.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				pack.p.Log().Debug("Transaction send failed", "err", err)
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-mm.quitSync:
			return
		}
	}
}

// syncer is responsible for periodically synchronising with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (mm *MasternodeManager) syncer() {
	// Start and ensure cleanup of sync mechanisms
	mm.fetcher.Start()
	defer mm.fetcher.Stop()
	defer mm.downloader.Terminate()

	// Wait for different events to fire synchronisation operations
	forceSync := time.NewTicker(forceSyncCycle)
	defer forceSync.Stop()

	for {
		select {
		case <-mm.newPeerCh:
			// Make sure we have peers to select from, then sync
			if mm.peers.Len() < minDesiredPeerCount {
				break
			}
			go mm.synchronise(mm.peers.BestPeer())

		case <-forceSync.C:
			// Force a sync even if not enough peers are present
			go mm.synchronise(mm.peers.BestPeer())

		case <-mm.noMorePeers:
			return
		}
	}
}

// synchronise tries to sync up our local block chain with a remote peer.
func (mm *MasternodeManager) synchronise(peer *peer) {
	// Short circuit if no peers are available
	if peer == nil {
		return
	}
	// Make sure the peer's TD is higher than our own
	currentBlock := mm.blockchain.CurrentBlock()
	td := mm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())

	pHead, pTd := peer.Head()
	if pTd.Cmp(td) <= 0 {
		return
	}
	// Otherwise try to sync with the downloader
	mode := downloader.FullSync
	if atomic.LoadUint32(&mm.fastSync) == 1 {
		// Fast sync was explicitly requested, and explicitly granted
		mode = downloader.FastSync
	} else if currentBlock.NumberU64() == 0 && mm.blockchain.CurrentFastBlock().NumberU64() > 0 {
		// The database seems empty as the current block is the genesis. Yet the fast
		// block is ahead, so fast sync was enabled for this node at a certain point.
		// The only scenario where this can happen is if the user manually (or via a
		// bad block) rolled back a fast sync node below the sync point. In this case
		// however it's safe to reenable fast sync.
		atomic.StoreUint32(&mm.fastSync, 1)
		mode = downloader.FastSync
	} else if currentBlock.NumberU64() == 0 && mm.blockchain.CurrentFastBlock().NumberU64() == 0 && mm.networkId == 88 {
		log.Info("Force fast sync until EthzeroBlock")
		atomic.StoreUint32(&mm.fastSync, 1)
		mode = downloader.FastSync
	}
	// Run the sync cycle, and disable fast sync if we've went past the pivot block
	if err := mm.downloader.Synchronise(peer.id, pHead, pTd, mode); err != nil {
		return
	}
	if atomic.LoadUint32(&mm.fastSync) == 1 {
		log.Info("Fast sync complete, auto disabling")
		atomic.StoreUint32(&mm.fastSync, 0)
	}
	atomic.StoreUint32(&mm.acceptTxs, 1) // Mark initial sync done
	if head := mm.blockchain.CurrentBlock(); head.NumberU64() > 0 {
		// We've completed a sync cycle, notify all peers of new state. This path is
		// essential in star-topology networks where a gateway node needs to notify
		// all its out-of-date peers of the availability of a new block. This failure
		// scenario will most often crop up in private and hackathon networks with
		// degenerate connectivity, but it should be healthy for the mainnet too to
		// more reliably update peers or the local TD state.
		go mm.BroadcastBlock(head, false)
	}
}