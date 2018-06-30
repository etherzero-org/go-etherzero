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
	"math/rand"
	"sort"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/trie"
)

type Controller struct {
	DevoteProtocol *types.DevoteProtocol
	statedb        *state.StateDB

	TimeStamp int64
}

// votes return  vote list in the Cycle.
func (self *Controller) votes() (votes map[common.Address]*big.Int, err error) {

	votes = map[common.Address]*big.Int{}
	cacheTrie := self.DevoteProtocol.CacheTrie()
	masternodeTrie := self.DevoteProtocol.MasternodeTrie()
	statedb := self.statedb

	iterMasternode := trie.NewIterator(masternodeTrie.NodeIterator(nil))
	existMasternode := iterMasternode.Next()
	if !existMasternode {
		return votes, errors.New("no masternodes")
	}

	for existMasternode {

		masternode := iterMasternode.Value
		masternodeAddr := common.BytesToAddress(masternode)
		cacheIterator := trie.NewIterator(cacheTrie.NodeIterator(masternode))
		existCache := cacheIterator.Next()

		if !existCache {
			votes[masternodeAddr] = new(big.Int)
			existMasternode = iterMasternode.Next()
			continue
		}
		for existCache {
			account := cacheIterator.Value
			score, ok := votes[masternodeAddr]
			if !ok {
				score = new(big.Int)
			}
			cacheAddr := common.BytesToAddress(account)
			weight := statedb.GetBalance(cacheAddr)

			score.Add(score, weight)
			votes[masternodeAddr] = score
			existCache = cacheIterator.Next()
		}
		existMasternode = iterMasternode.Next()
	}
	fmt.Printf("controller votes context:%x \n", votes)
	return votes, nil
}

//when a node does't work in the current cycle, delete.
func (ec *Controller) uncast(cycle int64) error {

	witnesses, err := ec.DevoteProtocol.GetWitnesses()
	if err != nil {
		return fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return errors.New("no witness could be uncast")
	}

	cycleDuration := cycleInterval
	// First cycle duration may lt cycle interval,
	// while the first block time wouldn't always align with cycle interval,
	// so caculate the first cycle duartion with first block time instead of cycle interval,
	// prevent the validators were uncast incorrectly.
	if ec.TimeStamp-timeOfFirstBlock < cycleInterval {
		cycleDuration = ec.TimeStamp - timeOfFirstBlock
	}

	needUncastWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(cycle))
		key = append(key, witness.Bytes()...)
		size := int64(0)
		if cntBytes := ec.DevoteProtocol.MintCntTrie().Get(key); cntBytes != nil {
			size = int64(binary.BigEndian.Uint64(cntBytes))
		}
		if size < cycleDuration / blockInterval / maxWitnessSize / 2 {
			// not active witnesses need uncast
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness, big.NewInt(size)})
		}
	}
	// no witnessees need uncast
	needUncastWitnessCnt := len(needUncastWitnesses)
	if needUncastWitnessCnt <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(needUncastWitnesses))

	masternodeCount := 0
	iter := trie.NewIterator(ec.DevoteProtocol.MasternodeTrie().NodeIterator(nil))
	for iter.Next() {
		masternodeCount++
		if masternodeCount >= needUncastWitnessCnt+safeSize {
			break
		}
	}

	for i, witness := range needUncastWitnesses {
		// ensure witness count greater than or equal to safeSize
		if masternodeCount <= safeSize {
			log.Info("No more masternode can be kickout", "prevCycleID", cycle, "masternodeCount", masternodeCount, "needKickoutCount", len(needUncastWitnesses)-i)
			return nil
		}
		if err := ec.DevoteProtocol.Unregister(witness.address); err != nil {
			return err
		}
		// if uncast success, masternode Count minus 1
		masternodeCount--
		log.Info("uncast masternode", "prevCycleID", cycle, "witness", witness.address.String(), "mintCnt", witness.weight.String())
	}
	return nil
}

func (ec *Controller) lookup(now int64) (witness common.Address, err error) {

	witness = common.Address{}
	offset := now % cycleInterval
	if offset % blockInterval != 0 {
		return common.Address{}, ErrInvalidMinerBlockTime
	}
	offset /= blockInterval

	witnesses, err := ec.DevoteProtocol.GetWitnesses()
	if err != nil {
		return common.Address{}, err
	}
	witnessSize := len(witnesses)
	if witnessSize == 0 {
		return common.Address{}, errors.New("failed to lookup witness")
	}
	offset %= int64(witnessSize)
	//return witnesses[offset], nil
	fmt.Printf("current witnesses count %d\n", len(witnesses))
	return common.HexToAddress("0xc5d725b7d19c6c7e2c50c85fb9cf5c0b78531da7"), nil
}

func (self *Controller) voting(genesis, parent *types.Header) error {

	genesisCycle := genesis.Time.Int64() / cycleInterval
	prevCycle := parent.Time.Int64() / cycleInterval
	currentCycle := self.TimeStamp / cycleInterval

	prevCycleIsGenesis := prevCycle == genesisCycle
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}

	prevCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevCycleBytes, uint64(prevCycle))
	iter := trie.NewIterator(self.DevoteProtocol.MintCntTrie().NodeIterator(prevCycleBytes))

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		if !prevCycleIsGenesis && iter.Next() {
			if err := self.uncast(prevCycle); err != nil {
				return err
			}
		}
		votes, err := self.votes()
		if err != nil {
			return err
		}
		masternodes := sortableAddresses{}
		for masternode, cnt := range votes {
			masternodes = append(masternodes, &sortableAddress{masternode, cnt})
		}
		if len(masternodes) < safeSize {
			return errors.New("too few masternodes")
		}
		sort.Sort(masternodes)
		if len(masternodes) > maxWitnessSize {
			masternodes = masternodes[:maxWitnessSize]
		}

		// disrupt the mastrnodes node to ensure the disorder of the node
		seed := int64(binary.LittleEndian.Uint32(crypto.Keccak512(parent.Hash().Bytes()))) + i
		r := rand.New(rand.NewSource(seed))
		for i := len(masternodes) - 1; i > 0; i-- {
			j := int(r.Int31n(int32(i + 1)))
			masternodes[i], masternodes[j] = masternodes[j], masternodes[i]
		}

		sortedWitnesses := make([]common.Address, 0)
		for _, masternode := range masternodes {
			sortedWitnesses = append(sortedWitnesses, masternode.address)
		}

		cycleTrie, _ := types.NewCycleTrie(common.Hash{}, self.DevoteProtocol.DB())
		self.DevoteProtocol.SetCycle(cycleTrie)
		self.DevoteProtocol.SetWitnesses(sortedWitnesses)
		log.Info("Come to new cycle", "prev", i, "next", i+1)
	}
	return nil
}

type sortableAddress struct {
	address common.Address
	weight  *big.Int
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
		return p[i].address.String() < p[j].address.String()
	}
}
