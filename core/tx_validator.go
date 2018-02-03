package core

import (
	"github.com/ethzero/go-ethzero/core/types"
	//"math/big"
	//"fmt"
	"fmt"
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

	//modify by roger on 2018-01-16
	// Ensure the transaction doesn't exceed the current block limit gas.

	if pool.currentMaxGas.Cmp(tx.Gas()) < 0 {
		return ErrGasLimit
	}
	// Make sure the transaction is signed properly

	//modify by roger on 2018-01-25
	//Ensure the Transaction used the normal ChainID
	if pool.chainconfig.IsEthzeroTOSBlock(pool.chain.CurrentBlock().Number()) || pool.chainconfig.IsEthzeroGenesisBlock(pool.chain.CurrentBlock().Number()) {
		pool.signer = types.NewEIP155Signer(pool.chainconfig.ChainId)
	}

	from, err := types.Sender(pool.signer, tx)
	if err != nil {
		return ErrInvalidSender
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	if pool.currentState.GetBalance(from).Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}

	// Drop non-local transactions under our own minimal accepted gas price
	//local = local || pool.locals.contains(from) // account may be local even if the transaction arrived from the network
	//if !local && pool.gasPrice.Cmp(tx.GasPrice()) > 0 {
	//	return ErrUnderpriced
	//}
	// Ensure the transaction adheres to nonce ordering
	if pool.currentState.GetNonce(from) > tx.Nonce() {
		return ErrNonceTooLow
	}

	count := pool.GetTransactionCountByFrom(from)
	balance := new(big.Int).Div(pool.currentState.GetBalance(from), big.NewInt(1e+18))
	if balance.Cmp(big.NewInt(1)) < 0 {
		balance = big.NewInt(1)
	}
	maxcount := new(big.Int).Mul(balance, big.NewInt(10))

	if maxcount.Cmp(DefaultCurrentMaxNonce) > 0 {
		maxcount = big.NewInt(500)
	}

	if big.NewInt(int64(count)).Cmp(maxcount) > 0 {
		return ErrTooTradeTimesInCurrentBlock
	}

	intrGas := IntrinsicGas(tx.Data(), tx.To() == nil, false)

	if tx.To() == nil && contractTxMaxGasSize.Cmp(intrGas) < 0 {
		fmt.Printf(" txvalidator.go intrGas %v and contractTxMaxGassize: %v", intrGas, contractTxMaxGasSize)
		return ErrContractTxIntrinsicGas
	}
	if tx.To() != nil && txMaxGasSize.Cmp(intrGas) < 0 {
		fmt.Printf(" txvalidator.go intrGas %v and contractTxMaxGassize: %v", intrGas, txMaxGasSize)
		return ErrIntrinsicGas
	}
	return nil
}
