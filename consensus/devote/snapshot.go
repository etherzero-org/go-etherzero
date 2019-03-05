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

	"encoding/json"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/hashicorp/golang-lru"
)

type Snapshot struct {
	devoteDB *devotedb.DevoteDB
	config   *params.DevoteConfig // Consensus engine parameters to fine tune behavior
	sigcache *lru.ARCCache        // Cache of recent block signatures to speed up ecrecover
	Hash     common.Hash          //Block hash where the snapshot was created
	Number   uint64               //Cycle number where the snapshot was created
	Cycle    uint64               //Cycle number where the snapshot was created
	Signers  map[string]struct{}  `json:"signers"` // Set of authorized signers at this moment
	Recents  map[uint64]string    // set of recent masternodes for spam protections

	TimeStamp uint64
	mu        sync.Mutex
}

//newSnapshot return snapshot by devoteDB
func newSnapshot(config *params.DevoteConfig, db *devotedb.DevoteDB) *Snapshot {
	snap := &Snapshot{
		config:   config,
		devoteDB: db,
		Signers:  make(map[string]struct{}),
		Recents:  make(map[uint64]string),
	}
	ary, err := db.GetWitnesses(db.GetCycle())
	if err != nil {
		log.Error("devote create Snapshot failed ", "cycle",db.GetCycle(),"err", err)
	}
	for _, s := range ary {
		snap.Signers[s] = struct{}{}
	}
	return snap
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
		// Remove any recent blocks on new cycle
		cycle := header.Time.Uint64()
		if cycle%params.Epoch == 0 {
			snap.Recents = make(map[uint64]string)
		}
		number := header.Number.Uint64()
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
			log.Error("devote apply  not in the current sigers:\n", "blockNumber", header.Number, "signer", header.Witness)
			return nil, errUnauthorizedSigner
		}
		if number%params.Epoch != 0 {
			snap.Recents[number] = signer
		}
	}
	//snap.Number += uint64(len(headers))
	snap.Number = headers[0].Number.Uint64()
	snap.Hash = headers[len(headers)-1].Hash()
	snap.Cycle = headers[len(headers)-1].Time.Uint64() / params.Epoch
	return snap, nil
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

// store inserts the snapshot into the database.
func (s *Snapshot) store(db ethdb.Database) error {
	blob, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return db.Put(append([]byte("devote-"), s.Hash[:]...), blob)
}

// masternodes return  masternode list in the Cycle.
// key   -- nodeid
// value -- votes count
func (self *Snapshot) calculate(parent *types.Header, isFirstCycle bool, nodes []string) (map[string]*big.Int, error) {
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

//Remove from candidate nodes when a node does't work in the current cycle
func (snap *Snapshot) uncast(cycle uint64, nodes []string) ([]string, error) {

	witnesses, err := snap.devoteDB.GetWitnesses(cycle)
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
		size = snap.devoteDB.GetStatsNumber(key)
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

func (snap *Snapshot) lookup(now uint64) (witness string, err error) {

	var (
		cycle uint64
	)
	offset := now % params.Epoch
	if offset%params.Period != 0 {
		err = ErrInvalidMinerBlockTime
		return
	}
	offset /= params.Period
	cycle = snap.devoteDB.GetCycle()
	witnesses, err := snap.devoteDB.GetWitnesses(cycle)
	if err != nil {
		log.Error("failed to get witness list", "cycle", cycle, "error", err)
		return
	}
	size := len(witnesses)
	if size == 0 {
		log.Error("failed to get witness list", "cycle", cycle, "error", err)
		err = errors.New("failed to lookup witness,size=0")
		return
	}
	offset %= uint64(size)
	witness = witnesses[offset]
	return
}

// signers retrieves the list of current cycle authorized signers
func (snap *Snapshot) signers() []string {
	signers := make([]string, 0, len(snap.Signers))
	for signer := range snap.Signers {
		signers = append(signers, signer)
	}
	return signers
}

func (snap *Snapshot) setSigners(ary []string) {
	for _, s := range ary {
		snap.Signers[s] = struct{}{}
	}
}

// Recording accumulating the total number of signature blocks each signer in the current cycle,return devoteDB hash
func (snap *Snapshot) recording(parent uint64, header uint64, witness string) *devotedb.DevoteProtocol {
	snap.devoteDB.Rolling(parent, header, witness)
	snap.devoteDB.Commit()
	return snap.devoteDB.Protocol()
}

//election record the current witness list into the Blockchain
func (snap *Snapshot) election(genesis, parent *types.Header, nodes []string, safeSize int, maxWitnessSize int64) ([]string, error) {

	var (
		sortedWitnesses []string
		genesiscycle    = genesis.Time.Uint64() / params.Epoch
		prevcycle       = parent.Time.Uint64() / params.Epoch
		currentcycle    = snap.TimeStamp / params.Epoch
	)
	preisgenesis := (prevcycle == genesiscycle)
	if preisgenesis && prevcycle < currentcycle {
		prevcycle = currentcycle - 1
	}
	for i := prevcycle; i < currentcycle; i++ {
		// if prevcycle is not genesis, uncast not active masternode
		list := make([]string, len(nodes))
		copy(list, nodes)
		if !preisgenesis {
			list, _ = snap.uncast(prevcycle, nodes)
		}

		count, err := snap.calculate(parent, preisgenesis, list)
		if err != nil {
			log.Error("snapshot init masternodes failed", "err", err)
			return nil, err
		}
		masternodes := sortableAddresses{}
		for masternode, cnt := range count {
			masternodes = append(masternodes, &sortableAddress{nodeid: masternode, weight: cnt})
		}
		if len(masternodes) < safeSize {
			return nil, fmt.Errorf(" too few masternodes ,cycle:%d, current :%d, safesize:%d",currentcycle, len(masternodes), safeSize)
		}
		sort.Sort(masternodes)
		if len(masternodes) > int(maxWitnessSize) {
			masternodes = masternodes[:maxWitnessSize]
		}
		for _, node := range masternodes {
			sortedWitnesses = append(sortedWitnesses, node.nodeid)
		}
		log.Debug("Initializing a new cycle ", "cycle", currentcycle, "count", len(sortedWitnesses), "sortedWitnesses", sortedWitnesses)
		snap.devoteDB.SetWitnesses(currentcycle, sortedWitnesses)
		snap.devoteDB.Commit()
	}
	return sortedWitnesses, nil
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
