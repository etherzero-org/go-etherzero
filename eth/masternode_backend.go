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
	"github.com/etherzero/go-etherzero/consensus/devote"
	"github.com/etherzero/go-etherzero/contracts/masternode/contract"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/crypto/secp256k1"
	"github.com/etherzero/go-etherzero/event"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p"
)

type MasternodeManager struct {
	votes          map[common.Hash]*types.Vote // vote hash -> vote
	devoteProtocol *types.DevoteProtocol
	active         *masternode.ActiveMasternode
	masternodes    *masternode.MasternodeSet
	voteFeed       event.Feed
	mu             sync.Mutex
	// channels for fetcher, syncer, txsyncLoop
	newPeerCh    chan *peer
	peers        *peerSet
	IsMasternode uint32
	engine       *devote.Devote
	srvr         *p2p.Server
	contract     *contract.Contract
}

func NewMasternodeManager(dp *types.DevoteProtocol) *MasternodeManager {

	// Create the masternode manager with its initial settings
	manager := &MasternodeManager{
		votes:          make(map[common.Hash]*types.Vote),
		masternodes:    &masternode.MasternodeSet{},
		devoteProtocol: dp,
	}

	return manager
}

func (self *MasternodeManager) AddVote() {

}

func (self *MasternodeManager) process(vote *types.Vote) error {

	h := vote.Hash()
	pubkey := self.active.PrivateKey.PublicKey
	if !vote.Verify(h[:], vote.Sign(), &pubkey) {
		return errors.New("vote valid failed")
	}

	return self.devoteProtocol.Voting(vote)

}

func (self *MasternodeManager) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.votes = make(map[common.Hash]*types.Vote)
}

func (self *MasternodeManager) Has(hash common.Hash) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	if vote := self.votes[hash]; vote != nil {
		return true
	}
	return false
}

func (self *MasternodeManager) Start(srvr *p2p.Server, contract *contract.Contract, peers *peerSet) {
	self.contract = contract
	self.srvr = srvr
	self.peers = peers
	log.Trace("MasternodeManqager start ")
	mns, err := masternode.NewMasternodeSet(contract)
	if err != nil {
		log.Error("masternode.NewMasternodeSet", "error", err)
	}
	self.masternodes = mns
	self.active = masternode.NewActiveMasternode(srvr, mns)
	fmt.Printf("MasternodeManager start active MasternodeId: %v\n", self.active.ID)

	go self.masternodeLoop()
}

func (self *MasternodeManager) Stop() {

}

func (mm *MasternodeManager) masternodeLoop() {
	mm.updateActiveMasternode()
	if mm.active.State() == masternode.ACTIVE_MASTERNODE_STARTED {
		fmt.Println("masternodeCheck true")
		atomic.StoreUint32(&mm.IsMasternode, 1)
		mm.checkPeers()
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

	mm.masternodes.Show()

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

	for {
		select {
		case join := <-joinCh:
			fmt.Println("join", common.Bytes2Hex(join.Id[:]))
			node, err := mm.masternodes.NodeJoin(join.Id)
			if err == nil {
				if bytes.Equal(join.Id[:], mm.srvr.Self().ID[0:8]) {
					mm.updateActiveMasternode()
					mm.active.Account = node.Account
				} else {
					mm.srvr.AddPeer(node.Node)
				}
				mm.masternodes.Show()
			}

		case quit := <-quitCh:
			fmt.Println("quit", common.Bytes2Hex(quit.Id[:]))
			mm.masternodes.NodeQuit(quit.Id)
			if bytes.Equal(quit.Id[:], mm.srvr.Self().ID[0:8]) {
				mm.updateActiveMasternode()
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
				continue
			}
			msg, err := mm.active.NewPingMsg()
			if err != nil {
				log.Error("NewPingMsg", "error", err)
				continue
			}
			peers := mm.peers.peers
			for _, peer := range peers {
				log.Debug("sending ping msg", "peer", peer.id)
				if err := peer.SendMasternodePing(msg); err != nil {
					log.Error("SendMasternodePing", "error", err)
				}
			}
			ping.Reset(masternode.MASTERNODE_PING_INTERVAL)

		case <-check.C:
			mm.masternodes.Check()
			check.Reset(masternode.MASTERNODE_CHECK_INTERVAL)
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

func (mm *MasternodeManager) updateActiveMasternode() {
	var state int

	n := mm.masternodes.Node(mm.active.ID)
	if n == nil {
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else if int(n.Node.TCP) != mm.active.Addr.Port {
		log.Error("updateActiveMasternode", "Port", n.Node.TCP, "active.Port", mm.active.Addr.Port)
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else if !n.Node.IP.Equal(mm.active.Addr.IP) {
		log.Error("updateActiveMasternode", "IP", n.Node.IP, "active.IP", mm.active.Addr.IP)
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE
	} else {
		state = masternode.ACTIVE_MASTERNODE_STARTED
	}
	fmt.Println("22222222222222222222", state)
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
