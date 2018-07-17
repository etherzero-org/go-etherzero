// Copyright 2018 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-etherzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/contracts/masternode/contract"
	"github.com/etherzero/go-etherzero/core"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/crypto/secp256k1"
	"github.com/etherzero/go-etherzero/event"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p"
	"github.com/etherzero/go-etherzero/params"
	"math/big"
)

var (
	statsReportInterval = 10 * time.Second // Time interval to report vote pool stats
)

type MasternodeManager struct {
	votes map[common.Hash]*types.Vote // vote hash -> vote
	beats map[common.Hash]time.Time   // Last heartbeat from each known vote

	devoteProtocol *types.DevoteProtocol
	active         *masternode.ActiveMasternode
	masternodes    *masternode.MasternodeSet
	mu             sync.Mutex
	// channels for fetcher, syncer, txsyncLoop
	newPeerCh    chan *peer
	peers        *peerSet
	IsMasternode uint32
	srvr         *p2p.Server
	contract     *contract.Contract
	blockchain   *core.BlockChain
	scope        event.SubscriptionScope
	voteFeed     event.Feed

	pingFeed     event.Feed
	currentCycle uint64        // Current vote of the block chain
	Lifetime     time.Duration // Maximum amount of time vote are queued
}

func NewMasternodeManager(dp *types.DevoteProtocol, blockchain *core.BlockChain, contract *contract.Contract) *MasternodeManager {

	// Create the masternode manager with its initial settings
	manager := &MasternodeManager{
		masternodes:    nil,
		devoteProtocol: dp,
		blockchain:     blockchain,
		votes:          make(map[common.Hash]*types.Vote),
		beats:          make(map[common.Hash]time.Time),
		Lifetime:       30 * time.Second,
		contract:       contract,
	}
	return manager
}

func (self *MasternodeManager) Voting() (*types.Vote, error) {

	if self.active == nil {
		return nil, errors.New("current node is not masternode ")
	}

	currentCycle := self.blockchain.CurrentBlock().Time().Uint64() / params.CycleInterval
	nextCycle := currentCycle + 1
	storeCycle := atomic.LoadUint64(&self.currentCycle)
	if storeCycle >= nextCycle {
		log.Debug("this masternode voted in the next cycle ", "cycle", nextCycle)
		return nil, nil
	}
	masternodeID := self.active.ID
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(nextCycle))
	key = append(key, []byte(masternodeID)...)
	masternodes := self.masternodes.AllNodes()
	weight := int64(0)
	best := self.active.ID
	for _, masternode := range masternodes {
		hash := self.blockchain.CurrentBlock().Hash()
		temp := int64(binary.LittleEndian.Uint32(crypto.Keccak512(hash.Bytes())))
		if temp > weight && masternode.ID != self.active.ID {
			weight = temp
			best = masternode.ID
		}
	}
	vote := types.NewVote(nextCycle, best, self.active.ID)

	vote.SignVote(self.active.PrivateKey)
	log.Info("masternode voting successfully ", "hash", vote.Hash(), "masternode", vote.Masternode, "poll", vote.Poll)
	//self.Add(vote)
	atomic.StoreUint64(&self.currentCycle, nextCycle)
	go self.PostVoteEvent(vote)
	return vote, nil
}

func (self *MasternodeManager) Process(vote *types.Vote) error {
	h := vote.NosigHash()
	masternode := self.masternodes.Node(vote.Masternode)
	if masternode == nil {
		log.Error("masternode not found", "masternodeId", vote.Masternode)
		return errors.New("masternode not found masternodeId" + vote.Masternode)
	}
	pubkey, err := masternode.Node.ID.Pubkey()
	if err != nil {
		log.Error("masternode pubkey not found ", "err", err)
		return err
	}

	if !vote.Verify(h[:], vote.Sign, pubkey) {
		return errors.New("vote valid failed")
	}
	self.Add(vote)
	return nil
}

func (self *MasternodeManager) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

}
func (self *MasternodeManager) Add(vote *types.Vote) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.votes[vote.Hash()] == nil {
		self.votes[vote.Hash()] = vote
		self.beats[vote.Hash()] = time.Now()
	}
}

func (self *MasternodeManager) RemoveVote(vote *types.Vote) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.votes[vote.Hash()] != nil {
		delete(self.votes, vote.Hash())
		delete(self.beats, vote.Hash())
	}
}
func (self *MasternodeManager) Votes() ([]*types.Vote, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	var votes []*types.Vote

	for _, vote := range self.votes {
		votes = append(votes, vote)
	}
	return votes, nil
}

func (self *MasternodeManager) Start(srvr *p2p.Server, peers *peerSet) {
	self.srvr = srvr
	self.peers = peers
	log.Trace("MasternodeManqager start ")
	mns, err := masternode.NewMasternodeSet(self.contract)
	if err != nil {
		log.Error("masternode.NewMasternodeSet", "error", err)
	}
	self.masternodes = mns
	self.active = masternode.NewActiveMasternode(srvr, mns)
	go self.masternodeLoop()

}

func (self *MasternodeManager) Stop() {

}

