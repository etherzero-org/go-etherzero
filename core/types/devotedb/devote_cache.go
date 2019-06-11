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
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
)

type DevoteCache struct {
	witness map[uint64][]string    // cycle -> witnessess
	stats   map[common.Hash]uint64 // cycle+witness -> count

	sTrie Trie
	cTrie Trie

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr    error
	devoteDB *DevoteDB
	mu       sync.Mutex
}

func newCache(db *DevoteDB) *DevoteCache {

	list := make(map[uint64][]string)
	stats := make(map[common.Hash]uint64)

	return &DevoteCache{
		witness:  list,
		stats:    stats,
		sTrie:    db.statsTrie,
		cTrie:    db.cycleTrie,
		devoteDB: db,
	}
}

func (self *DevoteCache) setError(err error) {
	self.dbErr = err
}

func (self *DevoteCache) getStatsTrie(db Database, root common.Hash) Trie {
	if self.sTrie == nil {
		var err error
		self.sTrie, err = db.OpenTrie(root)
		if err != nil {
			self.sTrie, _ = db.OpenTrie(root)
			self.setError(fmt.Errorf("can't create trie: %v", err))
		}
	}
	return self.sTrie
}

func (self *DevoteCache) getCycleTrie(db Database, root common.Hash) Trie {
	if self.cTrie == nil {
		var err error
		self.cTrie, err = db.OpenTrie(root)
		if err != nil {
			self.cTrie, _ = db.OpenTrie(root)
			self.setError(fmt.Errorf("can't create trie: %v", err))
		}
	}
	return self.cTrie
}

func (self *DevoteCache) deepCopy(db *DevoteDB) *DevoteCache {
	cache := &DevoteCache{devoteDB: db}

	if self.sTrie != nil {
		cache.sTrie = db.db.CopyTrie(self.sTrie)
	}
	if self.cTrie != nil {
		cache.cTrie = db.db.CopyTrie(self.cTrie)
	}
	return cache
}

func (self *DevoteCache) Hash() (h common.Hash) {
	self.mu.Lock()
	defer self.mu.Unlock()

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, self.sTrie.Hash())
	rlp.Encode(hw, self.cTrie.Hash())
	hw.Sum(h[:0])
	return h
}

// update counts in MinerRollingTrie for the miner of newBlock
func (self *DevoteCache) Rolling(db Database, parentBlockTime, currentBlockTime uint64, witness string) (Trie, error) {

	currentCycle := parentBlockTime / params.Epoch
	currentCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentCycleBytes, uint64(currentCycle))

	cnt := uint64(0)
	newCycle := currentBlockTime / params.Epoch
	key := common.Hash{}
	// still during the currentCycleID
	if currentCycle == newCycle {

		key.SetBytes(append(currentCycleBytes, []byte(witness)...))
		if _, ok := self.stats[key]; ok {
			self.stats[key]++
			cnt = self.stats[key]
		} else {
			self.stats[key] = 1
		}
	}

	newCntBytes := make([]byte, 8)
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(newCycle))
	binary.BigEndian.PutUint64(newCntBytes, uint64(cnt))
	statsTrie := self.getStatsTrie(db, key)
	if statsTrie == nil {
		return nil, fmt.Errorf("can't create trie")
	}
	fmt.Printf("DevoteCache newCn%d,newCntBytes%x\n", cnt, newCntBytes)
	err := statsTrie.TryUpdate(append(newCycleBytes, []byte(witness)...), newCntBytes)
	return statsTrie, err
}

func (self *DevoteCache) GetWitnesses(db Database, cycle uint64) ([]string, error) {
	if list, exists := self.witness[cycle]; exists {
		fmt.Printf("getwitnesses list in cache:%s\n", list)
		return list, nil
	}
	return nil, fmt.Errorf("cache is nil")
}

func (self *DevoteCache) SetWitnesses(cycle uint64, witnesses []string) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.witness[cycle] = witnesses
	return nil
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (self *DevoteCache) Finalise() {

	if len(self.witness) > maxPastTries {
		temp := uint64(0)
		for cycle, _ := range self.witness {
			if temp < cycle {
				temp = cycle
			}
		}
		delete(self.witness, temp)
	}

}

func (self *DevoteCache) updateStatsTrie(db Database, root common.Hash) Trie {
	tr := self.getStatsTrie(db, root)
	for key, value := range self.stats {
		delete(self.stats, key)
		if value == 0 {
			self.setError(tr.TryDelete(key[:]))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(value)
		self.setError(tr.TryUpdate(key[:], v))
	}
	return tr
}

func (self *DevoteCache) updateCycleTrie(db Database, root common.Hash) Trie {
	tr := self.getCycleTrie(db, root)
	for key, value := range self.witness {
		delete(self.witness, key)
		newCycleBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(newCycleBytes, uint64(key))
		if value == nil {
			self.setError(tr.TryDelete(newCycleBytes))
			continue
		}
		// Encoding []byte cannot fail, ok to ignore the error.
		v, _ := rlp.EncodeToBytes(value)
		self.setError(tr.TryUpdate(newCycleBytes, v))
	}
	return tr
}

func (self *DevoteCache) CommitStatsTrie(db Database, root common.Hash) error {
	self.updateStatsTrie(db, root)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err := self.sTrie.Commit(nil)

	return err
}

func (self *DevoteCache) CommitCycleTrie(db Database, root common.Hash) error {
	self.updateCycleTrie(db, root)
	if self.dbErr != nil {
		return self.dbErr
	}
	root, err := self.cTrie.Commit(nil)

	return err
}
