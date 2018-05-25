// Copyright 2015 The go-ethereum Authors
// Copyright 2018 The go-etherzero Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/masternode"
	"math/big"
)

const (
	MNPAYMENTS_SIGNATURES_REQUIRED = 6
	MNPAYMENTS_SIGNATURES_TOTAL    = 10

	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_1 = 70206
	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_2 = 70208
)

var (
	errInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	errECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

// Masternode Payments Class
// Keeps track of who should get paid for which blocks
type MasternodePayments struct {
	cachedBlockNumber *big.Int // Keep track of current block height
	minBlocksToStore  *big.Int
	storageCoeff      *big.Int //masternode count times nStorageCoeff payments blocks should be stored ...
	manager           *MasternodeManager

	votes      map[common.Hash]*MasternodePaymentVote
	blocks     map[uint64]*MasternodeBlockPayees
	lastVote   map[common.Hash]*big.Int
	didNotVote map[common.Hash]*big.Int
}

func NewMasternodePayments(manager *MasternodeManager, number *big.Int) *MasternodePayments {
	payments := &MasternodePayments{
		cachedBlockNumber: number,
		minBlocksToStore:  big.NewInt(1),
		storageCoeff:      big.NewInt(1),
		manager:           manager,
		votes:             make(map[common.Hash]*MasternodePaymentVote),
		blocks:            make(map[uint64]*MasternodeBlockPayees),
		lastVote:          make(map[common.Hash]*big.Int),
		didNotVote:        make(map[common.Hash]*big.Int),
	}
	return payments
}

//hash is blockHash,(!GetBlockHash(blockHash, vote.nBlockHeight - 101))
func (mp *MasternodePayments) Add(hash common.Hash, vote *MasternodePaymentVote) bool {

	if mp.Has(hash) {
		return false
	}
	mp.votes[hash] = vote

	if payee := mp.blocks[vote.number.Uint64()]; payee == nil {
		blockPayees := NewMasternodeBlockPayees(vote.number)
		blockPayees.Add(vote)
	} else {
		mp.blocks[vote.number.Uint64()].Add(vote)
	}

	return true
}

func (mp *MasternodePayments) VoteCount() int {
	return len(mp.votes)
}

func (mp *MasternodePayments) BlockCount() int {
	return len(mp.blocks)
}

func (mp *MasternodePayments) Has(hash common.Hash) bool {

	if vote := mp.votes[hash]; vote != nil {
		return vote.IsVerified()
	}
	return false
}

func (mp *MasternodePayments) Clear() {
	mp.blocks = make(map[uint64]*MasternodeBlockPayees)
	mp.votes = make(map[common.Hash]*MasternodePaymentVote)

}

func (mp *MasternodePayments) ProcessBlock(block *types.Block) bool {

	rank, ok := mp.manager.GetMasternodeRank(mp.manager.active.ID)

	if ok {
		log.Info("ProcessBlock -- Unknown Masternode")
		return false
	}
	if rank > MNPAYMENTS_SIGNATURES_TOTAL {
		log.Info("Masternode not in the top ", MNPAYMENTS_SIGNATURES_TOTAL, "( ", rank, ")")
		return false
	}
	// LOCATE THE NEXT MASTERNODE WHICH SHOULD BE PAID
	log.Info("ProcessBlock -- Start: nBlockHeight=", block.String(), " masternode=", mp.manager.active.ID)

	info, err := mp.manager.GetNextMasternodeInQueueForPayment(block.Hash())
	if err != nil {
		log.Info("ERROR: Failed to find masternode to pay", err)
		return false
	}

	vote := NewMasternodePaymentVote(block.Number(), info)
	mp.Add(block.Hash(), vote)

	return true
}

//Handle the voting of other masternodes
func (m *MasternodePayments) Vote(vote *MasternodePaymentVote) bool {

	if m.votes[vote.Hash()] != nil {
		log.Trace("ERROR:Avoid processing same vote multiple times", "hash=", vote.Hash().String(), " , Height:", vote.number.String())
		return false
	}

	m.votes[vote.Hash()] = vote
	// but first mark vote as non-verified,
	// AddPaymentVote() below should take care of it if vote is actually ok

	//vote out of range
	firstBlock := m.cachedBlockNumber.Sub(m.cachedBlockNumber, m.StorageLimit())
	if vote.number.Cmp(firstBlock) > 0 || vote.number.Cmp(m.cachedBlockNumber.Add(m.cachedBlockNumber, big.NewInt(20))) > 0 {
		log.Trace("ERROR:vote out of range: ", "FirstBlock=", firstBlock.String(), ", BlockHeight=", vote.number, " CacheHeight=", m.cachedBlockNumber.String())
		return false
	}

	if !vote.IsVerified() {
		log.Trace("ERROR: invalid message, error:")
		return false
	}

	//canvote

	//checkSignature
	//if vote.CheckSignature(vote.masternode.MasternodeInfo().ID)

	log.Info("masternode_winner vote: ", "address:", vote.masternode.Account.String(), "blockHeight:", vote.number, "cacheHeight:", m.cachedBlockNumber.String(), "Hash:", vote.Hash().String())

	if m.Add(vote.Hash(), vote) {
		//Relay

	}

	return true

}

func (m *MasternodePayments) StorageLimit() *big.Int {

	count := m.manager.masternodes.Len()
	size := big.NewInt(1).Mul(m.storageCoeff, big.NewInt(int64(count)))

	if size.Cmp(m.minBlocksToStore) > 0 {
		return size
	}
	return m.minBlocksToStore
}

type MasternodePayee struct {
	account common.Address
	votes   []*MasternodePaymentVote
}

func NewMasternodePayee(address common.Address, vote *MasternodePaymentVote) *MasternodePayee {

	mp := &MasternodePayee{
		account: address,
	}
	mp.votes = append(mp.votes, vote)
	return mp
}

func (mp *MasternodePayee) Add(vote *MasternodePaymentVote) {

	mp.votes = append(mp.votes, vote)
}

func (mp *MasternodePayee) Count() int {
	return len(mp.votes)
}

func (mp *MasternodePayee) Votes() []*MasternodePaymentVote {
	return mp.votes
}

type MasternodeBlockPayees struct {
	number *big.Int //blockHeight
	payees []*MasternodePayee
	//payees *set.Set
}

func NewMasternodeBlockPayees(number *big.Int) *MasternodeBlockPayees {

	payee := &MasternodeBlockPayees{
		number: number,
	}
	return payee
}

//vote
func (mbp *MasternodeBlockPayees) Add(vote *MasternodePaymentVote) {

	//When the masternode has been voted
	//info := vote.masternode.MasternodeInfo()
	for _, mp := range mbp.payees {
		if mp.account == vote.masternode.Account {
			mp.Add(vote)
			return
		}
	}
	payee := NewMasternodePayee(vote.masternode.Account, vote)
	mbp.payees = append(mbp.payees, payee)

}

//select the Masternode that has been voted the most
func (mbp *MasternodeBlockPayees) Best() (common.Address, bool) {

	if len(mbp.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	votes := -1
	hash := common.Address{}

	for _, payee := range mbp.payees {
		if votes < payee.Count() {
			hash = payee.account
			votes = payee.Count()
		}
	}
	return hash, votes > -1
}

//Used to record the last winning block of the masternode. At least 2 votes need to be satisfied
// Has(2,masternode.account)
func (mbp *MasternodeBlockPayees) Has(num int, address common.Address) bool {
	if len(mbp.payees) < 1 {
		log.Info("ERROR: ", "couldn't find any payee!")
	}
	for _, payee := range mbp.payees {
		if payee.Count() >= num && payee.account == address {
			return true
		}
	}
	return false
}

// vote for the winning payment
type MasternodePaymentVote struct {
	number     *big.Int //blockHeight
	masternode *masternode.Masternode

	KeySize int
}

//Voted block number,activeMasternode
func NewMasternodePaymentVote(blockHeight *big.Int, active *masternode.Masternode) *MasternodePaymentVote {

	vote := MasternodePaymentVote{
		number:     blockHeight,
		masternode: active,
		KeySize:    0,
	}
	return &vote
}

func (mpv *MasternodePaymentVote) Hash() common.Hash {

	tlvHash := rlpHash([]interface{}{
		mpv.number,
		mpv.masternode.ID,
	})
	return tlvHash
}

func (mpv *MasternodePaymentVote) CheckSignature(pubkey, signature []byte) bool {
	return crypto.VerifySignature(pubkey, mpv.Hash().Bytes(), signature)
}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *MasternodePaymentVote) Verify(sighash []byte, signature string, key interface{}) error {

	// Get the key
	var ecdsaKey *ecdsa.PublicKey
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		ecdsaKey = k
	default:
		return errInvalidKeyType
	}

	r := big.NewInt(0).SetBytes(sighash[:m.KeySize])
	s := big.NewInt(0).SetBytes(sighash[m.KeySize:])

	// Verify the signature
	if verifystatus := ecdsa.Verify(ecdsaKey, sighash, r, s); verifystatus == true {
		return nil
	} else {
		return errECDSAVerification
	}
}

