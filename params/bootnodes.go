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
	"enode://81e7f69a7990b9f2bc26abfba2b052e6fba389961ada8f60687acc1ac221997abc197bc9e56c0c7325b18438344234704e0429fabd75128d94b16d48586b18ee@127.0.0.1:30303", // Canada
	"enode://190adef951323157d8e1024fc685429599c03e05f1cd62ccd74bd82afaca78a94e80bc92e17227460a9e2126f45097204fb32d67373737017f0dfd348d230abc@127.0.0.1:20202", // Hongkong
}

var MainnetMasternodes = []string{
	"enode://81e7f69a7990b9f2bc26abfba2b052e6fba389961ada8f60687acc1ac221997abc197bc9e56c0c7325b18438344234704e0429fabd75128d94b16d48586b18ee", // [1]
	"enode://190adef951323157d8e1024fc685429599c03e05f1cd62ccd74bd82afaca78a94e80bc92e17227460a9e2126f45097204fb32d67373737017f0dfd348d230abc", // [2]
}

var MainnetInitIds = []string{
	"20fbe3f4885b2795",
	"e90b7b9c59200e86",
	"0af003cc6ad940ee",
	"da9bd1623e0eb15a",
	"a388a552859326da",
	"12a37c42bb749b85",
	"8349a1d45351f0b8",
	"8abbabf432ee79b4",
	"b3f2c4dc7bf22558",
	"b230915e87951825",
	"3b9b1020f5dbf112",
	"f130378ec8d78ff2",
	"87de9abd41e28914",
	"8940f5cd32fe86b5",
	"783eb3ba5dd71e6f",
	"232447ce7a26363b",
	"ed886091309e177b",
	"24355fa7bd81312c",
	"b77dcea1eaf7d3bd",
	"ea27dfb907d9383f",
	"cab342132aae345b",
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
