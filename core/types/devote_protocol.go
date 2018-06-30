package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/crypto/sha3"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
)

var (
	epochPrefix     = "epoch-"
	cachePrefix     = "cache-"
	votePrefix      = "vote-"
	candidatePrefix = "candidate-"
	mintCntPrefix   = "mintCnt-"
)

type DevoteProtocol struct {
	epochTrie     *trie.Trie
	cacheTrie     *trie.Trie
	voteTrie      *trie.Trie
	candidateTrie *trie.Trie
	mintCntTrie   *trie.Trie

	epochTriedb        *trie.Database
	cacheTriedb        *trie.Database
	voteTriedb        *trie.Database
	candidateTriedb        *trie.Database
	mintCntTriedb        *trie.Database

	diskdb ethdb.Database

}

func NewEpochTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {

	epochTriedb:=trie.NewDatabase(ethdb.NewTable(db, epochPrefix))
	return trie.New(root, epochTriedb)
}

func NewCacheTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	cacheTriedb:=trie.NewDatabase(ethdb.NewTable(db, cachePrefix))
	return trie.New(root, cacheTriedb)
}

func NewVoteTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	voteTriedb:=trie.NewDatabase(ethdb.NewTable(db, votePrefix))
	return trie.New(root, voteTriedb)
}

func NewCandidateTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	candidateTriedb:=trie.NewDatabase(ethdb.NewTable(db, candidatePrefix))
	return trie.New(root, candidateTriedb)
}

func NewMintCntTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	mintCntTriedb:=trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix))
	return trie.New(root, mintCntTriedb)

}

func NewDevoteProtocol(db ethdb.Database) (*DevoteProtocol, error) {

	epochTrie, err := NewEpochTrie(common.Hash{}, db)
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
	candidateTrie, err := NewCandidateTrie(common.Hash{}, db)

	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	return &DevoteProtocol{
		epochTrie:     epochTrie,
		cacheTrie:     cacheTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		diskdb:        db,
		epochTriedb:      trie.NewDatabase(ethdb.NewTable(db, epochPrefix)),
		cacheTriedb:      trie.NewDatabase(ethdb.NewTable(db, cachePrefix)),
		voteTriedb:      trie.NewDatabase(ethdb.NewTable(db, votePrefix)),
		candidateTriedb:      trie.NewDatabase(ethdb.NewTable(db, candidatePrefix)),
		mintCntTriedb:      trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix)),

	}, nil

}

func NewDevoteProtocolFromAtomic(db ethdb.Database, ctxAtomic *DevoteProtocolAtomic) (*DevoteProtocol, error) {

	epochTrie, err := NewEpochTrie(ctxAtomic.EpochHash, db)
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
	candidateTrie, err := NewCandidateTrie(ctxAtomic.CandidateHash, db)

	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(ctxAtomic.MintCntHash, db)

	if err != nil {
		return nil, err
	}
	return &DevoteProtocol{
		epochTrie:     epochTrie,
		cacheTrie:     cacheTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		diskdb:        db,
		epochTriedb:      trie.NewDatabase(ethdb.NewTable(db, epochPrefix)),
		cacheTriedb:      trie.NewDatabase(ethdb.NewTable(db, cachePrefix)),
		voteTriedb:      trie.NewDatabase(ethdb.NewTable(db, votePrefix)),
		candidateTriedb:      trie.NewDatabase(ethdb.NewTable(db, candidatePrefix)),
		mintCntTriedb:      trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix)),

	}, nil
}

