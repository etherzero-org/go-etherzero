package eth

import (
	"fmt"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/crypto/sha3"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/masternode"
	"github.com/ethzero/go-ethzero/rlp"
)

const (
	SIGNATURES_TOTAL = 10
)

type InstantSend struct {

	// maps for AlreadyHave
	lockRequestAccepted map[common.Hash]*types.TxLockRequest // tx hash - tx
	lockRequestRejected map[common.Hash]*types.TxLockRequest // tx hash - tx
	txLockedVotes       map[common.Hash]*types.TxLockVote    // vote hash - vote
	txLockVotesOrphan   map[common.Hash]*types.TxLockVote    // vote hash - vote

	Candidates map[common.Hash]*types.TxLockCondidate // tx hash - lock candidate

	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	voteds    map[common.Hash]int                  //用于缓存本地的投票对象，实际只有一笔
	lockedTxs map[common.Hash]*types.TxLockRequest //

	//track masternodes who voted with no txreq (for DOS protection)
	//追踪没有txreq投票的masternodes（用于DOS保护）
	masternodeOrphanVotes map[common.Hash]int
	//std::map<COutPoint, int64_t> mapMasternodeOrphanVotes; // mn outpoint - time
	log log.Logger

	active *masternode.Masternode
}

//received a consensus TxLockRequest
func (is *InstantSend) ProcessTxLockRequest(request *types.TxLockRequest) bool {

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
	is.TryToFinalizeLockCandidate(is.Candidates[txHash])

	return true
}

func (is *InstantSend) vote(condidate *types.TxLockCondidate) {

	txHash := condidate.Hash()
	if _, ok := is.lockRequestAccepted[txHash]; !ok {
		return
	}

	txlockRequest := condidate.TxLockRequest()
	nonce := txlockRequest.Tx().Nonce()
	if nonce < 1 {
		is.log.Info("nonce error")
		return
	}

	var alreadyVoted bool = false

	if _, ok := is.voteds[txHash]; !ok {
		txLockCondidate := is.Candidates[txHash] //找到当前交易的侯选人
		if txLockCondidate.HasMasternodeVoted(is.active.ID) {
			alreadyVoted = true
			is.log.Info("CInstantSend::Vote -- WARNING: We already voted for this outpoint, skipping: txHash=", txHash, ", masternodeid=", is.active.ID.String())
			return
		}
	}

	t := types.NewTxLockVote(txHash, is.active.ID) //构建一个投票对象

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
	txLock := is.Candidates[txHash]

	if txLock.AddVote(t) {
		is.log.Info("Vote created successfully, relaying: txHash=", txHash.String(), ", vote=", tvHash.String())
		is.voteds[txHash] = 1
	}

}

func (is *InstantSend) Vote(hash common.Hash) {

	txLockCondidate, ok := is.Candidates[hash]
	if !ok {
		return
	}
	is.vote(txLockCondidate)
	is.TryToFinalizeLockCandidate(txLockCondidate)
}

func (is *InstantSend) CreateTxLockCandidate(request *types.TxLockRequest) bool {

	if !request.IsValid() {
		return false
	}
	txhash := request.Hash()

	if is.Candidates == nil {
		is.log.Info("CreateTxLockCandidate -- new,txid=", txhash.String())
		txlockcondidate := types.NewTxLockCondidata(request)
		is.Candidates[txhash] = &txlockcondidate
	} else {
		is.log.Info("CreateTxLockCandidate -- seen, txid", txhash.String())
	}

	return true
}

func (is *InstantSend) ProcessTxLockVote(vote *types.TxLockVote) bool {

	txhash := vote.Hash()
	txLockCondidate := is.Candidates[txhash]

	is.log.Info("ProcessTxLockVote -- Transaction Lock Vote, txid=", txhash.String())
	if _, ok := is.voteds[txhash]; !ok {
		is.voteds[txhash]++
	}
	if txLockCondidate.AddVote(vote) {
		return false
	}

	signatures := txLockCondidate.CountVotes()
	txlockRequest := txLockCondidate.TxLockRequest()
	signaturesMax := txlockRequest.MaxSignatures()
	is.log.Info("ProcessTxLockVote Transaction Lock signatures count:", signatures, "/", signaturesMax, ",vote Hash:", vote.Hash().String())

	is.TryToFinalizeLockCandidate(txLockCondidate)

	return true
}

func (is *InstantSend) ProcessTxLockVotes(votes []*types.TxLockVote) bool {

	for i := range votes {
		if !is.ProcessTxLockVote(votes[i]) {
			is.log.Info("processTxLockVotes vote failed vote Hash:", votes[i].Hash())
		}
	}
	return true
}

func (is *InstantSend) IsLockedInstantSendTransaction(hash common.Hash) bool {

	_, ok := is.Candidates[hash]
	if !ok {
		return false
	}
	return is.lockedTxs[hash] != nil

}

func (is *InstantSend) TryToFinalizeLockCandidate(condidate *types.TxLockCondidate) {
	txLockRequest := condidate.TxLockRequest()

	txHash := txLockRequest.Hash()
	if condidate.IsReady() {
		is.lockedTxs[txHash] = txLockRequest
	}
}

func (is *InstantSend) Have(hash common.Hash) bool {
	return is.lockedTxs[hash] != nil
}

func (is *InstantSend) String() string {

	str := fmt.Sprintf("InstantSend Lock Candidates :", len(is.Candidates), ", Votes :", len(is.voteds))

	return str
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
