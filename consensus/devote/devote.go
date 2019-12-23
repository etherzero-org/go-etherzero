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
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/consensus"
	"github.com/etherzero/go-etherzero/consensus/misc"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/rpc"
	"github.com/hashicorp/golang-lru"

	"golang.org/x/crypto/sha3"

)

const (
	checkpointInterval = 600  // Number of blocks after which to save the snapshot to the database
	extraVanity        = 32   // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal          = 65   // Fixed number of extra-data suffix bytes reserved for signer seal
	inmemorySnapshots  = 128  // Number of recent snapshots to keep in memory
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory

	//maxWitnessSize uint64 = 0
	//safeSize              = maxWitnessSize*2/3 + 1
	//consensusSize         = maxWitnessSize*2/3 + 1
)

var (
	etherzeroBlockReward = big.NewInt(0.3375e+18) // Block reward in wei to masternode account when successfully mining a block
	rewardToCommunity    = big.NewInt(0.1125e+18) // Block reward in wei to community account when successfully mining a block

	timeOfFirstBlock   = uint64(0)
	confirmedBlockHead = []byte("confirmed-block-head")
	uncleHash          = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
)

var (
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")
	// errMissingVanity is returned if a block's extra-data section is shorter than
	// 32 bytes, which is required to store the signer vanity.
	errMissingVanity = errors.New("extra-data 32 byte vanity prefix missing")
	// errMissingSignature is returned if a block's extra-data section doesn't seem
	// to contain a 65 byte secp256k1 signature.
	errMissingSignature = errors.New("extra-data 65 byte suffix signature missing")
	// errInvalidMixDigest is returned if a block's mix digest is non-zero.
	errInvalidMixDigest = errors.New("non-zero mix digest")
	// errInvalidUncleHash is returned if a block contains an non-empty uncle list.
	errInvalidUncleHash  = errors.New("non empty uncle hash")
	errInvalidDifficulty = errors.New("invalid difficulty")
	// errUnauthorizedSigner is returned if a header is signed by a non-authorized entity.
	errUnauthorizedSigner = errors.New("unauthorized signer")

	// ErrInvalidTimestamp is returned if the timestamp of a block is lower than
	// the previous block's timestamp + the minimum block period.
	ErrInvalidTimestamp         = errors.New("invalid timestamp")
	ErrInvalidBlockWitness      = errors.New("invalid block witness")
	ErrMinerFutureBlock         = errors.New("miner the future block")
	ErrWaitForPrevBlock         = errors.New("wait for last block arrived")
	ErrNilBlockHeader           = errors.New("nil block header returned")
	ErrMismatchSignerAndWitness = errors.New("mismatch block signer and witness")
	ErrInvalidMinerBlockTime    = errors.New("invalid time to miner the block")
)

// SignerFn
// string:master node nodeid,[8]byte
// []byte,signature
type SignerFn func(string, []byte) ([]byte, error)

type MasternodeListFn func(number *big.Int) ([]string, error)

type GetGovernanceContractAddress func(number *big.Int) (common.Address, error)


// SealHash returns the hash of a block prior to it being sealed.
func SealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	encodeSigHeader(hasher, header)
	hasher.Sum(hash[:0])
	return hash
}

type Devote struct {
	config *params.DevoteConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints

	signer     string   // master node nodeid
	signFn     SignerFn // signature function
	recents    *lru.ARCCache   // Snapshots for recent block to speed up reorgs
	signatures *lru.ARCCache   // Signatures of recent blocks to speed up mining
	proposals  map[string]bool // Current list of proposals we are pushing

	confirmedBlockHeader        *types.Header
	masternodeListFn            MasternodeListFn             //get current all masternodes
	governanceContractAddressFn GetGovernanceContractAddress //get current GovernanceContractAddress

	mu   sync.RWMutex
	lock sync.RWMutex
	stop chan bool
}

