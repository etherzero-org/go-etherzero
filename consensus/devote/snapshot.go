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

type Controller struct {

	DevoteProtocol *types.DevoteProtocol
	statedb     *state.StateDB

	TimeStamp   int64
}

// votes return  vote list in the Epoch.
func (ec *Controller) votes() (votes map[common.Address]*big.Int, err error) {

	votes = map[common.Address]*big.Int{}
	cacheTrie := ec.DevoteProtocol.CacheTrie()
	candidateTrie := ec.DevoteProtocol.CandidateTrie()
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

func (ec *Controller) uncast(epoch int64) error {
	witnesses, err := ec.DevoteProtocol.GetWitnesses()
	if err != nil {
		return fmt.Errorf("failed to get witness: %s", err)
	}
	if len(witnesses) == 0 {
		return errors.New("no witness could be uncast")
	}

	epochDuration := epochInterval
	// First epoch duration may lt epoch interval,
	// while the first block time wouldn't always align with epoch interval,
	// so caculate the first epoch duartion with first block time instead of epoch interval,
	// prevent the validators were kickout incorrectly.
	if ec.TimeStamp-timeOfFirstBlock < epochInterval {
		epochDuration = ec.TimeStamp - timeOfFirstBlock
	}

	needUncastWitnesses := sortableAddresses{}
	for _, witness := range witnesses {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(epoch))
		key = append(key, witness.Bytes()...)
		cnt := int64(0)
		if cntBytes := ec.DevoteProtocol.MintCntTrie().Get(key); cntBytes != nil {
			cnt = int64(binary.BigEndian.Uint64(cntBytes))
		}
		if cnt < epochDuration/blockInterval/maxValidatorSize/2 {
			// not active witnesses need uncast
			needUncastWitnesses = append(needUncastWitnesses, &sortableAddress{witness, big.NewInt(cnt)})
		}
	}
	// no witnessees need uncast
	needUncastWitnessCnt := len(needUncastWitnesses)
	if needUncastWitnessCnt <= 0 {
		return nil
	}
	sort.Sort(sort.Reverse(needUncastWitnesses))

	candidateCount := 0
	iter := trie.NewIterator(ec.DevoteProtocol.CandidateTrie().NodeIterator(nil))
	for iter.Next() {
		candidateCount++
		if candidateCount >= needUncastWitnessCnt+safeSize {
			break
		}
	}

	for i, witness := range needUncastWitnesses {
		// ensure candidate count greater than or equal to safeSize
		if candidateCount <= safeSize {
			log.Info("No more candidate can be kickout", "prevEpochID", epoch, "candidateCount", candidateCount, "needKickoutCount", len(needUncastWitnesses)-i)
			return nil
		}

		if err := ec.DevoteProtocol.Uncast(witness.address); err != nil {
			return err
		}
		// if uncast success, candidate Count minus 1
		candidateCount--
		log.Info("uncast candidate", "prevEpochID", epoch, "candidate", witness.address.String(), "mintCnt", witness.weight.String())
	}
	return nil
}

func (ec *Controller) lookup(now int64) (witness common.Address, err error) {
	witness = common.Address{}
	offset := now % epochInterval
	if offset%blockInterval != 0 {
		return common.Address{}, ErrInvalidMintBlockTime
	}
	offset /= blockInterval

	witnesses, err := ec.DevoteProtocol.GetWitnesses()
	if err != nil {
		return common.Address{}, err
	}
	witnessSize := len(witnesses)
	if witnessSize == 0 {
		return common.Address{}, errors.New("failed to lookup witness")
	}
	offset %= int64(witnessSize)
	//return witnesses[offset], nil

	return common.HexToAddress("0x44655bd29f63eacf71e715c8b9fd4a4bcc561175"), nil
}

func (ec *Controller) voting(genesis, parent *types.Header) error {

	genesisEpoch := genesis.Time.Int64() / epochInterval
	prevEpoch := parent.Time.Int64() / epochInterval
	currentEpoch := ec.TimeStamp / epochInterval

	prevEpochIsGenesis := prevEpoch == genesisEpoch
	if prevEpochIsGenesis && prevEpoch < currentEpoch {
		prevEpoch = currentEpoch - 1
	}

	prevEpochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(prevEpochBytes, uint64(prevEpoch))
	iter := trie.NewIterator(ec.DevoteProtocol.MintCntTrie().PrefixIterator(prevEpochBytes))
	for i := prevEpoch; i < currentEpoch; i++ {
		// if prevEpoch is not genesis, uncast not active candidate
		if !prevEpochIsGenesis && iter.Next() {
			if err := ec.uncast(prevEpoch); err != nil {
				return err
			}
		}
		votes, err := ec.votes()
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

		epochTrie, _ := types.NewEpochTrie(common.Hash{}, ec.DevoteProtocol.DB())
		ec.DevoteProtocol.SetEpoch(epochTrie)
		ec.DevoteProtocol.SetWitnesses(sortedWitnesses)
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
