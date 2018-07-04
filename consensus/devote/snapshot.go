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
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
)

type Controller struct {
	devoteProtocol *types.DevoteProtocol
	statedb        *state.StateDB
	active         *masternode.ActiveMasternode
	TimeStamp      int64

	// update loop
	postVote PostVoteFn
}

func Newcontroller(devoteProtocol *types.DevoteProtocol) *Controller {

	controller := &Controller{
		devoteProtocol: devoteProtocol,
	}
	return controller
}

// votes return  vote list in the Cycle.
func (self *Controller) votes(currentCycle int64) (votes map[common.Address]*big.Int, err error) {

	votes = map[common.Address]*big.Int{}
	cacheTrie := self.devoteProtocol.CacheTrie()
	masternodeTrie := self.devoteProtocol.MasternodeTrie()
	//voteCntTrie:=self.DevoteProtocol.VoteCntTrie()
	//statedb := self.statedb

	//iterVoteCnt := trie.NewIterator(voteCntTrie.NodeIterator(nil))
	//existVoteCnt := iterVoteCnt.Next()
	//if !existVoteCnt {
	//	return votes,errors.New("no vote count in current cycle")
	//}

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
			//account := cacheIterator.Value
			score, ok := votes[masternodeAddr]
			if !ok {
				score = new(big.Int)
			}
			//cacheAddr := common.BytesToAddress(account)
			//weight := statedb.GetBalance(cacheAddr)
			//score.Add(score, weight)
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(currentCycle))
			key = append(key, masternodeAddr.Bytes()...)

			vote := new(types.Vote)
			if voteCntBytes := self.devoteProtocol.VoteCntTrie().Get(key); voteCntBytes != nil {
				if err := rlp.Decode(bytes.NewReader(voteCntBytes), vote); err != nil {
					log.Error("Invalid Vote body RLP", "masternode", masternodeAddr, "err", err)
					return nil, err
				}
			}
			score.Add(score, big.NewInt(1))
			votes[masternodeAddr] = score
			existCache = cacheIterator.Next()
		}
		existMasternode = iterMasternode.Next()
	}
	//fmt.Printf("controller votes context:%x \n", votes)
	return votes, nil
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
		if cntBytes := ec.devoteProtocol.MinerRollingTrie().Get(key); cntBytes != nil {
			size = int64(binary.BigEndian.Uint64(cntBytes))
		}
		if size < cycleDuration/blockInterval/maxWitnessSize/2 {
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
	iter := trie.NewIterator(ec.devoteProtocol.MasternodeTrie().NodeIterator(nil))
	for iter.Next() {
		masternodeCount++
		if masternodeCount >= needUncastWitnessCnt+safeSize {
			break
		}
	}

	for i, witness := range needUncastWitnesses {
		// ensure witness count greater than or equal to safeSize
		if masternodeCount <= safeSize {
			log.Info("No more masternode can be uncast", "prevCycleID", cycle, "masternodeCount", masternodeCount, "needKickoutCount", len(needUncastWitnesses)-i)
			return nil
		}
		if err := ec.devoteProtocol.Unregister(witness.address); err != nil {
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
	if offset%blockInterval != 0 {
		return common.Address{}, ErrInvalidMinerBlockTime
	}
	offset /= blockInterval

	witnesses, err := ec.devoteProtocol.GetWitnesses()
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
	return common.HexToAddress("0xc5c5b2c89e61d8e129f5f53a6697ae3b96d04204"), nil
}

func (self *Controller) election(genesis, parent *types.Header) error {

	genesisCycle := genesis.Time.Int64() / cycleInterval
	prevCycle := parent.Time.Int64() / cycleInterval
	currentCycle := self.TimeStamp / cycleInterval

	prevCycleIsGenesis := (prevCycle == genesisCycle)
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}

	prevCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevCycleBytes, uint64(prevCycle))
	iter := trie.NewIterator(self.devoteProtocol.MinerRollingTrie().NodeIterator(prevCycleBytes))

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		if !prevCycleIsGenesis && iter.Next() {
			if err := self.uncast(prevCycle); err != nil {
				return err
			}
		}
		votes, err := self.votes(currentCycle)
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

		cycleTrie, _ := types.NewCycleTrie(common.Hash{}, self.devoteProtocol.DB())
		self.devoteProtocol.SetCycle(cycleTrie)
		self.devoteProtocol.SetWitnesses(sortedWitnesses)
		self.Voting()
		log.Info("Come to new cycle", "prev", i, "next", i+1)
	}
	return nil
}

// Process save the vote result to the desk
func (self *Controller) Voting() (*types.Vote, error) {

	currentCycle := self.TimeStamp / cycleInterval
	nextCycle := currentCycle + 1
	nextCycleVoteId := make([]byte, 8)
	binary.BigEndian.PutUint64(nextCycleVoteId, uint64(nextCycle))

	if self.active == nil{
		return nil,errors.New(" the current node is not masternode")
	}

	masternodeBytes := self.active.Account.Bytes()
	voteCntInTrieBytes := self.devoteProtocol.VoteCntTrie().Get(append(nextCycleVoteId, masternodeBytes...))
	if voteCntInTrieBytes != nil {
		return nil, errors.New("vote already exists")
	}

	masternodes, err := self.votes(currentCycle)
	if err != nil {
		return nil, err
	}
	weight := int64(0)
	best := common.Address{}
	for account, _ := range masternodes {
		hash := append(masternodeBytes, account.Bytes()...)
		temp := int64(binary.LittleEndian.Uint32(crypto.Keccak512(hash)))
		if temp > weight && account != self.active.Account {
			weight = temp
			best = account
		}
	}
	fmt.Printf("best masternode:%x\n", best)
	vote := types.NewVote(nextCycle, best, self.active.ID)
	vote.SignVote(self.active.PrivateKey)
	voteRLP, err := rlp.EncodeToBytes(vote)
	if err != nil {
		return nil, err
	}
	self.postVote(vote)
	voteCntInTrieBytes = append(append(voteCntInTrieBytes, nextCycleVoteId...), best.Bytes()...)
	fmt.Printf("controller new voteCntbytes id %s\n", voteCntInTrieBytes)
	self.devoteProtocol.VoteCntTrie().TryUpdate(voteCntInTrieBytes, voteRLP)
	return vote, nil
}

// Voting save the vote result to the desk
func (self *Controller) Process(vote *types.Vote) error {

	currentVoteId := make([]byte, 8)
	binary.BigEndian.PutUint64(currentVoteId, uint64(vote.Cycle()))

	masternodeBytes := []byte(vote.Masternode())
	voteCntInTrieBytes := self.devoteProtocol.VoteCntTrie().Get(append(currentVoteId, masternodeBytes...))
	if voteCntInTrieBytes != nil {
		return errors.New("vote already exists")
	}
	voteRLP, err := rlp.EncodeToBytes(vote)
	if err != nil {
		return err
	}
	// Broadcast the vote and update votecnt trie event
	self.postVote(vote)
	self.devoteProtocol.VoteCntTrie().TryUpdate(voteCntInTrieBytes, voteRLP)

	return nil
}

func (self *Controller) PostVote(fn PostVoteFn) {
	self.postVote = fn
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