// Implements the Sign method from SigningMethod
// For this signing method, key must be an ecdsa.PrivateKey struct
func (m *MasternodePaymentVote) Sign(signingString common.Hash, key interface{}) (string, error) {
	// Get the key
	var ecdsaKey *ecdsa.PrivateKey
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		ecdsaKey = k
	default:
		return "", errInvalidKeyType
	}
	// Sign the string and return r, s
	if r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, signingString[:]); err == nil {
		curveBits := ecdsaKey.Curve.Params().BitSize
		keyBytes := curveBits / 8
		if curveBits%8 > 0 {
			keyBytes += 1
		}

		// We serialize the outpus (r and s) into big-endian byte arrays and pad
		// them with zeros on the left to make sure the sizes work out. Both arrays
		// must be keyBytes long, and the output must be 2*keyBytes long.
		rBytes := r.Bytes()
		rBytesPadded := make([]byte, keyBytes)
		copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

		sBytes := s.Bytes()
		sBytesPadded := make([]byte, keyBytes)
		copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

		out := append(rBytesPadded, sBytesPadded...)

		return string(out[:]), nil
	} else {
		return "", err
	}
}

func (v *MasternodePaymentVote) IsVerified() bool {
	return true
}

//TODO:Need to improve the judgment of vote validity in MasternodePayments and increase the validity of the voting master node
func (v *MasternodePaymentVote) CheckValid(height *big.Int) (bool, error) {

	// info := v.masternode.MasternodeInfo()

	var minRequiredProtocal uint = 0

	if v.number.Cmp(height) > 0 {
		minRequiredProtocal = MIN_MASTERNODE_PAYMENT_PROTO_VERSION_1
	} else {
		minRequiredProtocal = MIN_MASTERNODE_PAYMENT_PROTO_VERSION_2
	}

	if v.masternode.ProtocolVersion < minRequiredProtocal {
		return false, fmt.Errorf("Masternode protocol is too old: ProtocolVersion=%d, MinRequiredProtocol=%d", v.masternode.ProtocolVersion, minRequiredProtocal)
	}

	if v.number.Cmp(height) < 0 {
		return true, nil
	}
	//v.number

	//TODO:Voting validity check is not judged here

	// Only masternodes should try to check masternode rank for old votes - they need to pick the right winner for future blocks.
	// Regular clients (miners included) need to verify masternode rank for future block votes only.

	return true, nil
}
