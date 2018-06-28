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

type DevoteContext struct {
	epochTrie     *trie.Trie
	cacheTrie     *trie.Trie
	voteTrie      *trie.Trie
	candidateTrie *trie.Trie
	mintCntTrie   *trie.Trie

	diskdb ethdb.Database
}

func NewEpochTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	triedb := trie.NewDatabase(ethdb.NewTable(db, epochPrefix))
	return trie.New(root, triedb)
}

func NewCacheTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	triedb := trie.NewDatabase(ethdb.NewTable(db, cachePrefix))
	return trie.New(root, triedb)
}

func NewVoteTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	triedb := trie.NewDatabase(ethdb.NewTable(db, votePrefix))
	return trie.New(root, triedb)
}

func NewCandidateTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	triedb := trie.NewDatabase(ethdb.NewTable(db, candidatePrefix))
	return trie.New(root, triedb)
}

func NewMintCntTrie(root common.Hash, db ethdb.Database) (*trie.Trie, error) {
	triedb := trie.NewDatabase(ethdb.NewTable(db, mintCntPrefix))
	return trie.New(root, triedb)

}

func NewDevoteContext(db ethdb.Database) (*DevoteContext, error) {

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
	return &DevoteContext{
		epochTrie:     epochTrie,
		cacheTrie:     cacheTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		diskdb:        db,
	}, nil

}

func NewDevoteContextFromAtomic(db ethdb.Database, ctxAtomic *DevoteContextAtomic) (*DevoteContext, error) {

	fmt.Printf("devote_context epochTrie old value:%x\n",ctxAtomic.EpochHash)
	epochTrie, err := NewEpochTrie(ctxAtomic.EpochHash, db)
	if err != nil {
		return nil, err
	}

	fmt.Printf("devote_context epochTrie new value:%x\n",epochTrie.Hash())

	fmt.Printf("devote_context cacheTrie old value:%x\n",ctxAtomic.CacheHash)

	cacheTrie, err := NewCacheTrie(ctxAtomic.CacheHash, db)
	fmt.Printf("devote_context cacheTrie new value:%x\n",cacheTrie.Hash())

	if err != nil {
		return nil, err
	}
	voteTrie, err := NewVoteTrie(ctxAtomic.VoteHash, db)
	fmt.Printf("devote_context voteTrie new value:%x\n",voteTrie.Hash())

	if err != nil {
		return nil, err
	}
	candidateTrie, err := NewCandidateTrie(ctxAtomic.CandidateHash, db)
	fmt.Printf("devote_context candidateTrie new value:%x\n",candidateTrie.Hash())

	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(ctxAtomic.MintCntHash, db)
	fmt.Printf("devote_context mintCntTrie new value:%x\n",mintCntTrie.Hash())

	if err != nil {
		return nil, err
	}
	return &DevoteContext{
		epochTrie:     epochTrie,
		cacheTrie:     cacheTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		diskdb:        db,
	}, nil
}

