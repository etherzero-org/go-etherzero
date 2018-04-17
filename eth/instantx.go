package eth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/masternode"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"github.com/ethzero/go-ethzero/rlp"
	"math/big"
	"time"
)

const (
	SIGNATURES_REQUIRED = 6
	SIGNATURES_TOTAL    = 10
)

var (
	ErrInvalidKeyType = errors.New("key is of invalid type")
	// Sadly this is missing from crypto/ecdsa compared to crypto/rsa
	ErrECDSAVerification = errors.New("crypto/ecdsa: verification error")
)

type InstantSend struct {

	// maps for AlreadyHave
	lockRequestAccepted map[common.Hash]*TxLockRequest // tx hash - tx
	lockRequestRejected map[common.Hash]*TxLockRequest // tx hash - tx
	txLockedVotes       map[common.Hash]*TxLockVote    // vote hash - vote
	txLockVotesOrphan   map[common.Hash]*TxLockVote    // vote hash - vote

	txLockCandidates map[common.Hash]*TxLockCondidate // tx hash - lock candidate

	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	voteds    map[common.Hash]int //用于缓存本地的投票对象，实际只有一笔
	lockedTxs map[common.Hash]int //

	//track masternodes who voted with no txreq (for DOS protection)
	//追踪没有txreq投票的masternodes（用于DOS保护）
	masternodeOrphanVotes map[common.Hash]int
	//std::map<COutPoint, int64_t> mapMasternodeOrphanVotes; // mn outpoint - time
	log log.Logger

	active *masternode.Masternode

	mm *MasternodeManager
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

//received a consensus TxLockRequest
func (is *InstantSend) ProcessTxLockRequest(request *TxLockRequest) bool {

	txHash := request.Hash()

	//check to see if we conflict with existing completed lock

	if _, ok := is.lockedTxs[txHash]; !ok {
		// Conflicting with complete lock, proceed to see if we should cancel them both
		is.log.Info("WARNING: Found conflicting completed Transaction Lock", "InstantSend  txid=", txHash, "completed lock txid=", is.lockedTxs[txHash])
	}

	// Check to see if there are votes for conflicting request,
	// if so - do not fail, just warn user
	if _, ok := is.voteds[txHash]; !ok {
		is.log.Info("WARNING:Double spend attempt!", "InstantSend txid=", txHash, "Voted txid count :", is.voteds[txHash])
	}

	if !is.CreateTxLockCandidate(request) {
		is.log.Info("CreateTxLockCandidate failed, txid=", txHash)
		return false
	}
	// Masternodes will sometimes propagate votes before the transaction is known to the client.
	// If this just happened - lock inputs, resolve conflicting locks, update transaction status
	// forcing external script notification.
	is.TryToFinalizeLockCandidate(is.txLockCandidates[txHash])

	return true
}

func (is *InstantSend) vote(mm *MasternodeManager, condidate *TxLockCondidate) {

	txHash := condidate.Hash()
	if _, ok := is.lockRequestAccepted[txHash]; !ok {
		return
	}

	nonce := condidate.txLockRequest.tx.Nonce()
	if nonce < 1 {
		is.log.Info("nonce error")
		return
	}
	rank, ok := mm.GetMasternodeRank(is.active.ID)
	if !ok {
		is.log.Info("InstantSend::Vote -- Can't calculate rank for masternode ", is.active.ID.String()," rank: ", rank)
		return
	} else if rank > SIGNATURES_TOTAL {
		is.log.Info("InstantSend::Vote -- Masternode not in the top ", SIGNATURES_TOTAL, " (",rank,")")
		return
	}
	is.log.Info("InstantSend::Vote -- In the top ",SIGNATURES_TOTAL," (",rank,")")

	var alreadyVoted bool = false

	if _, ok := is.voteds[txHash]; !ok {
		txLockCondidate := is.txLockCandidates[txHash] //找到当前交易的侯选人
		if txLockCondidate.HasMasternodeVoted(txHash, is.active.ID) {
			alreadyVoted = true
			is.log.Info("CInstantSend::Vote -- WARNING: We already voted for this outpoint, skipping: txHash=",txHash,", masternodeid=",is.active.ID.String())
			return
		}
	}

	t := NewTxLockVote(txHash, is.active.ID) //构建一个投票对象

	if alreadyVoted {
		return
	}
	signByte, err := t.Sign(t.Hash(), is.active.Config().PrivateKey)

	if err != nil {
		return
	}
	sigErr := t.Verify(t.Hash().Bytes(), signByte, is.active.Config().PrivateKey.Public())

	if sigErr != nil {
		return
	}
	tvHash := t.Hash()

	is.txLockedVotes[tvHash] = t
	if is.txLockCandidates[txHash].txLock.AddVote(t) {
		is.log.Info("Vote created successfully, relaying: txHash=",txHash.String(),", vote=",tvHash.String())
		is.voteds[txHash] = 1
	}

}

func (is *InstantSend) Vote(hash common.Hash) {

	txLockCondidate, ok := is.txLockCandidates[hash]
	if !ok {
		return
	}
	is.vote(is.mm, txLockCondidate)
	is.TryToFinalizeLockCandidate(txLockCondidate)
}

func (is *InstantSend) CreateTxLockCandidate(request *TxLockRequest) bool {

	if !request.IsValid() {
		return false
	}
	txhash := request.Hash()

	if is.txLockCandidates == nil {
		is.log.Info("CreateTxLockCandidate -- new,txid=", txhash.String())
		txlockcondidate := NewTxLockCondidata(request)
		txlockcondidate.AddTxLock(txhash)
		is.txLockCandidates[txhash] = &txlockcondidate
	} else {
		is.log.Info("CreateTxLockCandidate -- seen, txid", txhash.String())
	}

	return true
}

func (is *InstantSend) ProcessTxLockVote(vote *TxLockVote) bool {

	txhash := vote.Hash()
	txLockCondidate := is.txLockCandidates[txhash]

	is.log.Info("ProcessTxLockVote -- Transaction Lock Vote, txid=", txhash.String())
	if _, ok := is.voteds[txhash]; !ok {
		is.voteds[txhash]++
	}
	if txLockCondidate.AddVote(vote) {
		return false
	}

	signatures := txLockCondidate.CountVotes()
	signaturesMax := txLockCondidate.txLockRequest.MaxSignatures()
	is.log.Info("ProcessTxLockVote Transaction Lock signatures count:", signatures, "/", signaturesMax, ",vote Hash:", vote.Hash().String())

	is.TryToFinalizeLockCandidate(txLockCondidate)

	return true
}

func (is *InstantSend) IsLockedInstantSendTransaction(hash common.Hash) bool {

	txLockCondidate, ok := is.txLockCandidates[hash]

	if !ok {
		return false
	}
	if txLockCondidate.txLock == nil || txLockCondidate.txLock.txhash != hash {
		return false
	}
	return true

}

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *TxLockCondidate) {
	txHash := condidate.txLockRequest.Hash()

	if condidate.IsAllTxReady() {
		condidate.txLock.txhash = txHash
	}

}

