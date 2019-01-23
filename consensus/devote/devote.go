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
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"encoding/binary"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/common/hexutil"
	"github.com/etherzero/go-etherzero/consensus"
	"github.com/etherzero/go-etherzero/consensus/misc"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/crypto/sha3"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/rpc"
	lru "github.com/hashicorp/golang-lru"
)

const (
	checkpointInterval = 600                    // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128                    // Number of recent vote snapshots to keep in memory
	inmemorySignatures = 4096                   // Number of recent block signatures to keep in memory
	maxSignersSize     = 11                     // Number of max singers in current cycle
	wiggleTime         = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

// Devote proof-of-authority protocol constants.
var (
	epochLength = uint64(600) // Default number of blocks after which to checkpoint and reset the pending votes

	extraVanity = 32 // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = 65 // Fixed number of extra-data suffix bytes reserved for signer seal

	nonceAuthVote = hexutil.MustDecode("0xffffffffffffffff") // Magic nonce number to vote on adding a new signer
	nonceDropVote = hexutil.MustDecode("0x0000000000000000") // Magic nonce number to vote on removing a signer.

	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.

	confirmedBlockHead = []byte("confirmed-block-head")

	etherzeroBlockReward = big.NewInt(0.3375e+18) // Block reward in wei to masternode account when successfully mining a block
	rewardToCommunity    = big.NewInt(0.1125e+18) // Block reward in wei to community account when successfully mining a block

	diffInTurn          = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn          = big.NewInt(1) // Block difficulty for out-of-turn signatures
	masternodeDifficult = big.NewInt(1) // Block difficult for masternode consensus
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	// errInvalidCheckpointBeneficiary is returned if a checkpoint/epoch transition
	// block has a beneficiary set to non-zeroes.
	errInvalidCheckpointBeneficiary = errors.New("beneficiary in checkpoint block non-zero")

	// errInvalidVote is returned if a nonce value is something else that the two
	// allowed constants of 0x00..0 or 0xff..f.
	errInvalidVote = errors.New("vote nonce not 0x00..0 or 0xff..f")

	// errInvalidCheckpointVote is returned if a checkpoint/epoch transition block
	// has a vote nonce set to non-zeroes.
	errInvalidCheckpointVote = errors.New("vote nonce in checkpoint block non-zero")

	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")

	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")

	// errExtraSigners is returned if non-checkpoint block contain signer data in
	// their extra-data fields.
	errExtraSigners = errors.New("non-checkpoint block contains extra signer list")

	// errInvalidCheckpointSigners is returned if a checkpoint block contains an
	// invalid list of signers (i.e. non divisible by 20 bytes).
	errInvalidCheckpointSigners = errors.New("invalid signer list on checkpoint block")

	// errMismatchingCheckpointSigners is returned if a checkpoint block contains a
	// list of signers different than the one the local node calculated.
	errMismatchingCheckpointSigners = errors.New("mismatching signer list on checkpoint block")

	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")

	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash = errors.New("non empty uncle hash")

	// errInvalidDifficulty is returned if the difficulty of a block neither 1 or 2.
	errInvalidDifficulty = errors.New("invalid difficulty")

	// errWrongDifficulty is returned if the difficulty of a block doesn't match the
	// turn of the signer.
	errWrongDifficulty = errors.New("wrong difficulty")

	// ErrInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// errRecentlySigned is returned if a header is signed by an authorized entity
	// that already signed a header recently, thus is temporarily not allowed to.
	errRecentlySigned = errors.New("recently signed")

	ErrNilBlockHeader = errors.New("nil block header returned")

	ErrMismatchSignerAndWitness = errors.New("mismatch block signer and witness")

	ErrWaitForPrevBlock = errors.New("wait for last block arrived")

	ErrMinerFutureBlock = errors.New("miner the future block")

	ErrInvalidMinerBlockTime = errors.New("invalid time to miner the block")

	ErrInvalidBlockWitness      = errors.New("invalid block witness")
)

// SignerFn is a signer callback function to request a hash to be signed by a
// backing account.
// SignerFn
// string:master node nodeid,[8]byte
// []byte,signature
type SignerFn func(string, []byte) ([]byte, error)

type MasternodeListFn func(number *big.Int) ([]string, error)

type GetGovernanceContractAddress func(number *big.Int) (common.Address, error)

// sigHash returns the hash which is used as input for the proof-of-authority
// signing. It is the hash of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewKeccak256()

	rlp.Encode(hasher, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Witness,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-65], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
		header.Protocol.Root(),
	})
	hasher.Sum(hash[:0])
	return hash
}

