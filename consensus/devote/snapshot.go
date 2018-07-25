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

// Package devote implements the proof-of-stake consensus engine.

package devote

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/trie"
)

type Controller struct {
	devoteProtocol *types.DevoteProtocol
	TimeStamp      uint64
	mu             sync.Mutex
}

func Newcontroller(devoteProtocol *types.DevoteProtocol) *Controller {
	controller := &Controller{
		devoteProtocol: devoteProtocol,
	}
	return controller
}

// masternodes return  masternode list in the Cycle.
// key   -- nodeid
// value -- votes count

func (self *Controller) masternodes(parent *types.Header, isFirstCycle bool, nodes []string) (map[string]*big.Int, error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	list := make(map[string]*big.Int)
	for i := 0; i < len(nodes); i++ {
		masternode := nodes[i]
		hash := make([]byte, 8)
		hash = append(hash, []byte(masternode)...)
		hash = append(hash, parent.Hash().Bytes()...)
		weight := int64(binary.LittleEndian.Uint32(crypto.Keccak512(hash)))

		score := big.NewInt(0)
		score.Add(score, big.NewInt(weight))
		log.Debug("masternodes ", "score", score.Uint64(), "masternode", masternode)
		list[masternode] = score
	}
	log.Debug("controller nodes ", "context", nodes, "count", len(nodes))
	return list, nil
}

//when a node does't work in the current cycle, delete.
func (ec *Controller) uncast(cycle uint64, nodes []string) ([]string, error) {

	witnesses, err := ec.devoteProtocol.GetWitnesses()
	if err != nil {
		return nodes, fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return nodes, errors.New("no witness could be uncast")
	}
	needUncastWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, cycle)
		// TODO
		key = append(key, []byte(witness)...)
		size := uint64(0)
		if cntBytes := ec.devoteProtocol.MinerRollingTrie().Get(key); cntBytes != nil {
			size = binary.BigEndian.Uint64(cntBytes)
		}
		if size < 1 {
			// not active witnesses need uncast
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness, big.NewInt(int64(size))})
		}
	}
	// no witnessees need uncast
	needUncastWitnessCnt := len(needUncastWitnesses)
	if needUncastWitnessCnt <= 0 {
		return nodes, nil
	}
	for _, witness := range needUncastWitnesses {
		j := 0
		for _, s := range nodes {
			if s != witness.nodeid {
				nodes[j] = s
				j++
			}
		}
		//do sth on masternode list
		//if err := ec.devoteProtocol.Unregister(witness.nodeid); err != nil {
		//	return err
		//}
		// if uncast success, masternode Count minus 1
		log.Info("uncast masternode", "prevCycleID", cycle, "witness", witness.nodeid, "miner count", witness.weight.String())
	}
	return nodes, nil
}

func (ec *Controller) lookup(now uint64) (witness string, err error) {

	offset := now % params.CycleInterval
	if offset%params.BlockInterval != 0 {
		err = ErrInvalidMinerBlockTime
		return
	}

	offset /= params.BlockInterval
	witnesses, err := ec.devoteProtocol.GetWitnesses()
	if err != nil {
		return
	}

	witnessSize := len(witnesses)
	if witnessSize == 0 {
		err = errors.New("failed to lookup witness")
		return
	}
	offset %= uint64(witnessSize)
	witness = witnesses[offset]
	return
}

func (self *Controller) election(genesis, first, parent *types.Header, nodes []string) error {

	genesisCycle := genesis.Time.Uint64() / params.CycleInterval
	prevCycle := parent.Time.Uint64() / params.CycleInterval
	currentCycle := self.TimeStamp / params.CycleInterval

	prevCycleIsGenesis := (prevCycle == genesisCycle)
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}
	prevCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevCycleBytes, uint64(prevCycle))

	it := trie.NewIterator(self.devoteProtocol.MinerRollingTrie().NodeIterator(prevCycleBytes))
	//fmt.Printf("election prevCycle :%d ,currentCycle:%d\n", prevCycle, currentCycle)

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		list:=make([]string,len(nodes))
		copy(list,nodes)
		if !prevCycleIsGenesis && it.Next() {
			list,_ = self.uncast(prevCycle, nodes)
		}

		votes, err := self.masternodes(parent, prevCycleIsGenesis, list)
		if err != nil {
			log.Error("init masternodes ", "err", err)
			return err
		}
		masternodes := sortableAddresses{}
		for masternode, cnt := range votes {
			masternodes = append(masternodes, &sortableAddress{nodeid: masternode, weight: cnt})
		}
		if len(masternodes) < int(safeSize) {
			return fmt.Errorf(" too few masternodes ", "current", len(masternodes), "safesize", safeSize)
		}
		sort.Sort(masternodes)
		if len(masternodes) > int(maxWitnessSize) {
			masternodes = masternodes[:maxWitnessSize]
		}
		var sortedWitnesses []string
		for _, node := range masternodes {
			sortedWitnesses = append(sortedWitnesses, node.nodeid)
		}
		log.Info("Controller election witnesses ", "sortedWitnesses", sortedWitnesses)
		cycleTrie, _ := types.NewCycleTrie(common.Hash{}, self.devoteProtocol.DB())
		self.devoteProtocol.SetCycle(cycleTrie)
		self.devoteProtocol.SetWitnesses(sortedWitnesses)
		log.Info("Initializing a new cycle", "witnesses count", len(sortedWitnesses), "prev", i, "next", i+1)
	}
	return nil
}

// nodeid  masternode nodeid
// weight the number of polls for one nodeid
type sortableAddress struct {
	nodeid string
	weight *big.Int
}

type sortableAddresses []*sortableAddress

func (p sortableAddresses) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p sortableAddresses) Len() int      { return len(p) }
func (p sortableAddresses) Less(i, j int) bool {
	if p[i].weight.Cmp(p[j].weight) < 0 {
		return false
	} else if p[i].weight.Cmp(p[j].weight) > 0 {
		return true
	} else {
		return p[i].nodeid > p[j].nodeid
	}
	return true
}
