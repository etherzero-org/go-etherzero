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
	"errors"
	"math/big"
	"sync"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/log"
)

var (
	errClosed            = errors.New("masternode set is closed")
	errAlreadyRegistered = errors.New("masternode is already registered")
	errNotRegistered     = errors.New("masternode is not registered")
)

const nodeIDBits = 512

const (
	masternodePreEnabled = 0x00
	masterNodeEnabled    = 0x01
	masterNodeExpired    = 0x02
	masterNodeVinSpent   = 0x03
	masterNodeRemove     = 0x04
	masterNodePOSError   = 0x05
)

const (
	coliateralOK            = 0x00
	coliateralInvalidAmount = 0x01
)

// NodeID is a unique identifier for each node.
// The node identifier is a marshaled elliptic curve public key.
type NodeID [nodeIDBits / 8]byte

type Masternode struct {
	id common.Hash `json:"id"` //
	//the enode URL of the P2P masternode running onMasternodes are the enode URL of the P2P nodes running on
	// the Masternode Etherzero network.
	url 			string 		   `json:"url"`

	name            string         `json:"name"`            // Name of the Masternode, just as a alise
	Address         common.Address `json:"address"`         // Ethereum Masternode account address derived from the key
	activeState     int            `json:"activestate"`     //Masternode active state
	protocalVersion int            `json:"protocalversion"` //Masternode protocalVersion
	lastDsq         int            `json:"lastdsq"`         //
	timeLastChecked int            `json:"timelastchecked"` //
	timeLasttxid    common.Hash    `json:"timeLasttxid"`    //
	timeLastPing    int            `json:"timelastping"`    //
	infoValid       bool           `json:"infovalid"`       //

	head common.Hash `json:"head"`

	log log.Logger

	lock sync.RWMutex

	td *big.Int

	knownTxs map[common.Hash]*types.Transaction // All currently processable transactions

}

func NewMasternode(id common.Hash, name string, url string, activeState int, protocalVersion int) *Masternode {

	return &Masternode{
		id:              id,
		name:            name,
		activeState:     activeState,
		url:             url,
		protocalVersion: protocalVersion,
		knownTxs:        make(map[common.Hash]*types.Transaction),
	}
}

func (m *Masternode) ID() common.Hash {
	return m.id
}

func (m *Masternode) URL() string {
	return m.url
}

func (m *Masternode) Name() string {
	return m.name
}

func (m *Masternode) ActiveState() int {
	return m.activeState
}

func (m *Masternode) ProtocalVersion() int {
	return m.protocalVersion
}

func (m *Masternode) TimeLastChecked() int {
	return m.timeLastChecked
}

func (m *Masternode) TimeLasttxid() common.Hash {
	return m.timeLasttxid
}

func (m *Masternode) TimeLastPing() int {
	return m.timeLastPing
}

func (m *Masternode) LastDsq() int {
	return m.lastDsq
}

func (m *Masternode) setID(id common.Hash) {
	m.id = id
}

func (m *Masternode) setURL(url string) {
	m.url = url
}

func (m *Masternode) SetLastDsq(lastdsq int) {
	m.lastDsq = lastdsq
}

func (m *Masternode) SetTimeLastPing(timeLastPing int) {
	m.timeLastPing = timeLastPing
}

func (m *Masternode) SetTimeLastChecked(timeLastChecked int) {
	m.timeLastChecked = timeLastChecked
}

func (m *Masternode) SetTimeLasttxid(timeLasttxid common.Hash) {
	m.timeLasttxid = timeLasttxid
}

func (m *Masternode) SetProtocalVersion(protocalVersion int) {
	m.protocalVersion = protocalVersion
}

func (m *Masternode) SetActiveState(activeState int) {
	m.activeState = activeState
}

// Head retrieves a copy of the current head hash and total difficulty of the
// masternode.
func (m *Masternode) Head() (hash common.Hash, td *big.Int) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	copy(hash[:], m.head[:])
	return hash, new(big.Int).Set(m.td)
}

// SetHead updates the head hash and total difficulty of the masternode.
func (m *Masternode) SetHead(hash common.Hash, td *big.Int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	copy(m.head[:], hash[:])
	m.td.Set(td)
}

type MasternodeCofing struct {
}

// masternodeSet represents the collection of masternode currently participating in
// the Etherzero  masternode-protocol.
type masternodeSet struct {
	masternodes map[string]*Masternode
	lock        sync.RWMutex
	closed      bool
}

// newMasternodeSet creates a new peer set to track the active participants.
func newMasternodeSet() *masternodeSet {
	return &masternodeSet{
		masternodes: make(map[string]*Masternode),
	}
}

// Len returns if the current number of peers in the set.
func (ms *masternodeSet) Len() int {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	return len(ms.masternodes)
}

// Register injects a new peer into the working set, or returns an error if the
// Masternode is already known.
func (ms *masternodeSet) Register(m *Masternode) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	if ms.closed {
		return errClosed
	}
	if _, ok := ms.masternodes[m.name]; ok {
		return errAlreadyRegistered
	}
	ms.masternodes[m.name] = m
	return nil
}

// Unregister removes a remote Masternode from the active set, disabling any further
// actions to/from that particular entity.
func (ms *masternodeSet) Unregister(id string) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	if _, ok := ms.masternodes[id]; !ok {
		return errNotRegistered
	}
	delete(ms.masternodes, id)
	return nil
}

// Masternode retrieves the registered masternode with the given id.
func (mn *masternodeSet) Masternode(id string) *Masternode {
	mn.lock.RLock()
	defer mn.lock.RUnlock()

	return mn.masternodes[id]
}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ms *masternodeSet) BestMasternode() *Masternode {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	var (
		bestMasternode *Masternode
		bestTd         *big.Int
	)
	for _, m := range ms.masternodes {
		if _, td := m.Head(); bestMasternode == nil || td.Cmp(bestTd) > 0 {
			bestMasternode, bestTd = m, td
		}
	}
	return bestMasternode
}
