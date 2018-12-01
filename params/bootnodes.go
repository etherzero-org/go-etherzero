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
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212",
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033",
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034",
}

var MainnetMasternodes = []string{
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212",
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033",
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034",
}

var TestnetMasternodes = []string{
	"enode://f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be6235fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3@127.0.0.1:21212",
	"enode://7ec780bcd5488bcf83cdabc52b1471e3efa2a5575a2e3c1b886766529d4d3d0036e5a36958b3445e657cfe716148a64f6768e7cf55c61ab58af6e11cf7c80443@127.0.0.1:20033",
	"enode://f550ac2228cb2bc8373ec740c7064876b60e71c66175075c617ed5988c500fcaa481bcdedd74b1f45ae76412ae5b7cf5ebc251e6e3c0c141d1f10ff3d67d5b6e@127.0.0.1:20034",
}
// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{

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

/*
miner.start()

*/