func (is *InstantSend) Have(hash common.Hash) bool {

	return is.lockedTxs[hash] > 0
}

func (is *InstantSend) String() string {

	str := fmt.Sprintf("InstantSend Lock Candidates :", len(is.txLockCandidates),", Votes :",len(is.voteds))

	return str
}

type TxLockVote struct {
	txHash          common.Hash
	masternodeId    discover.NodeID
	sig             []byte
	confirmedHeight int
	createdTime     time.Time
	txLocks         map[common.Hash]*TxLock
	KeySize         int
}

func (tlv *TxLockVote) MasternodeId() discover.NodeID {
	return tlv.masternodeId
}

func NewTxLockVote(hash common.Hash, id discover.NodeID) *TxLockVote {

	tv := &TxLockVote{
		txHash:          hash,
		masternodeId:    id,
		createdTime:     time.Now(),
		confirmedHeight: -1,
		KeySize:         256,
	}
	return tv
}

func (tlv *TxLockVote) Hash() common.Hash {

	tlvHash := rlpHash([]interface{}{
		tlv.txHash,
		tlv.masternodeId,
	})
	return tlvHash
}

func (tlv *TxLockVote) CheckSignature(pubkey, signature []byte) bool {
	return crypto.VerifySignature(pubkey, tlv.Hash().Bytes(), signature)
}

