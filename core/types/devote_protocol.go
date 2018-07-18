package types

import (
	"fmt"
	"sync"
	"encoding/binary"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto/sha3"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
)

var (
	cyclePrefix        = "cycle-"
	minerRollingPrefix = "mintCnt-"
)

type DevoteProtocol struct {
	cycleTrie        *trie.Trie
	minerRollingTrie *trie.Trie

	cycleTriedb        *trie.Database
	minerRollingTriedb *trie.Database

	diskdb ethdb.Database

	mu sync.Mutex
}

func NewCycleTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	cycleTriedb := trie.NewDatabase(ethdb.NewTable(db, cyclePrefix))
	return trie.New(root, cycleTriedb)
}


func NewMinerRollingTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	minerRollingTriedb := trie.NewDatabase(ethdb.NewTable(db, minerRollingPrefix))
	return trie.New(root, minerRollingTriedb)

}

func NewDevoteProtocolFromAtomic(db ethdb.Database, ctxAtomic *DevoteProtocolAtomic) (*DevoteProtocol, error) {
	cycleTrie, err := NewCycleTrie(ctxAtomic.CycleHash, db)
	if err != nil {
		return nil, err
	}
	minerRollingTrie, err := NewMinerRollingTrie(ctxAtomic.MinerRollingHash, db)
	if err != nil {
		return nil, err
	}
	return &DevoteProtocol{
		cycleTrie:        cycleTrie,
		minerRollingTrie: minerRollingTrie,
		diskdb:           db,

		cycleTriedb:        trie.NewDatabase(ethdb.NewTable(db, cyclePrefix)),
		minerRollingTriedb: trie.NewDatabase(ethdb.NewTable(db, minerRollingPrefix)),
	}, nil
}



func (d *DevoteProtocol) Copy() *DevoteProtocol {
	cycleTrie := *d.cycleTrie
	minerRollingTrie := *d.minerRollingTrie
	return &DevoteProtocol{
		cycleTrie:        &cycleTrie,
		minerRollingTrie: &minerRollingTrie,
	}
}

func (d *DevoteProtocol) Root() (h common.Hash) {
	d.mu.Lock()
	defer d.mu.Unlock()

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.cycleTrie.Hash())
	rlp.Encode(hw, d.minerRollingTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DevoteProtocol) Snapshot() *DevoteProtocol {
	return d.Copy()
}

func (d *DevoteProtocol) RevertToSnapShot(snapshot *DevoteProtocol) {
	d.cycleTrie = snapshot.cycleTrie
	d.minerRollingTrie = snapshot.minerRollingTrie
}

func (d *DevoteProtocol) FromAtomic(dpa *DevoteProtocolAtomic) error {
	var err error
	d.cycleTrie, err = NewCycleTrie(dpa.CycleHash, d.diskdb)
	if err != nil {
		return err
	}
	d.minerRollingTrie, err = NewMinerRollingTrie(dpa.MinerRollingHash, d.diskdb)
	return err
}

type DevoteProtocolAtomic struct {
	mu               sync.Mutex
	CycleHash        common.Hash `json:"cycleRoot"         gencodec:"required"`
	MinerRollingHash common.Hash `json:"minerRollingRoot"  gencodec:"required"`
}

func (d *DevoteProtocol) CycleTrie() *trie.Trie        { return d.cycleTrie }
func (d *DevoteProtocol) MinerRollingTrie() *trie.Trie { return d.minerRollingTrie }

func (d *DevoteProtocol) DB() ethdb.Database { return d.diskdb }

func (dc *DevoteProtocol) SetCycle(cycle *trie.Trie)              { dc.cycleTrie = cycle }
func (dc *DevoteProtocol) SetMinerRollingTrie(rolling *trie.Trie) { dc.minerRollingTrie = rolling }

func (d *DevoteProtocol) Commit(db ethdb.Database) (*DevoteProtocolAtomic, error) {
	cycleRoot, err := d.cycleTrie.CommitTo(d.cycleTriedb)
	if err != nil {
		return nil, err
	}
	d.cycleTriedb.Commit(cycleRoot, false)
	minerRollingRoot, err := d.minerRollingTrie.CommitTo(d.minerRollingTriedb)
	if err != nil {
		return nil, err
	}
	d.minerRollingTriedb.Commit(minerRollingRoot, false)
	a := &DevoteProtocolAtomic{
		CycleHash:        cycleRoot,
		MinerRollingHash: minerRollingRoot,
	}
	return a, nil
}

func (d *DevoteProtocol) ProtocolAtomic() *DevoteProtocolAtomic {
	d.mu.Lock()
	defer d.mu.Unlock()
	return &DevoteProtocolAtomic{
		CycleHash:        d.cycleTrie.Hash(),
		MinerRollingHash: d.minerRollingTrie.Hash(),
	}
}

func (p *DevoteProtocolAtomic) Root() (h common.Hash) {
	p.mu.Lock()
	defer p.mu.Unlock()

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, p.CycleHash)
	rlp.Encode(hw, p.MinerRollingHash)
	hw.Sum(h[:0])
	return h
}

func (self *DevoteProtocol) SetWitnesses(witnesses []string) error {
	key := []byte("witness")
	witnessesRLP, err := rlp.EncodeToBytes(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode witnesses to rlp bytes: %s", err)
	}
	self.cycleTrie.Update(key, witnessesRLP)
	return nil
}

func (self *DevoteProtocol) GetWitnesses() ([]string, error) {
	var witnesses []string
	key := []byte("witness")
	witnessRLP := self.cycleTrie.Get(key)
	if err := rlp.DecodeBytes(witnessRLP, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witnesses: %s", err)
	}
	return witnesses, nil
}

// update counts in MinerRollingTrie for the miner of newBlock
func (self *DevoteProtocol) Rolling(parentBlockTime, currentBlockTime uint64, witness string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	currentMinerRollingTrie := self.MinerRollingTrie()
	currentCycle := parentBlockTime / params.CycleInterval
	currentCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentCycleBytes, uint64(currentCycle))

	cnt := uint64(1)
	newCycle := currentBlockTime / params.CycleInterval
	// still during the currentCycleID
	if currentCycle == newCycle {
		iter := trie.NewIterator(currentMinerRollingTrie.NodeIterator(currentCycleBytes))
		// when current is not genesis, read last count from the MintCntTrie
		if iter.Next() {
			cntBytes := currentMinerRollingTrie.Get(append(currentCycleBytes, []byte(witness)...))
			// not the first time to mint
			if cntBytes != nil {
				cnt = binary.BigEndian.Uint64(cntBytes) + 1
			}
		}
	}

	newCntBytes := make([]byte, 8)
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(newCycle))
	binary.BigEndian.PutUint64(newCntBytes, uint64(cnt))
	self.MinerRollingTrie().TryUpdate(append(newCycleBytes, []byte(witness)...), newCntBytes)
}
