package masternode

import (
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"github.com/ethzero/go-ethzero/core/types"
	"time"
	"math/big"
)

const(
	SIGNATURES_REQUIRED        = 6;
	SIGNATURES_TOTAL           = 10;
)

type InstantSend struct {

	// maps for AlreadyHave
	lockRequestAccepted map[common.Hash]TxLockRequest // tx hash - tx
	lockRequestRejected map[common.Hash]TxLockRequest// tx hash - tx
	txLockVotes map[common.Hash]TxLockVote // vote hash - vote
	txLockVotesOrphan map[common.Hash]TxLockVote // vote hash - vote

	txLockCandidates map[common.Hash]TxLockCondidate // tx hash - lock candidate

	//std::map<COutPoint, std::set<uint256> > mapVotedOutpoints; // utxo - tx hash set
	//std::map<COutPoint, uint256> mapLockedOutpoints; // utxo - tx hash
	voteds map[common.Hash]*big.Int
	lockedTxs map[common.Hash]*big.Int

	//track masternodes who voted with no txreq (for DOS protection)
	//追踪没有txreq投票的masternodes（用于DOS保护）
	masternodeOrphanVotes map[common.Hash]*big.Int
	//std::map<COutPoint, int64_t> mapMasternodeOrphanVotes; // mn outpoint - time
}


//这个类是投票的辅助类，投票和创建侯选对象都需要用到
type TxLock struct{

	txhash common.Hash
	masternodeVotes map[discover.NodeID]*TxLockVote
	attacked bool
}

func (tl *TxLock) CountVotes() int{
	if(tl.attacked){
		return 0
	}else{
		return len(tl.masternodeVotes)
	}
}

func (tl *TxLock) IsReady() bool{

	return !tl.attacked && tl.CountVotes()>=SIGNATURES_REQUIRED
}

type TxLockRequest struct {

	tx *types.Transaction

}

func (tq *TxLockRequest) Hash() common.Hash{
	return tq.tx.Hash()
}

func (tq *TxLockRequest) MaxSignatures() int{
	return int(tq.tx.Size())*SIGNATURES_TOTAL
}

func (tq *TxLockRequest) IsValid() bool{

	return tq.tx.CheckNonce()
}

func (tq *TxLockRequest) Tx() *types.Transaction{
	return tq.tx
}

type TxLockCondidate struct {

	confirmedHeight int
	createdTime time.Time
	txLockRequest *TxLockRequest
	txLocks map[common.Hash]*TxLockRequest   // TxLockRequests by tx hash

}

func (tc *TxLockCondidate) Hash() common.Hash{

	return tc.txLockRequest.Hash()
}




type TxLockVote struct {

	txHash            common.Hash
	masternodeId discover.NodeID
	sig []byte
	confirmedHeight int
	createdTime time.Time
	txLocks map[common.Hash]*TxLock
}





