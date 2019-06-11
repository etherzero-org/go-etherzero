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

//The Masternode Class. For managing the InstantTX process. It contains the input of the 20000ETZ, signature to prove
// it's the one who own that ip address and code for calculating the payment election.

package devotedb

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto/sha3"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
	"github.com/hashicorp/golang-lru"
	"github.com/etherzero/go-etherzero/log"
)

type DevoteDB struct {
	db Database //etherzero db

	statsTrie Trie //statsTrie cycle+witness +cnt
	cycleTrie Trie //cycleTrie cyele key  ->  witnesses value

	dCache *DevoteCache

	cycle             uint64 //current cycle
	txhash, blockHash common.Hash
	codeSizeCache     *lru.Cache
	mu                sync.Mutex
}

// Create a new state from a given trie.
func New(db Database, cycleRoot, statsRoot common.Hash) (*DevoteDB, error) {
	csc, _ := lru.New(codeSizeCacheSize)

	cycleTrie, cerr := db.OpenTrie(cycleRoot)
	if cerr != nil {
		return nil, cerr
	}
	statsTrie, serr := db.OpenTrie(statsRoot)
	if serr != nil {
		return nil, serr
	}

	d := &DevoteDB{
		db:            db,
		cycleTrie:     cycleTrie,
		statsTrie:     statsTrie,
		codeSizeCache: csc,
	}
	cache := newCache(d)
	d.setDevoteCache(cache)
	return d, nil
}

func NewDevoteByProtocol(db Database, protocol *DevoteProtocol) (*DevoteDB, error) {

	cycleTrie, err := db.OpenTrie(protocol.CycleHash)
	if err != nil {
		return nil, err
	}
	statsTrie, err := db.OpenTrie(protocol.StatsHash)
	if err != nil {
		return nil, err
	}
	d := &DevoteDB{
		cycleTrie: cycleTrie,
		statsTrie: statsTrie,
		db:        db,
	}
	cache := newCache(d)
	d.setDevoteCache(cache)
	return d, nil
}

func (db *DevoteDB) Database() Database {
	return db.db
}

func (db *DevoteDB) Root() (h common.Hash) {
	db.mu.Lock()
	defer db.mu.Unlock()

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, db.cycleTrie.Hash())
	rlp.Encode(hw, db.statsTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DevoteDB) Commit() (*DevoteProtocol, error) {
	cycleRoot, err := d.cycleTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.db.TrieDB().Commit(cycleRoot, false)
	statsRoot, err := d.statsTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	d.db.TrieDB().Commit(statsRoot, false)
	a := &DevoteProtocol{
		CycleHash: cycleRoot,
		StatsHash: statsRoot,
	}
	return a, nil
}

func (d *DevoteDB) Copy() *DevoteDB {
	cycleTrie := d.cycleTrie
	statsTrie := d.statsTrie

	return &DevoteDB{
		cycleTrie: cycleTrie,
		statsTrie: statsTrie,
	}
}

func (d *DevoteDB) Snapshot() *DevoteDB {
	return d.Copy()
}

func (d *DevoteDB) RevertToSnapShot(snapshot *DevoteDB) {
	d.cycleTrie = snapshot.cycleTrie
	d.statsTrie = snapshot.statsTrie
}

func (d *DevoteDB) SetCycleTrie(trie Trie) {
	d.cycleTrie = trie
}

func (d *DevoteDB) SetStatsTrie(trie Trie) {
	d.statsTrie = trie
}

func (d *DevoteDB) GetStatsNumber(key []byte) uint64 {

	hash := common.Hash{}
	hash.SetBytes(key)
	if len(d.dCache.stats) < 1 {
		if cntBytes, _ := d.statsTrie.TryGet(key); cntBytes != nil {
			count := binary.BigEndian.Uint64(cntBytes)
			return count
		}
	}
	return d.dCache.stats[hash]
}

func (d *DevoteDB) GetWitnesses(cycle uint64) ([]string, error) {
	//dc := d.dCache
	//if dc != nil {
	//	list, err := dc.GetWitnesses(d.db, cycle)
	//	if err == nil {
	//		return list, nil
	//	}
	//}
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(cycle))
	// Load from DB in case it is missing.
	witnessRLP, err := d.cycleTrie.TryGet(newCycleBytes)
	if err != nil {
		return nil, err
	}

	var witnesses []string
	if err := rlp.DecodeBytes(witnessRLP, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witnesses: %s", err)
	}
	if err != nil {
		return nil, err
	}
	//dc.SetWitnesses(cycle, witnesses)
	return witnesses, nil
}

