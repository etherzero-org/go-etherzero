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

package masternode

import (
	"errors"
	"math/big"
	"net"
	"reflect"
	"sync"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/node"
	"github.com/ethzero/go-ethzero/rlp"
	"github.com/ethzero/go-ethzero/rpc"
)

var (
	errClosed            = errors.New("masternode set is closed")
	errAlreadyRegistered = errors.New("masternode is already registered")
	errNotRegistered     = errors.New("masternode is not registered")
)

// Constants to match up protocol versions and messages
const (
	etz64 = 64
)

// Node is a container on which services can be registered.
type Masternode struct {

	eventmux *event.TypeMux // Event multiplexer used between the services of a stack
	Stack    *node.Node     // Ethereum protocol stack
	account  common.Address //Masternode account information

	rpcAPIs       []rpc.API   // List of APIs currently provided by the node
	inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

	httpEndpoint  string        // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	httpWhitelist []string      // HTTP RPC modules to allow through this endpoint
	httpListener  net.Listener  // HTTP RPC listener socket to server API requests
	httpHandler   *rpc.Server   // HTTP RPC request handler to process the API requests
	stop          chan struct{} // Channel to wait for termination notifications
	lock          sync.RWMutex

	//etherzero masternode
	name string

	//last paid height
	Height *big.Int

	//protocolVersion should contain the version number of the protocol.
	protocolVersion uint

	//remember the hash of the block where masternode collateral had minimum required confirmations
	CollateralMinConfBlockHash common.Hash

	log log.Logger
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// New creates a new P2P node, ready for protocol registration.
func New(node *node.Node, name string) (*Masternode, error) {

	// Note: any interaction with Config that would create/touch files
	// in the data directory or instance directory is delayed until Start.
	return &Masternode{
		name:            name,
		Stack:           node,
		Height:          big.NewInt(-1),
		protocolVersion: etz64,
		eventmux:        new(event.TypeMux),
	}, nil
}

// Start create a live P2P node and starts running it.
func (n *Masternode) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()
	return nil
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (n *Masternode) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Terminate the API, services and the p2p server.
	n.rpcAPIs = nil
	failure := &StopError{
		Services: make(map[reflect.Type]error),
	}

	// unblock n.Wait
	close(n.stop)

	if len(failure.Services) > 0 {
		return failure
	}

	return nil
}

// Wait blocks the thread until the node is stopped. If the node is not running
// at the time of invocation, the method immediately returns.
func (n *Masternode) Wait() {
	n.lock.RLock()
	//if n.server == nil {
	//	n.lock.RUnlock()
	//	return
	//}
	stop := n.stop
	n.lock.RUnlock()

	<-stop
}

// Restart terminates a running node and boots up a new one in its place. If the
// node isn't running, an error is returned.
func (n *Masternode) Restart() error {
	if err := n.Stop(); err != nil {
		return err
	}
	if err := n.Start(); err != nil {
		return err
	}
	return nil
}

// Attach creates an RPC client attached to an in-process API handler.
func (n *Masternode) Attach() (*rpc.Client, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	//if n.server == nil {
	//	return nil, ErrNodeStopped
	//}
	return rpc.DialInProc(n.inprocHandler), nil
}

// RPCHandler returns the in-process RPC request handler.
func (n *Masternode) RPCHandler() (*rpc.Server, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if n.inprocHandler == nil {
		return nil, ErrNodeStopped
	}
	return n.inprocHandler, nil
}

// EventMux retrieves the event multiplexer used by all the network services in
// the current protocol stack.
func (n *Masternode) EventMux() *event.TypeMux {
	return n.eventmux
}


//TODO:TBA
// Deterministically calculate a given "score" for a Masternode depending on how close it's hash is to
// the proof of work for that block. The further away they are the better, the furthest will win the election
// and get paid this block
func (m *Masternode) CalculateScore(hash common.Hash) *big.Int {

	blockHash := rlpHash([]interface{}{
		hash,
		m.account,
		m.CollateralMinConfBlockHash,
	})

	return blockHash.Big()
}

// MasternodeInfo represents a short summary of the information known about the host.
type MasternodeInfo struct {
	ID              string         `json:"id"`    // Unique node identifier (also the encryption key)
	Name            string         `json:"name"`  // Name of the Masternode
	Enode           string         `json:"enode"` // Enode URL for adding this peer from remote peers
	Account         common.Address `json:"account"`
	IP              string         `json:"ip"` // IP address of the node
	ProtocolVersion uint           `json:"protocolVersion"`
	Height			*big.Int       `json:"paid"`  //last paid height
	TxHash          common.Hash    `json:"txHash"` //Send a transaction to the contract through the masternode account to prove that you own the account
	Ports           struct {
		Discovery int `json:"discovery"` // UDP listening port for discovery protocol
		Listener  int `json:"listener"`  // TCP listening port for RLPx
	} `json:"ports"`
	ListenAddr string                 `json:"listenAddr"`
	Protocols  map[string]interface{} `json:"protocols"`
}

func (m *Masternode) MasternodeInfo() *MasternodeInfo {

	node := m.Stack.Server().Self()
	srv := m.Stack.Server()

	info := &MasternodeInfo{
		Name:            m.name,
		ID:              node.ID.String(),
		IP:              node.IP.String(),
		Account:         m.account,
		Height:			 m.Height,
		ProtocolVersion: m.protocolVersion,
		ListenAddr:      srv.ListenAddr,
		Protocols:       make(map[string]interface{}),
	}
	info.Ports.Discovery = int(node.UDP)
	info.Ports.Listener = int(node.TCP)

	// Gather all the running protocol infos (only once per protocol type)
	for _, proto := range srv.Protocols {
		if _, ok := info.Protocols[proto.Name]; !ok {
			nodeInfo := interface{}("unknown")
			if query := proto.NodeInfo; query != nil {
				nodeInfo = proto.NodeInfo()
			}
			info.Protocols[proto.Name] = nodeInfo
		}
	}
	return info
}
