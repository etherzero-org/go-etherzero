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
	"enode://255532ad39595d512c7b7028815afefd00a4bc31941c56e683199a5874307f606619e269cc770b5f85afdb7723c304a4491a3499eb1689434f5905afab78958e@124.156.218.158:21212",
}

var MainnetMasternodes = []string{
	"enode://81e7f69a7990b9f2bc26abfba2b052e6fba389961ada8f60687acc1ac221997abc197bc9e56c0c7325b18438344234704e0429fabd75128d94b16d48586b18ee", // [1]
	"enode://190adef951323157d8e1024fc685429599c03e05f1cd62ccd74bd82afaca78a94e80bc92e17227460a9e2126f45097204fb32d67373737017f0dfd348d230abc", // [2]
}

var MainnetInitIds = []string{
	"cde8ff27b83eb1fd",
	"1a4aa55295ac86f5",
	"0096774dab8fa7ad",
	"baa490b7d73c43d5",
	"f422ea52a3f8a605",
	"a3f309069a1abd2c",
	"55ec79cddb871ec0",
	"e9bd0bf6f5777c17",
	"352d56d8ef23edfa",
	"0948bfddf3656a8a",
	"6079d71555734172",
	"55bb6a431d35afc6",
	"782dfb20fb36401f",
	"8cdfff66a74c0fb1",
	"d77597358e056f77",
	"a4889af17694e0e1",
	"8fdb077e972e6a90",
	"19c4e94c735db0ab",
	"9695e1a870a4ad13",
	"729af40261ba80f4",
	"87df82258895aae7",
	"6f3cf020b06c46da",
	"5428ad714fde5ced",
	"ca703be1b161ed9e",
	"a44c83732a87bf7f",
	"28d324899f27e768",
	"9cfd67a31406d80c",
	"495a54b73469ae0c",
	"0d8bb911a668adbd",
	"a6c8b8a25e615d67",
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
