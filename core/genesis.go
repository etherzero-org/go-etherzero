// Copyright 2014 The go-ethereum Authors
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

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"bufio"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/common/hexutil"
	"github.com/etherzero/go-etherzero/common/math"
	"github.com/etherzero/go-etherzero/core/rawdb"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/core/types/devotedb"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/p2p/discover"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/trie"
	"io"
	"os"
	"path/filepath"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *params.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	StateRoot  common.Hash         `json:"stateRoot"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.DevoteChainConfig, common.Hash{}, errGenesisNoConfig
	}
	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
			root, err := genesisAccounts(common.Hash{}, db)
			if err != nil {
				return params.DevoteChainConfig, common.Hash{}, err
			}
			genesis.StateRoot = root
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)
		return genesis.Config, block.Hash(), err
	}
	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}
	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)

		return newcfg, stored, nil
	}
	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	if genesis == nil && stored != params.MainnetGenesisHash {
		return storedcfg, stored, nil
	}
	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {

		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {

		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	default:
		return params.DevoteChainConfig
	}
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db ethdb.Database) *types.Block {
	if db == nil {
		db = ethdb.NewMemDatabase()
	}

	statedb, _ := state.New(g.StateRoot, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance, big.NewInt(1))
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root := statedb.IntermediateRoot(false)

	// add devote protocol
	devoteDB := initGenesisDevoteProtocol(g, db)
	// add devote protocol
	protcol, _ := devoteDB.Commit()

	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       new(big.Int).SetUint64(g.Timestamp),
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		Difficulty: g.Difficulty,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
		Protocol:   protcol,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)
	block := types.NewBlock(head, nil, nil, nil)
	block.DevoteDB = devoteDB

	return block
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)

	fmt.Printf("genesis devoteProtocol Commit begin block.DevoteProtocol :%x\n", block.DevoteDB)

	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())

	config := g.Config
	if config == nil {
		config = params.AllEthashProtocolChanges
	}
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db ethdb.Database) *types.Block {
	block, err := g.Commit(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db ethdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: GenesisAlloc{addr: {Balance: balance}}}
	return g.MustCommit(db)
}

func masternodeContractAccount(masternodes []string) GenesisAccount {
	addresses := []common.Address{
		common.HexToAddress("0x6b7f544158e4dacf3247125a491241889829a436"),
		common.HexToAddress("0x8cba568d021f63465c89c573b8deaa7368adedd4"),
		common.HexToAddress("0x0198e9b6ad2b83da9cadfe3b1332ace83c7ef409"),
		common.HexToAddress("0x47bf9d89d79d61aac08322a3d23946476424e6d2"),
		common.HexToAddress("0x0c3330af31eb804800d1876c9b5fcbadcf146e79"),
		common.HexToAddress("0x7d6c6fe0fc97a567c5ed7ff9f67c7cbdf5d8b6f5"),
		common.HexToAddress("0x47d8215f49fbd0ed1f3145fac25b5d1bbefd9e04"),
		common.HexToAddress("0x99a4e8ab60add45ef834f3e6b5e920bdc71d5a10"),
		common.HexToAddress("0xa22e4712c2747a3a2fdaf0a457f46c584eeb1d40"),
		common.HexToAddress("0xb13016d02efc64517e334b91f66357f58d549433"),
		common.HexToAddress("0x711a659ede41e097c644e382e9cb320e112e4a29"),
		common.HexToAddress("0x4a59b6f943afd1ce7053482ab69117ed763340fb"),
		common.HexToAddress("0xd6f6bec1c90dd258eb91bba3a4221fcabad729bf"),
		common.HexToAddress("0xafd73d8649ac2c500f9dda354963b1a723b27a61"),
		common.HexToAddress("0x2fa2d38ab70c9af20efc2458e267686206ea4df5"),
		common.HexToAddress("0xae068e98c621a1fe061ed513822aeb32a6f4a83b"),
		common.HexToAddress("0x4affc36e567af8986f0a750eb595258052aeaa66"),
		common.HexToAddress("0xb45ed055a7f748a567219b16184b0ff4d806070f"),
		common.HexToAddress("0xefc5c57b389974c113ceff29bdb6a79034cdfde1"),
		common.HexToAddress("0x4c2ea4e923dea28d38a75726d120ff66c0d4edcd"),
		common.HexToAddress("0x9db9d0b702134b660cec839decbfc686f9952caa"),
	}

	var (
		data    = make(map[common.Hash]common.Hash)
		lastKey common.Hash
		lastId  [8]byte
	)

	for index, n := range masternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}

		var contextId common.Hash
		copy(contextId[24:32], lastId[:8])

		id1 := common.BytesToHash(node.ID[:32])
		id2 := common.BytesToHash(node.ID[32:])
		copy(lastId[:8], id1[:8])

		if lastContextId, ok := data[lastKey]; ok {
			copy(lastContextId[16:24], id1[:8])
			data[lastKey] = lastContextId
		}

		var nodeKey [64]byte
		copy(nodeKey[:8], id1[:8])
		nodeKey[63] = 2

		var contextAddress common.Hash
		copy(contextAddress[12:32], addresses[index].Bytes())

		key := new(big.Int).SetBytes(crypto.Keccak256(nodeKey[:]))
		key1 := common.BytesToHash(key.Bytes())                         // id1
		key2 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes()) // id2
		key3 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes()) // nextId,preId
		key4 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes()) // account
		lastKey = key3
		data[key1] = id1
		data[key2] = id2
		data[key3] = contextId
		data[key4] = contextAddress

		pubkey, err := node.ID.Pubkey()
		if err != nil {
			panic(err)
		}
		addr := crypto.PubkeyToAddress(*pubkey)

		var nodeAddressToIdKey [64]byte
		copy(nodeAddressToIdKey[12:32], addr[:20])
		nodeAddressToIdKey[63] = 4
		nodeAddressToIdKey1 := common.BytesToHash(crypto.Keccak256(nodeAddressToIdKey[:]))
		data[nodeAddressToIdKey1] = common.BytesToHash(id1[:8])
	}

	data[common.HexToHash("00")] = common.BytesToHash(lastId[:8])
	data[common.HexToHash("01")] = common.BytesToHash(big.NewInt(int64(len(masternodes))).Bytes())

	return GenesisAccount{
		Balance: big.NewInt(2),
		Nonce:   1,
		Storage: data,
		Code:    hexutil.MustDecode("0x6080604052600436106100e55763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306661abd811461066b57806316e7f171146106925780632c103c79146106c85780632f926732146106dd57806365f68c89146106ed578063795053d31461072b578063c1292cc31461075c578063c27cabb514610771578063c4e3ed9314610786578063c808021c146107ff578063dc1e30da14610814578063e3596ce01461086c578063e7b895b614610881578063e8c74af214610896578063f834f524146108b7578063ff5ecad2146108cb575b6000806000806100f3611046565b6100fb611061565b60008080341561010a57600080fd5b3360009081526004602052604090205460c060020a029850600160c060020a031989161580159061013f575061013f896108e0565b1561022957600160c060020a0319891660009081526002602052604081206006015498508811156101be578743039650610e1087111561019b57600160c060020a031989166000908152600260205260408120600501556101be565b600160c060020a0319891660009081526002602052604090206005018054880190555b600160c060020a03198916600081815260026020908152604091829020436006820181905560059091015483519485529184019190915282820152517fb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c29181900360600190a1610660565b3360009081526003602090815260408083205460c060020a02600160c060020a0319811684526002909252909120549099509550341580156102745750600160c060020a0319891615155b801561027f57508515155b8015610296575069043c339e0c82f4bf0000303110155b80156102a457506000600154115b15156102af57600080fd5b858552600160c060020a031989166000908152600260209081526040822060010154818801529085906080908890600b600019f115156102ee57600080fd5b50508151600160a060020a0381166000908152600460209081526040808320805467ffffffffffffffff19169055600160c060020a03198b811684526002928390529220015491925060c060020a80830292680100000000000000009004029082161561039d57600160c060020a0319821660009081526002602081905260409091200180546fffffffffffffffff000000000000000019166801000000000000000060c060020a8404021790555b600160c060020a03198116156103e657600160c060020a03198116600090815260026020819052604090912001805467ffffffffffffffff191660c060020a8404179055610400565b6000805467ffffffffffffffff191660c060020a84041790555b6101006040519081016040528060006001026000191681526020016000600102600019168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a03191681526020016000600160a060020a0316815260200160008152602001600081526020016000815250600260008b600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600060c060020a026003600033600160a060020a0316600160a060020a0316815260200190815260200160002060006101000a81548167ffffffffffffffff021916908360c060020a90040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd89336040518083600160c060020a031916600160c060020a031916815260200182600160a060020a0316600160a060020a031681526020019250505060405180910390a1604051339060009069043c339e0c82f4bf00009082818181858883f1935050505015801561065e573d6000803e3d6000fd5b505b505050505050505050005b34801561067757600080fd5b506106806108fe565b60408051918252519081900360200190f35b34801561069e57600080fd5b506106b4600160c060020a0319600435166108e0565b604080519115158252519081900360200190f35b3480156106d457600080fd5b50610680610904565b6106eb60043560243561090b565b005b3480156106f957600080fd5b5061070e600160a060020a0360043516610ce1565b60408051600160c060020a03199092168252519081900360200190f35b34801561073757600080fd5b50610740610d02565b60408051600160a060020a039092168252519081900360200190f35b34801561076857600080fd5b5061070e610d11565b34801561077d57600080fd5b50610680610d1d565b34801561079257600080fd5b506107a8600160c060020a031960043516610d29565b604080519889526020890197909752600160c060020a0319958616888801529390941660608701526080860191909152600160a060020a031660a085015260c084019190915260e083015251908190036101000190f35b34801561080b57600080fd5b50610680610d8f565b34801561082057600080fd5b50610835600160a060020a0360043516610d9a565b60408051958652602086019490945284840192909252600160a060020a039081166060850152166080830152519081900360a00190f35b34801561087857600080fd5b50610680610dd7565b34801561088d57600080fd5b50610740610ddd565b3480156108a257600080fd5b506106eb600160a060020a0360043516610dec565b6106eb600160a060020a0360043516610f2e565b3480156108d757600080fd5b50610680611038565b600160c060020a031916600090815260026020526040902054151590565b60015481565b62124f8081565b6000610915611046565b61091d611061565b8492506000831580159061093057508415155b80156109455750600160c060020a0319841615155b801561096e57503360009081526003602052604090205460c060020a02600160c060020a031916155b80156109915750600160c060020a03198416600090815260026020526040902054155b80156109a6575069043c33c193756480000034145b15156109b157600080fd5b8583526020808401869052826080856000600b600019f115156109d357600080fd5b508051600160a060020a03811615156109eb57600080fd5b836003600033600160a060020a0316600160a060020a0316815260200190815260200160002060006101000a81548167ffffffffffffffff021916908360c060020a900402179055506101006040519081016040528087600019168152602001866000191681526020016000809054906101000a900460c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a031916815260200133600160a060020a031681526020014381526020016000815260200160008152506002600086600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600060c060020a02600160c060020a0319166000809054906101000a900460c060020a02600160c060020a031916141515610c235760008054600160c060020a031960c060020a918202168252600260208190526040909220909101805491860468010000000000000000026fffffffffffffffff0000000000000000199092169190911790555b6000805460c060020a860467ffffffffffffffff19918216811783556001805481019055600160a060020a03841680845260046020526040808520805490941690921790925551909190662386f26fc100009082818181858883f19350505050158015610c94573d6000803e3d6000fd5b5060408051600160c060020a03198616815233602082015281517ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389929181900390910190a1505050505050565b600160a060020a031660009081526003602052604090205460c060020a0290565b600554600160a060020a031681565b60005460c060020a0281565b678ac7230489e8000081565b600160c060020a03191660009081526002602081905260409091208054600182015492820154600483015460038401546005850154600690950154939660c060020a808502966801000000000000000090950402949293600160a060020a039092169290565b662386f26fc1000081565b600160a060020a03908116600090815260086020526040902080546001820154600283015460038401546004909401549295919490938116921690565b610e1081565b600654600160a060020a031681565b600160a060020a038116600090815260086020526040812090610e0e33610ce1565b905060008260010154118015610e275750816001015443115b8015610e365750816002015443105b8015610e4b5750600160c060020a0319811615155b8015610e785750600160c060020a0319811660009081526002602052604090206004015460014391909103115b8015610ea85750600160a060020a038316600090815260076020908152604080832033845290915290205460ff16155b1515610eb357600080fd5b600160a060020a03831660009081526007602090815260408083203384529091529020805460ff19166001908117909155825481018355546002900482600001541115610f29574360028301556005805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0385161790555b505050565b600160a060020a038116600090815260086020526040902054158015610f6d5750600160a060020a038116600090815260086020526040902060010154155b8015610f805750678ac7230489e8000034145b1515610f8b57600080fd5b6040805160a081018252600080825243602080840182815262124f80909201848601908152336060860190815260068054600160a060020a03908116608089019081529981168088526008909552979095209551865592516001860155516002850155905160038401805491861673ffffffffffffffffffffffffffffffffffffffff19928316179055945160049093018054939094169285169290921790925581549092169091179055565b69043c33c193756480000081565b60408051808201825290600290829080388339509192915050565b60206040519081016040528060019060208202803883395091929150505600a165627a7a72305820735cec9f639b15b842e04720633bb78329eb9d3409f4215db7cdff6340609b850029"),
	}
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	alloc := decodePrealloc(mainnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.MainnetMasternodes)
	alloc[common.HexToAddress("0x6b7f544158e4dacf3247125a491241889829a436")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(1e+16), big.NewInt(1e+15)),
	}

	alloc[common.HexToAddress("0x281A16dbBE7810eDc892DD365eE377CC0Fee9AC9")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x5a5a12E5AAAD081367301E49d429Eee37EC68B9E")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x91e51bcb44C9FF0F41d05560936e369027A6942f")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xd4Dcff6AcfdBbF4a22437c0897d4Ca2688c24FE8")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xDB61C948a51c68B6B1092B7c891C1eb5E11381C1")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x0C01739bC45FC63f1Ce524a465b5865F301DC03D")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xa3b2FDC0d193f4A18eD383063ECAb2452B32E21f")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xCBD0A40Dcf146B74fcF368cc66c693802f0fB479")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xA2F7EeB6800FfD24b9F5a0939afae57B33268112")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xA0cdbe530F33c5368ED2B714415CDf9183293d48")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x8A23a7712a5A156f030D4C87D503e02e41B71bF1")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x6adfc3e09bab6a854537129fb6ff6062A59E821A")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x2B15c7cCedbae9d750Cd477D870Cd73A50062e9e")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x2A2bE4FE883544EfFc5F6efF8D6334184463afD7")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x435B527E6b13f65c079160d8A2312C3064B34C02")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x884B87FD59CEe8F56ffafdAC739325513FAedf39")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x9589359e0C97471D0e8F0a002B27916Ce31B0d36")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x40E48EBF166Af172AC17DDe1fA4E70c09bd46925")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xbffA822C3D4dE45d82c5dC3db82521c8eeE48048")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x03F3C1F292c4cD64625D6Ba69529973639D848Cd")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x0797a068b3f65304104a61E1900e617952E95eCA")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x1a0FB32c69Eba29787222c8BcEd6eB34400c292F")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xbC8bc5f5174b0ef35dCeEd2580adc55d60293A12")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x9685E1FC92B4e2F3CF0aa3F60d452f24Ee3183Ea")] = GenesisAccount{Balance: big.NewInt(1e+16)}

	return &Genesis{
		Config:     params.DevoteChainConfig,
		Nonce:      66,
		Timestamp:  1531551970,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   30000000,
		Difficulty: big.NewInt(1),
		Alloc:      alloc,
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	alloc := decodePrealloc(testnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.TestnetMasternodes)
	alloc[common.HexToAddress("0x6b7f544158e4dacf3247125a491241889829a436")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(1e+15), big.NewInt(1e+15)),
	}
	return &Genesis{
		Config:     params.TestnetChainConfig,
		Nonce:      66,
		Timestamp:  1531551970,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1048576),
		Alloc:      alloc,
	}
}

// DefaultRinkebyGenesisBlock returns the Rinkeby network genesis block.
func DefaultRinkebyGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.RinkebyChainConfig,
		Timestamp:  1492009146,
		ExtraData:  hexutil.MustDecode("0x52657370656374206d7920617574686f7269746168207e452e436172746d616e42eb768f2244c8811c63729a21a3569731535f067ffc57839b00206d1ad20c69a1981b489f772031b279182d99e65703f0076e4812653aab85fca0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc:      decodePrealloc(rinkebyAllocData),
	}
}

// DeveloperGenesisBlock returns the 'geth --dev' genesis block. Note, this must
// be seeded with the
func DeveloperGenesisBlock(period uint64, faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params.AllCliqueProtocolChanges
	config.Clique.Period = period
	alloc := decodePrealloc(testnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.TestnetMasternodes)
	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, 65)...),
		GasLimit:   6283185,
		Difficulty: big.NewInt(1),
		Alloc: map[common.Address]GenesisAccount{
			common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
			common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
			common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
			common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
			common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
			common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
			common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
			common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
			faucet: {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
		},
	}
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}

	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}

func initGenesisDevoteProtocol(g *Genesis, db ethdb.Database) *devotedb.DevoteDB {

	devoteDB, err := devotedb.NewDevoteByProtocol(devotedb.NewDatabase(db), &devotedb.DevoteProtocol{})
	if err != nil {
		return nil
	}
	if g.Config != nil && g.Config.Devote != nil && g.Config.Devote.Witnesses != nil {
		genesisCycle := g.Timestamp / params.CycleInterval
		devoteDB.SetWitnesses(genesisCycle, g.Config.Devote.Witnesses)
	}
	return devoteDB
}

func genesisAccounts(root common.Hash, db ethdb.Database) (common.Hash, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return common.Hash{}, err
	}
	path := strings.Replace(dir, "\\", "/", -1) + "/init.bin"
	file, err := os.Open(path)
	if err != nil {
		return common.Hash{}, err
	}
	defer file.Close()

	triedb := trie.NewDatabase(db)
	tr, err := trie.New(root, triedb)
	if err != nil {
		return common.Hash{}, err
	}

	bufReader := bufio.NewReader(file)
	buf := make([]byte, 43)
	accountCount := 0
	log.Info("Import initial accounts, waitting ...")
	emptyRoot := common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	emptyState := crypto.Keccak256Hash(nil).Bytes()
	for {
		readNum, err := bufReader.Read(buf[0:43])
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == readNum {
			break
		}
		for readNum < 43 {
			n, err := bufReader.Read(buf[readNum:43])
			if err != nil && err != io.EOF {
				panic(err)
			}
			readNum += n
		}
		var account = state.Account{
			Balance:     new(big.Int).SetBytes(buf[32:43]),
			Power:       common.Big0,
			BlockNumber: common.Big0,
			Root:        emptyRoot,
			CodeHash:    emptyState,
		}
		encodeData, err := rlp.EncodeToBytes(&account)
		if err != nil {
			panic(err)
		}
		tr.TryUpdate(buf[0:32], encodeData)
		if accountCount%100000 == 0 {
			root1, err := tr.Commit(nil)
			if err != nil {
				panic(err)
			}
			triedb.Commit(root1, true)
			log.Info("Import initial accounts", "count", accountCount)
		}
		accountCount++
	}
	log.Info("Import initial accounts", "count", accountCount)
	root2, err := tr.Commit(nil)
	if err != nil {
		panic(err)
	}
	triedb.Commit(root2, true)
	return root2, nil
}
