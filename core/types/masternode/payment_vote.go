package masternode

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/crypto"
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

// vote for the winning payment
type MasternodePaymentVote struct {
	Number     *big.Int //blockHeight
	MasternodeAccount common.Address
	KeySize int
}

//Voted block number,activeMasternode
func NewMasternodePaymentVote(blockHeight *big.Int, account common.Address) *MasternodePaymentVote {

	vote := MasternodePaymentVote{
		Number:     blockHeight,
		MasternodeAccount: account,
		KeySize:    0,
	}

	return &vote
}

func (mpv *MasternodePaymentVote) Hash() common.Hash {

	tlvHash := rlpHash([]interface{}{
		mpv.Number,
		mpv.MasternodeAccount,
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

	//var minRequiredProtocal uint = 0

	//if v.Number.Cmp(height) > 0 {
	//	minRequiredProtocal = MIN_MASTERNODE_PAYMENT_PROTO_VERSION_1
	//} else {
	//	minRequiredProtocal = MIN_MASTERNODE_PAYMENT_PROTO_VERSION_2
	//}

	//if v.Masternode.ProtocolVersion < minRequiredProtocal {
	//	return false, fmt.Errorf("Masternode protocol is too old: ProtocolVersion=%d, MinRequiredProtocol=%d", v.Masternode.ProtocolVersion, minRequiredProtocal)
	//}

	if v.Number.Cmp(height) < 0 {
		return true, nil
	}
	//v.number

	//TODO:Voting validity check is not judged here

	// Only masternodes should try to check masternode rank for old votes - they need to pick the right winner for future blocks.
	// Regular clients (miners included) need to verify masternode rank for future block votes only.

	return true, nil
}
