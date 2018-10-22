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
	"github.com/etherzero/go-etherzero/p2p/discover"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/rpc"
	"github.com/hashicorp/golang-lru"
	//"math/rand"
	"math/rand"
)

const (
	extraVanity        = 32   // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal          = 65   // Fixed number of extra-data suffix bytes reserved for signer seal
	inmemorySignatures = 4096 // Number of recent block signatures to keep in memory
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	//maxWitnessSize uint64 = 0
	//safeSize              = maxWitnessSize*2/3 + 1
	//consensusSize         = maxWitnessSize*2/3 + 1
	wiggleTime = 500 * time.Millisecond // Random delay (per signer) to allow concurrent signers
)

var (
	etherzeroBlockReward = big.NewInt(0.3375e+18) // Block reward in wei to masternode account when successfully mining a block
	rewardToCommunity    = big.NewInt(0.1125e+18) // Block reward in wei to community account when successfully mining a block
	diffInTurn           = big.NewInt(2)          // Block difficulty for in-turn signatures
	diffNoTurn           = big.NewInt(1)          // Block difficulty for out-of-turn signatures

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

	// errUnauthorized is returned if a header is signed by a non-authorized entity.
	errUnauthorized = errors.New("unauthorized")

	// errInvalidVotingChain is returned if an authorization list is attempted to
	// be modified via out-of-range or non-contiguous headers.
	errInvalidVotingChain = errors.New("invalid voting chain")

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

	signer string   // masternode nodeid
	signFn SignerFn // signature function

	recents                     *lru.ARCCache // Snapshots for recent block to speed up reorgs
	signatures                  *lru.ARCCache // Signatures of recent blocks to speed up mining
	confirmedBlockHeader        *types.Header
	masternodeListFn            MasternodeListFn             //get current all masternodes
	governanceContractAddressFn GetGovernanceContractAddress //get current GovernanceContractAddress

	blockTimes map[common.Hash][]int64 //save arrived system time of lastblock
	mu         sync.RWMutex
	stop       chan bool
}