// Devote is the proof-of-authority consensus engine proposed to support the
// Ethereum testnet following the Ropsten attacks.
type Devote struct {
	config *params.DevoteConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	recents    *lru.ARCCache // Snapshots for recent block to speed up reorgs
	signatures *lru.ARCCache // Signatures of recent blocks to speed up mining

	proposals            map[string]bool // Current list of proposals we are pushing
	confirmedBlockHeader *types.Header

	signer string       // Masternode 's Id
	signFn SignerFn     // Signer function to authorize hashes with
	lock   sync.RWMutex // Protects the signer fields

	masternodeListFn            MasternodeListFn             //get current all masternodes
	governanceContractAddressFn GetGovernanceContractAddress //get current GovernanceContractAddress

}

// ecrecover extracts the Masternode account ID from a signed header.
func ecrecover(header *types.Header, sigcache *lru.ARCCache) (string, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(string), nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return "", errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]
	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(sigHash(header).Bytes(), signature)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%x", pubkey[1:9])
	sigcache.Add(hash, id)
	return id, nil
}

// New creates a Clique proof-of-authority consensus engine with the initial
// signers set to the ones provided by the user.
func New(config *params.DevoteConfig, db ethdb.Database) *Devote {
	// Set any missing consensus parameters to their defaults
	conf := *config
	if conf.Epoch == 0 {
		conf.Epoch = epochLength
	}
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)

	return &Devote{
		config:     &conf,
		db:         db,
		recents:    recents,
		signatures: signatures,
		proposals:  make(map[string]bool),
	}
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *Devote) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
	//return ecrecover(header, c.signatures)
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *Devote) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *Devote) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (d *Devote) verifyHeader(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()

	// Don't waste time checking blocks from the future
	if header.Time.Cmp(big.NewInt(time.Now().Unix())) > 0 {
		return consensus.ErrFutureBlock
	}
	// Checkpoint blocks need to enforce zero beneficiary
	checkpoint := (number % d.config.Epoch) == 0
	// Nonces must be 0x00..0 or 0xff..f, zeroes enforced on checkpoints
	if !bytes.Equal(header.Nonce[:], nonceAuthVote) && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidVote
	}
	if checkpoint && !bytes.Equal(header.Nonce[:], nonceDropVote) {
		return errInvalidCheckpointVote
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in devote
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	if chain.Config().IsDevote(header.Number) {
		// Ensure that the block's difficulty is meaningful (may not be correct at this point)
		if number > 0 {
			if header.Difficulty == nil || (header.Difficulty.Cmp(diffInTurn) != 0 && header.Difficulty.Cmp(diffNoTurn) != 0) {
				return errInvalidDifficulty
			}
		}
	} else {
		if header.Difficulty.Cmp(masternodeDifficult) != 0 {
			return errInvalidDifficulty
		}
	}

	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
		return err
	}
	// All basic checks passed, verify cascading fields
	return d.verifyCascadingFields(chain, header, parents)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (d *Devote) verifyCascadingFields(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}
	// Ensure that the block's timestamp isn't too close to it's parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	if parent.Time.Uint64()+d.config.Period > header.Time.Uint64() {
		return ErrInvalidTimestamp
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := d.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}
	// If the block is a checkpoint block, verify the signer list
	if number%d.config.Epoch == 0 {
		signers := make([]string, len(snap.Signers))
		for i, signer := range snap.signers() {
			signers[i] = signer
		}
	}
	// All basic checks passed, verify the seal and return
	return d.verifySeal(chain, header, parents)
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (d *Devote) snapshot(chain consensus.ChainReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {

		// If we're at an checkpoint block, make a snapshot if it's known
		if number == 0 || number%d.config.Epoch == 0 {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				hash := checkpoint.Hash()
				cycle := number / d.config.Epoch
				all, err := d.masternodeListFn(big.NewInt(int64(number)))
				if err != nil {
					return nil, fmt.Errorf("get current masternodes err:%s", err)
				}
				stabilization := number - 100
				stableBlock := chain.GetHeaderByNumber(stabilization)
				if stableBlock != nil {
					hash = stableBlock.Hash()
				}
				result, err := masternodes(hash, all)
				masternodes := sortableAddresses{}
				for masternode, cnt := range result {
					masternodes = append(masternodes, &sortableAddress{nodeid: masternode, weight: cnt})
				}
				sort.Sort(masternodes)
				if len(masternodes) > int(maxSignersSize) {
					masternodes = masternodes[:maxSignersSize]
				}
				var sortedWitnesses []string
				for _, node := range masternodes {
					sortedWitnesses = append(sortedWitnesses, node.nodeid)
				}
				sort.Strings(sortedWitnesses)
				context := []interface{}{
					"cycle", cycle,
					"signers", sortedWitnesses,
					"hash", hash,
					"number", number,
				}
				log.Debug("Elected new cycle signers", context...)
				snap = newSnapshot(d.config, number, cycle, d.signatures, hash, sortedWitnesses)
				if err := snap.store(d.db); err != nil {
					return nil, err
				}
				d.recents.Add(snap.Hash, snap)
				log.Trace("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
				break
			}
		}

		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}
		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}
	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}

	snap, err := snap.apply(headers)
	if err != nil {
		return nil, err
	}
	d.recents.Add(snap.Hash, snap)
	// If we've generated a new checkpoint snapshot, save to disk
	if snap.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = snap.store(d.db); err != nil {
			return nil, err
		}
		log.Info("Stored snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *Devote) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal implements consensus.Engine, checking whether the signature contained
// in the header satisfies the consensus protocol requirements.
func (c *Devote) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return c.verifySeal(chain, header, nil)
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (d *Devote) verifySeal(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := d.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}

	// Resolve the authorization key and check against signers
	signer, err := ecrecover(header, d.signatures)
	if err != nil {
		return err
	}
	if _, ok := snap.Signers[signer]; !ok {
		return errUnauthorizedSigner
	}
	if chain.Config().IsDevote(header.Number) {
		for seen, recent := range snap.Recents {
			if recent == signer {
				// Signer is among recents, only fail if the current block doesn't shift it out
				if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
					return errRecentlySigned
				}
			}
		}

		// Ensure that the difficulty corresponds to the turn-ness of the signer
		inturn := snap.inturn(header.Number.Uint64(), signer)
		if inturn && header.Difficulty.Cmp(diffInTurn) != 0 {
			return errWrongDifficulty
		}
		if !inturn && header.Difficulty.Cmp(diffNoTurn) != 0 {
			return errWrongDifficulty
		}
	} else {
		witness, err := lookup(snap.signers(), header.Time.Uint64())
		if err != nil {
			return err
		}
		if err := d.verifyBlockSigner(witness, header); err != nil {
			return err
		}

		if header.Difficulty.Cmp(masternodeDifficult) != 0 {
			return errWrongDifficulty
		}
	}

	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (d *Devote) Prepare(chain consensus.ChainReader, header *types.Header) error {
	// If the block isn't a checkpoint, cast a random vote (good enough for now)
	header.Nonce = types.BlockNonce{}

	number := header.Number.Uint64()
	// Assemble the voting snapshot to check which votes make sense
	snap, err := d.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}
	if number%d.config.Epoch != 0 {
		d.lock.RLock()

		// Gather all the proposals that make sense voting on
		witnesses := make([]string, 0, len(d.proposals))
		for witness, authorize := range d.proposals {
			if snap.validWitness(witness, authorize) {
				witnesses = append(witnesses, witness)
			}
		}
		// If there's pending proposals, cast a vote on them
		if len(witnesses) > 0 {
			header.Witness = witnesses[rand.Intn(len(witnesses))]
			if d.proposals[header.Witness] {
				copy(header.Nonce[:], nonceAuthVote)
			} else {
				copy(header.Nonce[:], nonceDropVote)
			}
		}
		d.lock.RUnlock()
	}
	// Set the correct difficulty
	d.lock.Lock()
	if chain.Config().IsDevote(header.Number){
		header.Difficulty = CalcDifficulty(snap, d.signer)
	}else{
		header.Difficulty=big.NewInt(1)
	}
	header.Witness = d.signer
	d.lock.Unlock()

	// Ensure the extra data has all it's components
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]

	if number%d.config.Epoch == 0 {
		for _, signer := range snap.signers() {
			header.Extra = append(header.Extra, signer[:]...)
		}
	}
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)

	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}

	// Ensure the timestamp has the correct delay
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Time = new(big.Int).Add(parent.Time, new(big.Int).SetUint64(d.config.Period))
	if header.Time.Int64() < time.Now().Unix() {
		header.Time = big.NewInt(time.Now().Unix())
	}
	return nil
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward.  The devote consensus allowed uncle block .
func AccumulateRewards(govAddress common.Address, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain progression
	blockReward := etherzeroBlockReward
	// Accumulate the rewards for the masternode and any included uncles
	reward := new(big.Int).Set(blockReward)
	state.AddBalance(header.Coinbase, reward, header.Number)

	//  Accumulate the rewards to community account
	rewardForCommunity := new(big.Int).Set(rewardToCommunity)
	state.AddBalance(govAddress, rewardForCommunity, header.Number)
}