func (mm *MasternodeManager) masternodeLoop() {
	mm.updateActiveMasternode()
	if mm.active.State() == masternode.ACTIVE_MASTERNODE_STARTED {
		fmt.Println("masternode check true")
		atomic.StoreUint32(&mm.IsMasternode, 1)
		// mm.checkPeers()
	} else if !mm.srvr.MasternodeAddr.IP.Equal(net.IP{}) {
		var misc [32]byte
		misc[0] = 1
		copy(misc[1:17], mm.srvr.Config.MasternodeAddr.IP)
		binary.BigEndian.PutUint16(misc[17:19], uint16(mm.srvr.Config.MasternodeAddr.Port))

		var buf bytes.Buffer
		buf.Write(mm.srvr.Self().ID[:])
		buf.Write(misc[:])
		d := "0x4da274fd" + common.Bytes2Hex(buf.Bytes())
		fmt.Println("Masternode transaction data:", d)
	}

	joinCh := make(chan *contract.ContractJoin, 32)
	quitCh := make(chan *contract.ContractQuit, 32)
	joinSub, err1 := mm.contract.WatchJoin(nil, joinCh)
	if err1 != nil {
		// TODO: exit
		return
	}
	quitSub, err2 := mm.contract.WatchQuit(nil, quitCh)
	if err2 != nil {
		// TODO: exit
		return
	}

	ping := time.NewTimer(masternode.MASTERNODE_PING_INTERVAL)
	check := time.NewTimer(masternode.MASTERNODE_CHECK_INTERVAL)

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()
	voting := time.NewTicker(masternode.MASTERNODE_VOTING_ENABLE)
	defer voting.Stop()

	for {
		select {
		case join := <-joinCh:
			node, err := mm.masternodes.NodeJoin(join.Id)
			if err == nil {
				if bytes.Equal(join.Id[:], mm.srvr.Self().ID[0:8]) {
					mm.updateActiveMasternode()
					// TODO
					err := mm.Register(node)
					if err != nil {
						fmt.Println("err when register ", err)
					}
					mm.active.Account = node.Account
				} else {
					mm.srvr.AddPeer(node.Node)
				}
				mm.masternodes.Show()
			}

		case quit := <-quitCh:
			nodeid := common.Bytes2Hex(quit.Id[:])
			fmt.Println("quit", nodeid)
			node := mm.masternodes.Node(nodeid)
			if node != nil {
				err := mm.Unregister(node)
				if err != nil {
					fmt.Println("err when register ", err)
				}
				mm.masternodes.NodeQuit(quit.Id)
				if bytes.Equal(quit.Id[:], mm.srvr.Self().ID[0:8]) {
					mm.updateActiveMasternode()
				}
			}

			mm.masternodes.Show()

		case err := <-joinSub.Err():
			joinSub.Unsubscribe()
			fmt.Println("eventJoin err", err.Error())
		case err := <-quitSub.Err():
			quitSub.Unsubscribe()
			fmt.Println("eventQuit err", err.Error())

		case <-ping.C:
			if mm.active.State() != masternode.ACTIVE_MASTERNODE_STARTED {
				break
			}
			msg, err := mm.active.NewPingMsg()
			if err != nil {
				log.Error("NewPingMsg", "error", err)
				break
			}
			peers := mm.peers.peers
			for _, peer := range peers {
				fmt.Println("sending ping msg", "peer", peer.id)
				if err := peer.SendMasternodePing(msg); err != nil {
					log.Error("SendMasternodePing", "error", err)
					break
				}
			}
			go mm.PostPingEvent(msg)
			ping.Reset(masternode.MASTERNODE_PING_INTERVAL)
		case <-check.C:
			mm.masternodes.Check()
			check.Reset(masternode.MASTERNODE_CHECK_INTERVAL)
		case <-report.C:
			for _, vote := range mm.votes {
				if time.Since(mm.beats[vote.Hash()]) > mm.Lifetime {
					log.Debug("clean vote pool", "hash", vote.Hash())
					mm.RemoveVote(vote)
				}
			}
		}
	}
}

func (mm *MasternodeManager) ProcessPingMsg(pm *masternode.PingMsg) error {
	if mm.masternodes == nil {
		return nil
	}
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

	// mark the ping message
	for _, v := range mm.peers.peers { //
		v.markPingMsg(id, pm.Time)
	}
	mm.masternodes.RecvPingMsg(id, pm.Time)
	return nil
}

func (mm *MasternodeManager) updateActiveMasternode() {
	var state int

	n := mm.masternodes.Node(mm.active.ID)
	if n == nil {
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
		//} else if int(n.Node.TCP) != mm.active.Addr.Port {
		//	log.Error("updateActiveMasternode", "Port", n.Node.TCP, "active.Port", mm.active.Addr.Port)
		//	state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
		//} else if !n.Node.IP.Equal(mm.active.Addr.IP) {
		//	log.Error("updateActiveMasternode", "IP", n.Node.IP, "active.IP", mm.active.Addr.IP)
		//	state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else {
		state = masternode.ACTIVE_MASTERNODE_STARTED
	}
	mm.active.SetState(state)
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

func (self *MasternodeManager) PostVoteEvent(vote *types.Vote) {
	self.voteFeed.Send(core.NewVoteEvent{vote})
}

// SubscribeVoteEvent registers a subscription of VoteEvent and
// starts sending event to the given channel.
func (self *MasternodeManager) SubscribeVoteEvent(ch chan<- core.NewVoteEvent) event.Subscription {
	return self.scope.Track(self.voteFeed.Subscribe(ch))
}

func (self *MasternodeManager) Register(masternode *masternode.Masternode) error {
	return self.devoteProtocol.Register(masternode.ID)
}

func (self *MasternodeManager) Unregister(masternode *masternode.Masternode) error {
	return self.devoteProtocol.Unregister(masternode.ID)
}

func (self *MasternodeManager) MasternodeList(number *big.Int) ([]string, error) {
	return masternode.GetIdsByBlockNumber(self.contract, number)
}

func (self *MasternodeManager) PostPingEvent(pingMsg *masternode.PingMsg) {
	self.pingFeed.Send(core.PingEvent{pingMsg})
}

// SubscribePingEvent registers a subscription of PingEvent and
// starts sending event to the given channel.
func (self *MasternodeManager) SubscribePingEvent(ch chan<- core.PingEvent) event.Subscription {
	return self.scope.Track(self.pingFeed.Subscribe(ch))
}
