// Copyright 2018 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-eth library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"math/rand"

	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/consensus/ethash"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/node"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto"
)

const (
	testInstance = "console-tester"
	testAddress1 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

func newEtherrum() *Ethereum {
	// Create a temporary storage for the node keys and initialize it
	workspace, err := ioutil.TempDir("", "console-tester-")
	if err != nil {
		fmt.Printf("failed to create temporary keystore: %v", err)
	}

	// Create a networkless protocol stack and start an Ethereum service within
	stack, err := node.New(&node.Config{DataDir: workspace, UseLightweightKDF: true, Name: testInstance})
	if err != nil {
		fmt.Printf("failed to create node: %v", err)
	}
	ethConf := &Config{
		Genesis:   core.DeveloperGenesisBlock(15, common.Address{}),
		Etherbase: common.HexToAddress(testAddress1),
		Ethash: ethash.Config{
			PowMode: ethash.ModeTest,
		},
	}

	if err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) { return New(ctx, ethConf) }); err != nil {
		fmt.Printf("failed to register Ethereum protocol: %v", err)
	}
	// Start the node and assemble the JavaScript console around it
	if err = stack.Start(); err != nil {
		fmt.Printf("failed to start test stack: %v", err)
	}
	_, err = stack.Attach()
	if err != nil {
		fmt.Printf("failed to attach to node: %v", err)
	}

	// Create the final tester and return
	var ethereum *Ethereum
	err = stack.Service(&ethereum)
	if err != nil {
		fmt.Printf("failed to as a service: %v", err)
	}

	ethereum.blockchain = newBlockChain()
	return ethereum
}
func returnMasternodeManager() *MasternodeManager {
	//// initial the parameter may needed during this test function
	eth := newEtherrum()
	return &MasternodeManager{
		networkId:   uint64(88),
		eventMux:    nil,
		blockchain:  nil,
		chainconfig: nil,
		newPeerCh:   make(chan *peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		masternodes: &masternode.MasternodeSet{},
		is:          NewInstantx(newChainConfig(), eth),
	}
}

// TestMasternodeManager_BestMasternode
// Test function for choose BestMasternode
func TestMasternodeManager_BestMasternode(t *testing.T) {
	//// initial the parameter may needed during this test function
	manager := returnMasternodeManager()
	ranksFn := func(height *big.Int) map[int64]*masternode.Masternode {
		return manager.GetMasternodeRanks(height)
	}
	manager.winner = NewMasternodePayments(big.NewInt(10), ranksFn)

	// init new hash
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(rand.Intn(0x1000))
	}
	// init new account
	var accounts []common.Address
	seed := make([]byte, common.AddressLength)
	for i := 0; i < 10; i++ {
		rand.Read(seed)
		accounts = append(accounts, crypto.CreateAddress(common.BytesToAddress(seed), uint64(rand.Int63())))
		fmt.Printf("account value:%s", accounts[i].Hex())
	}

	// generate a new block data
	tx1 := types.NewTransaction(1, accounts[1], big.NewInt(111), 1111, big.NewInt(111111), []byte{0x11, 0x11, 0x11})
	tx2 := types.NewTransaction(2, accounts[2], big.NewInt(222), 2222, big.NewInt(222222), []byte{0x22, 0x22, 0x22})
	tx3 := types.NewTransaction(3, accounts[3], big.NewInt(333), 333, big.NewInt(33333), []byte{0x33, 0x33, 0x33})
	txs := []*types.Transaction{tx1, tx2, tx3}
	//block := types.NewBlock(&types.Header{Number: big.NewInt(31415926)}, txs, nil, nil)

	// generate a new masternode set and deploy the new nodeset
	ms := newMasternodeSet(true)

	//// begin to test
	testsbody := []struct {
		ms                  *masternode.MasternodeSet // input ms, MasternodeSet
		voteNum             *big.Int                  // 投票的区块
		err                 error                     // return type, error or nil
		numberofmasternodes uint32                    //  is numberofmasternodes==0 ,return the error
	}{
		//{
		//	nil,
		//	big.NewInt(31415926),
		//	nil,
		//	1,
		//},
		//
		//{
		//	ms,
		//	nil,
		//	errors.New("no masternode detected"),
		//	20,
		//},
		//{
		//	ms,
		//	nil,
		//	errors.New("The number of local masternodes is too less to obtain the best Masternode"),
		//	0,
		//},
		{
			ms,
			nil,
			nil,
			20,
		},
	}
	number := big.NewInt(31415926)
	// show the test process
	for _, v := range testsbody {
		// first two line test
		if v.voteNum != nil {
			manager.winner.blocks[31415926] = NewMasternodeBlockPayees(number)
			masternodePayee := NewMasternodePayee(*(new(common.Address)), nil)
			manager.winner.blocks[31415926].payees = append(manager.winner.blocks[31415926].payees, masternodePayee)
		}
		//if v.numberofmasternodes == 0 {
		//	fmt.Printf("numberofmasternodes %s",v.numberofmasternodes)
		//	v.ms = newMasternodeSet(false)
		//}

		manager.masternodes = v.ms
		i:=0
		for key, node := range manager.masternodes.AllNodes() {
			//node.Height=big.NewInt(int64(3141591+rand.Intn(10)))
			node.Height=big.NewInt(int64(3141591+i))
			fmt.Printf("AllNodes ,key:%s,node.accounts:%s\n", key, node.Account.Hex())
			i++
		}

		for i := 0; i < 10; i++ {
			height := int64(31415921)
			block := types.NewBlock(&types.Header{Number: big.NewInt(height)}, txs, nil, nil)
			addr, err := manager.BestMasternode(block)
			if err != nil{
				fmt.Printf("\n Masternode_Manager_test err %s\n", addr.String(), err.Error())
			}else {
				fmt.Printf("\n Masternode_Manager_test addr.string()%s\n", addr.String())
			}
		}

		//if err != nil {
		//	fmt.Println("Masternode_Manager_test addr.string()",addr.String())
		//	if !strings.EqualFold(err.Error(), v.err.Error()) {
		//		t.Errorf("test failed %v", err)
		//	}
		//}

		//if addr != *(new(common.Address)) {
		//	t.Logf("winnerid is %v", addr.String())
		//}
	}
}

func TestMasternodeManager_GetMasternodeScores(t *testing.T) {

}

func TestMasternodeManager_GetMasternodeRank(t *testing.T) {

}
