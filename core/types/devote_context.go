package types

import (
	"fmt"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
)

type DevoteContext struct {
	epochTrie     *trie.Trie
	cacheTrie     *trie.Trie
	voteTrie      *trie.Trie
	candidateTrie *trie.Trie
	mintCntTrie   *trie.Trie

	db *trie.Database
}

var (
	epochPrefix     = []byte("epoch-")
	cachePrefix     = []byte("cache-")
	votePrefix      = []byte("vote-")
	candidatePrefix = []byte("candidate-")
	mintCntPrefix   = []byte("mintCnt-")
)

func NewEpochTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, epochPrefix, db)
}

func NewCacheTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, cachePrefix, db)
}

func NewVoteTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, votePrefix, db)
}

func NewCandidateTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, candidatePrefix, db)
}

func NewMintCntTrie(root common.Hash, db *trie.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, mintCntPrefix, db)
}

func NewDevoteContext(db *trie.Database) (*DevoteContext, error) {
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
		db:            db,
	}, nil
}

func NewDevoteContextFromProto(db *trie.Database, ctxProto *DevoteContextProto) (*DevoteContext, error) {
	epochTrie, err := NewEpochTrie(ctxProto.EpochHash, db)
	if err != nil {
		return nil, err
	}
	cacheTrie, err := NewCacheTrie(ctxProto.CacheHash, db)
	if err != nil {
		return nil, err
	}
	voteTrie, err := NewVoteTrie(ctxProto.VoteHash, db)
	if err != nil {
		return nil, err
	}
	candidateTrie, err := NewCandidateTrie(ctxProto.CandidateHash, db)
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(ctxProto.MintCntHash, db)
	if err != nil {
		return nil, err
	}
	return &DevoteContext{
		epochTrie:     epochTrie,
		cacheTrie:     cacheTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		db:            db,
	}, nil
}

func (d *DevoteContext) CandidateTrie() *trie.Trie { return d.candidateTrie }
func (d *DevoteContext) CacheTrie() *trie.Trie     { return d.cacheTrie }
func (d *DevoteContext) VoteTrie() *trie.Trie      { return d.voteTrie }
func (d *DevoteContext) EpochTrie() *trie.Trie     { return d.epochTrie }
func (d *DevoteContext) MintCntTrie() *trie.Trie   { return d.mintCntTrie }

func (d *DevoteContext) Commit() (*DevoteContextProto, error) {
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
	return &DevoteContextProto{
		EpochHash:     epochRoot,
		CacheHash:     cacheRoot,
		VoteHash:      voteRoot,
		CandidateHash: candidateRoot,
		MintCntHash:   mintCntRoot,
	}, nil
}

func (d *DevoteContext) ContextProto() *DevoteContextProto {
	return &DevoteContextProto{
		EpochHash:     d.epochTrie.Hash(),
		CacheHash:     d.cacheTrie.Hash(),
		CandidateHash: d.candidateTrie.Hash(),
		VoteHash:      d.voteTrie.Hash(),
		MintCntHash:   d.mintCntTrie.Hash(),
	}
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
