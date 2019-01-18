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
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"sync"

	"encoding/binary"
	"fmt"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/consensus"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/hashicorp/golang-lru"
	"sort"
)

const (
	Epoch  = 600
	Period = 1
)

type Snapshot struct {
	config    *params.DevoteConfig // Consensus engine parameters to fine tune behavior
	TimeStamp uint64
	devoteDB  *devotedb.DevoteDB

	mu       sync.Mutex
	sigcache *lru.ARCCache       // Cache of recent block signatures to speed up ecrecover
	Hash     common.Hash         //Block hash where the snapshot was created
	Number   uint64              //Cycle number where the snapshot was created
	Cycle    uint64              //Cycle number where the snapshot was created
	Signers  map[string]struct{} `json:"signers"` // Set of authorized signers at this moment
	Recents  map[uint64]string   // set of recent masternodes for spam protections

}

func newSnapshot(config *params.DevoteConfig, number uint64, cycle uint64, signatures *lru.ARCCache, hash common.Hash, signers []string) *Snapshot {

	snapshot := &Snapshot{
		config:   config,
		Signers:  make(map[string]struct{}),
		Recents:  make(map[uint64]string),
		Number:   number,
		Cycle:    cycle,
		Hash:     hash,
		sigcache: signatures,
	}

	for _, signer := range signers {
		snapshot.Signers[signer] = struct{}{}
	}
	return snapshot
}

// loadSnapshot loads an existing snapshot from the database.
func loadSnapshot(config *params.DevoteConfig, sigcache *lru.ARCCache, db ethdb.Database, hash common.Hash) (*Snapshot, error) {
	blob, err := db.Get(append([]byte("devote-"), hash[:]...))
	if err != nil {
		return nil, err
	}
	snap := new(Snapshot)
	if err := json.Unmarshal(blob, snap); err != nil {
		return nil, err
	}
	snap.config = config
	snap.sigcache = sigcache

	return snap, nil
}

// copy creates a deep copy of the snapshot, though not the individual votes.
func (s *Snapshot) copy() *Snapshot {
	cpy := &Snapshot{
		sigcache: s.sigcache,
		Number:   s.Number,
		Hash:     s.Hash,
		Signers:  make(map[string]struct{}),
		Recents:  make(map[uint64]string),
	}
	for signer := range s.Signers {
		cpy.Signers[signer] = struct{}{}
	}
	for block, signer := range s.Recents {
		cpy.Recents[block] = signer
	}

	return cpy
}

// apply creates a new authorization snapshot by applying the given headers to
// the original one.
func (s *Snapshot) apply(headers []*types.Header) (*Snapshot, error) {
	// Allow passing in no headers for cleaner code
	if len(headers) == 0 {
		return s, nil
	}
	// Iterate through the headers and create a new snapshot
	snap := s.copy()

	for _, header := range headers {
		// Remove any votes on checkpoint blocks
		number := header.Number.Uint64()
		if number%Epoch == 0 {
			snap.Recents = make(map[uint64]string)
		}
		// Delete the oldest signer from the recent list to allow it signing again
		if limit := uint64(len(snap.Signers)/2 + 1); number >= limit {
			delete(snap.Recents, number-limit)
		}
		// Resolve the authorization key and check against signers
		signer, err := ecrecover(header, s.sigcache)
		if err != nil {
			return nil, err
		}
		if _, ok := snap.Signers[signer]; !ok {
			return nil, errUnauthorizedSigner
		}
		if number%Epoch != 0 {
			snap.Recents[number] = signer
		}
	}
	snap.Number += uint64(len(headers))
	snap.Hash = headers[len(headers)-1].Hash()
	snap.Cycle = headers[len(headers)-1].Number.Uint64() / Epoch
	return snap, nil
}

// store inserts the snapshot into the database.
func (c *Snapshot) store(db ethdb.Database) error {
	blob, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return db.Put(append([]byte("devote-"), c.Hash[:]...), blob)
}

// signers retrieves the list of authorized signers in ascending order.
func (s *Snapshot) signers() []string {
	signers := make([]string, 0, len(s.Signers))
	for signer := range s.Signers {
		signers = append(signers, signer)
	}
	for i := 0; i < len(signers); i++ {
		for j := i + 1; j < len(signers); j++ {
			if strings.Compare(signers[i], signers[j]) > 0 {
				signers[i], signers[j] = signers[j], signers[i]
			}
		}
	}
	return signers
}

