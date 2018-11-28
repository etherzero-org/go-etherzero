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

	"github.com/etherzero/go-etherzero/aux"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/contracts/masternode/contract"
	contract2 "github.com/etherzero/go-etherzero/contracts/enodeinfo/contract"
	"github.com/etherzero/go-etherzero/core"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/event"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/eth/downloader"
	"github.com/etherzero/go-etherzero/p2p/discover"
	"net"
)

var (
	statsReportInterval = 10 * time.Second // Time interval to report vote pool stats
)

type MasternodeManager struct {
	beats map[common.Hash]time.Time // Last heartbeat from each known vote

	active *masternode.ActiveMasternode
	mu     sync.Mutex
	// channels for fetcher, syncer, txsyncLoop
	newPeerCh         chan *peer
	IsMasternode      uint32
	srvr              *p2p.Server
	contract          *contract.Contract
	enodeinfoContract *contract2.Contract
	blockchain        *core.BlockChain
	scope             event.SubscriptionScope

	currentCycle uint64        // Current vote of the block chain
	Lifetime     time.Duration // Maximum amount of time vote are queued

	txPool *core.TxPool

	downloader *downloader.Downloader
}

func NewMasternodeManager(blockchain *core.BlockChain, contract *contract.Contract, enodeinfoContract *contract2.Contract, txPool *core.TxPool) *MasternodeManager {

	// Create the masternode manager with its initial settings
	manager := &MasternodeManager{
		blockchain:        blockchain,
		beats:             make(map[common.Hash]time.Time),
		Lifetime:          30 * time.Second,
		contract:          contract,
		enodeinfoContract: enodeinfoContract,
		txPool:            txPool,
	}
	return manager
}

func (self *MasternodeManager) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

}

func (self *MasternodeManager) Start(srvr *p2p.Server, peers *peerSet, downloader *downloader.Downloader) {
	self.srvr = srvr
	self.downloader = downloader
	log.Trace("MasternodeManqager start ")
	self.active = masternode.NewActiveMasternode(srvr)
	go self.masternodeLoop()
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
		mm.updateActiveMasternode(true)
	} else if mm.srvr.IsMasternode {
		mm.updateActiveMasternode(false)
		data := "0x2f926732" + common.Bytes2Hex(xy[:])
		fmt.Printf("### Masternode Transaction Data: %s\n", data)
	}

	//time.AfterFunc(masternode.MASTERNODE_PING_INTERVAL, func() {
	//	mm.SaveNodeIpToContract()
	//})
	mm.SaveNodeIpToContract()
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
				mm.updateActiveMasternode(true)
				fmt.Println("### It's already been a masternode! ")
			}
		case quit := <-quitCh:
			if bytes.Equal(quit.Id[:], xy[0:8]) {
				atomic.StoreUint32(&mm.IsMasternode, 0)
				mm.updateActiveMasternode(false)
				fmt.Println("### Remove masternode! ")
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
			if mm.active.State() != masternode.ACTIVE_MASTERNODE_STARTED {
				break
			}
			if mm.downloader.Synchronising() {
				break
			}

			address := mm.active.NodeAccount
			stateDB, _ := mm.blockchain.State()
			if stateDB.GetBalance(address).Cmp(big.NewInt(1e+16)) < 0 {
				fmt.Println("Failed to deposit 0.01 etz to ", address.String())
				break
			}
			if stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).Cmp(minPower) < 0 {
				fmt.Println("Insufficient power for ping transaction.", address.Hex(), mm.blockchain.CurrentBlock().Number().String(), stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).String())
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
			signed, err := types.SignTx(tx, types.NewEIP155Signer(mm.blockchain.Config().ChainID), mm.active.PrivateKey)
			if err != nil {
				fmt.Println("SignTx error:", err)
				break
			}

			if err := mm.txPool.AddLocal(signed); err != nil {
				fmt.Println("send ping to txpool error:", err)
				break
			}
			fmt.Println("Send ping message ...")
		}
	}
}

func (mm *MasternodeManager) updateActiveMasternode(isMasternode bool) {
	var state int
	if isMasternode {
		state = masternode.ACTIVE_MASTERNODE_STARTED
	} else {
		state = masternode.ACTIVE_MASTERNODE_NOT_CAPABLE

	}
	mm.active.SetState(state)
}

