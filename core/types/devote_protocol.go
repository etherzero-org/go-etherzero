package types

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/binary"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto/sha3"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
)

const (
	CycleInterval = int64(3600)
)

var (
	cyclePrefix      = "cycle-"
	cachePrefix      = "cache-"
	votePrefix       = "vote-"
	masternodePrefix = "masternode-"
	mintCntPrefix    = "mintCnt-"
	voteCntPrefix    = "voteCnt-"
)

type DevoteProtocol struct {
	cycleTrie      *trie.Trie
	cacheTrie      *trie.Trie
	voteTrie       *trie.Trie
	masternodeTrie *trie.Trie
	mintCntTrie    *trie.Trie
	voteCntTrie    *trie.Trie

	cycleTriedb      *trie.Database
	cacheTriedb      *trie.Database
	voteTriedb       *trie.Database
	masternodeTriedb *trie.Database
	mintCntTriedb    *trie.Database
	voteCntTriedb    *trie.Database

	diskdb ethdb.Database
}

func NewCycleTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {

	cycleTriedb := trie.NewDatabase(ethdb.NewTable(db, cyclePrefix))
	return trie.New(root, cycleTriedb)
}

func NewCacheTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	cacheTriedb := trie.NewDatabase(ethdb.NewTable(db, cachePrefix))
	return trie.New(root, cacheTriedb)
}

func NewVoteTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	voteTriedb := trie.NewDatabase(ethdb.NewTable(db, votePrefix))
	return trie.New(root, voteTriedb)
}

func NewMasternodeTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	masternodeTriedb := trie.NewDatabase(ethdb.NewTable(db, masternodePrefix))
	return trie.New(root, masternodeTriedb)
}

func NewMintCntTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	mintCntTriedb := trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix))
	return trie.New(root, mintCntTriedb)

}

func NewVoteCntTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	voteCntTriedb := trie.NewDatabase(ethdb.NewTable(db, voteCntPrefix))
	return trie.New(root, voteCntTriedb)

}
func NewDevoteProtocol(db ethdb.Database) (*DevoteProtocol, error) {

	cycleTrie, err := NewCycleTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	cacheTrie, err := NewCacheTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	voteTrie, err := NewVoteTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	masternodeTrie, err := NewMasternodeTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	mintCntTrie, err := NewMintCntTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	voteCntTrie, err := NewVoteCntTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}

	return &DevoteProtocol{
		cycleTrie:        cycleTrie,
		cacheTrie:        cacheTrie,
		voteTrie:         voteTrie,
		masternodeTrie:   masternodeTrie,
		mintCntTrie:      mintCntTrie,
		voteCntTrie:      voteCntTrie,
		diskdb:           db,
		cycleTriedb:      trie.NewDatabase(ethdb.NewTable(db, cyclePrefix)),
		cacheTriedb:      trie.NewDatabase(ethdb.NewTable(db, cachePrefix)),
		voteTriedb:       trie.NewDatabase(ethdb.NewTable(db, votePrefix)),
		masternodeTriedb: trie.NewDatabase(ethdb.NewTable(db, masternodePrefix)),
		mintCntTriedb:    trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix)),
		voteCntTriedb:    trie.NewDatabase(ethdb.NewTable(db, voteCntPrefix)),
	}, nil

}

func NewDevoteProtocolFromAtomic(db ethdb.Database, ctxAtomic *DevoteProtocolAtomic) (*DevoteProtocol, error) {

	cycleTrie, err := NewCycleTrie(ctxAtomic.CycleHash, db)
	if err != nil {
		return nil, err
	}

	cacheTrie, err := NewCacheTrie(ctxAtomic.CacheHash, db)
	if err != nil {
		return nil, err
	}

	voteTrie, err := NewVoteTrie(ctxAtomic.VoteHash, db)
	if err != nil {
		return nil, err
	}

	masternodeTrie, err := NewMasternodeTrie(ctxAtomic.MasternodeHash, db)
	if err != nil {
		return nil, err
	}

	mintCntTrie, err := NewMintCntTrie(ctxAtomic.MintCntHash, db)
	if err != nil {
		return nil, err
	}

	voteCntTrie, err := NewVoteCntTrie(ctxAtomic.VoteCntHash, db)
	if err != nil {
		return nil, err
	}
	return &DevoteProtocol{
		cycleTrie:        cycleTrie,
		cacheTrie:        cacheTrie,
		voteTrie:         voteTrie,
		masternodeTrie:   masternodeTrie,
		mintCntTrie:      mintCntTrie,
		voteCntTrie:      voteCntTrie,
		diskdb:           db,
		cycleTriedb:      trie.NewDatabase(ethdb.NewTable(db, cyclePrefix)),
		cacheTriedb:      trie.NewDatabase(ethdb.NewTable(db, cachePrefix)),
		voteTriedb:       trie.NewDatabase(ethdb.NewTable(db, votePrefix)),
		masternodeTriedb: trie.NewDatabase(ethdb.NewTable(db, masternodePrefix)),
		mintCntTriedb:    trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix)),
		voteCntTriedb:    trie.NewDatabase(ethdb.NewTable(db, voteCntPrefix)),
	}, nil
}

