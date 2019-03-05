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

//
//import (
//	"net"
//	"sync"
//	"crypto/ecdsa"
//	"errors"
//
//	"github.com/etherzero/go-etherzero/p2p"
//	"github.com/etherzero/go-etherzero/crypto"
//	"github.com/etherzero/go-etherzero/common"
//
//	"fmt"
//)
//
//const (
//	ACTIVE_MASTERNODE_INITIAL         = 0 // initial state
//	ACTIVE_MASTERNODE_SYNCING         = 2
//	ACTIVE_MASTERNODE_NOT_CAPABLE     = 3
//	ACTIVE_MASTERNODE_STARTED         = 4
//)
//
//// ErrUnknownMasternode is returned for any requested operation for which no backend
//// provides the specified masternode.
//var ErrUnknownMasternode = errors.New("unknown masternode")
//
////Responsible for activating the Masternode and pinging the network
//type ActiveMasternode struct {
//	ID          string
//	NodeID      [64]byte
//	NodeAccount common.Address
//	PrivateKey  *ecdsa.PrivateKey
//	activeState int
//	Addr        net.TCPAddr
//
//	mu sync.RWMutex
//}
//
//func NewActiveMasternode(srvr *p2p.Server) *ActiveMasternode {
//	x8 := srvr.Self().X8()
//	id := fmt.Sprintf("%x", x8[:])
//	am := &ActiveMasternode{
//		ID:          id,
//		NodeID:      srvr.Self().XY(),
//		activeState: ACTIVE_MASTERNODE_INITIAL,
//		PrivateKey:  srvr.Config.PrivateKey,
//		NodeAccount: crypto.PubkeyToAddress(srvr.Config.PrivateKey.PublicKey),
//	}
//	return am
//}
//
//func (am *ActiveMasternode) State() int {
//	return am.activeState
//}
//
//func (am *ActiveMasternode) SetState(state int) {
//	am.activeState = state
//}
//
//// SignHash calculates a ECDSA signature for the given hash. The produced
//// signature is in the [R || S || V] format where V is 0 or 1.
//func (a *ActiveMasternode) SignHash(id string, hash []byte) ([]byte, error) {
//	// Look up the key to sign with and abort if it cannot be found
//	a.mu.RLock()
//	defer a.mu.RUnlock()
//
//	if id != a.ID{
//		return nil, ErrUnknownMasternode
//	}
//	// Sign the hash using plain ECDSA operations
//	return crypto.Sign(hash, a.PrivateKey)
//}
