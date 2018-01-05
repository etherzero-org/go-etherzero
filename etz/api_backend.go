// Copyright 2015 The go-ethzero Authors
// This file is part of the go-ethzero library.
//
// The go-ethzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethzero library. If not, see <http://www.gnu.org/licenses/>.

package etz

import (
	"context"
	"math/big"

	"github.com/ethzero/go-ethzero/accounts"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/common/math"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/bloombits"
	"github.com/ethzero/go-ethzero/core/state"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/vm"
	"github.com/ethzero/go-ethzero/etz/downloader"
	"github.com/ethzero/go-ethzero/etz/gasprice"
	"github.com/ethzero/go-ethzero/etzdb"
	"github.com/ethzero/go-ethzero/event"
	"github.com/ethzero/go-ethzero/params"
	"github.com/ethzero/go-ethzero/rpc"
)

// EtzApiBackend implements etzapi.Backend for full nodes
type EtzApiBackend struct {
	etz *Ethzero
	gpo *gasprice.Oracle
}

func (b *EtzApiBackend) ChainConfig() *params.ChainConfig {
	return b.etz.chainConfig
}

func (b *EtzApiBackend) CurrentBlock() *types.Block {
	return b.etz.blockchain.CurrentBlock()
}

func (b *EtzApiBackend) SetHead(number uint64) {
	b.etz.protocolManager.downloader.Cancel()
	b.etz.blockchain.SetHead(number)
}

func (b *EtzApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.etz.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.etz.blockchain.CurrentBlock().Header(), nil
	}
	return b.etz.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *EtzApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.etz.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.etz.blockchain.CurrentBlock(), nil
	}
	return b.etz.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *EtzApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.etz.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.etz.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *EtzApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.etz.blockchain.GetBlockByHash(blockHash), nil
}

func (b *EtzApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.etz.chainDb, blockHash, core.GetBlockNumber(b.etz.chainDb, blockHash)), nil
}

func (b *EtzApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.etz.blockchain.GetTdByHash(blockHash)
}

func (b *EtzApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.etz.BlockChain(), nil)
	return vm.NewEVM(context, state, b.etz.chainConfig, vmCfg), vmError, nil
}

func (b *EtzApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.etz.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *EtzApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.etz.BlockChain().SubscribeChainEvent(ch)
}

func (b *EtzApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.etz.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *EtzApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.etz.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *EtzApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.etz.BlockChain().SubscribeLogsEvent(ch)
}

func (b *EtzApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.etz.txPool.AddLocal(signedTx)
}

func (b *EtzApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.etz.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *EtzApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.etz.txPool.Get(hash)
}

func (b *EtzApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.etz.txPool.State().GetNonce(addr), nil
}

func (b *EtzApiBackend) Stats() (pending int, queued int) {
	return b.etz.txPool.Stats()
}

func (b *EtzApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.etz.TxPool().Content()
}

func (b *EtzApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.etz.TxPool().SubscribeTxPreEvent(ch)
}

func (b *EtzApiBackend) Downloader() *downloader.Downloader {
	return b.etz.Downloader()
}

func (b *EtzApiBackend) ProtocolVersion() int {
	return b.etz.EthVersion()
}

func (b *EtzApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *EtzApiBackend) ChainDb() etzdb.Database {
	return b.etz.ChainDb()
}

func (b *EtzApiBackend) EventMux() *event.TypeMux {
	return b.etz.EventMux()
}

func (b *EtzApiBackend) AccountManager() *accounts.Manager {
	return b.etz.AccountManager()
}

func (b *EtzApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.etz.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *EtzApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.etz.bloomRequests)
	}
}
