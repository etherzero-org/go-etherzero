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
	"fmt"
	"math/big"
	"time"

	"crypto/ecdsa"
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/contracts/masternode/contract"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p/discv5"
	"github.com/etherzero/go-etherzero/p2p/enode"
)

const (
	MasternodeInit = iota
)

const (
	MASTERNODE_PING_INTERVAL = 1200 * time.Second
)

var (
	errClosed            = errors.New("masternode set is closed")
	errAlreadyRegistered = errors.New("masternode is already registered")
	errNotRegistered     = errors.New("masternode is not registered")
)

type Masternode struct {
	ENode *enode.Node

	ID          string
	NodeID      discv5.NodeID
	Account     common.Address
	OriginBlock *big.Int
	State       int

	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
}

func newMasternode(nodeId discv5.NodeID, account common.Address, block, blockOnlineAcc, blockLastPing *big.Int) *Masternode {

	id := GetMasternodeID(nodeId)
	p := &ecdsa.PublicKey{Curve: crypto.S256(), X: new(big.Int), Y: new(big.Int)}
	p.X.SetBytes(nodeId[:32])
	p.Y.SetBytes(nodeId[32:])
	if !p.Curve.IsOnCurve(p.X, p.Y) {
		return &Masternode{}
	}
	node := enode.NewV4(p, nil, 0, 0)
	return &Masternode{
		ENode:          node,
		ID:             id,
		NodeID:         nodeId,
		Account:        account,
		OriginBlock:    block,
		State:          MasternodeInit,
		BlockOnlineAcc: blockOnlineAcc,
		BlockLastPing:  blockLastPing,
	}
}

func (n *Masternode) String() string {
	return fmt.Sprintf("Node: %s\n", n.NodeID.String())
}

func GetGovernanceAddress(contract *contract.Contract, blockNumber *big.Int) (common.Address, error) {
	if blockNumber == nil {
		blockNumber = new(big.Int)
	}
	opts := new(bind.CallOpts)
	opts.BlockNumber = blockNumber
	addr, err := contract.GovernanceAddress(opts)
	return addr, err
}

func GetIdsByBlockNumber(contract *contract.Contract, blockNumber *big.Int) ([]string, error) {
	if blockNumber == nil {
		blockNumber = new(big.Int)
	}
	opts := new(bind.CallOpts)
	opts.BlockNumber = blockNumber
	var (
		lastId [8]byte
		ctx    *MasternodeContext
		ids    []string
	)
	lastId, err := contract.LastId(opts)
	if err != nil {
		return ids, err
	}
	for lastId != ([8]byte{}) {
		ctx, err = GetMasternodeContext(opts, contract, lastId)
		if err != nil {
			log.Error("GetIdsByBlockNumber", "error", err)
			break
		}
		lastId = ctx.pre
		if ctx.Node.BlockLastPing.Cmp(common.Big0) > 0 {
			if new(big.Int).Sub(blockNumber, ctx.Node.BlockLastPing).Cmp(big.NewInt(3600)) > 0 {
				continue
			}
		} else if ctx.Node.OriginBlock.Cmp(common.Big0) > 0 {
			continue
		}
		ids = append(ids, ctx.Node.ID)
	}
	if len(ids) < 21 {
		lastId, err = contract.LastId(opts)
		if err != nil {
			return ids, err
		}
		for lastId != ([8]byte{}) {
			ctx, err = GetMasternodeContext(opts, contract, lastId)
			if err != nil {
				log.Error("GetIdsByBlockNumber", "error", err)
				break
			}
			lastId = ctx.pre
			if ctx.Node.BlockLastPing.Cmp(common.Big0) > 0 {
				if new(big.Int).Sub(blockNumber, ctx.Node.BlockLastPing).Cmp(big.NewInt(3600)) <= 0 {
					repeat := false
					for _, n := range ids {
						if n == ctx.Node.ID {
							repeat = true
						}
					}
					if !repeat {
						ids = append(ids, ctx.Node.ID)
					}
				}
			}
		}
	}
	return ids, nil
}

func GetMasternodeID(ID discv5.NodeID) string {
	return fmt.Sprintf("%x", ID[:8])
}

type MasternodeContext struct {
	Node *Masternode
	pre  [8]byte
	next [8]byte
}

func GetMasternodeContext(opts *bind.CallOpts, contract *contract.Contract, id [8]byte) (*MasternodeContext, error) {

	data, err := contract.ContractCaller.GetInfo(opts, id)
	if err != nil {
		return &MasternodeContext{}, err
	}
	id2 := append(data.Id1[:], data.Id2[:]...)
	var nodeId discv5.NodeID
	copy(nodeId[:], id2[:])
	node := newMasternode(nodeId, data.Account, data.BlockNumber, data.BlockOnlineAcc, data.BlockLastPing)

	return &MasternodeContext{
		Node: node,
		pre:  data.PreId,
		next: data.NextId,
	}, nil
}
