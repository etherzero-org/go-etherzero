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
	"github.com/etherzero/go-etherzero/params"
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

func (self *Controller) Active(activeMasternode *masternode.ActiveMasternode) {
	self.active = activeMasternode
	fmt.Printf("controller active id :%s\n", self.active.ID)
}

// masternodes return  masternode list in the Cycle.
func (self *Controller) masternodes(isFirstCycle bool) (nodes map[common.Address]*big.Int, err error) {
	currentCycle := self.TimeStamp / cycleInterval

	nodes = map[common.Address]*big.Int{}
	masternodeTrie := self.devoteProtocol.MasternodeTrie()
	it := trie.NewIterator(masternodeTrie.NodeIterator(nil))

	for it.Next() {
		if isFirstCycle {
			fmt.Printf("masternodes isFirstCycle\n")
			address := common.BytesToAddress(it.Value)
			nodes[address] = big.NewInt(0)
		} else {
			fmt.Printf("add masternodes  , masternodeId:%v  Account:%x \n", string(it.Key), common.BytesToAddress(it.Value))
			masternodeId := it.Key
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, uint64(currentCycle))
			key = append(key, masternodeId...)
			vote := new(types.Vote)
			if voteCntBytes := self.devoteProtocol.VoteCntTrie().Get(key); voteCntBytes != nil {
				fmt.Printf("vote is not nil vote hash:%x,vote account:%x\n", vote.Hash(), vote.Account())
				if err := rlp.Decode(bytes.NewReader(voteCntBytes), vote); err != nil {
					log.Error("Invalid Vote body RLP", "masternodeId", masternodeId, "err", err)
					return nil, err
				}
				score, ok := nodes[vote.Account()]
				if !ok {
					score = new(big.Int)
				}
				score.Add(score, big.NewInt(1))
				nodes[vote.Account()] = score
			}

		}
	}
	//fmt.Printf("controller nodes context:%x \n", nodes)
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
		// TODO
		key = append(key, witness.Addr.Bytes()...)
		size := int64(0)
		if cntBytes := ec.devoteProtocol.MinerRollingTrie().Get(key); cntBytes != nil {
			size = int64(binary.BigEndian.Uint64(cntBytes))
		}
		if size < cycleDuration/blockInterval/maxWitnessSize/2 {
			// not active witnesses need uncast
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness.ID, witness.Addr, big.NewInt(size)})
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
			log.Info("No more masternode can be uncast", "prevCycleID", cycle, "masternodeCount", masternodeCount, "needUncastCount", len(needUncastWitnesses)-i)
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
	fmt.Printf("current witnesses count %d\n", len(witnesses))
	account := witnesses[offset].Addr
	return account, nil

	//return common.HexToAddress("0xc5c5b2c89e61d8e129f5f53a6697ae3b96d04204"), nil
	//return common.HexToAddress("0x37f672cc4885162b520193533546253e117acd63"), nil
}

func (self *Controller) election(genesis, first, parent *types.Header) error {

	genesisCycle := genesis.Time.Int64() / cycleInterval
	prevCycle := parent.Time.Int64() / cycleInterval
	currentCycle := self.TimeStamp / cycleInterval
	firstCycle := int64(0)

	if first != nil {
		firstCycle = first.Time.Int64() / cycleInterval
	}
	isFirstCycle := currentCycle == firstCycle

	fmt.Printf("election isFirstCycle %v \n", isFirstCycle)

	prevCycleIsGenesis := (prevCycle == genesisCycle)
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}

	prevCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevCycleBytes, uint64(prevCycle))
	it := trie.NewIterator(self.devoteProtocol.MinerRollingTrie().NodeIterator(prevCycleBytes))

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		if !prevCycleIsGenesis && it.Next() {
			//if err := self.uncast(prevCycle); err != nil {
			//	return err
			//}
		}
		votes, err := self.masternodes(isFirstCycle)
		if err != nil {
			return err
		}
		masternodes := sortableAddresses{}
		for masternode, cnt := range votes {
			masternodes = append(masternodes, &sortableAddress{address: masternode, weight: cnt})

		}
		fmt.Printf("snapshot.go election masternodes %d\n", len(masternodes))

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

		var sortedWitnesses []*params.Account
		for _, masternode_ := range masternodes {

			singlesortedWitnesses := &params.Account{Addr: masternode_.address}
			sortedWitnesses = append(sortedWitnesses, singlesortedWitnesses)
		}

		cycleTrie, _ := types.NewCycleTrie(common.Hash{}, self.devoteProtocol.DB())
		self.devoteProtocol.SetCycle(cycleTrie)
		self.devoteProtocol.SetWitnesses(sortedWitnesses)
		log.Info("Come to new cycle", "prev", i, "next", i+1)
	}
	self.Voting(isFirstCycle)
	return nil
}

// Process save the vote result to the desk
func (self *Controller) Voting(isFirstCycle bool) (*types.Vote, error) {

	fmt.Printf("come to voting begin\n")
	currentCycle := self.TimeStamp / cycleInterval
	nextCycle := currentCycle + 1
	nextCycleVoteId := make([]byte, 8)
	binary.BigEndian.PutUint64(nextCycleVoteId, uint64(nextCycle))

	if self.active == nil {

		fmt.Printf("voting check active masternode failed \n")
		return nil, errors.New(" the current node is not masternode")
	}

	fmt.Printf("voting check active masternode end \n")
	masternodeBytes := self.active.ID
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(nextCycle))
	key = append(key, []byte(masternodeBytes)...)

	voteCntInTrieBytes := self.devoteProtocol.VoteCntTrie().Get(key)
	if voteCntInTrieBytes != nil {
		fmt.Printf("vote already exists!\n")
		return nil, errors.New("vote already exists")
	}
	masternodes, err := self.masternodes(isFirstCycle)
	if err != nil {
		return nil, err
	}
	weight := int64(0)
	best := common.Address{}
	for account, _ := range masternodes {
		hash := make([]byte, 8)
		binary.BigEndian.PutUint64(hash, uint64(self.TimeStamp))
		hash = append(hash, account.Bytes()...)
		temp := int64(binary.LittleEndian.Uint32(crypto.Keccak512(hash)))
		if temp > weight && account != self.active.Account {
			weight = temp
			best = account
		}
	}
	fmt.Printf("best masternode:%x\n", best)
	vote := types.NewVote(nextCycle, best, self.active.ID)
	vote.SignVote(self.active.PrivateKey)
	fmt.Printf("voting signvote end vote.sign:%x\n", vote.Sign())
	voteRLP, err := rlp.EncodeToBytes(vote)
	if err != nil {
		fmt.Printf("voting rlp.EncodeTobytes error err%x\n", err)
		return nil, err
	}
	self.postVote(vote)
	voteCntInTrieBytes = append(append(voteCntInTrieBytes, nextCycleVoteId...), best.Bytes()...)
	fmt.Printf("controller new vote hash: %x\n", vote.Hash())
	self.devoteProtocol.VoteCntTrie().TryUpdate(voteCntInTrieBytes, voteRLP)

	fmt.Printf("controller new vote save end %x\n", voteCntInTrieBytes)
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
