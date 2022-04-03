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
	"enode://a81aef156c6a3b6f7df516af0a01f9e8c9da41cf07fd3d519cdf7802b9c30f8243e0a729a4b92210720e456006920db309001b0e50164a729d250cee44187096@47.242.227.99:21212",//HongKong
	"enode://c9a8506f69758800a2d87d40418e18c6d228b92be6679649dfba84ae06f5af781e6de1e8ad0584f843cfed42395ad8b5e724bb4304993bb3ce52b3ab57c99dda@47.242.226.47:21212",//HongKong
	"enode://9b1f954f77a3f846f6e79e47a445fef52d992620ff113af8a58e102e31b72b971029eef7e9f3496bc6c91fcb6e762bc7c91b2f6cde55a4d5d31274d9d0d947e4@8.210.30.31:21212",//HongKong
	"enode://3b9471c1b4d93a45a1f7aff368d027dc7eeac7c526d80848d9848773b0426f41931ecdafa6f513800f68b5425f5b1a482ccbd6eb4b0f39982c4d3ff0cefe085e@35.182.48.79:21212", // Canada
	"enode://c4fed9426ad1355c845edbc6442ef6bc3e9edb9eacdc42f613c297e8b986dc8099956a1b79c3d2ad12118d77e20c71fc22e2876773d5508cd0d7707b91767e35@52.47.202.205:21212", // Paris
	"enode://33e3f4ea45c3b20d1703be686ce6f6e1726fe6279451eda864a8e63d57324c89ebe3ba0167c457850e50a9ec153e063e4f4ff6344c821db93d06c0ab87d092f5@3.0.240.110:21212", // Singapore
	"enode://3e5655447985a71d2c46097fe790c95310a7d075f7333ed3019897cbbde057624f31285999da192aa27ac48f801d24bbb61fac5bf1bacd30439018843d676e59@13.52.109.169:21212", // California
	"enode://a81aef156c6a3b6f7df516af0a01f9e8c9da41cf07fd3d519cdf7802b9c30f8243e0a729a4b92210720e456006920db309001b0e50164a729d250cee44187096@47.242.227.99:21212",
	"enode://c9a8506f69758800a2d87d40418e18c6d228b92be6679649dfba84ae06f5af781e6de1e8ad0584f843cfed42395ad8b5e724bb4304993bb3ce52b3ab57c99dda@47.242.226.47:21212",
	"enode://9b1f954f77a3f846f6e79e47a445fef52d992620ff113af8a58e102e31b72b971029eef7e9f3496bc6c91fcb6e762bc7c91b2f6cde55a4d5d31274d9d0d947e4@8.210.30.31:21212",
	"enode://d843cf2f1b180685a363bf8e5adf7aa875894c5bec3e97bfeed0d8b391b9e72944bb275347c21908a5b70e766a081370829ea82468673a710eab2bd4507ca349@95.217.60.229:23212",
	"enode://ac0cefbfd864e16ac7ccfca042172f46a818803205798e4a51d623f29e8ed6e648d3e4b2c366b98e1d95ab362c22e06bb130652bb0a552a340b70090eab4865e@95.217.60.229:23414",
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