func (d *DevoteDB) SetWitnesses(cycle uint64, witnesses []string) error {
	if d.dCache != nil {
		d.dCache.SetWitnesses(cycle, witnesses)
	}
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(cycle))
	witnessesRLP, err := rlp.EncodeToBytes(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode witnesses to rlp bytes: %s", err)
	}

	return d.cycleTrie.TryUpdate(newCycleBytes, witnessesRLP)
}

func (d *DevoteDB) setDevoteCache(cache *DevoteCache) {
	d.dCache = cache
}

func (d *DevoteDB) getDevoteCache() *DevoteCache {
	return d.dCache
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *DevoteDB) StorageCycleTrie(hash common.Hash) Trie {

	if self.dCache == nil {
		return nil
	}
	cpy := self.dCache.deepCopy(self)
	return cpy.updateCycleTrie(self.db, self.dCache.cTrie.Hash())
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *DevoteDB) StorageStatsTrie(hash common.Hash) Trie {

	if self.dCache == nil {
		return nil
	}
	cpy := self.dCache.deepCopy(self)
	return cpy.updateStatsTrie(self.db, self.dCache.sTrie.Hash())
}

func (d *DevoteDB) Rolling(parentBlockTime, currentBlockTime uint64, witness string) {

	if d.dCache == nil {
		return
	}
	currentCycle := parentBlockTime / params.Epoch
	currentCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentCycleBytes, uint64(currentCycle))

	cnt := uint64(1)
	newCycle := currentBlockTime / params.Epoch
	hash := common.Hash{}
	// still during the currentCycleID
	if currentCycle == newCycle {
		hash.SetBytes(append(currentCycleBytes, []byte(witness)...))

		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, currentCycle)
		// TODO
		key = append(key, []byte(witness)...)
		if cntBytes, _ := d.statsTrie.TryGet(key); cntBytes != nil {
			cnt = binary.BigEndian.Uint64(cntBytes) + 1
		}
	}

	newCntBytes := make([]byte, 8)
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(newCycle))
	binary.BigEndian.PutUint64(newCntBytes, uint64(cnt))

	d.statsTrie.TryUpdate(append(newCycleBytes, []byte(witness)...), newCntBytes)
}

// Exist reports whether the given Devote hash exists in the state.
// Notably this also returns true for suicided Devotes.
func (d *DevoteDB) Exists() bool {
	return d.dCache != nil
}

func (d *DevoteDB) Protocol() *DevoteProtocol {
	return &DevoteProtocol{
		CycleHash: d.cycleTrie.Hash(),
		StatsHash: d.statsTrie.Hash(),
	}
}

func (d *DevoteDB) SetCycle(cycle uint64) {
	d.cycle = cycle
}

func (d *DevoteDB) GetCycle() uint64 {
	return d.cycle
}

type DevoteProtocol struct {
	mu sync.Mutex

	CycleHash common.Hash `json:"cyclehash"  gencodec:"required"`
	StatsHash common.Hash `json:"statshash"  gencodec:"required"`
}

func (d *DevoteProtocol) Root() (h common.Hash) {
	d.mu.Lock()
	defer d.mu.Unlock()

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.CycleHash)
	rlp.Encode(hw, d.StatsHash)
	hw.Sum(h[:0])
	return h
}

type statsTrie struct {
	db *trie.Database
	*trie.SecureTrie
}

func (self statsTrie) Commit(onleaf trie.LeafCallback) (common.Hash, error) {
	var lock sync.RWMutex
	lock.Lock()
	root, err := self.SecureTrie.Commit(onleaf)
	if err != nil {
		// do sth
		log.Error("statsTrie commit was failed ", "message", err)
	}
	lock.Unlock()
	return root, err
}

func (self statsTrie) Prove(key []byte, fromLevel uint, proofDb ethdb.Putter) error {
	return self.SecureTrie.Prove(key, fromLevel, proofDb)
}

type cycleTrie struct {
	db *trie.Database
	*trie.SecureTrie
}

func (self cycleTrie) Commit(onleaf trie.LeafCallback) (common.Hash, error) {
	var lock sync.RWMutex
	lock.Lock()
	root, err := self.SecureTrie.Commit(onleaf)
	if err != nil {
		// do sth
		log.Error("cycleTrie commit was failed", "message", err)
	}
	lock.Unlock()
	return root, err
}

func (self cycleTrie) Prove(key []byte, fromLevel uint, proofDb ethdb.Putter) error {
	return self.SecureTrie.Prove(key, fromLevel, proofDb)
}