// Implements the Verify method from SigningMethod
// For this verify method, key must be an ecdsa.PublicKey struct
func (m *TxLockVote) Verify(sighash []byte, signature string, key interface{}) error {

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
func (m *TxLockVote) Sign(signingString common.Hash, key interface{}) (string, error) {
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

//这个类是投票的辅助类，投票和创建侯选对象都需要用到
type TxLock struct {
	txhash          common.Hash
	masternodeVotes map[discover.NodeID]*TxLockVote
	attacked        bool
}

func NewTxLock(txid common.Hash) *TxLock {

	txlock := &TxLock{
		txhash:          txid,
		masternodeVotes: make(map[discover.NodeID]*TxLockVote),
		attacked:        false}
	return txlock
}

func (tl *TxLock) CountVotes() int {
	if tl.attacked {
		return 0
	} else {
		return len(tl.masternodeVotes)
	}
}

func (tl *TxLock) HasMasternodeVoted(id discover.NodeID) bool {

	return tl.masternodeVotes[id] != nil
}

func (tl *TxLock) IsReady() bool {

	return !tl.attacked && tl.CountVotes() >= SIGNATURES_REQUIRED
}

func (tl *TxLock) AddVote(vote *TxLockVote) bool {

	if tl.masternodeVotes[vote.MasternodeId()] == nil {
		tl.masternodeVotes[vote.MasternodeId()] = vote
		return true
	}
	return false
}

//主要目的是为了获取投票对象对应的交易，需要相关的内容以及投票的相应规则参数
type TxLockRequest struct {
	tx *types.Transaction
}

func (tq *TxLockRequest) Hash() common.Hash {
	return tq.tx.Hash()
}

func (tq *TxLockRequest) MaxSignatures() int {
	return int(tq.tx.Size()) * SIGNATURES_TOTAL
}

func (tq *TxLockRequest) IsValid() bool {

	return tq.tx.CheckNonce()
}

func (tq *TxLockRequest) Tx() *types.Transaction {
	return tq.tx
}

type TxLockCondidate struct {
	confirmedHeight int
	createdTime     time.Time
	txLockRequest   *TxLockRequest
	txLock          *TxLock // TxLockRequests by tx hash

}

func NewTxLockCondidata(request *TxLockRequest) TxLockCondidate {

	txlockcondidata := TxLockCondidate{confirmedHeight: -1, createdTime: time.Now(), txLockRequest: request}
	return txlockcondidata

}

func (tc *TxLockCondidate) HasMasternodeVoted(hash common.Hash, id discover.NodeID) bool {

	return tc.txLock.HasMasternodeVoted(id)
}

func (tc *TxLockCondidate) Hash() common.Hash {

	return tc.txLockRequest.Hash()
}

func (tc *TxLockCondidate) AddTxLock(txid common.Hash) {
	txlock := NewTxLock(txid)
	tc.txLock = txlock
}

func (tc *TxLockCondidate) AddVote(vote *TxLockVote) bool {
	txlock := tc.txLock
	return txlock.AddVote(vote)
}

func (tc *TxLockCondidate) IsAllTxReady() bool {
	if tc.txLock == nil {
		return false
	}
	return true
}

func (tc *TxLockCondidate) CountVotes() int {
	return tc.txLock.CountVotes()
}