// Finalize implements consensus.Engine, ensuring no uncles are set, nor block
// rewards given, and returns the final block.
func (d *Devote) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, db *devotedb.DevoteDB) (*types.Block, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	parent := chain.GetHeaderByHash(header.ParentHash)
	stableBlockNumber := new(big.Int).Sub(parent.Number, big.NewInt(maxSignersSize))
	if stableBlockNumber.Cmp(big.NewInt(0)) < 0 {
		stableBlockNumber = big.NewInt(0)
	}

	// Accumulate block rewards and commit the final state root
	govaddress, err := d.governanceContractAddressFn(stableBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("got current governance address err:%s", err)
	}
	nodes, merr := d.masternodeListFn(stableBlockNumber)
	if merr != nil {
		return nil, fmt.Errorf("got current masternodes err:%s", merr)
	}

	AccumulateRewards(govaddress, state, header, nil)

	// No block rewards in PoA, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)
	protocol :=d.GenerateProtocol(chain , header , db ,nodes)
	header.Protocol = protocol

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts), nil
}

func(d *Devote)  GenerateProtocol(chain consensus.ChainReader, header *types.Header, db *devotedb.DevoteDB, nodes []string) *devotedb.DevoteProtocol {

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
		fmt.Printf("devote finalize protocol statsHash value:%x,witness :%s,height: %d \n",db.Protocol().StatsHash,header.Witness,header.Number)
		return db.Protocol()

	}
}


// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (d *Devote) Authorize(signer string, signFn SignerFn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.signer = signer
	d.signFn = signFn
}

func (d *Devote) verifyBlockSigner(witness string, header *types.Header) error {
	signer, err := ecrecover(header, d.signatures)
	if err != nil {
		return err
	}
	if signer != witness {
		return fmt.Errorf("invalid block witness have: %s,got: %s,time:%d \n", witness, signer, header.Time)
	}
	if signer != header.Witness {
		return ErrMismatchSignerAndWitness
	}
	return nil
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (d *Devote) Seal(chain consensus.ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {

	header := block.Header()
	// Sealing the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	// Don't hold the signer fields for the entire sealing procedure
	d.lock.RLock()
	signer, signFn := d.signer, d.signFn
	d.lock.RUnlock()
	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(header.Time.Int64(), 0).Sub(time.Now()) // nolint: gosimple
	// Bail out if we're unauthorized to sign a block
	snap, err := d.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}

	if chain.Config().IsDevote(header.Number) {
		singerMap := snap.Signers
		if _, ok := singerMap[signer]; !ok {
			return errUnauthorizedSigner
		}
		// If we're amongst the recent signers, wait for the next block
		for seen, recent := range snap.Recents {
			if recent == signer {
				// Signer is among recents, only wait if the current block doesn't shift it out
				if limit := uint64(len(singerMap)/2 + 1); number < limit || seen > number-limit {
					log.Info("Signed recently, must wait for others, ", "signer", signer, "seen", seen, "number", number, "limit", limit)
					return nil
				}
				log.Info("Passed Signed recently, ", "signer", signer, "seen", seen, "number", number, "limit", uint64(len(singerMap)/2+1))
			}
		}
		if header.Difficulty.Cmp(diffNoTurn) == 0 {
			// It's not our turn explicitly to sign, delay it a bit
			wiggle := time.Duration(15) * wiggleTime
			delay += time.Duration(rand.Int63n(int64(wiggle)))

			log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
		}
	}else {
		witness, err := lookup(snap.signers(), header.Time.Uint64())
		if err != nil {
			return err
		}
		if witness != d.signer {
			return fmt.Errorf("it's not our turn,current witness:%s,d.signer%s",witness,d.signer)
		}
	}

	// Sign all the things!
	sighash, err := signFn(signer, sigHash(header).Bytes())
	if err != nil {
		return err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	// Wait until sealing is terminated or delay timeout.
	log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))
	go func() {
		select {
		case <-stop:
			return
		case <-time.After(delay):
		}

		select {
		case results <- block.WithSeal(header):
		default:
			log.Warn("Sealing result is not read by miner", "sealhash", d.SealHash(header))
		}
	}()

	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have based on the previous blocks in the chain and the
// current signer.
func (d *Devote) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	snap, err := d.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
	if err != nil {
		return nil
	}
	return CalcDifficulty(snap, d.signer)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have based on the previous blocks in the chain and the
// current signer.
func CalcDifficulty(snap *Snapshot, signer string) *big.Int {
	if snap.inturn(snap.Number+1, signer) {
		return new(big.Int).Set(diffInTurn)
	}
	return new(big.Int).Set(diffNoTurn)
}

func (d *Devote) Masternodes(masternodeListFn MasternodeListFn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.masternodeListFn = masternodeListFn
}

func (d *Devote) GetGovernanceContractAddress(goveAddress GetGovernanceContractAddress) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.governanceContractAddressFn = goveAddress
}

