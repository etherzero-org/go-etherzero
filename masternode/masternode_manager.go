package masternode

import (
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/eth/downloader"
	"github.com/ethzero/go-ethzero/eth/fetcher"
	"github.com/ethzero/go-ethzero/ethdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/params"
)

const (
	DSEG_UPDATE_SECONDS   = 3 * 60 * 60
	LAST_PAID_SCAN_BLOCKS = 100
)
const (
	MIN_POSE_PROTO_VERSION = 70203
	MAX_POSE_CONNECTIONS   = 10
	MAX_POSE_RANK          = 10
	MAX_POSE_BLOCKS        = 10
)
const (
	MNB_RECOVERY_QUORUM_TOTAL    = 10
	MNB_RECOVERY_QUORUM_REQUIRED = 6
	MNB_RECOVERY_MAX_ASK_ENTRIES = 10
	MNB_RECOVERY_WAIT_SECONDS    = 60
	MNB_RECOVERY_RETRY_SECONDS   = 3 * 60 * 60
)

type MasternodeManager struct {
	networkId   uint64
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	downloader  *downloader.Downloader
	fetcher     *fetcher.Fetcher
	// map to hold all Masternodes
	masternodes *masternodeSet
	/// Set when masternodes are added, cleared when CGovernanceManager is notified
	masternodesAdded bool `json:"masternodes_added"`
	/// Set when masternodes are removed, cleared when CGovernanceManager is notified
	masternodesRemoved bool `json:"masternodes_removed"`
	// who's asked for the Masternode list and the last time
	mAskedUsForMasternodes *masternodeSet
	// who we asked for the Masternode list and the last time
	mWeAskedForMasternode *masternodeSet

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txsyncCh chan *txsync
	// channels for fetcher, syncer, txsyncLoop
	newMasternodeCh   chan *Masternode
	noMoreMasternodes chan struct{}

	//mWeAskedForMasternodeListEntry *masternodeSet

}

func NewMasternodeManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, engine consensus.Engine, blockchain *core.BlockChain, chaindb ethdb.Database, added bool, removed bool) (*MasternodeManager, error) {

	manager := &MasternodeManager{
		networkId:              networkId,
		chainconfig:            config,
		blockchain:             blockchain,
		masternodesAdded:       added,
		masternodesRemoved:     removed,
		masternodes:            newMasternodeSet(),
		mAskedUsForMasternodes: newMasternodeSet(),
		mWeAskedForMasternode:  newMasternodeSet(),
	}

	return manager, nil
}

func (mm *MasternodeManager) Add(m *Masternode) bool {

	m.log.Info("masternode", "CMasternodeMan::Add -- Adding new Masternode: addr=%s, %i now\n", m.URL(), mm.Size()+1)
	mm.masternodes.Register(m)
	mm.masternodesAdded = true
	return true
}

/// Return the number of (unique) Masternodes
func (mm *MasternodeManager) Size() int {
	return mm.masternodes.Len()
}

func (mm *MasternodeManager) Start() {

	//mm.maxPeers = maxPeers
	//// broadcast transactions
	//mm.txCh = make(chan core.TxPreEvent, txChanSize)
	//mm.txSub = pm.txpool.SubscribeTxPreEvent(pm.txCh)
	//go mm.txBroadcastLoop()
	//
	//// broadcast mined blocks
	//mm.minedBlockSub = mm.eventMux.Subscribe(core.NewMinedBlockEvent{})
	//go mm.minedBroadcastLoop()
	//
	//// start sync handlers
	//go mm.syncer()
	//go mm.txsyncLoop()
}

func (mm *MasternodeManager) Stop() {
	log.Info("Stopping Etherzero Masternode protocol")

	//mm.txSub.Unsubscribe()         // quits txBroadcastLoop
	//mm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop
	//
	//// Quit the sync loop.
	//// After this send has completed, no new peers will be accepted.
	//mm.noMorePeers <- struct{}{}
	//
	//// Quit fetcher, txsyncLoop.
	//close(mm.quitSync)
	//
	//// Disconnect existing sessions.
	//// This also closes the gate for any new registrations on the peer set.
	//// sessions which are already established but not added to pm.peers yet
	//// will exit when they try to register.
	//mm.peers.Close()
	//
	//// Wait for all peer handler goroutines and the loops to come down.
	//mm.wg.Wait()

	log.Info("Etherzero Masternode stopped")
}

// MasternodeInfo represents a short summary of the Etherzero Masternode-protocol metadata
// known about the host Masternode.
type MasternodeInfo struct {
	Alias   string      `json:"alias"`
	Url     string      `json:"url"`
	Account common.Hash `json:"account"`
}
