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
	"enode://3b9471c1b4d93a45a1f7aff368d027dc7eeac7c526d80848d9848773b0426f41931ecdafa6f513800f68b5425f5b1a482ccbd6eb4b0f39982c4d3ff0cefe085e@35.182.48.79:21212", // Canada
	"enode://58b0c020981e62942f1a724aa060cabd504098e49f3e4ab028da7aa0f7e12ca581224692052fbd0448bd83b4f1a45be83083f4929855cef08db37e2625adf664@47.90.101.201:21212", // Hongkong
	"enode://c4fed9426ad1355c845edbc6442ef6bc3e9edb9eacdc42f613c297e8b986dc8099956a1b79c3d2ad12118d77e20c71fc22e2876773d5508cd0d7707b91767e35@52.47.202.205:21212", // Paris
	"enode://33e3f4ea45c3b20d1703be686ce6f6e1726fe6279451eda864a8e63d57324c89ebe3ba0167c457850e50a9ec153e063e4f4ff6344c821db93d06c0ab87d092f5@3.0.240.110:21212", // Singapore
	"enode://3e5655447985a71d2c46097fe790c95310a7d075f7333ed3019897cbbde057624f31285999da192aa27ac48f801d24bbb61fac5bf1bacd30439018843d676e59@13.52.109.169:21212", // California
}

var MainnetMasternodes = []string{
	"enode://81e4a6821a4d7117b4740ffd50826b823acef0e5161df7cfca3bde8764e294cbe70eabbc99e85cd01e8456fd2c66b5ec900fb658910360ebe0648e3e47c5dc26", // [1]
	"enode://754ae5a877ff913ec0f735c462fa4edd81dc6be347d9eea6beeb04cbdf6ff50ae7e276adf89e44f11389c6b3224f7ce991347e564d5f3c004b4d85513040088b", // [2]
	"enode://9e2ebf310caf945d5e3abfa9968413755a6f3cb9fe044ac154d306726c9e54eae5c6266e42f5d5c6c81ec43afe48f8b7ad9c153e95be2131922edb0724b66825", // [3]
	"enode://5ca084808222e3c8fc213a3d7c5ab4a15a54726f6801489311b9de355267086478c67ce82b4eba47d691a440ff7dcb91ba3e6392b9de15ff283bb30ddfea73fb", // [4]
	"enode://b553bcf51ae5f42bf93c8ace9970f8fc56846c1ba3cba425bb13bdfefec4d1a2c37691f900455f0c7020890d163edcd85efea95977d6e76d79083ffbebe9b4e1", // [5]
	"enode://bbdf34418a2630cf456eec69565e53d7462869a5c672312fc64126acda848437a05269fee2fe313a16f77bb0e871f7a4caf3abd7695d6ac3f19324a62bbe9979", // [6]
	"enode://7f0dcd95f323b9f28d9ba631ffa3738318e4234190483c7e29a9cb61795ccc1f8eff96619e8e94a4547af4837b99b2fbfea49a713d323801b5369e35d4121c37", // [7]
	"enode://35b680724f0cdfed0e1e8687d3627fa7fe7e6e0d46028938449952e460abd02beff51d135410b2d154c22a0211ce9a2067cd6f891ef0af90af999a42accfd0fe", // [8]
	"enode://65fed3a560cd231a3a56490776a0af02bd1de92821033dd17cc7ad98a63a59aabfeeb3022e2e6b0ca19f62dde1c7524928ddc0d421d1b8ac6e3e472945d85078", // [9]
	"enode://2cbd44f1b7b4d8bec0460c8d2770b8bbd514cc9b8a023a615fd16d32fb7b8aae5f07d13655b4895f8a05b59ed9daaae178aac68a62e8ca96132ec283df368013", // [10]
	"enode://3b9639dadd18a258f9615166861693897049bf6dd1ce2da4f3b78394b6c7f44a32163f4e39320455dddebea62f0f8690fb57205198c761ed7cf3d8e13ef72f99", // [11]
	"enode://53da176be1538aed11ca42608c8576c3591e2894c5aebb0c35bb49ff66710f63ca87668d0310f29a3ad23da2db87bbc540071b485e5b8cffddeb07b831c2bdd6", // [12]
	"enode://017c131b2ae66403717d218906840b0c374c0c75337e5573e0abe3eb63c5dad6e1decd3bfb31941f6cc6dbbcdc820cc103a0dabf7cc65763a508c295c1cf30af", // [13]
	"enode://4fa67b765794778372a86e7f6678d2044c9933b14c645e321718b280b74e0452dcd9d92199e90f38fca83023745ff91c8cda65f09d44bf1bc45797da97136d88", // [14]
	"enode://4717417a9605535deb1e2f5241e2682b8165b06807d4a3bf0ab9ec4e4b2b49a039968cedeba98b1a8e43782e9c7b740d810de8a046251353f267401e0ba902e3", // [15]
	"enode://b371992843eaf7cdedeccd6bd422e3229a86b639ffddfcaf45fc0b57cefdb6baf1a2b461c2e062f718130f3366d2b6df68dfdbdec21eeb6ec525111fc8570e31", // [16]
	"enode://98da296630899f29522c90824cbba193e6d799eca04f20160031ef1f3f9fd711c9e8d84f91b598d49a619809979e8ed00b9a8cb3a8dfe41d998fb6d2b61c802f", // [17]
	"enode://3beabd9d5fceccf6cb2c9d7f408f65df16065967c45d3d97882f9f2394430512b4a19036553d84950c8a5a6efe20bfba38e6921e94f2954f253c215a8c58caf8", // [18]
	"enode://6f5381470fb24553e54e75a28216e6a630f93b81f8dcb0c8ab43796100d7e8f97b869b119ced8a3f22e5a0475d173528fea1e2ac61ab31f0d82db9ee788c1e85", // [19]
	"enode://3b9471c1b4d93a45a1f7aff368d027dc7eeac7c526d80848d9848773b0426f41931ecdafa6f513800f68b5425f5b1a482ccbd6eb4b0f39982c4d3ff0cefe085e", // [20]
	"enode://8375c6b34607d06b5d5b4df1a375cecc1df1237f420cb201c37900f856260e7b90d6fe8f64a30a01a4216c9c9627e22baa0089dee385f27aa0398f6fd2f085e4", // [21]
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
