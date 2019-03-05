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

// Package devote implements the proof-of-authority consensus engine.
package devote

import (
	"math/big"

	"github.com/etherzero/go-etherzero/consensus"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/rpc"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/common"
)
// API is a user facing RPC API to allow controlling the delegate and voting
// mechanisms of the delegated-proof-of-stake
type API struct {
	chain consensus.ChainReader
	devote  *Devote
}

// GetSigners retrieves the list of the Witnesses at specified block
func (api *API) GetSigners(number *rpc.BlockNumber) ([]string, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	currentEpoch:=header.Time.Uint64()/params.Epoch
	devoteDB,_:=devotedb.New(devotedb.NewDatabase(api.devote.db),header.Protocol.CycleHash,header.Protocol.StatsHash)
	signers, err := devoteDB.GetWitnesses(currentEpoch)
	if err != nil {
		return nil, err
	}
	return signers, nil
}

// GetSignersByEpoch retrieves the list of the Witnesses by round
func (api *API) GetSignersByEpoch(epoch uint64) ([]string, error) {
	var header *types.Header
	header = api.chain.CurrentHeader()
	currentEpoch:=header.Time.Uint64()/params.Epoch
	if epoch > currentEpoch{
		return []string{} , nil
	}
	devoteDB,_:=devotedb.New(devotedb.NewDatabase(api.devote.db), header.Protocol.CycleHash, header.Protocol.StatsHash)
	signers, err := devoteDB.GetWitnesses(epoch)
	if err != nil {
		return nil, err
	}
	return signers, nil
}

// GetConfirmedBlockNumber retrieves the latest irreversible block
func (api *API) GetConfirmedBlockNumber() (*big.Int, error) {
	var err error
	header := api.devote.confirmedBlockHeader
	if header == nil {
		header, err = api.devote.loadConfirmedBlockHeader(api.chain)
		if err != nil {
			return nil, err
		}
	}
	return header.Number, nil
}

// GetSnapshot retrieves the state snapshot at a given block.
func (api *API) GetSnapshot(number *rpc.BlockNumber) (*Snapshot, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	// Ensure we have an actually valid block and return its snapshot
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.devote.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSnapshotAtHash retrieves the state snapshot at a given block.
func (api *API) GetSnapshotAtHash(hash common.Hash) (*Snapshot, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	return api.devote.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
}

// GetSignersAtHash retrieves the list of authorized signers at the specified block.
func (api *API) GetSignersAtHash(hash common.Hash) ([]string, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, errUnknownBlock
	}
	snap, err := api.devote.snapshot(api.chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, err
	}
	return snap.signers(), nil
}

// Proposals returns the current proposals the node tries to uphold and vote on.
func (api *API) Proposals() map[string]bool {
	api.devote.lock.RLock()
	defer api.devote.lock.RUnlock()

	proposals := make(map[string]bool)
	for signer, auth := range api.devote.proposals {
		proposals[signer] = auth
	}
	return proposals
}

// Propose injects a new authorization proposal that the signer will attempt to
// push through.
func (api *API) Propose(signer string, auth bool) {
	api.devote.lock.Lock()
	defer api.devote.lock.Unlock()

	api.devote.proposals[signer] = auth
}

// Discard drops a currently running proposal, stopping the signer from casting
// further votes (either for or against).
func (api *API) Discard(signer string) {
	api.devote.lock.Lock()
	defer api.devote.lock.Unlock()

	delete(api.devote.proposals, signer)
}