func (d *DevoteContext) KickoutCandidate(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	err := d.candidateTrie.TryDelete(candidate)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	iter := trie.NewIterator(d.cacheTrie.PrefixIterator(candidate))
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

func (d *DevoteContext) Copy() *DevoteContext {

	epochTrie := *d.epochTrie
	cacheTrie := *d.cacheTrie
	voteTrie := *d.voteTrie
	candidateTrie := *d.candidateTrie
	mintCntTrie := *d.mintCntTrie

	return &DevoteContext{
		epochTrie:     &epochTrie,
		cacheTrie:     &cacheTrie,
		voteTrie:      &voteTrie,
		candidateTrie: &candidateTrie,
		mintCntTrie:   &mintCntTrie,
	}
}

func (d *DevoteContext) Root() (h common.Hash) {

	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.epochTrie.Hash())
	rlp.Encode(hw, d.cacheTrie.Hash())
	rlp.Encode(hw, d.candidateTrie.Hash())
	rlp.Encode(hw, d.voteTrie.Hash())
	rlp.Encode(hw, d.mintCntTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DevoteContext) Snapshot() *DevoteContext {
	return d.Copy()
}

func (d *DevoteContext) RevertToSnapShot(snapshot *DevoteContext) {

	d.epochTrie = snapshot.epochTrie
	d.cacheTrie = snapshot.cacheTrie
	d.candidateTrie = snapshot.candidateTrie
	d.voteTrie = snapshot.voteTrie
	d.mintCntTrie = snapshot.mintCntTrie
}

func (d *DevoteContext) FromAtomic(dcp *DevoteContextAtomic) error {

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

type DevoteContextAtomic struct {
	EpochHash     common.Hash `json:"epochRoot"        gencodec:"required"`
	CacheHash     common.Hash `json:"cacheRoot"        gencodec:"required"`
	CandidateHash common.Hash `json:"candidateRoot"    gencodec:"required"`
	VoteHash      common.Hash `json:"voteRoot"         gencodec:"required"`
	MintCntHash   common.Hash `json:"mintCntRoot"      gencodec:"required"`
}

func (d *DevoteContext) CandidateTrie() *trie.Trie { return d.candidateTrie }
func (d *DevoteContext) CacheTrie() *trie.Trie     { return d.cacheTrie }
func (d *DevoteContext) VoteTrie() *trie.Trie      { return d.voteTrie }
func (d *DevoteContext) EpochTrie() *trie.Trie     { return d.epochTrie }
func (d *DevoteContext) MintCntTrie() *trie.Trie   { return d.mintCntTrie }

func (d *DevoteContext) DB() ethdb.Database { return d.diskdb }

func (dc *DevoteContext) SetEpoch(epoch *trie.Trie)         { dc.epochTrie = epoch }
func (dc *DevoteContext) SetCache(cache *trie.Trie)         { dc.cacheTrie = cache }
func (dc *DevoteContext) SetVote(vote *trie.Trie)           { dc.voteTrie = vote }
func (dc *DevoteContext) SetCandidate(candidate *trie.Trie) { dc.candidateTrie = candidate }
func (dc *DevoteContext) SetMintCnt(mintCnt *trie.Trie)     { dc.mintCntTrie = mintCnt }

func (d *DevoteContext) Commit(db *trie.Database) (*DevoteContextAtomic, error) {
	fmt.Printf("************ devote_context commit begin ************")
	epochRoot, err := d.epochTrie.Commit(nil)

	if err != nil {
		return nil, err
	}
	cacheRoot, err := d.cacheTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	voteRoot, err := d.voteTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	candidateRoot, err := d.candidateTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	mintCntRoot, err := d.mintCntTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("************ devote_context commit end ************")

	return &DevoteContextAtomic{
		EpochHash:     epochRoot,
		CacheHash:     cacheRoot,
		VoteHash:      voteRoot,
		CandidateHash: candidateRoot,
		MintCntHash:   mintCntRoot,
	}, nil
}

func (d *DevoteContext) ContextAtomic() *DevoteContextAtomic {
	return &DevoteContextAtomic{
		EpochHash:     d.epochTrie.Hash(),
		CacheHash:     d.cacheTrie.Hash(),
		CandidateHash: d.candidateTrie.Hash(),
		VoteHash:      d.voteTrie.Hash(),
		MintCntHash:   d.mintCntTrie.Hash(),
	}
}

func (p *DevoteContextAtomic) Root() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, p.EpochHash)
	rlp.Encode(hw, p.CacheHash)
	rlp.Encode(hw, p.CandidateHash)
	rlp.Encode(hw, p.VoteHash)
	rlp.Encode(hw, p.MintCntHash)
	hw.Sum(h[:0])
	return h
}

func (dc *DevoteContext) SetWitnesses(witnesses []common.Address) error {

	key := []byte("witness")
	witnessesRLP, err := rlp.EncodeToBytes(witnesses)
	if err != nil {
		return fmt.Errorf("failed to encode Witnesses to rlp bytes: %s", err)
	}
	dc.epochTrie.Update(key, witnessesRLP)
	return nil
}

func (dc *DevoteContext) GetWitnesses() ([]common.Address, error) {

	var witnesses []common.Address
	key := []byte("witness")
	witnessRLP := dc.epochTrie.Get(key)
	if err := rlp.DecodeBytes(witnessRLP, &witnesses); err != nil {
		return nil, fmt.Errorf("failed to decode witnesses: %s", err)
	}
	return witnesses, nil
}

func (d *DevoteContext) BecomeCandidate(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	return d.candidateTrie.TryUpdate(candidate, candidate)
}

func (d *DevoteContext) UnDelegate(delegatorAddr, candidateAddr common.Address) error {
	delegator, candidate := delegatorAddr.Bytes(), candidateAddr.Bytes()

	// the candidate must be candidate
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
