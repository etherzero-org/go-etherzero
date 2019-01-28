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
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/etherzero/go-etherzero/common"
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
	"github.com/hashicorp/golang-lru"
)

const (
	extraVanity        = 32   // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal          = 65   // Fixed number of extra-data suffix bytes reserved for signer seal
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory
	maxSignersSize     = 21   // Number of max singers in current cycle
	testNetSignerSize  = 1
	safeSignerSize     = 16
	wiggleTime         = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers

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

	diffInTurn          = big.NewInt(2) // Block difficulty for in-turn signatures
	diffNoTurn          = big.NewInt(1) // Block difficulty for out-of-turn signatures
	masternodeDifficult = big.NewInt(1) // Block difficult for masternode consensus

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

// NOTE: sigHash was copy from clique
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

type Devote struct {
	config *params.DevoteConfig // Consensus engine configuration parameters
	db     ethdb.Database       // Database to store and retrieve snapshot checkpoints
	stop   chan bool
	lock   sync.RWMutex // Protects the signer fields

	signer                      string        // master node nodeid
	signFn                      SignerFn      // signature function
	sigcache                    *lru.ARCCache // Cache of current witness
	recents                     *lru.ARCCache // Snapshots for recent block to speed up reorgs
	confirmedBlockHeader        *types.Header
	masternodeListFn            MasternodeListFn             //get current all masternodes
	governanceContractAddressFn GetGovernanceContractAddress //get current GovernanceContractAddress

}

func NewDevote(config *params.DevoteConfig, db ethdb.Database) *Devote {

	sigcache, _ := lru.NewARC(inmemorySignatures)
	return &Devote{
		config:   config,
		db:       db,
		sigcache: sigcache,
	}
}

// calculate return  masternode list in the Cycle.
// key   -- nodeid
// value -- votes count
func calculate(hash common.Hash, origins []string) (map[string]*big.Int, error) {

	result := make(map[string]*big.Int)
	for i := 0; i < len(origins); i++ {
		item := origins[i]
		bytes := make([]byte, 8)
		bytes = append(bytes, []byte(item)...)
		bytes = append(bytes, hash[:]...)
		weight := int64(binary.LittleEndian.Uint32(crypto.Keccak512(bytes)))

		score := big.NewInt(0)
		score.Add(score, big.NewInt(weight))
		result[item] = score
	}
	log.Debug("snapshot nodes ", "context", origins, "count", len(origins))
	return result, nil
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
		curCycle := curHeader.Time.Uint64() / d.config.Period
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

// store inserts the snapshot into the database.
func (s *Devote) storeConfirmedBlockHeader(db ethdb.Database) error {
	db.Put(confirmedBlockHead, s.confirmedBlockHeader.Hash().Bytes())
	return nil
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
	header.Witness = d.signer
	header.Difficulty = d.CalcDifficulty(chain, header.Time.Uint64(), parent)

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
	uncles []*types.Header, receipts []*types.Receipt, db *devotedb.DevoteDB) (*types.Block, error) {

	var (
		maxWitnessSize = 21
	)
	if chain.Config().ChainID.Cmp(big.NewInt(90)) != 0 {
		maxWitnessSize = 1
	}
	number := maxWitnessSize

	parent := chain.GetHeaderByHash(header.ParentHash)
	stableBlockNumber := new(big.Int).Sub(parent.Number, big.NewInt(int64(number)))
	if stableBlockNumber.Cmp(big.NewInt(0)) < 0 {
		stableBlockNumber = big.NewInt(0)
	}
	// Accumulate block rewards and commit the final state root
	govaddress, gerr := d.governanceContractAddressFn(stableBlockNumber)
	if gerr != nil {
		return nil, fmt.Errorf("get current governance address err:%s", gerr)
	}
	AccumulateRewards(govaddress, state, header, uncles)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

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
	if header.Time.Cmp(big.NewInt(time.Now().Unix())) > 0 {
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
	// Masternode Difficulty always 1
	if header.Difficulty.Cmp(masternodeDifficult) != 0 {
		return errInvalidDifficulty
	}
	// Ensure that the block doesn't contain any uncles which are meaningless in devote
	if header.UncleHash != uncleHash {
		return errInvalidUncleHash
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyForkHashes(chain.Config(), header, false); err != nil {
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
	if parent.Time.Uint64()+params.BlockInterval > header.Time.Uint64() {
		return ErrInvalidTimestamp
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
	if err != nil {
		// log.Debug("devote verifySeal failed ", "cycle Hash", devoteProtocol.CycleTrie())
		return err
	}

	currentCycle := parent.Time.Uint64() / params.CycleInterval
	devoteDB.SetCycle(currentCycle)
	witnesses, err := devoteDB.GetWitnesses(currentCycle)
	if err != nil {
		return err
	}
	witness, err := lookup(header.Time.Uint64(), witnesses)
	if err != nil {
		return err
	}
	if err := d.verifyBlockSigner(witness, header); err != nil {
		return err
	}
	return d.updateConfirmedBlockHeader(chain)
}

func (d *Devote) verifyBlockSigner(witness string, header *types.Header) error {
	signer, err := ecrecover(header)
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

func lookup(now uint64, witnesses []string) (witness string, err error) {

	offset := now % Epoch
	if offset%Period != 0 {
		err = ErrInvalidMinerBlockTime
		return
	}
	offset /= Period
	size := len(witnesses)
	if size == 0 {
		err = errors.New("failed to lookup witness")
		return
	}
	offset %= uint64(size)
	witness = witnesses[offset]
	return
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

func (d *Devote) CheckWitness(chain consensus.ChainReader,lastBlock *types.Block, now int64) error {

	snap, err := d.snapshot(chain, lastBlock.Number().Uint64(), lastBlock.Hash(), nil)
	if err := d.checkTime(lastBlock, uint64(now)); err != nil {
		return err
	}
	witnesses := snap.signers()
	witness, err := lookup(uint64(now), witnesses)
	if err != nil {
		return err
	}
	log.Info("devote checkWitness lookup", " witness", witness, "signer", d.signer)
	if (witness == "") || witness != d.signer {
		return ErrInvalidBlockWitness
	}
	return nil
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (d *Devote) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {

	var safeSize = 5
	if chain.Config().ChainID.Cmp(big.NewInt(90)) != 0 {
		safeSize = 1
	}
	log.Info("safeSize", safeSize)
	header := block.Header()
	number := header.Number.Uint64()
	// Don't hold the signer fields for the entire sealing procedure
	d.lock.RLock()
	signer, signFn := d.signer, d.signFn
	d.lock.RUnlock()
	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(header.Time.Int64(), 0).Sub(time.Now()) // nolint: gosimple

	// Sealing the genesis block is not supported
	if number == 0 {
		return nil, errUnknownBlock
	}
	blockTime := time.Now().Unix()
	block.Header().Time.SetInt64(blockTime)
	parent := chain.GetHeaderByHash(header.ParentHash)
	stableBlockNumber := new(big.Int).Sub(parent.Number, big.NewInt(maxSignersSize))
	if stableBlockNumber.Cmp(big.NewInt(0)) < 0 {
		stableBlockNumber = big.NewInt(0)
	}

	all, merr := d.masternodeListFn(stableBlockNumber)
	if merr != nil {
		return nil, fmt.Errorf("get current masternodes err:%s", merr)
	}
	log.Debug("finalize get masternode ", "stableBlockNumber", stableBlockNumber, "nodes", all)
	cycle := header.Time.Uint64() / params.CycleInterval
	snap, err := d.snapshot(chain, header.Number.Uint64(), header.Hash(), nil)
	if err != nil {
		return nil, fmt.Errorf("got error when voting next cycle, err: %s", err)
	}
	singerMap := snap.Signers
	if _, ok := singerMap[signer]; !ok {
		return nil, errUnauthorizedSigner
	}
	// If we're amongst the recent signers, wait for the next block
	for seen, recent := range snap.Recents {
		if recent == signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(singerMap)/2 + 1); number < limit || seen > number-limit {
				log.Info("Signed recently, must wait for others, ", "signer", signer, "seen", seen, "number", number, "limit", limit)
				return nil, fmt.Errorf("Signed recently, must wait for others")
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

	w, know := d.sigcache.Get(cycle)
	if know {
		return nil, fmt.Errorf("got error when voting next cycle, err: %s", err)
	}
	db, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), parent.Protocol)
	if err != nil {
		// log.Debug("devote verifySeal failed ", "cycle Hash", devoteProtocol.CycleTrie())
		return nil, err
	}

	witness := w.([]string)
	db.SetWitnesses(cycle, witness)
	db.Commit()
	//snap := newSnapshot(d.config, header.Number.Uint64(), cycle, header.Hash(), witness, db)
	//log.Debug("Initializing a new cycle", "witnesses count", len(witness), "current", cycle, "next", cycle+1)

	//miner Rolling
	log.Debug("rolling ", "Number", header.Number, "parnetTime", parent.Time.Uint64(),
		"headerTime", header.Time.Uint64(), "witness", header.Witness)
	db.Rolling(parent.Time.Uint64(), header.Time.Uint64(), header.Witness)
	db.Commit()
	header.Protocol = db.Protocol()

	// Sign all the things!
	sighash, err := signFn(signer, sigHash(header).Bytes())
	if err != nil {
		return nil, err
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
	}()
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	return block.WithSeal(header), nil
}

func (d *Devote) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	if chain.Config().IsDevote(parent.Number) {
		snap, err := d.snapshot(chain, parent.Number.Uint64(), parent.Hash(), nil)
		if err != nil {
			return nil
		}
		return CalcDifficulty(snap, d.signer)
	} else {
		return big.NewInt(1)
	}
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

func (d *Devote) Authorize(signer string, signFn SignerFn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.signer = signer
	d.signFn = signFn
	log.Info("devote Authorize ", "signer", signer)
}

func (d *Devote) SetGetMasternodesFn(masternodeListFn MasternodeListFn) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.masternodeListFn = masternodeListFn
}

func (d *Devote) SetGetGovernanceContractAddress(goveAddress GetGovernanceContractAddress) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.governanceContractAddressFn = goveAddress
}

// ecrecover extracts the Masternode account ID from a signed header.
func ecrecover(header *types.Header) (string, error) {
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
	return id, nil
}

func PrevSlot(now uint64) uint64 {
	return (now - 1) / params.BlockInterval * params.BlockInterval
}

func NextSlot(now uint64) uint64 {
	return ((now + params.BlockInterval - 1) / params.BlockInterval) * params.BlockInterval
}

//election create a new cycle witnesses by block hash to
// the original one.
func (d *Devote) election(chain consensus.ChainReader, parent *types.Header, nodes []string, safeSize int, maxWitnessSize uint64) ([]string, error) {

	var sortedWitnesses []string
	current := chain.CurrentHeader()
	genesis := chain.GetHeaderByNumber(0)
	genesisCycle := genesis.Time.Uint64() / d.config.Epoch
	prevCycle := parent.Time.Uint64() / d.config.Epoch
	currentCycle := current.Time.Uint64() / d.config.Epoch

	prevCycleIsGenesis := (prevCycle == genesisCycle)
	if prevCycleIsGenesis && prevCycle < currentCycle {
		prevCycle = currentCycle - 1
	}
	//If the witnesses's already cached, return that
	if w, know := d.sigcache.Get(currentCycle); know {
		sortedWitnesses := w.([]string)
		return sortedWitnesses, nil
	} else {
		for i := prevCycle; i < currentCycle; i++ {
			all := make([]string, len(nodes))
			copy(all, nodes)
			votes, err := calculate(parent.Hash(), all)
			if err != nil {
				log.Error("init masternodes ", "err", err)
				return nil, err
			}
			masternodes := sortableAddresses{}
			for masternode, cnt := range votes {
				masternodes = append(masternodes, &sortableAddress{nodeid: masternode, weight: cnt})
			}
			if len(masternodes) < safeSize {
				return nil, fmt.Errorf(" too few masternodes current :%d, safesize:%d", len(masternodes), safeSize)
			}
			sort.Sort(masternodes)
			if len(masternodes) > int(maxWitnessSize) {
				masternodes = masternodes[:maxWitnessSize]
			}

			for _, node := range masternodes {
				sortedWitnesses = append(sortedWitnesses, node.nodeid)
			}
			log.Debug("Snapshot election witnesses ", "currentCycle", currentCycle, "sortedWitnesses", sortedWitnesses)
		}
	}
	return sortedWitnesses, nil
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (d *Devote) snapshot(chain consensus.ChainReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers         []*types.Header
		snap            *Snapshot
		currentCycle    uint64
		sortedWitnesses []string
	)

	current := chain.CurrentHeader()
	currentCycle = current.Time.Uint64() / d.config.Epoch
	//generate snap begin
	for snap == nil {
		//parentNumber:=new(big.Int).Sub(chain.CurrentHeader().Number,big.NewInt(1))
		//parent:=chain.GetHeaderByNumber(parentNumber.Uint64())
		//
		//genesis := chain.GetHeaderByNumber(0)
		//genesisCycle := genesis.Time.Uint64() / d.config.Epoch
		//prevCycle := parent.Time.Uint64() / d.config.Epoch

		// If we're at an checkpoint block, make a snapshot if it's known
		if w, know := d.sigcache.Get(currentCycle); know {
			sortedWitnesses = w.([]string)
		} else {
			all, err := d.masternodeListFn(big.NewInt(int64(number)))
			if err != nil {
				return nil, fmt.Errorf("get current masternodes err:%s", err)
			}
			stabilization := number - 100
			stableBlock := chain.GetHeaderByNumber(stabilization)
			if stableBlock != nil {
				hash = stableBlock.Hash()
			}
			result, err := calculate(hash, all)
			masternodes := sortableAddresses{}
			for masternode, cnt := range result {
				masternodes = append(masternodes, &sortableAddress{nodeid: masternode, weight: cnt})
			}
			sort.Sort(masternodes)
			if len(masternodes) > int(maxSignersSize) {
				masternodes = masternodes[:maxSignersSize]
			}
			for _, node := range masternodes {
				sortedWitnesses = append(sortedWitnesses, node.nodeid)
			}
		}
		cycle := number / d.config.Epoch
		context := []interface{}{
			"cycle", cycle,
			"signers", sortedWitnesses,
			"hash", hash,
			"number", number,
		}
		log.Debug("Elected new cycle signers", context...)
		snap = newSnapshot(d.config, number, cycle, hash, sortedWitnesses)
		if err := snap.store(d.db); err != nil {
			return nil, err
		}
		d.sigcache.Add(cycle, sortedWitnesses)
		log.Trace("Stored checkpoint snapshot to disk", "number", number, "hash", hash)

		var header *types.Header
		if number == 0 && (header.Time.Uint64()%d.config.Epoch == 0) {
			checkpoint := chain.GetHeaderByNumber(number)
			if checkpoint != nil {
				// No Epoch for this header, gather the header and move backward
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
		}
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
	if snap.Number%d.config.Epoch == 0 && len(headers) > 0 {
		if err = snap.store(d.db); err != nil {
			return nil, err
		}
		log.Info("Stored snapshot to disk", "number", snap.Number, "hash", snap.Hash)
	}
	return snap, err
}

// APIs implements consensus.Engine, returning the user facing RPC APIs.
func (d *Devote) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{{
		Namespace: "devote",
		Version:   "1.0",
		Service:   &API{chain: chain, devote: d},
		Public:    true,
	}}
}

// Close implements consensus.Engine. It's a noop for Devote as there is are no background threads.
func (c *Devote) Close() error {
	return nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (c *Devote) SealHash(header *types.Header) common.Hash {
	return sigHash(header)
}
