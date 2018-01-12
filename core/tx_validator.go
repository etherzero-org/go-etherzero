
package core

import(
	"github.com/ethzero/go-ethzero/core/types"
	"math/big"
)

// validateTx checks whether a transaction is valid according to the consensus
// rules and adheres to some heuristic limits of the local node (price and size).
func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error {
	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Size() > 32*1024 {
		return ErrOversizedData
	}

	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction doesn't exceed the current block limit gas.
	if pool.currentMaxGas.Cmp(tx.Gas()) < 0 {
		return ErrGasLimit
	}
	// Make sure the transaction is signed properly
	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}
	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if pool.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}
	//确保交易金额能有满足执行交易需要的规则。
	//fmt.Println("from.Address 's vlaue:%s",from.String())
	heightCount:= pool.currentState.HeightTxCount(from)
	//fmt.Println("heightCount 's value:%s",heightCount)

	blockheight := pool.currentState.TxBlockHeight(from)

	balance := pool.currentState.GetBalance(from)

	//可执行步数等于当前余额*10
	tradeNumber := new(big.Int).Div(balance,TradeTimesCount)

	currentBlockNumber := pool.chain.CurrentBlock().Number()
	if blockheight.Cmp(currentBlockNumber) == 0{
		if int64(heightCount) > tradeNumber.Int64(){
			return ErrHeightTxTooMuch
		}
		heightTxCount:=pool.currentState.HeightTxCount(from)
	//	fmt.Println("heightTxCount :%s, blockheight:%s,currentBlockNumber:%s",heightTxCount,blockheight,currentBlockNumber)
		pool.currentState.SetHeightTxCount(from,heightTxCount+1)
	}else{
		pool.currentState.SetTxBlockHeight(from,*currentBlockNumber)
		pool.currentState.SetHeightTxCount(from,1)
	}

	// Drop non-local transactions under our own minimal accepted gas price
	local = local || pool.locals.contains(from) // account may be local even if the transaction arrived from the network
	if !local && pool.gasPrice.Cmp(tx.GasPrice()) > 0 {
		return ErrUnderpriced
	}
	// Ensure the transaction adheres to nonce ordering
	if pool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}


	//modify by roger on 2017-01-12
	//intrGas := IntrinsicGas(tx.Data(), tx.To() == nil, pool.homestead)
	//if tx.Gas().Cmp(intrGas) < 0 {
	//	return ErrIntrinsicGas
	//}
	return nil
}