// masternodes return  masternode list in the Cycle.
// key   -- nodeid
// value -- votes count
func masternodes(hash common.Hash, nodes []string) (map[string]*big.Int, error) {

	result := make(map[string]*big.Int)
	for i := 0; i < len(nodes); i++ {
		masternode := nodes[i]
		bytes := make([]byte, 8)
		bytes = append(bytes, []byte(masternode)...)
		bytes = append(bytes, hash[:]...)
		weight := int64(binary.LittleEndian.Uint32(crypto.Keccak512(bytes)))

		score := big.NewInt(0)
		score.Add(score, big.NewInt(weight))
		result[masternode] = score
	}
	log.Debug("snapshot nodes ", "context", nodes, "count", len(nodes))
	return result, nil
}

// store inserts the snapshot into the database.
func (s *Devote) storeConfirmedBlockHeader(db ethdb.Database) error {
	db.Put(confirmedBlockHead, s.confirmedBlockHeader.Hash().Bytes())
	return nil
}

func (s *Devote) loadConfirmedBlockHeader(chain consensus.ChainReader) (*types.Header, error) {

	key, err := s.db.Get(confirmedBlockHead)
	if err != nil {
		return nil, err
	}
	header := chain.GetHeaderByHash(common.BytesToHash(key))
	if header == nil {
		return nil, ErrNilBlockHeader
	}
	return header, nil
}

