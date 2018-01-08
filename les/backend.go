// Copyright 2016 The go-ethzero Authors
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

// Package les implements the Light Ethzero Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/ethzero/go-ethzero/accounts"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/common/hexutil"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/bloombits"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/etz"
	"github.com/ethzero/go-ethzero/etz/downloader"
	"github.com/ethzero/go-ethzero/etz/filters"
	"github.com/ethzero/go-ethzero/etz/gasprice"
	"github.com/ethzero/go-ethzero/etzdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/internal/etzapi"
	"github.com/ethzero/go-ethzero/light"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/node"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/p2p/discv5"
	"github.com/ethzero/go-ethzero/params"
	rpc "github.com/ethzero/go-ethzero/rpc"
)

type LightEthzero struct {
	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb etzdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *etzapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *etz.Config) (*LightEthzero, error) {
	chainDb, err := etz.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	letz := &LightEthzero{
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           etz.CreateConsensusEngine(ctx, &config.Ethash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     etz.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	letz.relay = NewLesTxRelay(peers, letz.reqDist)
	letz.serverPool = newServerPool(chainDb, quitSync, &letz.wg)
	letz.retriever = newRetrieveManager(peers, letz.reqDist, letz.serverPool)
	letz.odr = NewLesOdr(chainDb, letz.chtIndexer, letz.bloomTrieIndexer, letz.bloomIndexer, letz.retriever)
	if letz.blockchain, err = light.NewLightChain(letz.odr, letz.chainConfig, letz.engine); err != nil {
		return nil, err
	}
	letz.bloomIndexer.Start(letz.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		letz.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	letz.txPool = light.NewTxPool(letz.chainConfig, letz.blockchain, letz.relay)
	if letz.protocolManager, err = NewProtocolManager(letz.chainConfig, true, ClientProtocolVersions, config.NetworkId, letz.eventMux, letz.engine, letz.peers, letz.blockchain, nil, chainDb, letz.odr, letz.relay, quitSync, &letz.wg); err != nil {
		return nil, err
	}
	letz.ApiBackend = &LesApiBackend{letz, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	letz.ApiBackend.gpo = gasprice.NewOracle(letz.ApiBackend, gpoParams)
	return letz, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Etzerbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Etzerbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Etzerbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the ethzero package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightEthzero) APIs() []rpc.API {
	return append(etzapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "etz",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "etz",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "etz",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightEthzero) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightEthzero) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightEthzero) TxPool() *light.TxPool              { return s.txPool }
func (s *LightEthzero) Engine() consensus.Engine           { return s.engine }
func (s *LightEthzero) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightEthzero) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightEthzero) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightEthzero) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethzero protocol implementation.
func (s *LightEthzero) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = etzapi.NewPublicNetAPI(srvr, s.networkId)
	// search the topic belonging to the oldest supported protocol because
	// servers always advertise all supported protocols
	protocolVersion := ClientProtocolVersions[len(ClientProtocolVersions)-1]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethzero protocol.
func (s *LightEthzero) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
