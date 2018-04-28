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
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/masternode"
	"math/big"
)

const (
	MNPAYMENTS_SIGNATURES_REQUIRED         = 6
	MNPAYMENTS_SIGNATURES_TOTAL            = 10
	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_1 = 70206
	MIN_MASTERNODE_PAYMENT_PROTO_VERSION_2 = 70208
)

var (
	ErrInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	ErrECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

type MasternodePayee struct {
	account common.Hash
	votes   []*MasternodePaymentVote
}

func NewMasternodePayee(hash common.Hash, vote *MasternodePaymentVote) *MasternodePayee {

	mp := &MasternodePayee{
		account: hash,
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

	log log.Logger
}

//vote
func (mbp *MasternodeBlockPayees) Add(vote *MasternodePaymentVote) {

	//When the masternode has been voted
	for _, mp := range mbp.payees {
		if mp.account == vote.Hash() {
			mp.Add(vote)
			return
		}
	}
	payee := NewMasternodePayee(vote.Hash(), vote)
	mbp.payees = append(mbp.payees, payee)

}

//select the Masternode that has been voted the most
func (mbp *MasternodeBlockPayees) Best() (common.Hash, bool) {

	if len(mbp.payees) < 1 {
		mbp.log.Info("ERROR: ", "couldn't find any payee!")
	}
	votes := -1
	hash := common.Hash{}

	for _, payee := range mbp.payees {
		if votes < payee.Count() {
			hash = payee.account
			votes = payee.Count()
		}
	}
	return hash, votes > -1
}

//Used to record the last winning block of the masternode. At least 2 votes need to be satisfied
// Have(2,masternode.account)
func (mbp *MasternodeBlockPayees) Have(votes int, hash common.Hash) bool {
	if len(mbp.payees) < 1 {
		mbp.log.Info("ERROR: ", "couldn't find any payee!")
	}
	for _, payee := range mbp.payees {
		if payee.Count() >= votes && payee.account == hash {
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
	log     log.Logger
}

func (mpv *MasternodePaymentVote) Hash() common.Hash {

	tlvHash := rlpHash([]interface{}{
		mpv.number,
		mpv.masternode.MasternodeInfo().ID,
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
		return ErrInvalidKeyType
	}

	r := big.NewInt(0).SetBytes(sighash[:m.KeySize])
	s := big.NewInt(0).SetBytes(sighash[m.KeySize:])

	// Verify the signature
	if verifystatus := ecdsa.Verify(ecdsaKey, sighash, r, s); verifystatus == true {
		return nil
	} else {
		return ErrECDSAVerification
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
		return "", ErrInvalidKeyType
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

type MasternodePayments struct {

	cachedBlockNumber *big.Int
	minBlocksToStore *big.Int
	storageCoeff *big.Int

	votes map[common.Hash]*MasternodePaymentVote
	blocks map[*big.Int]*MasternodeBlockPayees
	lastVote map[common.Hash]*big.Int
	didNotVote map[common.Hash]*big.Int

}

func NewMasternodePayments(){

}

func (mp *MasternodePayments) Add(hash common.Hash,vote *MasternodePaymentVote){

	mp.votes[vote.Hash()]=vote
}

func(mp *MasternodePayments) VoteCount()int{
	return len(mp.votes)
}

func (mp*MasternodePayments) BlockCount()int{
	return len(mp.blocks)
}

func(mp *MasternodePayments) Have() bool{

	return true
}