// Unregister If the masternode does not complete the packing action during the current block cycle,
// and no block has been generated during the entire cycle, the masternode is removed from the network.
func (d *DevoteProtocol) Unregister(masternodeAddr common.Address) error {
	masternode := masternodeAddr.Bytes()
	err := d.masternodeTrie.TryDelete(masternode)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	iter := trie.NewIterator(d.cacheTrie.NodeIterator(masternode))
	for iter.Next() {
		delegator := iter.Value
		key := append(masternode, delegator...)
		err = d.cacheTrie.TryDelete(key)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		v, err := d.voteTrie.TryGet(delegator)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		if err == nil && bytes.Equal(v, masternode) {
			err = d.voteTrie.TryDelete(delegator)
			if err != nil {
				if _, ok := err.(*trie.MissingNodeError); !ok {
					return err
				}
			}
		}
	}
	return nil
}

func (d *DevoteProtocol) Copy() *DevoteProtocol {

	cycleTrie := *d.cycleTrie
	cacheTrie := *d.cacheTrie
	voteTrie := *d.voteTrie
	masternodeTrie := *d.masternodeTrie
	mintCntTrie := *d.mintCntTrie

	return &DevoteProtocol{
		cycleTrie:      &cycleTrie,
		cacheTrie:      &cacheTrie,
		voteTrie:       &voteTrie,
		masternodeTrie: &masternodeTrie,
		mintCntTrie:    &mintCntTrie,
	}
}

func (d *DevoteProtocol) Root() (h common.Hash) {

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.cycleTrie.Hash())
	rlp.Encode(hw, d.cacheTrie.Hash())
	rlp.Encode(hw, d.masternodeTrie.Hash())
	rlp.Encode(hw, d.voteTrie.Hash())
	rlp.Encode(hw, d.mintCntTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DevoteProtocol) Snapshot() *DevoteProtocol {
	return d.Copy()
}

func (d *DevoteProtocol) RevertToSnapShot(snapshot *DevoteProtocol) {

	d.cycleTrie = snapshot.cycleTrie
	d.cacheTrie = snapshot.cacheTrie
	d.masternodeTrie = snapshot.masternodeTrie
	d.voteTrie = snapshot.voteTrie
	d.mintCntTrie = snapshot.mintCntTrie
}

func (d *DevoteProtocol) FromAtomic(dcp *DevoteProtocolAtomic) error {

	var err error
	d.cycleTrie, err = NewCycleTrie(dcp.CycleHash, d.diskdb)
	if err != nil {
		return err
	}
	d.cacheTrie, err = NewCacheTrie(dcp.CacheHash, d.diskdb)
	if err != nil {
		return err
	}
	d.masternodeTrie, err = NewMasternodeTrie(dcp.MasternodeHash, d.diskdb)
	if err != nil {
		return err
	}
	d.voteTrie, err = NewVoteTrie(dcp.VoteHash, d.diskdb)
	if err != nil {
		return err
	}
	d.mintCntTrie, err = NewMintCntTrie(dcp.MintCntHash, d.diskdb)
	return err
}

type DevoteProtocolAtomic struct {
	CycleHash      common.Hash `json:"cycleRoot"        gencodec:"required"`
	CacheHash      common.Hash `json:"cacheRoot"        gencodec:"required"`
	MasternodeHash common.Hash `json:"masternodeRoot"    gencodec:"required"`
	VoteHash       common.Hash `json:"voteRoot"         gencodec:"required"`
	MintCntHash    common.Hash `json:"mintCntRoot"      gencodec:"required"`
	VoteCntHash    common.Hash `json:"voteCntRoot"      gencodec:"required"`
}