func NewDevote(config *params.DevoteConfig, db ethdb.Database) *Devote {
	// Allocate the snapshot caches and create the engine
	recents, _ := lru.NewARC(inmemorySnapshots)
	signatures, _ := lru.NewARC(inmemorySignatures)
	return &Devote{
		config:     config,
		db:         db,
		signatures: signatures,
		recents:    recents,
		proposals:  make(map[string]bool),
	}
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (d *Devote) snapshot(chain consensus.ChainReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		checkpoint := chain.GetHeaderByNumber(number)
		if checkpoint != nil {
			// If an in-memory snapshot was found, use that
			if s, ok := d.recents.Get(hash); ok {
				log.Debug("Loaded snapshot from Cache", "number", number, "hash", hash)
				snap = s.(*Snapshot)
				break
			}
			// If we're at an checkpoint block, make a snapshot if it's known
			if number == params.GenesisBlockNumber || checkpoint.Time%params.Epoch == 0 {
				hash := checkpoint.Hash()
				devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), checkpoint.Protocol)
				if err != nil || devoteDB == nil {
					log.Info("Snapshot of devote create devoteDB failed by checkpoint.Protocol", "Number", checkpoint.Number, "err", err)
					return nil, err
				}
				newcycle := checkpoint.Time / params.Epoch
				devoteDB.SetCycle(newcycle)
				snap = &Snapshot{
					Number:    number,
					Cycle:     newcycle,
					Hash:      checkpoint.Hash(),
					TimeStamp: checkpoint.Time,
					config:    d.config,
					devoteDB:  devoteDB,
					Signers:   make(map[string]struct{}),
					Recents:   make(map[uint64]string),
				}
				if err := snap.store(d.db); err != nil {
					return nil, err
				}
				log.Info("Stored checkpoint snapshot to disk", "number", number, "hash", hash)
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
			header = chain.GetHeaderByNumber(number)
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

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (d *Devote) Prepare(chain consensus.ChainReader, header *types.Header) error {
	header.Nonce = types.BlockNonce{}
	number := header.Number.Uint64()
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	header.Difficulty = d.CalcDifficulty(chain, header.Time, parent)
	header.Witness = d.signer
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

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (d *Devote) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	maxWitnessSize := int64(21)
	safeSize := int(15)
	if chain.Config().ChainID.Cmp(big.NewInt(90)) != 0 {
		maxWitnessSize = 1
		safeSize = 1
	}
	parent := chain.GetHeaderByHash(header.ParentHash)
	stableBlockNumber := new(big.Int).Sub(parent.Number, big.NewInt(maxWitnessSize))
	if stableBlockNumber.Cmp(big.NewInt(int64(params.GenesisBlockNumber))) < 0 {
		stableBlockNumber = big.NewInt(int64(params.GenesisBlockNumber))
	}
	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), parent.Protocol)
	if err != nil || devoteDB == nil {
		return nil, fmt.Errorf("Can't create DevoteDB by header Protocol , Header.number : %d ", header.Number)
	}
	// Accumulate block rewards and commit the final state root
	govaddress, err := d.governanceContractAddressFn(stableBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("get current gov address failed from contract, err:%s", err)
	}
	AccumulateRewards(govaddress, state, header, uncles)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	cycle := header.Time / params.Epoch
	devoteDB.SetCycle(cycle)
	snap := &Snapshot{config: d.config, devoteDB: devoteDB}
	snap.TimeStamp = header.Time

	if timeOfFirstBlock == 0 {
		if firstBlockHeader := chain.GetHeaderByNumber(params.GenesisBlockNumber + 1); firstBlockHeader != nil {
			timeOfFirstBlock = firstBlockHeader.Time
		}
	}

	nodes, err := d.masternodeListFn(stableBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("get current masternodes failed from contract, err:%s", err)
	}
	genesis := chain.GetHeaderByNumber(params.GenesisBlockNumber)
	//Record the current witness list into the blockchain
	list, err := snap.election(genesis, parent, nodes, safeSize, maxWitnessSize)
	if err != nil {
		return nil, err
	}
	d.signatures.Add(cycle, list)

	//accumulating the signer of block
	log.Debug("rolling ", "Number", header.Number, "parentTime", parent.Time, "headerTime", header.Time, "witness", header.Witness)
	header.Protocol = snap.recording(parent.Time, header.Time, header.Witness)
	return types.NewBlock(header, txs, uncles, receipts), nil
}

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-stake verified author of the block.
func (d *Devote) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Etherzero devote engine.
func (d *Devote) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	return d.verifyHeader(chain, header, nil)
}

