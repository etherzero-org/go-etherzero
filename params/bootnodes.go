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

var MainnetBootnodes = []string{
	"enode://96ff5320efa01fd6ca18c00df808664353ee536cadc79ff642bd06ca133ad3bb884b632dc8793ac622f207b928f063ab9fc526714d2bea08fb9621dd646ddfab@150.109.6.165:21212",
}

var MainnetMasternodes = []string{
	"enode://81e7f69a7990b9f2bc26abfba2b052e6fba389961ada8f60687acc1ac221997abc197bc9e56c0c7325b18438344234704e0429fabd75128d94b16d48586b18ee", // [1]
	"enode://190adef951323157d8e1024fc685429599c03e05f1cd62ccd74bd82afaca78a94e80bc92e17227460a9e2126f45097204fb32d67373737017f0dfd348d230abc", // [2]
}

var MainnetInitIds = []string{
	"8f5292507f858ef4",
	"041515e0266f1265",
	"f5ad3fe12339603f",
	"550c6eee655f1893",
	"6f77099f79db13ee",
	"f5e2ea87dd8d2091",
	"a5f2ab321dae4afa",
	"254d4493a2241e4b",
	"45bccc38587203ce",
	"bc51a1c43bbec4ed",
	"c036811c5c2b7f10",
	"f3b52a628bc27590",
	"da4ed95859514002",
	"4a8ffd0c571c4c7c",
	"f5d8ce70915284f9",
	"accfe0634ae0a8f5",
	"ea2f1d02f638806f",
	"48ba739314695e32",
	"5fac251f8d9e66da",
	"018b2a04e21a8f7e",
	"d083475716830cfb",
	"bb253a75325ca93d",
	"cd8a6c20e9f456bf",
	"140cb7a5716e9c44",
	"99f1f663f88b2733",
	"9427022379910691",
}

var TestnetBootnodes = []string{
	"enode://59ca967b2c9c1442e81026f5ffc2b24f4b3787512194a41e4ab14dfac97e75b700988cac80f973641d40cd65f775f41955b93d2e843ebb03555b16dd9bf983d4@127.0.0.1:9646",
}

var TestnetMasternodes = []string{
	"enode://59ca967b2c9c1442e81026f5ffc2b24f4b3787512194a41e4ab14dfac97e75b700988cac80f973641d40cd65f775f41955b93d2e843ebb03555b16dd9bf983d4", // nodekey: a9b50794ab7a9987aa416c455c13aa6cc8c0448c501a3ce8e4840efe47cb5c29
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{}
