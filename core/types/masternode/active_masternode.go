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

//The Masternode Class. For managing the InstantTX process. It contains the input of the 20000ETZ, signature to prove
// it's the one who own that ip address and code for calculating the payment election.

package masternode

import (
	"crypto/ecdsa"
	"github.com/ethzero/go-ethzero/p2p"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"net"
	"time"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/crypto"
	"encoding/binary"
	"github.com/ethzero/go-ethzero/common"
)

const (
	ACTIVE_MASTERNODE_INITIAL         = 0 // initial state
	ACTIVE_MASTERNODE_SYNC_IN_PROCESS = 1
	ACTIVE_MASTERNODE_INPUT_TOO_NEW   = 2
	ACTIVE_MASTERNODE_NOT_CAPABLE     = 3
	ACTIVE_MASTERNODE_STARTED         = 4
)

//Responsible for activating the Masternode and pinging the network
type ActiveMasternode struct {
	ID          string
	NodeID      discover.NodeID
	Account     common.Address
	PrivateKey  *ecdsa.PrivateKey
	activeState int
	Addr        net.TCPAddr
}

func NewActiveMasternode(srvr *p2p.Server, mns *MasternodeSet) *ActiveMasternode {
	nodeId := srvr.Self().ID
	id := GetMasternodeID(nodeId)
	am := &ActiveMasternode{
		ID:          id,
		NodeID:      nodeId,
		activeState: ACTIVE_MASTERNODE_INITIAL,
		PrivateKey:  srvr.Config.PrivateKey,
		Addr:        srvr.MasternodeAddr,
	}
	if n := mns.Node(id); n != nil {
		am.Account = n.Account
	}
	return am
}

func (am *ActiveMasternode) State() int {
	return am.activeState
}

func (am *ActiveMasternode) SetState(state int) {
	am.activeState = state
}

func (am *ActiveMasternode) NewPingMsg() (*PingMsg, error) {
	sec := uint64(time.Now().Unix())
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], sec)
	sig, err := crypto.Sign(crypto.Keccak256(b[:]), am.PrivateKey)
	if err != nil {
		log.Error("Can't sign PingMsg packet", "err", err)
		return nil, err
	}
	return &PingMsg{
		Time: sec,
		Sig: sig,
	}, nil
}
