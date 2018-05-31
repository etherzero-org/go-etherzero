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
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/consensus"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/crypto/secp256k1"
	"github.com/ethzero/go-ethzero/eth/downloader"
	"github.com/ethzero/go-ethzero/eth/fetcher"
	"github.com/ethzero/go-ethzero/ethdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/params"
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

	masternodes *masternode.MasternodeSet

	enableds map[string]*masternode.Masternode //id -> masternode

	is *InstantSend

	winner *MasternodePayments

	active *masternode.ActiveMasternode

	scope event.SubscriptionScope

	voteFeed event.Feed

	winnerFeed event.Feed

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

	contract *contract.Contract
	srvr     *p2p.Server
}

// NewProtocolManager returns a new Masternode sub protocol manager. The Masternode sub protocol manages peers capable
// with the ETZ-Masternode network.
func NewMasternodeManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, txpool txPool, engine consensus.Engine, blockchain *core.BlockChain, chaindb ethdb.Database) (*MasternodeManager, error) {
	// Create the protocol manager with the base fields
	manager := &MasternodeManager{
		networkId:   networkId,
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chainconfig: config,
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		masternodes: &masternode.MasternodeSet{},
	}

	manager.is = NewInstantx()
	manager.winner = NewMasternodePayments(manager, blockchain.CurrentBlock().Number())

	return manager, nil
}

func (mm *MasternodeManager) removePeer(id string) {
	mm.masternodes.SetState(id, masternode.MasternodeDisconnected)
}

func (mm *MasternodeManager) Start(srvr *p2p.Server, contract *contract.Contract, peers *peerSet) {
	mm.contract = contract
	mm.srvr = srvr
	mm.peers = peers
	log.Trace("MasternodeManqager start ")
	mns, err := masternode.NewMasternodeSet(contract)
	if err != nil {
		log.Error("masternode.NewMasternodeSet", "error", err)
	}
	mm.masternodes = mns

	mm.active = masternode.NewActiveMasternode(srvr)

	mm.is.Active = mm.active

	go mm.masternodeLoop()
}

func (mm *MasternodeManager) Stop() {

}

// SubscribeTxPreEvent registers a subscription of VoteEvent and
// starts sending event to the given channel.
func (self *MasternodeManager) SubscribeVoteEvent(ch chan<- core.VoteEvent) event.Subscription {
	return self.is.SubscribeVoteEvent(ch)
}

// SubscribeWinnerVoteEvent registers a subscription of PaymentVoteEvent and
// starts sending event to the given channel.
func (self *MasternodeManager) SubscribeWinnerVoteEvent(ch chan<- core.PaymentVoteEvent) event.Subscription {
	return self.winner.SubscribeWinnerVoteEvent(ch)
}

func (mm *MasternodeManager) newPeer(p *peer) {
	p.SetMasternode(true)
	mm.masternodes.SetState(p.id, masternode.MasternodeEnable)
}

