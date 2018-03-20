package masternode

import (

	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/eth/downloader"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/ethdb"
	"github.com/ethzero/go-ethzero/eth/fetcher"
)

type MasternodeManager struct{

	networkId uint64
	blockchain  *core.BlockChain
	chainconfig *params.ChainConfig
	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
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


	txsyncCh    chan *txsync
	// channels for fetcher, syncer, txsyncLoop
	newMasternodeCh   chan *Masternode
	noMoreMasternodes chan struct{}


	//mWeAskedForMasternodeListEntry *masternodeSet

}


func NewMasternodeManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, engine consensus.Engine, blockchain *core.BlockChain, chaindb ethdb.Database ,added bool,removed bool) (*MasternodeManager,error){

	manager := &MasternodeManager{
		networkId : networkId,
		chainconfig:config,
		blockchain: blockchain,
		masternodesAdded:    added,
		masternodesRemoved:      removed,
		masternodes: newMasternodeSet(),
		mAskedUsForMasternodes : newMasternodeSet(),
		mWeAskedForMasternode :newMasternodeSet(),
	}

	return manager,nil
}

func (mn *MasternodeManager) Add(m *Masternode) bool{

	m.log.Info("masternode", "CMasternodeMan::Add -- Adding new Masternode: addr=%s, %i now\n", mn.addr.ToString(), size() + 1);
	mn.masternodes.Register(m)
	mn.masternodesAdded = true
	return true;
}

func (mn *MasternodeManager) clear(){

}