func (d *Devote) verifyHeader(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}
	number := header.Number.Uint64()
	// Unnecssary to verify the block from feature
	if int64(header.Time) > time.Now().Unix() {
		return consensus.ErrFutureBlock
	}
	// Check that the extra-data contains both the vanity and signature
	if len(header.Extra) < extraVanity {
		return errMissingVanity
	}
	if len(header.Extra) < extraVanity+extraSeal {
		return errMissingSignature
	}
	// Ensure that the mix digest is zero as we don't have fork protection currently
	if header.MixDigest != (common.Hash{}) {
		return errInvalidMixDigest
	}
	// Difficulty always 1
	if header.Difficulty.Uint64() != 1 {
		return errInvalidDifficulty
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in devote
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
		log.Error("devote consensus verifyHeader was failed ", "err", err)
		return err
	}

	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}
	return nil
}

func (d *Devote) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := d.verifyHeader(chain, header, headers[:i])
			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (d *Devote) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal implements consensus.Engine, checking whether the signature contained
// in the header satisfies the consensus protocol requirements.
func (d *Devote) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return d.verifySeal(chain, header, nil)
}

func (d *Devote) verifySeal(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}

	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), parent.Protocol)
	if err != nil || devoteDB == nil {
		log.Error("devote consensus verifySeal failed", "err", err)
		return err
	}

	currentcycle := parent.Time / params.Epoch
	devoteDB.SetCycle(currentcycle)
	snap := newSnapshot(d.config, devoteDB)
	snap.sigcache = d.signatures

	witness, err := snap.lookup(header.Time)
	if err != nil {
		return err
	}
	if err := d.verifyBlockSigner(witness, header); err != nil {
		return err
	}
	return d.updateConfirmedBlockHeader(chain)
}

func (d *Devote) verifyBlockSigner(witness string, header *types.Header) error {
	signer, err := ecrecover(header, d.signatures)
	if err != nil {
		return err
	}
	if signer != witness {
		return fmt.Errorf("invalid block witness signer: %s,witness: %s\n", signer, witness)
	}
	if signer != header.Witness {
		return ErrMismatchSignerAndWitness
	}
	return nil
}

func (d *Devote) checkTime(lastBlock *types.Block, now uint64) error {
	prevSlot := PrevSlot(now)
	nextSlot := NextSlot(now)
	if lastBlock.Time() >= nextSlot {
		return ErrMinerFutureBlock
	}
	// last block was arrived, or time's up
	if lastBlock.Time() == prevSlot || nextSlot-now <= 1 {
		return nil
	}
	return ErrWaitForPrevBlock
}

func (d *Devote) CheckWitness(lastBlock *types.Block, now int64) error {
	if err := d.checkTime(lastBlock, uint64(now)); err != nil {
		return err
	}
	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), lastBlock.Header().Protocol)
	if err != nil || devoteDB == nil {
		log.Error("CheckWitness Failed ", "BlockNumber", lastBlock.Number(), "err", err)
		return err
	}
	currentCycle := lastBlock.Time() / params.Epoch
	devoteDB.SetCycle(currentCycle)
	snap := newSnapshot(d.config, devoteDB)
	snap.sigcache = d.signatures

	witness, err := snap.lookup(uint64(now))
	if err != nil {
		return err
	}
	log.Info("devote checkWitness lookup", " witness", witness, "signer", d.signer, "cycle", currentCycle, "blockNumber", lastBlock.Number())
	if (witness == "") || witness != d.signer {
		return ErrInvalidBlockWitness
	}
	logTime := time.Now().Format("[2006-01-02 15:04:05]")
	fmt.Printf("%s [CheckWitness] Found my witness(%s)\n", logTime, witness)
	return nil
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (d *Devote) Seal(chain consensus.ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) (*types.Block, error) {
	header := block.Header()
	number := header.Number.Uint64()
	// Sealing the genesis block is not supported
	if number == 0 {
		return nil, errUnknownBlock
	}
	// Don't hold the signer fields for the entire sealing procedure
	d.lock.RLock()
	signer, signFn := d.signer, d.signFn
	d.lock.RUnlock()
	// Bail out if we're unauthorized to sign a block
	snap, err := d.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return nil, err
	}

	last := chain.CurrentHeader()
	now := time.Now().Unix()
	diff := now - int64(last.Time)
	if diff > 30 {
		snap.Recents = make(map[uint64]string)
	}
	singerMap := snap.Signers
	// If we're amongst the recent signers, wait for the next block
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(singerMap)/2 + 1); number < limit || seen > number-limit {
				log.Info("Signed recently, must wait for others, ", "signer", signer, "seen", seen, "number", number, "limit", limit)
				return nil, nil
			}
			log.Info("Passed Signed recently, ", "signer", signer, "seen", seen, "number", number, "limit", uint64(len(singerMap)/2+1))
		}
	}

	// time's up, sign the block
	sighash, err := signFn(d.signer, SealHash(header).Bytes())
	if err != nil {
		return nil, err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	return block.WithSeal(header), nil
}

