package masternode

import (
	"fmt"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"time"
)

const (
	SIGNATURES_REQUIRED = 6
	SIGNATURES_TOTAL    = 10
)

type InstantSend struct {

	// maps for AlreadyHave
	lockRequestAccepted map[common.Hash]*TxLockRequest // tx hash - tx
	lockRequestRejected map[common.Hash]*TxLockRequest // tx hash - tx
	txLockVotes         map[common.Hash]*TxLockVote    // vote hash - vote
	txLockVotesOrphan   map[common.Hash]*TxLockVote    // vote hash - vote

	txLockCandidates map[common.Hash]*TxLockCondidate // tx hash - lock candidate

	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	voteds    map[common.Hash]int
	lockedTxs map[common.Hash]int

	//track masternodes who voted with no txreq (for DOS protection)
	//追踪没有txreq投票的masternodes（用于DOS保护）
	masternodeOrphanVotes map[common.Hash]int
	//std::map<COutPoint, int64_t> mapMasternodeOrphanVotes; // mn outpoint - time
	log log.Logger
}

//received a consensus TxLockRequest
func (is *InstantSend) ProcessTxLockRequest(request *TxLockRequest) bool {

	txHash := request.Hash()

	//check to see if we conflict with existing completed lock
	if is.lockedTxs != nil || is.lockedTxs[txHash] >= 0 {
		// Conflicting with complete lock, proceed to see if we should cancel them both
		is.log.Info("WARNING: Found conflicting completed Transaction Lock", "InstantSend  txid=", txHash, "completed lock txid=", is.lockedTxs[txHash])
	}

	// Check to see if there are votes for conflicting request,
	// if so - do not fail, just warn user
	if is.voteds != nil || is.voteds[txHash] >= 0 {
		is.log.Info("WARNING:Double spend attempt!", "InstantSend txid=", txHash, "Voted txid=", is.voteds[txHash])
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

func (is *InstantSend) vote(condidate *TxLockCondidate) {

	txHash:=condidate.Hash()
	if is.lockRequestAccepted[txHash] == nil{
		return
	}

	nonce:=condidate.txLockRequest.tx.Nonce()
	if nonce <1 {
		is.log.Info("nonce error")
		return
	}

}

func (is *InstantSend) Vote(hash common.Hash) {

	txLockCondidate := is.txLockCandidates[hash]
	if txLockCondidate == nil {
		return
	}
	is.vote(txLockCondidate)
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

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *TxLockCondidate) {

}

func (is *InstantSend) Have(hash common.Hash) bool {

	return is.lockedTxs[hash] > 0
}

func (is *InstantSend) String() string {

	str := fmt.Sprintf("InstantSend Lock Candidates :%v , Votes %v:", len(is.txLockCandidates), len(is.voteds))

	return str
}

type TxLockVote struct {
	txHash          common.Hash
	masternodeId    discover.NodeID
	sig             []byte
	confirmedHeight int
	createdTime     time.Time
	txLocks         map[common.Hash]*TxLock
}

func (tlv *TxLockVote) MasternodeId() discover.NodeID {
	return tlv.masternodeId
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

//主要目的是为了获取投票对象对应的交易需要相关的内容以及投票的相应规则参数
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
	txLock        *TxLock // TxLockRequests by tx hash

}

func NewTxLockCondidata(request *TxLockRequest) TxLockCondidate {

	txlockcondidata := TxLockCondidate{confirmedHeight: -1, createdTime: time.Now(), txLockRequest: request}
	return txlockcondidata

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
