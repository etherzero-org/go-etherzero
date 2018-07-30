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
	var (
		data    = make(map[common.Hash]common.Hash)
		lastKey common.Hash
		lastId  [8]byte
	)

	for _, n := range masternodes {
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

		key := new(big.Int).SetBytes(crypto.Keccak256(nodeKey[:]))
		key1 := common.BytesToHash(key.Bytes())                         // id1
		key2 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes()) // id2
		key3 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes()) // nextId,preId
		lastKey = key3
		data[key1] = id1
		data[key2] = id2
		data[key3] = contextId

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
		Code:    hexutil.MustDecode("0x608060405260043610610099576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd14610d5e57806316e7f17114610d895780632f92673214610de957806365f68c8914610e1b578063c1292cc314610ea8578063c4e3ed9314610f09578063c808021c1461103e578063e3596ce014611069578063ff5ecad214611094575b6000806000806100a7611cf1565b6100af611d13565b600080600080341415156100c257600080fd5b600460003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029850600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168977ffffffffffffffffffffffffffffffffffffffffffffffff191614158015610190575061018f896110bf565b5b156103ce57600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060060154975060008811156102af578743039650610e10871115610253576000600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501819055506102ae565b86600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501600082825401925050819055505b5b43600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601819055507fb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c289600260008c77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206005015443604051808477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001838152602001828152602001935050505060405180910390a1610d53565b600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029850600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015495506000341480156104e657508877ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156104fe5750856000191660006001026000191614155b80156105335750662386f26fc100006801158e460913d00000033073ffffffffffffffffffffffffffffffffffffffff163110155b801561054157506000600154115b151561054c57600080fd5b8585600060028110151561055c57fe5b60200201906000191690816000191681525050600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600101548560016002811015156105cb57fe5b602002019060001916908160001916815250506020846080876000600b600019f115156105f757600080fd5b83600060018110151561060657fe5b60200201516001900492506000780100000000000000000000000000000000000000000000000002600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a90047801000000000000000000000000000000000000000000000000029150600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a90047801000000000000000000000000000000000000000000000000029050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168277ffffffffffffffffffffffffffffffffffffffffffffffff191614151561086e5780600260008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff19161415156109535781600260008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555061098e565b816000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b6101006040519081016040528060006001026000191681526020016000600102600019168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081526020016000815250600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060a0820151816004015560c0820151816005015560e082015181600601559050506000780100000000000000000000000000000000000000000000000002600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd8933604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100006801158e460913d00000039081150290604051600060405180830381858888f19350505050158015610d51573d6000803e3d6000fd5b505b505050505050505050005b348015610d6a57600080fd5b50610d73611123565b6040518082815260200191505060405180910390f35b348015610d9557600080fd5b50610dcf600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff191690602001909291905050506110bf565b604051808215151515815260200191505060405180910390f35b610e1960048036038101908080356000191690602001909291908035600019169060200190929190505050611129565b005b348015610e2757600080fd5b50610e5c600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611942565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610eb457600080fd5b50610ebd6119b0565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610f1557600080fd5b50610f4f600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff191690602001909291905050506119da565b60405180896000191660001916815260200188600019166000191681526020018777ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018381526020018281526020019850505050505050505060405180910390f35b34801561104a57600080fd5b50611053611cd3565b6040518082815260200191505060405180910390f35b34801561107557600080fd5b5061107e611cde565b6040518082815260200191505060405180910390f35b3480156110a057600080fd5b506110a9611ce4565b6040518082815260200191505060405180910390f35b60008060010260001916600260008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001546000191614159050919050565b60015481565b6000611133611cf1565b61113b611d13565b60008593508560001916600060010260001916141580156111685750846000191660006001026000191614155b80156111c657508377ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156112875750600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916145b80156112ea5750600260008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015460001916600060010260001916145b80156112fe57506801158e460913d0000034145b151561130957600080fd5b8583600060028110151561131957fe5b602002019060001916908160001916815250508483600160028110151561133c57fe5b602002019060001916908160001916815250506020826080856000600b600019f1151561136857600080fd5b81600060018110151561137757fe5b6020020151600190049050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16141515156113be57600080fd5b83600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506101006040519081016040528087600019168152602001866000191681526020016000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020013373ffffffffffffffffffffffffffffffffffffffff168152602001438152602001600081526020016000815250600260008677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19166000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19161415156117895783600260008060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b836000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506001806000828254019250508190555083600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055508073ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100009081150290604051600060405180830381858888f19350505050158015611898573d6000803e3d6000fd5b507ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac3898433604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a1505050505050565b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029050919050565b6000809054906101000a900478010000000000000000000000000000000000000000000000000281565b600080600080600080600080600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001549750600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600101549650600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a90047801000000000000000000000000000000000000000000000000029550600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a90047801000000000000000000000000000000000000000000000000029450600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600401549350600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169250600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501549150600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601549050919395975091939597565b662386f26fc1000081565b610e1081565b6801158e460913d0000081565b6040805190810160405280600290602082028038833980820191505090505090565b6020604051908101604052806001906020820280388339808201915050905050905600a165627a7a723058209d766c0d6af2154380de216af5d42ea3c847e183a359218291354ce12b197fa20029"),
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