func (d *DevoteProtocol) MasternodeTrie() *trie.Trie { return d.masternodeTrie }
func (d *DevoteProtocol) CacheTrie() *trie.Trie      { return d.cacheTrie }
func (d *DevoteProtocol) VoteTrie() *trie.Trie       { return d.voteTrie }
func (d *DevoteProtocol) CycleTrie() *trie.Trie      { return d.cycleTrie }
func (d *DevoteProtocol) MintCntTrie() *trie.Trie    { return d.mintCntTrie }
func (d *DevoteProtocol) VoteCntTrie() *trie.Trie    { return d.voteCntTrie }

func (d *DevoteProtocol) DB() ethdb.Database { return d.diskdb }

func (dc *DevoteProtocol) SetCycle(cycle *trie.Trie)           { dc.cycleTrie = cycle }
func (dc *DevoteProtocol) SetCache(cache *trie.Trie)           { dc.cacheTrie = cache }
func (dc *DevoteProtocol) SetVote(vote *trie.Trie)             { dc.voteTrie = vote }
func (dc *DevoteProtocol) SetMasternode(masternode *trie.Trie) { dc.masternodeTrie = masternode }
func (dc *DevoteProtocol) SetMintCnt(mintCnt *trie.Trie)       { dc.mintCntTrie = mintCnt }
func (dc *DevoteProtocol) SetVoteCnt(voteCnt *trie.Trie)       { dc.voteCntTrie = voteCnt }

func (d *DevoteProtocol) Commit(db ethdb.Database) (*DevoteProtocolAtomic, error) {

	cycleRoot, err := d.cycleTrie.CommitTo(d.cycleTriedb)
	if err != nil {
		return nil, err
	}
	dberr := d.cycleTriedb.Commit(cycleRoot, false)
	if dberr != nil {
		return nil, err
	}
	cacheRoot, err := d.cacheTrie.CommitTo(d.cacheTriedb)
	if err != nil {
		return nil, err
	}
	d.cacheTriedb.Commit(cacheRoot, false)
	voteRoot, err := d.voteTrie.CommitTo(d.voteTriedb)
	if err != nil {
		return nil, err
	}
	d.voteTriedb.Commit(voteRoot, false)

	masternodeRoot, err := d.masternodeTrie.CommitTo(d.masternodeTriedb)
	if err != nil {
		return nil, err
	}
	d.masternodeTriedb.Commit(masternodeRoot, false)

	mintCntRoot, err := d.mintCntTrie.CommitTo(d.mintCntTriedb)
	if err != nil {
		return nil, err
	}
	d.mintCntTriedb.Commit(mintCntRoot, false)

	return &DevoteProtocolAtomic{
		CycleHash:      cycleRoot,
		CacheHash:      cacheRoot,
		VoteHash:       voteRoot,
		MasternodeHash: masternodeRoot,
		MintCntHash:    mintCntRoot,
	}, nil
}

func (d *DevoteProtocol) ProtocolAtomic() *DevoteProtocolAtomic {
	return &DevoteProtocolAtomic{
		CycleHash:      d.cycleTrie.Hash(),
		CacheHash:      d.cacheTrie.Hash(),
		MasternodeHash: d.masternodeTrie.Hash(),
		VoteHash:       d.voteTrie.Hash(),
		MintCntHash:    d.mintCntTrie.Hash(),
		VoteCntHash:    d.voteCntTrie.Hash(),
	}
}

func (p *DevoteProtocolAtomic) Root() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, p.CycleHash)
	rlp.Encode(hw, p.CacheHash)
	rlp.Encode(hw, p.MasternodeHash)
	rlp.Encode(hw, p.VoteHash)
	rlp.Encode(hw, p.MintCntHash)
	rlp.Encode(hw, p.VoteCntHash)
	hw.Sum(h[:0])
	return h
}

func (self *DevoteProtocol) SetWitnesses(witnesses []common.Address) error {

	key := []byte("witness")
	witnessesRLP, err := rlp.EncodeToBytes(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode witnesses to rlp bytes: %s", err)
	}
	self.cycleTrie.Update(key, witnessesRLP)
	return nil
}