func NewDevote(config *params.DevoteConfig, db ethdb.Database) *Devote {

	// Allocate the snapshot caches and create the engine
	signatures, _ := lru.NewARC(inmemorySignatures)
	recents, _ := lru.NewARC(inmemorySnapshots)

	return &Devote{
		config:     config,
		db:         db,
		recents:    recents,
		signatures: signatures,
		blockTimes: make(map[common.Hash][]int64),
	}
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (c *Devote) snapshot(chain consensus.ChainReader, number uint64, cycle uint64, witnesses []string, hash common.Hash, parents []*types.Header) (*Controller, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Controller
	)
	new := chain.GetHeaderByNumber(number)
	// If we've generated a new checkpoint snapshot, save to disk
	if new.Time.Uint64()%params.CycleInterval == 0 {
		snap = newController(number, cycle, c.signatures, new.Hash(), witnesses)
		if err := snap.store(c.db); err != nil {
			return nil, err
		}
		log.Info("saved voting snapshot to disk", "number", snap.Number, "hash", snap.Hash)
		c.recents.Add(snap.Number, snap)
		return snap, nil
	}

	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := c.recents.Get(number); ok {
			snap = s.(*Controller)
			log.Info("Locaded voting snapshot from cache ","number",snap.Number,"hash",snap.Hash)
			break
		}
		h := chain.GetHeaderByNumber(number)
		if h != nil {
			// If an on-disk checkpoint snapshot can be found, use that
			if h.Time.Uint64()%params.CycleInterval == 0 {
				if s, err := loadSnapshot(c.signatures, c.db, hash); err == nil {
					log.Info("Loaded voting snapshot from disk", "number", number, "hash", hash)
					snap = s
					break
				}
			}
		}

		// If we're at block zero, make a snapshot
		if number == 0 {
			genesis := chain.GetHeaderByNumber(0)
			//if err := c.VerifyHeader(chain, genesis, false); err != nil {
			//	return nil, err
			//}
			signers := make([]string, 21)
			for i := 0; i < len(signers); i++ {
				signers[i] = genesis.Witness
			}
			snap = newController(number, cycle, c.signatures, genesis.Hash(), witnesses)
			if err := snap.store(c.db); err != nil {
				return nil, err
			}
			log.Info("Stored genesis voting snapshot to disk")
			break
		}
		// If we've generated a new checkpoint snapshot, save to disk
		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		//fmt.Printf("devote.go seal parents size:%d \n ", len(parents))
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
	return snap, nil
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
	header.Difficulty = d.CalcDifficulty(chain, header.Time.Uint64(), parent)
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
	uncles []*types.Header, receipts []*types.Receipt, devoteDB *devotedb.DevoteDB) (*types.Block, error) {
	maxWitnessSize := uint64(20)
	safeSize := int(11)
	if chain.Config().ChainID.Cmp(big.NewInt(90)) != 0 {
		maxWitnessSize = 1
		safeSize = 1
	}
	parent := chain.GetHeaderByHash(header.ParentHash)
	number := maxWitnessSize
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

	controller := &Controller{
		devoteDB:  devoteDB,
		TimeStamp: header.Time.Uint64(),
		Recents:   make(map[uint64]string),
	}
	if timeOfFirstBlock == 0 {
		if firstBlockHeader := chain.GetHeaderByNumber(1); firstBlockHeader != nil {
			timeOfFirstBlock = firstBlockHeader.Time.Uint64()
		}
	}

	nodes, merr := d.masternodeListFn(stableBlockNumber)
	if merr != nil {
		return nil, fmt.Errorf("get current masternodes err:%s", merr)
	}
	log.Debug("finalize get masternode ", "stableBlockNumber", stableBlockNumber, "nodes", nodes)
	genesis := chain.GetHeaderByNumber(0)
	first := chain.GetHeaderByNumber(1)
	err := controller.election(genesis, first, parent, nodes, safeSize, maxWitnessSize)
	if err != nil {
		return nil, fmt.Errorf("got error when voting next cycle, err: %s", err)
	}
	//miner Rolling
	log.Debug("rolling ", "Number", header.Number, "parentTime", parent.Time.Uint64(), "headerTime", header.Time.Uint64(), "witness", header.Witness)
	devoteDB.Rolling(parent.Time.Uint64(), header.Time.Uint64(), header.Witness)
	devoteDB.Commit()
	header.Protocol = devoteDB.Protocol()
	//controller.Recents[header.Number.Uint64()] = header.Witness
	fmt.Printf("devote 's Finalize recode last number of signer :number %d,signer :%s,Rencts size:%d \n", header.Number.Uint64(), header.Witness, len(controller.Recents))
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
		//log.Debug("devote verifySeal failed ", "cycle Hash", devoteProtocol.CycleTrie())
		return err
	}

	currentcycle := parent.Time.Uint64() / params.CycleInterval
	devoteDB.SetCycle(currentcycle)
	//controller := &Controller{devoteDB: devoteDB}

	//witness, err := controller.lookup(header.Time.Uint64())
	//if err != nil {
	//	return err
	//}
	witnesses, e := devoteDB.GetWitnesses(currentcycle)
	if e != nil {
		return e
	}

	snap, err := d.snapshot(chain, number-1, currentcycle, witnesses, header.ParentHash, parents)
	if err != nil {
		return err
	}
	// Resolve the authorization key and check against signers
	witness, err := ecrecover(header, d.signatures)
	if err != nil {
		return err
	}

	if _, ok := snap.Signers[header.Witness]; !ok {
		fmt.Printf("devote verifySeal snap.Signers size:%d,value:%s \n ", len(snap.Signers), snap.Signers)
		return errUnauthorized
	}
	for seen, recent := range snap.Recents {
		if recent == witness {
			// Signer is among recents, only fail if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); seen > number-limit {
				fmt.Printf("devote verifySeal snap.Signers size:%d,limit%d \n ", len(snap.Signers), limit)
				return errUnauthorized
			}
		}
	}

	//if err := d.verifyBlockSigner(witness, header); err != nil {
	//	return err
	//}
	return d.updateConfirmedBlockHeader(chain)
}