func (d *Devote) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return big.NewInt(1)
}

func (d *Devote) Authorize(signer string, signFn SignerFn) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.signer = signer
	d.signFn = signFn
	log.Info("devote Authorize ", "signer", signer)
}

func (d *Devote) Masternodes(masternodeListFn MasternodeListFn) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.masternodeListFn = masternodeListFn
}

func (d *Devote) GovernanceContract(fn GetGovernanceContractAddress) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.governanceContractAddressFn = fn
}

// ecrecover extracts the Masternode account ID from a signed header.
func ecrecover(header *types.Header, sigcache *lru.ARCCache) (string, error) {
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return "", errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]
	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(SealHash(header).Bytes(), signature)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%x", pubkey[1:9])
	return id, nil
}

func (d *Devote) updateConfirmedBlockHeader(chain consensus.ChainReader) error {
	if d.confirmedBlockHeader == nil {
		header, err := d.loadConfirmedBlockHeader(chain)
		if err != nil {
			header = chain.GetHeaderByNumber(params.GenesisBlockNumber)
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
		curCycle := curHeader.Time / params.Epoch
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

// FinalizeAndAssemble implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state and assembling the block.
func (d *Devote) FinalizeAndAssemble(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	accumulateRewards(chain.Config(), state, header, uncles)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, uncles, receipts), nil
}

// store inserts the snapshot into the database.
func (d *Devote) storeConfirmedBlockHeader(db ethdb.Database) error {
	db.Put(confirmedBlockHead, d.confirmedBlockHeader.Hash().Bytes())
	return nil
}

func (d *Devote) loadConfirmedBlockHeader(chain consensus.ChainReader) (*types.Header, error) {

	key, err := d.db.Get(confirmedBlockHead)
	if err != nil {
		return nil, err
	}
	header := chain.GetHeaderByHash(common.BytesToHash(key))
	if header == nil {
		return nil, ErrNilBlockHeader
	}
	return header, nil
}

func PrevSlot(now uint64) uint64 {
	return (now - 1) / params.Period * params.Period
}

func NextSlot(now uint64) uint64 {
	return ((now + params.Period - 1) / params.Period) * params.Period
}

// APIs implements consensus.Engine, returning the user facing RPC APIs.
func (d *Devote) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{{
		Namespace: "devote",
		Version:   "1.0",
		Service:   &API{chain: chain, devote: d},
		Public:    true,
	},}
}

// Close implements consensus.Engine. It's a noop for Devote as there is are no background threads.
func (d *Devote) Close() error {
	return nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (d *Devote) SealHash(header *types.Header) common.Hash {
	return SealHash(header)
}

func (d *Devote) SetDevoteDB(db ethdb.Database) {
	d.db = db
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	err := rlp.Encode(w, []interface{}{
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
	if err != nil {
		panic("can't encode: " + err.Error())
	}
}

// Some weird constants to avoid constant memory allocs for them.
var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain progression
	blockReward := etherzeroBlockReward
	// Accumulate the rewards for the miner and any included uncles
	reward := new(big.Int).Set(blockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, big8)
		state.AddBalance(uncle.Coinbase, r, header.Number)

		r.Div(blockReward, big32)
		reward.Add(reward, r)
	}
	state.AddBalance(header.Coinbase, reward, header.Number)
}



