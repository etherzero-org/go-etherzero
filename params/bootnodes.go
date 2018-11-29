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
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212", //
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033", //
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034", //
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212", //
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033", //
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034", //
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212", //
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033", //
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034", //
}