// inturn returns if a signer at a given block height is in-turn or not.
func (d *Snapshot) inturn(number uint64, signer string) bool {

	signers := d.signers()
	offset := 0
	for offset < len(signers) && signers[offset] != signer {
		offset++
	}
	return (number % uint64(len(signers))) == uint64(offset)
}

// validVote returns whether it makes sense to cast the specified vote in the
// given snapshot context (e.g. don't try to add an already authorized signer).
func (s *Snapshot) validWitness(witness string, authorize bool) bool {
	_, signer := s.Signers[witness]
	return (signer && !authorize) || (!signer && authorize)
}

func (s *Snapshot) lookup(now uint64) (witness string, err error) {

	offset := now % params.CycleInterval
	if offset%params.BlockInterval != 0 {
		err = ErrInvalidMinerBlockTime
		return
	}
	offset /= params.BlockInterval
	witnesses, err := s.devoteDB.GetWitnesses(s.devoteDB.GetCycle())
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

//when a node does't work in the current cycle, Remove from candidate nodes.
func (ec *Snapshot) uncast(cycle uint64, nodes []string) ([]string, error) {

	witnesses, err := ec.devoteDB.GetWitnesses(cycle)
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
		size = ec.devoteDB.GetStatsNumber(key)
		if size < 1 {
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness, big.NewInt(int64(size))})
		}
		log.Debug("uncast masternode", "prevCycleID", cycle, "witness", witness, "miner count", int64(size))
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
	}
	return nodes, nil
}

func (self *Snapshot) election(genesis, first, parent *types.Header, nodes []string) error {

	maxWitnessSize := uint64(21)
	safeSize := int(15)

	genesisCycle := genesis.Time.Uint64() / params.CycleInterval
	prevCycle := parent.Time.Uint64() / params.CycleInterval
	currentCycle := self.TimeStamp / params.CycleInterval

	prevCycleIsGenesis := (prevCycle == genesisCycle)
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}
	prevCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevCycleBytes, uint64(prevCycle))

	for i := prevCycle; i < currentCycle; i++ {
		// if prevCycle is not genesis, uncast not active masternode
		list := make([]string, len(nodes))
		copy(list, nodes)
		if !prevCycleIsGenesis {
			list, _ = self.uncast(prevCycle, nodes)
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
		if len(masternodes) < safeSize {
			return fmt.Errorf(" too few masternodes current :%d, safesize:%d", len(masternodes), safeSize)
		}
		sort.Sort(masternodes)
		if len(masternodes) > int(maxWitnessSize) {
			masternodes = masternodes[:maxWitnessSize]
		}
		var sortedWitnesses []string
		for _, node := range masternodes {
			sortedWitnesses = append(sortedWitnesses, node.nodeid)
		}
		log.Info("Controller election witnesses ", "currentCycle", currentCycle, "sortedWitnesses", sortedWitnesses)
		self.devoteDB.SetWitnesses(currentCycle, sortedWitnesses)
		self.devoteDB.Commit()
		log.Info("Initializing a new cycle", "witnesses count", len(sortedWitnesses), "prev", i, "next", i+1)
	}
	return nil
}

// masternodes return  masternode list in the Cycle.
// key   -- nodeid
// value -- votes count

func (self *Snapshot) masternodes(parent *types.Header, isFirstCycle bool, nodes []string) (map[string]*big.Int, error) {
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

func GenerateProtocol(chain consensus.ChainReader, header *types.Header, db *devotedb.DevoteDB, nodes []string) *devotedb.DevoteProtocol {

	if chain.Config().IsDevote(header.Number) {
		protocol := &devotedb.DevoteProtocol{
			CycleHash: header.Root,
			StatsHash: header.Root,
		}
		return protocol
	} else {
		snap := &Snapshot{
			devoteDB:  db,
			TimeStamp: header.Time.Uint64(),
		}
		parent := chain.GetHeaderByHash(header.ParentHash)
		genesis := chain.GetHeaderByNumber(0)
		first := chain.GetHeaderByNumber(1)
		snap.election(genesis, first, parent, nodes)
		//miner Rolling
		log.Debug("rolling ", "Number", header.Number, "parnetTime", parent.Time.Uint64(),
			"headerTime", header.Time.Uint64(), "witness", header.Witness)
		db.Rolling(parent.Time.Uint64(), header.Time.Uint64(), header.Witness)
		db.Commit()
		return db.Protocol()

	}
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
}
