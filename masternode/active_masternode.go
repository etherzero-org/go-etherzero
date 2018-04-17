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

//The Masternode Class. For managing the InstantTX process. It contains the input of the 20000ETZ, signature to prove
// it's the one who own that ip address and code for calculating the payment election.

package masternode

import (
	"github.com/ethzero/go-ethzero/common"
	"net"
)

const(
	ACTIVE_MASTERNODE_INITIAL          = 0; // initial state
	ACTIVE_MASTERNODE_SYNC_IN_PROCESS  = 1;
	ACTIVE_MASTERNODE_INPUT_TOO_NEW    = 2;
	ACTIVE_MASTERNODE_NOT_CAPABLE      = 3;
	ACTIVE_MASTERNODE_STARTED          = 4;

)

const(
	masternodetype_unknow =0
	masternodetype_remote =1
)
//Responsible for activating the Masternode and pinging the network
type ActiveMasternode struct{

	masternodeType int

	activeState	int
	// Keys for the active Masternode
	MasternodeKey string
	//MasternodeID       NodeID // the Masternode's public key
	// Initialized while registering Masternode
	IP       net.IP // len 4 for IPv4 or 16 for IPv6
	UDP, TCP uint16 // port numbers
	txid common.Hash
	// This is a cached copy of sha3(ID) which is used for Masternode
	// distance calculations. This is part of Node in order to make it
	// possible to write tests that need a node at a certain distance.
	// In those tests, the content of sha will not actually correspond
	// with ID.
	sha common.Hash

}

func (am *ActiveMasternode) Type() (int){
	return am.masternodeType
}

func (am *ActiveMasternode)State()(int){
	return am.activeState
}

func (am *ActiveMasternode) ManageStateInitial(){

}

func (am *ActiveMasternode) manageStateRemote(){}


func (am *ActiveMasternode) UpdateSentinelPing()(bool,error){

	return true,nil
}

func (am *ActiveMasternode)SendMasternodePing()(bool,error){

	return true,nil
}


