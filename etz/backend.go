// Copyright 2014 The go-ethzero Authors
// This file is part of the go-ethzero library.
//
// The go-ethzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethzero library. If not, see <http://www.gnu.org/licenses/>.

// Package etz implements the Ethzero protocol.
package etz

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/ethzero/go-ethzero/accounts"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/common/hexutil"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/consensus/clique"
	"github.com/ethzero/go-ethzero/consensus/ethash"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/bloombits"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/vm"
	"github.com/ethzero/go-ethzero/etz/downloader"
	"github.com/ethzero/go-ethzero/etz/filters"
	"github.com/ethzero/go-ethzero/etz/gasprice"
	"github.com/ethzero/go-ethzero/etzdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/internal/etzapi"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/miner"
	"github.com/ethzero/go-ethzero/node"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/rlp"
	"github.com/ethzero/go-ethzero/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Ethzero implements the Ethzero full node service.
type Ethzero struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the ethzero
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb etzdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *EtzApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	etzerbase common.Address

	networkId     uint64
	netRPCService *etzapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etzerbase)
}

func (s *Ethzero) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Ethzero object (including the
// initialisation of the common Ethzero object)
func New(ctx *node.ServiceContext, config *Config) (*Ethzero, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run etz.Ethzero in light sync mode, use les.LightEthzero")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	etz := &Ethzero{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Ethash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		etzerbase:      config.Etzerbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Ethzero protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run getz upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	etz.blockchain, err = core.NewBlockChain(chainDb, etz.chainConfig, etz.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		etz.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	etz.bloomIndexer.Start(etz.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	etz.txPool = core.NewTxPool(config.TxPool, etz.chainConfig, etz.blockchain)

	if etz.protocolManager, err = NewProtocolManager(etz.chainConfig, config.SyncMode, config.NetworkId, etz.eventMux, etz.txPool, etz.engine, etz.blockchain, chainDb); err != nil {
		return nil, err
	}
	etz.miner = miner.New(etz, etz.chainConfig, etz.EventMux(), etz.engine)
	etz.miner.SetExtra(makeExtraData(config.ExtraData))

	etz.ApiBackend = &EtzApiBackend{etz, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	etz.ApiBackend.gpo = gasprice.NewOracle(etz.ApiBackend, gpoParams)

	return etz, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"getz",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (etzdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*etzdb.LDBDatabase); ok {
		db.Meter("etz/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Ethzero service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ethash.Config, chainConfig *params.ChainConfig, db etzdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == ethash.ModeFake:
		log.Warn("Ethash used in fake mode")
		return ethash.NewFaker()
	case config.PowMode == ethash.ModeTest:
		log.Warn("Ethash used in test mode")
		return ethash.NewTester()
	case config.PowMode == ethash.ModeShared:
		log.Warn("Ethash used in shared mode")
		return ethash.NewShared()
	default:
		engine := ethash.New(ethash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the ethzero package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ethzero) APIs() []rpc.API {
	apis := etzapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthzeroAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Ethzero) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Ethzero) Etzerbase() (eb common.Address, err error) {
	s.lock.RLock()
	etzerbase := s.etzerbase
	s.lock.RUnlock()

	if etzerbase != (common.Address{}) {
		return etzerbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			return accounts[0].Address, nil
		}
	}
	return common.Address{}, fmt.Errorf("etzerbase address must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Ethzero) SetEtzerbase(etzerbase common.Address) {
	self.lock.Lock()
	self.etzerbase = etzerbase
	self.lock.Unlock()

	self.miner.SetEtzerbase(etzerbase)
}

func (s *Ethzero) StartMining(local bool) error {
	eb, err := s.Etzerbase()
	if err != nil {
		log.Error("Cannot start mining without etzerbase", "err", err)
		return fmt.Errorf("etzerbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Etzerbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Ethzero) StopMining()         { s.miner.Stop() }
func (s *Ethzero) IsMining() bool      { return s.miner.Mining() }
func (s *Ethzero) Miner() *miner.Miner { return s.miner }

func (s *Ethzero) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Ethzero) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Ethzero) TxPool() *core.TxPool               { return s.txPool }
func (s *Ethzero) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Ethzero) Engine() consensus.Engine           { return s.engine }
func (s *Ethzero) ChainDb() etzdb.Database            { return s.chainDb }
func (s *Ethzero) IsListening() bool                  { return true } // Always listening
func (s *Ethzero) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Ethzero) NetVersion() uint64                 { return s.networkId }
func (s *Ethzero) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Ethzero) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethzero protocol implementation.
func (s *Ethzero) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = etzapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		maxPeers -= s.config.LightPeers
		if maxPeers < srvr.MaxPeers/2 {
			maxPeers = srvr.MaxPeers / 2
		}
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethzero protocol.
func (s *Ethzero) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
