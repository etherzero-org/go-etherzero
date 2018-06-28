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
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sort"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/trie"
)

type EpochContext struct {
	TimeStamp   int64
	DevoteContext *types.DevoteContext
	statedb     *state.StateDB
}

// countVotes
func (ec *EpochContext) countVotes() (votes map[common.Address]*big.Int, err error) {

	votes = map[common.Address]*big.Int{}
	cacheTrie := ec.DevoteContext.CacheTrie()
	candidateTrie := ec.DevoteContext.CandidateTrie()
	statedb := ec.statedb

	iterCandidate := trie.NewIterator(candidateTrie.NodeIterator(nil))
	existCandidate := iterCandidate.Next()
	if !existCandidate {
		return votes, errors.New("no candidates")
	}

	for existCandidate {

		candidate := iterCandidate.Value
		candidateAddr := common.BytesToAddress(candidate)
		cacheIterator := trie.NewIterator(cacheTrie.PrefixIterator(candidate))
		existCache := cacheIterator.Next()

		if !existCache {
			votes[candidateAddr] = new(big.Int)
			existCandidate = iterCandidate.Next()
			continue
		}
		for existCache {
			cache := cacheIterator.Value
			score, ok := votes[candidateAddr]
			if !ok {
				score = new(big.Int)
			}
			cacheAddr := common.BytesToAddress(cache)
			weight := statedb.GetBalance(cacheAddr)

			score.Add(score, weight)
			votes[candidateAddr] = score
			existCache = cacheIterator.Next()
		}
		existCandidate = iterCandidate.Next()
	}
	return votes, nil
}

func (ec *EpochContext) kickoutWitness(epoch int64) error {
	witnesses, err := ec.DevoteContext.GetWitnesses()
	if err != nil {
		return fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return errors.New("no witness could be kickout")
	}

	epochDuration := epochInterval
	// First epoch duration may lt epoch interval,
	// while the first block time wouldn't always align with epoch interval,
	// so caculate the first epoch duartion with first block time instead of epoch interval,
	// prevent the validators were kickout incorrectly.
	if ec.TimeStamp-timeOfFirstBlock < epochInterval {
		epochDuration = ec.TimeStamp - timeOfFirstBlock
	}

	needKickoutWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(epoch))
		key = append(key, witness.Bytes()...)
		cnt := int64(0)
		if cntBytes := ec.DevoteContext.MintCntTrie().Get(key); cntBytes != nil {
			cnt = int64(binary.BigEndian.Uint64(cntBytes))
		}
		if cnt < epochDuration/blockInterval/maxValidatorSize/2 {
			// not active validators need kickout
			needKickoutWitnesses = append(needKickoutWitnesses, &sortableAddress{witness, big.NewInt(cnt)})
		}
	}
	// no validators need kickout
	needKickoutValidatorCnt := len(needKickoutWitnesses)
	if needKickoutValidatorCnt <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(needKickoutWitnesses))

	candidateCount := 0
	iter := trie.NewIterator(ec.DevoteContext.CandidateTrie().NodeIterator(nil))
	for iter.Next() {
		candidateCount++
		if candidateCount >= needKickoutValidatorCnt+safeSize {
			break
		}
	}

	for i, witness := range needKickoutWitnesses {
		// ensure candidate count greater than or equal to safeSize
		if candidateCount <= safeSize {
			log.Info("No more candidate can be kickout", "prevEpochID", epoch, "candidateCount", candidateCount, "needKickoutCount", len(needKickoutWitnesses)-i)
			return nil
		}

		if err := ec.DevoteContext.KickoutCandidate(witness.address); err != nil {
			return err
		}
		// if kickout success, candidateCount minus 1
		candidateCount--
		log.Info("Kickout candidate", "prevEpochID", epoch, "candidate", witness.address.String(), "mintCnt", witness.weight.String())
	}
	return nil
}

func (ec *EpochContext) lookupWitness(now int64) (validator common.Address, err error) {
	validator = common.Address{}
	offset := now % epochInterval
	if offset%blockInterval != 0 {
		return common.Address{}, ErrInvalidMintBlockTime
	}
	offset /= blockInterval

	witnesses, err := ec.DevoteContext.GetWitnesses()
	if err != nil {
		return common.Address{}, err
	}
	witnessSize := len(witnesses)
	if witnessSize == 0 {
		return common.Address{}, errors.New("failed to lookup witness")
	}
	offset %= int64(witnessSize)
	//return validators[offset], nil

	return common.HexToAddress("0x44655bd29f63eacf71e715c8b9fd4a4bcc561175"), nil
}

func (ec *EpochContext) tryElect(genesis, parent *types.Header) error {

	genesisEpoch := genesis.Time.Int64() / epochInterval
	prevEpoch := parent.Time.Int64() / epochInterval
	currentEpoch := ec.TimeStamp / epochInterval

	prevEpochIsGenesis := prevEpoch == genesisEpoch
	if prevEpochIsGenesis && prevEpoch < currentEpoch {
		prevEpoch = currentEpoch - 1
	}

	prevEpochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevEpochBytes, uint64(prevEpoch))
	iter := trie.NewIterator(ec.DevoteContext.MintCntTrie().PrefixIterator(prevEpochBytes))
	for i := prevEpoch; i < currentEpoch; i++ {
		// if prevEpoch is not genesis, kickout not active candidate
		if !prevEpochIsGenesis && iter.Next() {
			if err := ec.kickoutWitness(prevEpoch); err != nil {
				return err
			}
		}
		votes, err := ec.countVotes()
		if err != nil {
			return err
		}
		candidates := sortableAddresses{}
		for candidate, cnt := range votes {
			candidates = append(candidates, &sortableAddress{candidate, cnt})
		}
		if len(candidates) < safeSize {
			return errors.New("too few candidates")
		}
		sort.Sort(candidates)
		if len(candidates) > maxValidatorSize {
			candidates = candidates[:maxValidatorSize]
		}

		// shuffle candidates
		seed := int64(binary.LittleEndian.Uint32(crypto.Keccak512(parent.Hash().Bytes()))) + i
		r := rand.New(rand.NewSource(seed))
		for i := len(candidates) - 1; i > 0; i-- {
			j := int(r.Int31n(int32(i + 1)))
			candidates[i], candidates[j] = candidates[j], candidates[i]
		}
		sortedWitnesses := make([]common.Address, 0)
		for _, candidate := range candidates {
			sortedWitnesses = append(sortedWitnesses, candidate.address)
		}

		epochTrie, _ := types.NewEpochTrie(common.Hash{}, ec.DevoteContext.DB())
		ec.DevoteContext.SetEpoch(epochTrie)
		ec.DevoteContext.SetWitnesses(sortedWitnesses)
		log.Info("Come to new epoch", "prevEpoch", i, "nextEpoch", i+1)
	}
	return nil
}

type sortableAddress struct {
	address common.Address
	weight  *big.Int
}
type sortableAddresses []*sortableAddress

func (p sortableAddresses) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p sortableAddresses) Len() int      { return len(p) }
func (p sortableAddresses) Less(i, j int) bool {
	if p[i].weight.Cmp(p[j].weight) < 0 {
		return false
	} else if p[i].weight.Cmp(p[j].weight) > 0 {
		return true
	} else {
		return p[i].address.String() < p[j].address.String()
	}
}