//Uncast Candidate
func (d *DevoteProtocol) Uncast(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	err := d.candidateTrie.TryDelete(candidate)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	iter := trie.NewIterator(d.cacheTrie.NodeIterator(candidate))
	for iter.Next() {
		delegator := iter.Value
		key := append(candidate, delegator...)
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
		if err == nil && bytes.Equal(v, candidate) {
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

	epochTrie := *d.epochTrie
	cacheTrie := *d.cacheTrie
	voteTrie := *d.voteTrie
	candidateTrie := *d.candidateTrie
	mintCntTrie := *d.mintCntTrie

	return &DevoteProtocol{
		epochTrie:     &epochTrie,
		cacheTrie:     &cacheTrie,
		voteTrie:      &voteTrie,
		candidateTrie: &candidateTrie,
		mintCntTrie:   &mintCntTrie,
	}
}

func (d *DevoteProtocol) Root() (h common.Hash) {

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.epochTrie.Hash())
	rlp.Encode(hw, d.cacheTrie.Hash())
	rlp.Encode(hw, d.candidateTrie.Hash())
	rlp.Encode(hw, d.voteTrie.Hash())
	rlp.Encode(hw, d.mintCntTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DevoteProtocol) Snapshot() *DevoteProtocol {
	return d.Copy()
}

func (d *DevoteProtocol) RevertToSnapShot(snapshot *DevoteProtocol) {

	d.epochTrie = snapshot.epochTrie
	d.cacheTrie = snapshot.cacheTrie
	d.candidateTrie = snapshot.candidateTrie
	d.voteTrie = snapshot.voteTrie
	d.mintCntTrie = snapshot.mintCntTrie
}

func (d *DevoteProtocol) FromAtomic(dcp *DevoteProtocolAtomic) error {

	var err error
	d.epochTrie, err = NewEpochTrie(dcp.EpochHash, d.diskdb)
	if err != nil {
		return err
	}
	d.cacheTrie, err = NewCacheTrie(dcp.CacheHash, d.diskdb)
	if err != nil {
		return err
	}
	d.candidateTrie, err = NewCandidateTrie(dcp.CandidateHash, d.diskdb)
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
	EpochHash     common.Hash `json:"epochRoot"        gencodec:"required"`
	CacheHash     common.Hash `json:"cacheRoot"        gencodec:"required"`
	CandidateHash common.Hash `json:"candidateRoot"    gencodec:"required"`
	VoteHash      common.Hash `json:"voteRoot"         gencodec:"required"`
	MintCntHash   common.Hash `json:"mintCntRoot"      gencodec:"required"`
}

func (d *DevoteProtocol) CandidateTrie() *trie.Trie { return d.candidateTrie }
func (d *DevoteProtocol) CacheTrie() *trie.Trie     { return d.cacheTrie }
func (d *DevoteProtocol) VoteTrie() *trie.Trie      { return d.voteTrie }
func (d *DevoteProtocol) EpochTrie() *trie.Trie     { return d.epochTrie }
func (d *DevoteProtocol) MintCntTrie() *trie.Trie   { return d.mintCntTrie }

func (d *DevoteProtocol) DB() ethdb.Database { return d.diskdb }

func (dc *DevoteProtocol) SetEpoch(epoch *trie.Trie)         { dc.epochTrie = epoch }
func (dc *DevoteProtocol) SetCache(cache *trie.Trie)         { dc.cacheTrie = cache }
func (dc *DevoteProtocol) SetVote(vote *trie.Trie)           { dc.voteTrie = vote }
func (dc *DevoteProtocol) SetCandidate(candidate *trie.Trie) { dc.candidateTrie = candidate }
func (dc *DevoteProtocol) SetMintCnt(mintCnt *trie.Trie)     { dc.mintCntTrie = mintCnt }

func (d *DevoteProtocol) Commit(db ethdb.Database) (*DevoteProtocolAtomic, error) {

	epochRoot, err := d.epochTrie.CommitTo(d.epochTriedb)
	if err != nil {
		return nil, err
	}
	dberr:=d.epochTriedb.Commit(epochRoot,false)
	if dberr != nil{
		return nil,err
	}
	cacheRoot, err := d.cacheTrie.CommitTo(d.cacheTriedb)
	if err != nil {
		return nil, err
	}
	d.cacheTriedb.Commit(cacheRoot,false)
	voteRoot, err := d.voteTrie.CommitTo(d.voteTriedb)
	if err != nil {
		return nil, err
	}
	d.voteTriedb.Commit(voteRoot,false)

	candidateRoot, err := d.candidateTrie.CommitTo(d.candidateTriedb)
	if err != nil {
		return nil, err
	}
	d.candidateTriedb.Commit(candidateRoot,false)

	mintCntRoot, err := d.mintCntTrie.CommitTo(d.mintCntTriedb)
	if err != nil {
		return nil, err
	}
	d.mintCntTriedb.Commit(mintCntRoot,false)

	return &DevoteProtocolAtomic{
		EpochHash:     epochRoot,
		CacheHash:     cacheRoot,
		VoteHash:      voteRoot,
		CandidateHash: candidateRoot,
		MintCntHash:   mintCntRoot,
	}, nil
}

func (d *DevoteProtocol) ProtocolAtomic() *DevoteProtocolAtomic {
	return &DevoteProtocolAtomic{
		EpochHash:     d.epochTrie.Hash(),
		CacheHash:     d.cacheTrie.Hash(),
		CandidateHash: d.candidateTrie.Hash(),
		VoteHash:      d.voteTrie.Hash(),
		MintCntHash:   d.mintCntTrie.Hash(),
	}
}

func (p *DevoteProtocolAtomic) Root() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, p.EpochHash)
	rlp.Encode(hw, p.CacheHash)
	rlp.Encode(hw, p.CandidateHash)
	rlp.Encode(hw, p.VoteHash)
	rlp.Encode(hw, p.MintCntHash)
	hw.Sum(h[:0])
	return h
}

func (dc *DevoteProtocol) SetWitnesses(witnesses []common.Address) error {

	key := []byte("witness")
	witnessesRLP, err := rlp.EncodeToBytes(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode witnesses to rlp bytes: %s", err)
	}
	dc.epochTrie.Update(key, witnessesRLP)
	return nil
}

func (dc *DevoteProtocol) GetWitnesses() ([]common.Address, error) {

	var witnesses []common.Address
	key := []byte("witness")
	witnessRLP := dc.epochTrie.Get(key)
	if err := rlp.DecodeBytes(witnessRLP, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witnesses: %s", err)
	}
	return witnesses, nil
}

// cast candidate
func (d *DevoteProtocol) Cast(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	return d.candidateTrie.TryUpdate(candidate, candidate)
}

func (d *DevoteProtocol) UnDelegate(delegatorAddr, candidateAddr common.Address) error {
	delegator, candidate := delegatorAddr.Bytes(), candidateAddr.Bytes()

	// the delegate must be cast candidate
	candidateInTrie, err := d.candidateTrie.TryGet(candidate)
	if err != nil {
		return err
	}
	if candidateInTrie == nil {
		return errors.New("invalid candidate to undelegate")
	}

	oldCandidate, err := d.voteTrie.TryGet(delegator)
	if err != nil {
		return err
	}
	if !bytes.Equal(candidate, oldCandidate) {
		return errors.New("mismatch candidate to undelegate")
	}

	if err = d.cacheTrie.TryDelete(append(candidate, delegator...)); err != nil {
		return err
	}
	return d.voteTrie.TryDelete(delegator)
}