// Deterministically select the oldest/best masternode to pay on the network
// Pass in the hash value of the block that participates in the calculation.
// Dash is the Hash passed to the first 100 blocks.
// If use the current block Hash, there is a risk that the current block will be discarded.
func (mm *MasternodeManager) BestMasternode(block common.Hash) (*masternode.Masternode, error) {
	// masternodes is nil
	if mm.masternodes == nil {
		return nil, errors.New("no masternode detected")
	}

	var (
		enableMasternodeNodes = mm.masternodes.EnableNodes()
		paids                 []int
		tenthNetWork          = len(enableMasternodeNodes) / 10 // TODO: when len < 10
		countTenth            = 0
		highest               = big.NewInt(0)
		winner                *masternode.Masternode
	)

	sortMap := make(map[int]*masternode.Masternode)
	if enableMasternodeNodes == nil {
		return nil, errors.New("no masternode detected")
	}
	log.Trace(" The number of local cached masternode ", "EnablesMasternodes", len(enableMasternodeNodes))
	if len(enableMasternodeNodes) < 1 {
		return nil, fmt.Errorf("The number of local masternodes is too less to obtain the best Masternode")
	}

	for _, node := range enableMasternodeNodes {
		i := int(node.Height.Int64())
		paids = append(paids, i)
		sortMap[i] = node
	}
	// Sort them low to high
	sort.Sort(sort.IntSlice(paids))

	for _, i := range paids {
		fmt.Printf("CalculateScore result index: %d \t  Score :%d \n", i, sortMap[i].CalculateScore(block))
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

	if winner == nil {
		return nil, fmt.Errorf("The number of local masternodes is too less to obtain the best Masternode")
	}
	return winner, nil
}

func (mm *MasternodeManager) GetMasternodeRank(id string) (int, bool) {

	var rank int = 0
	mm.syncer()
	block := mm.blockchain.CurrentBlock()

	if block == nil {
		log.Error("ERROR: GetBlockHash() failed at BlockHeight:%d ", block.Number())
		return rank, false
	}
	masternodeScores := mm.GetMasternodeScores(block.Hash(), 1)

	tRank := 0
	for _, masternode := range masternodeScores {
		//info := MasternodeInfo()
		tRank++
		if id == masternode.ID {
			rank = tRank
			break
		}
	}
	return rank, true
}

func (mm *MasternodeManager) GetMasternodeScores(blockHash common.Hash, minProtocol int) map[*big.Int]*masternode.Masternode {

	masternodeScores := make(map[*big.Int]*masternode.Masternode)

	for _, m := range mm.masternodes.EnableNodes() {
		masternodeScores[m.CalculateScore(blockHash)] = m
	}
	return masternodeScores
}

func (mm *MasternodeManager) ProcessTxLockVotes(votes []*masternode.TxLockVote) bool {

	rank, ok := mm.GetMasternodeRank(mm.active.ID)
	if !ok {
		log.Info("InstantSend::Vote -- Can't calculate rank for masternode ", mm.active.ID, " rank: ", rank)
		return false
	} else if rank > SIGNATURES_TOTAL {
		log.Info("InstantSend::Vote -- Masternode not in the top ", SIGNATURES_TOTAL, " (", rank, ")")
		return false
	}
	log.Info("InstantSend::Vote -- In the top ", SIGNATURES_TOTAL, " (", rank, ")")

	for i := range votes {
		if !mm.is.ProcessTxLockVote(votes[i]) {
			log.Info("processTxLockVotes vote failed vote Hash:", votes[i].Hash())
		} else {
			//Vote valid, let us forward it
			mm.winner.winnerFeed.Send(core.VoteEvent{votes[i]})
		}
	}

	return mm.is.ProcessTxLockVotes(votes)
}

func (mm *MasternodeManager) ProcessPaymentVotes(votes []*masternode.MasternodePaymentVote) bool {

	for i, vote := range votes {
		if !mm.winner.Vote(vote) {
			log.Info("Payment Winner vote :: Block Payment winner vote failed ", "vote hash:", vote.Hash().String(), "i:%s", i)
			return false
		}
	}
	return true
}

func (mn *MasternodeManager) ProcessTxVote(tx *types.Transaction) bool {

	mn.is.ProcessTxLockRequest(tx)
	log.Info("Transaction Lock Request accepted,", "txHash:", tx.Hash().String(), "MasternodeId", mn.active.ID)
	mn.is.Accept(tx)
	mn.is.Vote(tx.Hash())

	return true
}

// If server is masternode, connect one masternode at least
func (mm *MasternodeManager) checkPeers() {
	if mm.active.State() != masternode.ACTIVE_MASTERNODE_STARTED {
		return
	}
	for _, p := range mm.peers.peers {
		if p.isMasternode {
			return
		}
	}

	nodes := make(map[int]*masternode.Masternode)
	var i int = 0
	for _, p := range mm.masternodes.EnableNodes() {
		if p.State == masternode.MasternodeEnable && p.ID != mm.active.ID {
			nodes[i] = p
			i++
		}
	}
	if i <= 0 {
		return
	}
	key := rand.Intn(i - 1)
	mm.srvr.AddPeer(nodes[key].Node)
}

func (mn *MasternodeManager) updateActiveMasternode() {
	var state int

	n := mn.masternodes.Node(mn.active.ID)
	if n == nil {
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else if int(n.Node.TCP) != mn.active.Addr.Port {
		log.Error("updateActiveMasternode", "Port", n.Node.TCP, "active.Port", mn.active.Addr.Port)
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else if !n.Node.IP.Equal(mn.active.Addr.IP) {
		log.Error("updateActiveMasternode", "IP", n.Node.IP, "active.IP", mn.active.Addr.IP)
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else {
		state = masternode.ACTIVE_MASTERNODE_STARTED
	}

	mn.active.SetState(state)
}
func (mn *MasternodeManager) masternodeLoop() {
	mn.updateActiveMasternode()
	if mn.active.State() == masternode.ACTIVE_MASTERNODE_STARTED {
		fmt.Println("masternodeCheck true")
		mn.checkPeers()
	} else if !mn.srvr.MasternodeAddr.IP.Equal(net.IP{}) {
		var misc [32]byte
		misc[0] = 1
		copy(misc[1:17], mn.srvr.Config.MasternodeAddr.IP)
		binary.BigEndian.PutUint16(misc[17:19], uint16(mn.srvr.Config.MasternodeAddr.Port))

		var buf bytes.Buffer
		buf.Write(mn.srvr.Self().ID[:])
		buf.Write(misc[:])
		d := "0x4da274fd" + common.Bytes2Hex(buf.Bytes())
		fmt.Println("Masternode transaction data:", d)
	}

	mn.masternodes.Show()

	joinCh := make(chan *contract.ContractJoin, 32)
	quitCh := make(chan *contract.ContractQuit, 32)
	joinSub, err1 := mn.contract.WatchJoin(nil, joinCh)
	if err1 != nil {
		// TODO: exit
		return
	}
	quitSub, err2 := mn.contract.WatchQuit(nil, quitCh)
	if err2 != nil {
		// TODO: exit
		return
	}

	//pingMsg := &masternode.PingMsg{
	//	ID: self.node.ID,
	//	IP: self.node.IP,
	//	Port: self.node.TCP,
	//}
	//t := time.NewTimer(time.Second * 5)

	for {
		select {
		case join := <-joinCh:
			fmt.Println("join", common.Bytes2Hex(join.Id[:]))
			node, err := mn.masternodes.NodeJoin(join.Id)
			if err == nil {
				if bytes.Equal(join.Id[:], mn.srvr.Self().ID[0:32]) {
					mn.updateActiveMasternode()
				} else {
					mn.srvr.AddPeer(node.Node)
				}
				mn.masternodes.Show()
			}

		case quit := <-quitCh:
			fmt.Println("quit", common.Bytes2Hex(quit.Id[:]))
			mn.masternodes.NodeQuit(quit.Id)
			if bytes.Equal(quit.Id[:], mn.srvr.Self().ID[0:32]) {
				mn.updateActiveMasternode()
			}
			mn.masternodes.Show()

		case err := <-joinSub.Err():
			joinSub.Unsubscribe()
			fmt.Println("eventJoin err", err.Error())
		case err := <-quitSub.Err():
			quitSub.Unsubscribe()
			fmt.Println("eventQuit err", err.Error())

			//case <-t.C:
			//	pingMsg.Update(self.privateKey)
			//	peers := self.peers.peers
			//	for _, peer := range peers {
			//		fmt.Println("peer", peer.ID())
			//		if err := peer.SendMasternodePing(pingMsg); err != nil {
			//			fmt.Println("err:", err)
			//		}
			//	}
			//	t.Reset(time.Second * 100)
		}
	}
}

func (mm *MasternodeManager) DealPingMsg(pm *masternode.PingMsg) error {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], pm.Time)
	key, err := secp256k1.RecoverPubkey(crypto.Keccak256(b[:]), pm.Sig)
	if err != nil || len(key) != 65 {
		return err
	}
	id := fmt.Sprintf("%x", key[1:9])
	node := mm.masternodes.Node(id)
	if node == nil {
		return fmt.Errorf("error id %s", id)
	}
	if node.LastPingTime > pm.Time {
		return fmt.Errorf("error ping time: %d > %d", node.LastPingTime, pm.Time)
	}
	mm.masternodes.RecvPingMsg(id, pm.Time)
	return nil
}
