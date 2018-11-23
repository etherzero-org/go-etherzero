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
	"enode://7428587a4839162af1c2bc93629bcf45162abae8772ad93027dfdd94fddc7c653e085b5f58fd0def2533ca61a3e380f4f5e9e891d26eea8cc11df7dcf4b77189@172.18.188.159:21213", //
	"enode://c84c9860a017cadda359c2b63c29555811d02bf5839938107878ccce856447f67cc72adbad6837c18f823f4ee0a29d48405082ffb47fd490fa5d8d9f80b8ae78@172.18.188.159:21212", //
	"enode://fc7f473e7ccbdee93fce0749c4df890a8c3dd81a79678a939fb622f061ee48112d3cde8ed41f4000a21a21084ca382e339b5bd46eb58afa7962ee0f638d3342c@172.18.188.159:21214", //
	"enode://29416b59c5ce177413ab98f03fb307928e477c5437ec658cdd127633b19dc3968399990a58557e62d6ee843bb93f0d18639e42f3ec7ba0ec0869507d8b6ea551@172.18.188.159:21215", //
	"enode://c2a69b3dbdc8943052e74da95be2d1977fbde6eee2fa1537fff6286debd325ae6f9c49ef2ed09e47cc28bc88469c81645c9bcd499a6752c37c8f1fe77e2d9c50@172.18.188.159:21212",

}

var MainnetMasternodes = []string{
	"enode://c84c9860a017cadda359c2b63c29555811d02bf5839938107878ccce856447f67cc72adbad6837c18f823f4ee0a29d48405082ffb47fd490fa5d8d9f80b8ae78", //
	"enode://7428587a4839162af1c2bc93629bcf45162abae8772ad93027dfdd94fddc7c653e085b5f58fd0def2533ca61a3e380f4f5e9e891d26eea8cc11df7dcf4b77189", //
	"enode://fc7f473e7ccbdee93fce0749c4df890a8c3dd81a79678a939fb622f061ee48112d3cde8ed41f4000a21a21084ca382e339b5bd46eb58afa7962ee0f638d3342c", //
	"enode://29416b59c5ce177413ab98f03fb307928e477c5437ec658cdd127633b19dc3968399990a58557e62d6ee843bb93f0d18639e42f3ec7ba0ec0869507d8b6ea551", //
	"enode://c2a69b3dbdc8943052e74da95be2d1977fbde6eee2fa1537fff6286debd325ae6f9c49ef2ed09e47cc28bc88469c81645c9bcd499a6752c37c8f1fe77e2d9c50",

}

var TestnetMasternodes = []string{
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://30b7ab30a01c124a6cceca36863ece12c4f5fa68e3ba9b0b51407ccc002eeed3b3102d20a88f1c1d3c3154e2449317b8ef95090e77b312d5cc39354f86d5d606@52.176.7.10:21212",    // US-Azure geth
	"enode://865a63255b3bb68023b6bffd5095118fcc13e79dcf014fe4e47e065c350c7cc72af2e53eff895f11ba1bbb6a2b33271c1116ee870f266618eadfc2e78aa7349c@52.176.100.77:21212",  // US-Azure parity
	"enode://6332792c4a00e3e4ee0926ed89e0d27ef985424d97b6a45bf0f23e51f0dcb5e66b875777506458aea7af6f9e4ffb69f43f3778ee73c81ed9d34c51c4b16b0b0f@52.232.243.152:21212", // Parity
	"enode://94c15d1b9e2fe7ce56e458b9a3b672ef11894ddedd0c6f247e0f1d3487f52b66208fb4aeb8179fce6e3a749ea93ed147c37976d67af557508d199d9594c35f09@192.81.208.223:21212", // @gpip
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://c00b9f537c2b009228f95bf5a84a5f0c0e0407675efeea136e57cd8a67e16c8abdaa16661a4933bc852cc2424603d7d8f57a1199847cc0f8064dcc58814fe174@172.18.188.159:21211", //
	"enode://7a3188be1968bb029a8c0591d18364dcb32b394dc872b3239dfffffb59d753b8a2c428b08abf877b85d3c72afd26cd99a8a85b8f45e178764ff61d65a7356b40@172.18.188.159:21212", //
	"enode://883082b86d3f9224dcf51cd2d577ae049dc63b11d2e25a63ece48b4d27e571320061553bd069210388d23392b98f35e61fcbf1a3d5831a2a187ef979bbf726fe@172.18.188.159:21213", //
	"enode://92170f06c38080f71d2252faeda773348d5d2469893613ae826d03efcf780495c00e312602b9c3be91733898946ed3f6464eb469e47670cf744709da893ac278@172.18.188.159:21214", //
	"enode://a85a8b59e68df8aceb0249ad73f9e19ab707bad3bf0fe7723b9f23e9577902d095231e38053d02880536ce7f43d749e21952bb792e871e3bada99faef9342655@172.18.188.159:21215", //
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://06051a5573c81934c9554ef2898eb13b33a34b94cf36b202b69fde139ca17a85051979867720d4bdae4323d4943ddf9aeeb6643633aa656e0be843659795007a@35.177.226.168:21212",
	"enode://0cc5f5ffb5d9098c8b8c62325f3797f56509bff942704687b6530992ac706e2cb946b90a34f1f19548cd3c7baccbcaea354531e5983c7d1bc0dee16ce4b6440b@40.118.3.223:30304",
	"enode://1c7a64d76c0334b0418c004af2f67c50e36a3be60b5e4790bdac0439d21603469a85fad36f2473c9a80eb043ae60936df905fa28f1ff614c3e5dc34f15dcd2dc@40.118.3.223:30306",
	"enode://85c85d7143ae8bb96924f2b54f1b3e70d8c4d367af305325d30a61385a432f247d2c75c45c6b4a60335060d072d7f5b35dd1d4c45f76941f62a4f83b6e75daaf@40.118.3.223:30307",
}