func (d *Devote) updateConfirmedBlockHeader(chain consensus.ChainReader) error {
	if d.confirmedBlockHeader == nil {
		header, err := d.loadConfirmedBlockHeader(chain)
		if err != nil {
			header = chain.GetHeaderByNumber(0)
			if header == nil {
				return err
			}
		}
		d.confirmedBlockHeader = header
	}

	curHeader := chain.CurrentHeader()
	cycle := uint64(0)
	witnessMap := make(map[string]bool)
	consensusSize := int(15)
	if chain.Config().ChainID.Cmp(big.NewInt(90)) != 0 {
		consensusSize = 1
	}
	for d.confirmedBlockHeader.Hash() != curHeader.Hash() &&
		d.confirmedBlockHeader.Number.Uint64() < curHeader.Number.Uint64() {
		curCycle := curHeader.Time.Uint64() / params.CycleInterval
		if curCycle != cycle {
			cycle = curCycle
			witnessMap = make(map[string]bool)
		}
		// fast return
		// if block number difference less consensusSize-witnessNum
		// there is no need to check block is confirmed
		if curHeader.Number.Int64()-d.confirmedBlockHeader.Number.Int64() < int64(consensusSize-len(witnessMap)) {
			log.Debug("Devote fast return", "current", curHeader.Number.String(), "confirmed", d.confirmedBlockHeader.Number.String(), "witnessCount", len(witnessMap))
			return nil
		}
		witnessMap[curHeader.Witness] = true
		if len(witnessMap) >= consensusSize {
			d.confirmedBlockHeader = curHeader
			if err := d.storeConfirmedBlockHeader(d.db); err != nil {
				return err
			}
			log.Debug("devote set confirmed block header success", "currentHeader", curHeader.Number.String())
			return nil
		}
		curHeader = chain.GetHeaderByHash(curHeader.ParentHash)
		if curHeader == nil {
			return ErrNilBlockHeader
		}
	}
	return nil
}

func PrevSlot(now uint64) uint64 {
	return (now - 1) / params.BlockInterval * params.BlockInterval
}

func NextSlot(now uint64) uint64 {
	return ((now + params.BlockInterval - 1) / params.BlockInterval) * params.BlockInterval
}

func (d *Devote) checkTime(lastBlock *types.Block, now uint64) error {
	prevSlot := PrevSlot(now)
	nextSlot := NextSlot(now)
	if lastBlock.Time().Uint64() >= nextSlot {
		return ErrMinerFutureBlock
	}
	// last block was arrived, or time's up
	if lastBlock.Time().Uint64() == prevSlot || nextSlot-now <= 1 {
		return nil
	}
	return ErrWaitForPrevBlock
}

func (d *Devote) CheckWitness(lastBlock *types.Block, now int64) error {
	if err := d.checkTime(lastBlock, uint64(now)); err != nil {
		return err
	}
	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), lastBlock.Header().Protocol)
	if err != nil {
		return err
	}
	currentCycle := lastBlock.Time().Uint64() / params.CycleInterval
	devoteDB.SetCycle(currentCycle)
	snap := &Snapshot{devoteDB: devoteDB}

	witness, err := snap.lookup(uint64(now))
	if err != nil {
		return err
	}
	log.Info("devote checkWitness lookup", " witness", witness, "signer", d.signer)
	if (witness == "") || witness != d.signer {
		return errUnauthorizedSigner
	}
	return nil
}

func lookup(witnesses []string, now uint64) (witness string, err error) {

	offset := now % params.CycleInterval
	if offset%params.BlockInterval != 0 {
		err = ErrInvalidMinerBlockTime
		return
	}
	offset /= params.BlockInterval

	witnessSize := len(witnesses)
	if witnessSize == 0 {
		err = errors.New("failed to lookup witness")
		return
	}
	offset %= uint64(witnessSize)
	witness = witnesses[offset]
	return
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Devote) SealHash(header *types.Header) common.Hash {
	return sigHash(header)
}

// Close implements consensus.Engine. It's a noop for Devote as there is are no background threads.
func (c *Devote) Close() error {
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *Devote) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{{
		Namespace: "devote",
		Version:   "1.0",
		Service:   &API{chain: chain, devote: c},
		Public:    false,
	}}
}