func (d *Devote) verifyBlockSigner(witness string, header *types.Header) error {
	signer, err := ecrecover(header, d.signatures)
	if err != nil {
		return err
	}
	if signer != witness {
		return fmt.Errorf("invalid block witness signer: %s,witness: %s,time:%d \n", signer, witness, header.Time)
	}
	if signer != header.Witness {
		return ErrMismatchSignerAndWitness
	}
	return nil
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

func (d *Devote) saveNextBlockTime(lastBlock *types.Block) int64 {

	if _, ok := d.blockTimes[lastBlock.Hash()]; !ok {
		d.blockTimes[lastBlock.Hash()] = append(d.blockTimes[lastBlock.Hash()], time.Now().UnixNano())
	}
	i := len(d.blockTimes[lastBlock.Hash()])
	return d.blockTimes[lastBlock.Hash()][0] - d.blockTimes[lastBlock.Hash()][i]
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
	controller := &Controller{devoteDB: devoteDB}

	witness, err := controller.lookup(uint64(now))
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
	header := block.Header()
	number := header.Number.Uint64()
	// Sealing the genesis block is not supported
	if number == 0 {
		return nil, errUnknownBlock
	}
	now := time.Now().Unix()
	//NextSlot := int64(NextSlot(uint64(now)))
	//delay := NextSlot - now
	//
	//log.Info("Devote Seal delay time :", "delay", delay, "NextSlot", NextSlot, "now", now)
	//if delay > 0 {
	//	select {
	//	case <-stop:
	//		return nil, nil
	//	case <-time.After(time.Duration(delay) * time.Second):
	//	}
	//}
	drift := time.Duration(discover.NanoDrift())
	log.Info("Devote Seal delay time :", "NextSlot", NextSlot, "now", now, "drift", drift)
	blockTime := time.Now().Add(-drift).Unix()
	block.Header().Time.SetInt64(blockTime)

	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(d.db), block.Header().Protocol)
	if err != nil {
		return nil, err
	}
	currentCycle := uint64(header.Time.Uint64()) / params.CycleInterval
	devoteDB.SetCycle(currentCycle)

	//	c := Controller{devoteDB: devoteDB}
	//signers, err := c.Witnesses(currentCycle)
	//if err != nil {
	//	return nil, err
	//}
	signers, err := devoteDB.GetWitnesses(currentCycle)
	if err != nil {
		fmt.Printf("get witnesses err :%s\n", err)
		return nil, err
	}

	snap, e := d.snapshot(chain, number-1, currentCycle, signers, header.ParentHash, nil)
	if e != nil {
		fmt.Printf(" create snapshot err%s\n", e)
		return nil, e
	}
	for seen, recent := range snap.Recents {
		fmt.Printf("devote Seal signer's %s in Recents Recents size %d \n", recent, len(snap.Recents))
		if recent == d.signer {
			// Signer is among recents, only wait if the current block doesn't shift it out
			if limit := uint64(len(snap.Signers)/2 + 1); number < limit || seen > number-limit {
				log.Info("Signed recently, must wait for others", "limit", limit, "number", number, "seen", seen)
				<-stop
				return nil, nil
			}
		}
	}

	// Sweet, the protocol permits us to sign the block, wait for our time
	delay := time.Unix(header.Time.Int64(), 0).Sub(time.Now()) // nolint: gosimple
	if header.Difficulty.Cmp(diffNoTurn) == 0 {
		// It's not our turn explicitly to sign, delay it a bit
		wiggle := time.Duration(len(snap.Signers)/2+1) * wiggleTime
		delay += time.Duration(rand.Int63n(int64(wiggle)))

		log.Trace("Out-of-turn signing requested", "wiggle", common.PrettyDuration(wiggle))
	}
	log.Trace("Waiting for slot to sign and propagate", "delay", common.PrettyDuration(delay))

	select {
	case <-stop:
		return nil, nil
	case <-time.After(delay):
	}
	fmt.Printf("devote Seal signer's in Recents Recents  after stop\n")
	// time's up, sign the block
	sighash, err := d.signFn(d.signer, sigHash(header).Bytes())
	if err != nil {
		return nil, err
	}
	copy(header.Extra[len(header.Extra)-extraSeal:], sighash)
	return block.WithSeal(header), nil
}

func (d *Devote) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return big.NewInt(1)
	//currentCycle := uint64(chain.CurrentHeader().Time.Uint64()) / params.CycleInterval
	//snap, err := d.snapshot(chain, parent.Number.Uint64(),currentCycle, parent.Hash(), nil)
	//if err != nil {
	//	return nil
	//}
	//if snap.inturn(snap.Number+1, d.signer) {
	//	return new(big.Int).Set(diffInTurn)
	//}
	//return new(big.Int).Set(diffNoTurn)
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

func (d *Devote) GetGovernanceContractAddress(goveAddress GetGovernanceContractAddress) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.governanceContractAddressFn = goveAddress
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

func PrevSlot(now uint64) uint64 {
	return (now - 1) / params.BlockInterval * params.BlockInterval
}

func NextSlot(now uint64) uint64 {
	return ((now + params.BlockInterval - 1) / params.BlockInterval) * params.BlockInterval
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
