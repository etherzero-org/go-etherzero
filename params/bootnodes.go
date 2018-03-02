// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Ethereum network.
var MainnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	"enode://99646ce50bc153c4347035cacceffe3d09e53b366cd450bbfa3fa8eb01973354cda4904837f9e8b8b5b73e4020e9be6a4d15bc76955775b4e56b97d80fffb213@54.219.114.31:21212",  // US
	"enode://608ffb991e8c56df678a36a796f7e05159aca38ddd13a5b5f048367de5d95f78453eaace240ef8ef31dd743d786eebc8e6dca91621fc32872d7f02d54a80a68a@13.57.140.26:21212",   // US
	"enode://8c8a2a1f7ca9a378f5e64d8e8f3fd569aa13d462caa53f97cd660445968c0e78ec75e877d8b06d5c6da10b2d8401ae13e1c57adf8756a7cd6529b4cd7723869b@52.221.166.174:21212", // SG
	"enode://aab3cfe75ab355853f0cc3c57525e86dce34da581d92c0063c6e1262ed6ac459e535ea7b3a17dd4440d31418363c1ff0a0dcd81fd798c82653239b14628343a5@52.74.3.64:21212",     // SG
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
}

// RinkebyV5Bootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network for the experimental RLPx v5 topic-discovery network.
var RinkebyV5Bootnodes = []string{
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
}