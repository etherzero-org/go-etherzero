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
)
// API is a user facing RPC API to allow controlling the delegate and voting
// mechanisms of the delegated-proof-of-stake
type API struct {
	chain consensus.ChainReader
	devote  *Devote
}


// GetWitnesses retrieves the list of the Witnesses at specified block
func (api *API) GetWitnesses(number *rpc.BlockNumber) ([]string, error) {
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	if header == nil {
		return nil, errUnknownBlock
	}
	currentcycle:=header.Time.Uint64()/params.Epoch
	devoteDB,_:=devotedb.New(devotedb.NewDatabase(api.devote.db),header.Protocol.CycleHash,header.Protocol.StatsHash)
	witnesses, err := devoteDB.GetWitnesses(currentcycle)
	if err != nil {
		return nil, err
	}
	return witnesses, nil
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
