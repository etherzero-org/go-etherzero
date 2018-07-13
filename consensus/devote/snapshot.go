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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
	"sync"
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

func (self *Controller) masternodes(isFirstCycle bool) (nodes map[string]*big.Int, err error) {
	self.mu.Lock()
	defer  self.mu.Unlock()

	currentCycle := self.TimeStamp / params.CycleInterval

	nodes = make(map[string]*big.Int)
	masternodeTrie := self.devoteProtocol.MasternodeTrie()
	it := trie.NewIterator(masternodeTrie.NodeIterator(nil))

	for it.Next() {
		nodes[string(it.Key)] = big.NewInt(1)
	}
	if !isFirstCycle {
		voteCntTrie := self.devoteProtocol.VoteCntTrie()
		itvote := trie.NewIterator(voteCntTrie.NodeIterator(nil))

		for itvote.Next() {
			masternodeId := itvote.Key
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(currentCycle))
			key = append(key, masternodeId...)
			fmt.Printf("add masternodes Id:%v Account:%x  ,key:%x \n", string(itvote.Key), common.BytesToAddress(itvote.Value), key)
			vote := new(types.Vote)
			if voteCntBytes := self.devoteProtocol.VoteCntTrie().Get(key); voteCntBytes != nil {
				if err := rlp.Decode(bytes.NewReader(voteCntBytes), vote); err != nil {
					log.Error("Invalid Vote body RLP", "masternodeId", masternodeId, "err", err)
					return nil, err
				}
				log.Info("vote is not nil ", "hash", vote.Hash(), "poll", vote.Poll, "masternodeid", vote.Masternode)

				score, ok := nodes[vote.Masternode]
				if !ok {
					score = big.NewInt(0)
				}
				score.Add(score, big.NewInt(1))
				nodes[vote.Masternode] = score
			}
		}
	}
	fmt.Printf("controller nodes context:%v \n", nodes)
	return nodes, nil
}

//when a node does't work in the current cycle, delete.
func (ec *Controller) uncast(cycle int64) error {

	witnesses, err := ec.devoteProtocol.GetWitnesses()
	if err != nil {
		return fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return errors.New("no witness could be uncast")
	}
	cycleDuration := params.CycleInterval
	// First cycle duration may lt cycle interval,
	// while the first block time wouldn't always align with cycle interval,
	// so caculate the first cycle duartion with first block time instead of cycle interval,
	// prevent the validators were uncast incorrectly.
	if ec.TimeStamp-timeOfFirstBlock < params.CycleInterval {
		cycleDuration = ec.TimeStamp - timeOfFirstBlock
	}
	needUncastWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(cycle))
		// TODO
		key = append(key, witness.Addr.Bytes()...)
		size := uint64(0)
		if cntBytes := ec.devoteProtocol.MinerRollingTrie().Get(key); cntBytes != nil {
			size = binary.BigEndian.Uint64(cntBytes)
		}
		if size < cycleDuration/params.BlockInterval/maxWitnessSize/2 {
			// not active witnesses need uncast
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness.ID, witness.Addr, big.NewInt(int64(size))})
		}
	}
	// no witnessees need uncast
	needUncastWitnessCnt := len(needUncastWitnesses)
	if needUncastWitnessCnt <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(needUncastWitnesses))
	masternodeCount := 0
	iter := trie.NewIterator(ec.devoteProtocol.MasternodeTrie().NodeIterator(nil))
	for iter.Next() {
		masternodeCount++
		if masternodeCount >= needUncastWitnessCnt+int(safeSize) {
			break
		}
	}
	for i, witness := range needUncastWitnesses {
		// ensure witness count greater than or equal to safeSize
		if masternodeCount <= int(safeSize) {
			log.Info("No more masternode can be uncast", "prevCycleID", cycle, "masternodeCount", masternodeCount,
				"needUncastCount", len(needUncastWitnesses)-i)
			return nil
		}
		if err := ec.devoteProtocol.Unregister(witness.address); err != nil {
			return err
		}
		// if uncast success, masternode Count minus 1
		masternodeCount--
		log.Info("uncast masternode", "prevCycleID", cycle, "witness", witness.address.String(), "miner count", witness.weight.String())
	}
	return nil
}