func (self *MasternodeManager) MasternodeList(number *big.Int) ([]string, error) {
	return masternode.GetIdsByBlockNumber(self.contract, number)
}

func (self *MasternodeManager) GetGovernanceContractAddress(number *big.Int) (common.Address, error) {
	return masternode.GetGovernanceAddress(self.contract, number)
}

// SaveNodeIpToContract
// only masyernode need to save ip to contract
func (mm *MasternodeManager) SaveNodeIpToContract() (err error) {
	fmt.Printf("time.now is %v\n", time.Now())

	// not initialize
	if mm.srvr.Self() == nil {
		return
	}
	fmt.Printf("mm.srvr.IsMasternode  %v,mm.active.State()  %v", mm.srvr.IsMasternode, mm.active.State())
	if !mm.srvr.IsMasternode || mm.active.State() != masternode.ACTIVE_MASTERNODE_STARTED {
		return
	}
	minPower := big.NewInt(20e+14)
	// // send myself node info
	address := mm.active.NodeAccount
	fmt.Println("NodeAccountNodeAccount", address.String())
	stateDB, _ := mm.blockchain.State()
	if stateDB.GetBalance(address).Cmp(big.NewInt(1e+16)) < 0 {
		err = errors.New(fmt.Sprintf("Failed to deposit 0.01 etz to %v ", address.String()))
		return
	}

	if stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).Cmp(minPower) < 0 {
		err = errors.New(fmt.Sprintf("Insufficient power for send masternode transaction %v  %v",
			address.String(), stateDB.GetPower(address, mm.blockchain.CurrentBlock().Number()).String()))
		return
	}

	var dataRaw string
	dataRaw, err = mm.genData()
	if err != nil {
		fmt.Printf("gen node info err :%v\n", err)
		return
	}

	data := common.Hex2Bytes(dataRaw)
	fmt.Printf("dataRaw is %v,data is %v\n", dataRaw, data)
	tx := types.NewTransaction(
		mm.txPool.State().GetNonce(address),
		params.EnodeinfoAddress, //
		big.NewInt(0),
		270000,
		big.NewInt(20e+9),
		data,
	)

	signed, err := types.SignTx(tx, types.NewEIP155Signer(mm.blockchain.Config().ChainID), mm.active.PrivateKey)
	if err != nil {
		fmt.Println("SignTx error:", err)
		return
	}

	err = mm.txPool.AddLocal(signed)
	if err != nil {
		fmt.Println("send  ip to txpool error:", err)
		return
	}
	// add to send ping message
	fmt.Println("Send ip message ...")
	return
}

func (mm *MasternodeManager) genData() (data string, err error) {

	// Just for safe check
	if mm.srvr.Self() == nil || mm.srvr.MasternodeIP == "" {
		err = errors.New("Nil node info")
		return
	}

	selfnode := mm.srvr.Self()
	// return node info
	xy := selfnode.XY()

	var (
		ip         uint32
		ip_port    uint64
		funcSha3   = "c0e64821" // web3.sha3("register(bytes32,bytes32,bytes32)") in enodeinfo.sol
		bytes64len = uint32(64)
	)
	nodeid := common.Bytes2Hex(xy[:])
	fmt.Printf("nodeidnodeidnodeid is %v\n", nodeid)
	fmt.Println("mm.srvr.MasternodeIP", mm.srvr.MasternodeIP)
	ip = aux.Netiptoipnr(net.ParseIP(mm.srvr.MasternodeIP))

	// encode to string
	ip_port = aux.EncodeIpPort(ip, uint32(selfnode.TCP()))
	ip_portStr := fmt.Sprintf("%x", ip_port)
	fmt.Println("ip_portStr", ip_portStr)
	prevZero := aux.PrefixZeroString(bytes64len - uint32(len(ip_portStr)))
	encodeIp_port := fmt.Sprintf("%v%s", prevZero, ip_portStr)
	fmt.Printf("ip_port %v , nodeid %v ,encodeIp_port %s\n", ip_port, nodeid, encodeIp_port)
	data = fmt.Sprintf("%v%v%v", funcSha3, nodeid, encodeIp_port)
	fmt.Println("data string is ", data)
	return
}