func (self *DevoteProtocol) GetWitnesses() ([]common.Address, error) {

	var witnesses []common.Address
	key := []byte("witness")
	witnessRLP := self.cycleTrie.Get(key)
	if err := rlp.DecodeBytes(witnessRLP, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witnesses: %s", err)
	}
	return witnesses, nil
}

// register as a master node for saving to a block
func (d *DevoteProtocol) Register(masternodeAddr common.Address) error {
	masternode := masternodeAddr.Bytes()
	return d.masternodeTrie.TryUpdate(masternode, masternode)
}

func (d *DevoteProtocol) Delegate(delegatorAddr, masternodeAddr common.Address) error {
	delegator, masternode := delegatorAddr.Bytes(), masternodeAddr.Bytes()

	// the candidate must be masternode
	masternodeInTrie, err := d.masternodeTrie.TryGet(masternode)
	if err != nil {
		return err
	}
	if masternodeInTrie == nil {
		return errors.New("invalid masternode to delegate")
	}

	// delete old masternode if exists
	oldMasternode, err := d.voteTrie.TryGet(delegator)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	if oldMasternode != nil {
		d.cacheTrie.Delete(append(oldMasternode, delegator...))
	}

	if err = d.cacheTrie.TryUpdate(append(masternode, delegator...), delegator); err != nil {
		return err
	}
	return d.voteTrie.TryUpdate(delegator, masternode)
}

func (d *DevoteProtocol) UnDelegate(delegatorAddr, masternodeAddr common.Address) error {
	delegator, masternode := delegatorAddr.Bytes(), masternodeAddr.Bytes()

	// the delegate must be register a masternode
	masternodeInTrie, err := d.masternodeTrie.TryGet(masternode)
	if err != nil {
		return err
	}
	if masternodeInTrie == nil {
		return errors.New("invalid masternode to undelegate")
	}

	oldMasternode, err := d.voteTrie.TryGet(delegator)
	if err != nil {
		return err
	}
	if !bytes.Equal(masternode, oldMasternode) {
		return errors.New("mismatch masternode to undelegate")
	}

	if err = d.cacheTrie.TryDelete(append(masternode, delegator...)); err != nil {
		return err
	}
	return d.voteTrie.TryDelete(delegator)
}

// Voting save the vote result to the desk
func (d *DevoteProtocol) Voting(vote *Vote) error{

	currentVoteId := make([]byte, 8)
	binary.BigEndian.PutUint64(currentVoteId, uint64(vote.cycle))
	masternode:=vote.Masternode().Bytes()
	voteCntInTrieBytes := d.voteCntTrie.Get(append(currentVoteId, masternode...))
	if voteCntInTrieBytes != nil{
		return errors.New("vote already exists")
	}
	votebytes:=vote.Hash().Bytes()

	d.MintCntTrie().TryUpdate(voteCntInTrieBytes, votebytes)
	return nil

}

// update counts in MintCntTrie for the miner of newBlock
func (self *DevoteProtocol) UpdateMintCnt(parentBlockTime, currentBlockTime int64, witness common.Address) {

	currentMintCntTrie := self.MintCntTrie()
	currentCycle := parentBlockTime / CycleInterval
	currentCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentCycleBytes, uint64(currentCycle))

	cnt := int64(1)
	newCycle := currentBlockTime / CycleInterval
	// still during the currentCycleID
	if currentCycle == newCycle {
		iter := trie.NewIterator(currentMintCntTrie.NodeIterator(currentCycleBytes))

		// when current is not genesis, read last count from the MintCntTrie
		if iter.Next() {
			cntBytes := currentMintCntTrie.Get(append(currentCycleBytes, witness.Bytes()...))
			// not the first time to mint
			if cntBytes != nil {
				cnt = int64(binary.BigEndian.Uint64(cntBytes)) + 1
			}
		}
	}

	newCntBytes := make([]byte, 8)
	newCycleBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(newCycleBytes, uint64(newCycle))
	binary.BigEndian.PutUint64(newCntBytes, uint64(cnt))
	self.MintCntTrie().TryUpdate(append(newCycleBytes, witness.Bytes()...), newCntBytes)

}