func (ec *Controller) lookup(now uint64) (witness string, err error) {

	witness = ""
	offset := now % params.CycleInterval
	if offset%params.BlockInterval != 0 {
		return "", ErrInvalidMinerBlockTime
	}
	offset /= params.BlockInterval

	witnesses, err := ec.devoteProtocol.GetWitnesses()
	if err != nil {
		return "", err
	}
	witnessSize := len(witnesses)
	if witnessSize == 0 {
		return "", errors.New("failed to lookup witness")
	}
	offset %= uint64(witnessSize)
	fmt.Printf("current witnesses offset%d ,id:%s,count %d value:%v\n", offset, witnesses[offset].ID, len(witnesses), witnesses)
	id := witnesses[offset].ID
	return id, nil
}

func (self *Controller) election(genesis, first, parent *types.Header) error {

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
	fmt.Printf("election prevCycle :%d ,currentCycle:%d\n", prevCycle, currentCycle)

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		if !prevCycleIsGenesis && it.Next() {
			//if err := self.uncast(prevCycle); err != nil {
			//	return err
			//}
		}
		votes, err := self.masternodes(prevCycleIsGenesis)
		if err != nil {
			log.Error("get masternodes ", "err", err)
			return err
		}
		masternodes := sortableAddresses{}
		for masternode, cnt := range votes {
			masternodes = append(masternodes, &sortableAddress{id: masternode, address: common.Address{}, weight: cnt})
		}
		if len(masternodes) < int(safeSize) {
			return errors.New("too few masternodes")
		}
		sort.Sort(masternodes)
		if len(masternodes) > int(maxWitnessSize) {
			masternodes = masternodes[:maxWitnessSize]
		}
		// disrupt the mastrnodes node to ensure the disorder of the node
		seed := uint64(binary.LittleEndian.Uint32(crypto.Keccak512(parent.Hash().Bytes()))) + i
		r := rand.New(rand.NewSource(int64(seed)))
		for i := len(masternodes) - 1; i > 0; i-- {
			j := int(r.Int31n(int32(i + 1)))
			masternodes[i], masternodes[j] = masternodes[j], masternodes[i]
		}
		var sortedWitnesses []*params.Account
		for _, masternode_ := range masternodes {
			singlesortedWitnesses := &params.Account{ID: masternode_.id, Addr: masternode_.address}
			sortedWitnesses = append(sortedWitnesses, singlesortedWitnesses)
		}
		cycleTrie, _ := types.NewCycleTrie(common.Hash{}, self.devoteProtocol.DB())
		self.devoteProtocol.SetCycle(cycleTrie)
		self.devoteProtocol.SetWitnesses(sortedWitnesses)
		log.Info("Come to new cycle", "prev", i, "next", i+1)
	}
	return nil
}

func (self *Controller) ApplyVote(votes []*types.Vote) error {

	for i, vote := range votes {
		masternodeBytes := []byte(vote.Masternode)
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(vote.Cycle))
		key = append(key, masternodeBytes...)

		voteCntInTrieBytes := self.devoteProtocol.VoteCntTrie().Get(key)
		if voteCntInTrieBytes != nil {
			log.Error("vote already exists vote hash:", "hash:", votes[i].Hash())
			continue
		}
		voteRLP, err := rlp.EncodeToBytes(vote)
		if err != nil {
			return err
		}

		votecnttrie := self.devoteProtocol.VoteCntTrie()
		fmt.Printf("before votecnttrie root hash:%x\n", votecnttrie.Hash())
		votecnttrie.TryUpdate(key, voteRLP)
		fmt.Printf("after  votecnttrie root hash:%x\n", votecnttrie.Hash())
		// update votecnt trie event
		fmt.Printf("controller ApplyVote vote end\n")
	}
	return nil
}

// Voting save the vote result to the desk
func (self *Controller) Process(vote *types.Vote) error {

	masternodeBytes := []byte(vote.Masternode)
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(vote.Cycle))
	key = append(key, masternodeBytes...)

	voteCntInTrieBytes := self.devoteProtocol.VoteCntTrie().Get(key)
	if voteCntInTrieBytes != nil {
		log.Error("vote already exists")
		return errors.New("vote already exists")
	}
	voteRLP, err := rlp.EncodeToBytes(vote)
	if err != nil {
		return err
	}
	voteCntTrie, _ := types.NewVoteCntTrie(common.Hash{}, self.devoteProtocol.DB())
	voteCntTrie.TryUpdate(key, voteRLP)
	// update votecnt trie event
	self.devoteProtocol.SetVoteCnt(voteCntTrie)
	return nil
}

type sortableAddress struct {
	id      string
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
