// Copyright 2018 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-etherzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
	"errors"

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/contracts/masternode/contract"
	"github.com/etherzero/go-etherzero/core"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/event"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/p2p/discover"
	"crypto/ecdsa"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/eth/downloader"
)

var (
	statsReportInterval  = 10 * time.Second // Time interval to report vote pool stats
	ErrUnknownMasternode = errors.New("unknown masternode")
)

type MasternodeManager struct {
	blockchain *core.BlockChain
	// channels for fetcher, syncer, txsyncLoop
	IsMasternode uint32
	srvr         *p2p.Server
	contract     *contract.Contract

	txPool *core.TxPool
	mux    *event.TypeMux

	syncing int32

	mu          sync.RWMutex
	ID          string
	NodeAccount common.Address
	PrivateKey  *ecdsa.PrivateKey
}

func NewMasternodeManager(blockchain *core.BlockChain, contract *contract.Contract, txPool *core.TxPool) *MasternodeManager {

	// Create the masternode manager with its initial settings
	manager := &MasternodeManager{
		blockchain: blockchain,
		contract:   contract,
		txPool:     txPool,
	}
	return manager
}

func (self *MasternodeManager) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

}

func (self *MasternodeManager) Start(srvr *p2p.Server, mux *event.TypeMux) {
	self.srvr = srvr
	self.mux = mux
	log.Trace("MasternodeManqager start ")
	x8 := srvr.Self().X8()
	self.ID = fmt.Sprintf("%x", x8[:])
	self.NodeAccount = crypto.PubkeyToAddress(srvr.Config.PrivateKey.PublicKey)
	self.PrivateKey = srvr.Config.PrivateKey

	go self.masternodeLoop()
	go self.checkSyncing()
}

func (self *MasternodeManager) Stop() {

}

func (mm *MasternodeManager) masternodeLoop() {
	xy := mm.srvr.Self().XY()
	has, err := mm.contract.Has(nil, mm.srvr.Self().X8())
	if err != nil {
		log.Error("contract.Has", "error", err)
	}
	if has {
		fmt.Println("### It's already been a masternode! ")
		atomic.StoreUint32(&mm.IsMasternode, 1)
	} else {
		atomic.StoreUint32(&mm.IsMasternode, 0)
		if mm.srvr.IsMasternode {
			data := "0x2f926732" + common.Bytes2Hex(xy[:])
			fmt.Printf("### Masternode Transaction Data: %s\n", data)
		}
	}

	joinCh := make(chan *contract.ContractJoin, 32)
	quitCh := make(chan *contract.ContractQuit, 32)
	joinSub, err1 := mm.contract.WatchJoin(nil, joinCh)
	if err1 != nil {
		// TODO: exit
		return
	}
	quitSub, err2 := mm.contract.WatchQuit(nil, quitCh)
	if err2 != nil {
		// TODO: exit
		return
	}

	ping := time.NewTimer(masternode.MASTERNODE_PING_INTERVAL)
	defer ping.Stop()
	ntp := time.NewTimer(time.Second)
	defer ntp.Stop()
	minPower := big.NewInt(20e+14)

	report := time.NewTicker(statsReportInterval)
	defer report.Stop()

	for {
		select {
		case join := <-joinCh:
			if bytes.Equal(join.Id[:], xy[:]) {
				atomic.StoreUint32(&mm.IsMasternode, 1)
				fmt.Println("### Become a masternode! ")
			}
		case quit := <-quitCh:
			if bytes.Equal(quit.Id[:], xy[0:8]) {
				atomic.StoreUint32(&mm.IsMasternode, 0)
				fmt.Println("### Remove a masternode! ")
			}
		case err := <-joinSub.Err():
			joinSub.Unsubscribe()
			fmt.Println("eventJoin err", err.Error())
		case err := <-quitSub.Err():
			quitSub.Unsubscribe()
			fmt.Println("eventQuit err", err.Error())

		case <-ntp.C:
			ntp.Reset(10 * time.Minute)
			go discover.CheckClockDrift()
		case <-ping.C:
			ping.Reset(masternode.MASTERNODE_PING_INTERVAL)
			if atomic.LoadUint32(&mm.IsMasternode) == 0 {
				break
			}
			logTime := time.Now().Format("2006-01-02 15:04:05")
			if atomic.LoadInt32(&mm.syncing) == 1 {
				fmt.Println(logTime, " syncing...")
				break
			}
			address := mm.NodeAccount
			stateDB, _ := mm.blockchain.State()
			if stateDB.GetBalance(address).Cmp(big.NewInt(1e+16)) < 0 {
				fmt.Println(logTime, "Failed to deposit 0.01 etz to ", address.String())
				break
			}
			if stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).Cmp(minPower) < 0 {
				fmt.Println(logTime, "Insufficient power for ping transaction.", address.Hex(), mm.blockchain.CurrentBlock().Number().String(), stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).String())
				break
			}
			tx := types.NewTransaction(
				mm.txPool.State().GetNonce(address),
				params.MasterndeContractAddress,
				big.NewInt(0),
				90000,
				big.NewInt(20e+9),
				nil,
			)
			signed, err := types.SignTx(tx, types.NewEIP155Signer(mm.blockchain.Config().ChainID), mm.PrivateKey)
			if err != nil {
				fmt.Println(logTime, "SignTx error:", err)
				break
			}

			if err := mm.txPool.AddLocal(signed); err != nil {
				fmt.Println(logTime, "send ping to txpool error:", err)
				break
			}
			fmt.Println(logTime, "Send ping message!")
		}
	}
}

// SignHash calculates a ECDSA signature for the given hash. The produced
// signature is in the [R || S || V] format where V is 0 or 1.
func (self *MasternodeManager) SignHash(id string, hash []byte) ([]byte, error) {
	// Look up the key to sign with and abort if it cannot be found
	self.mu.RLock()
	defer self.mu.RUnlock()

	if id != self.ID {
		return nil, ErrUnknownMasternode
	}
	// Sign the hash using plain ECDSA operations
	return crypto.Sign(hash, self.PrivateKey)
}

func (self *MasternodeManager) checkSyncing() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			atomic.StoreInt32(&self.syncing, 1)
		case downloader.DoneEvent, downloader.FailedEvent:
			atomic.StoreInt32(&self.syncing, 0)
		}
	}
}


func (self *MasternodeManager) MasternodeList(number *big.Int) ([]string, error) {
	return masternode.GetIdsByBlockNumber(self.contract, number)
}


func (self *MasternodeManager) GetGovernanceContractAddress(number *big.Int) (common.Address, error) {
	return masternode.GetGovernanceAddress(self.contract, number)
}
