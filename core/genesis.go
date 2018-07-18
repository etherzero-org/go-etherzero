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

	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/common/hexutil"
	"github.com/etherzero/go-etherzero/common/math"
	"github.com/etherzero/go-etherzero/core/rawdb"
	"github.com/etherzero/go-etherzero/core/state"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/log"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/rlp"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/p2p/discover"
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
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
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
	devoteProtocol := initGenesisDevoteProtocol(g, db)
	devoteProtocolAtomic := devoteProtocol.ProtocolAtomic()

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
		Protocol:   devoteProtocolAtomic,
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
	block.DevoteProtocol = devoteProtocol

	return block
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)

	fmt.Printf("genesis devoteProtocol Commit begin block.DevoteProtocol :%x\n", block.DevoteProtocol)
	// add devote protocol
	if _, err := block.DevoteProtocol.Commit(db); err != nil {
		return nil, err
	}

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
		data = make(map[common.Hash]common.Hash)
		lastKey common.Hash
		lastId [8]byte
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
		Code:    hexutil.MustDecode("0x6080604052600436106100985763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306661abd811461061c57806316e7f171146106435780632f9267321461067957806365f68c8914610689578063c1292cc3146106c7578063c4e3ed93146106dc578063c808021c14610755578063e3596ce01461076a578063ff5ecad21461077f575b6000806000806100a6610c3e565b6100ae610c59565b6000808034156100bd57600080fd5b3360009081526004602052604090205460c060020a029850600160c060020a03198916158015906100f257506100f289610794565b156101dc57600160c060020a03198916600090815260026020526040812060060154985088111561017157874303965061012c87111561014e57600160c060020a03198916600090815260026020526040812060050155610171565b600160c060020a0319891660009081526002602052604090206005018054880190555b600160c060020a03198916600081815260026020908152604091829020436006820181905560059091015483519485529184019190915282820152517fb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c29181900360600190a1610611565b3360009081526003602090815260408083205460c060020a02600160c060020a0319811684526002909252909120549099509550341580156102275750600160c060020a0319891615155b801561023257508515155b801561024857506801156abf16a40f0000303110155b801561025657506000600154115b151561026157600080fd5b858552600160c060020a031989166000908152600260209081526040822060010154818801529085906080908890600b600019f115156102a057600080fd5b50508151600160a060020a0381166000908152600460209081526040808320805467ffffffffffffffff19169055600160c060020a03198b811684526002928390529220015491925060c060020a80830292680100000000000000009004029082161561034f57600160c060020a0319821660009081526002602081905260409091200180546fffffffffffffffff000000000000000019166801000000000000000060c060020a8404021790555b600160c060020a031981161561039857600160c060020a03198116600090815260026020819052604090912001805467ffffffffffffffff191660c060020a84041790556103b2565b6000805467ffffffffffffffff191660c060020a84041790555b6101006040519081016040528060006001026000191681526020016000600102600019168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a03191681526020016000600160a060020a0316815260200160008152602001600081526020016000815250600260008b600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600060c060020a026003600033600160a060020a0316600160a060020a0316815260200190815260200160002060006101000a81548167ffffffffffffffff021916908360c060020a90040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd89336040518083600160c060020a031916600160c060020a031916815260200182600160a060020a0316600160a060020a031681526020019250505060405180910390a160405133906000906801156abf16a40f00009082818181858883f1935050505015801561060f573d6000803e3d6000fd5b505b505050505050505050005b34801561062857600080fd5b506106316107b2565b60408051918252519081900360200190f35b34801561064f57600080fd5b50610665600160c060020a031960043516610794565b604080519115158252519081900360200190f35b6106876004356024356107b8565b005b34801561069557600080fd5b506106aa600160a060020a0360043516610b8d565b60408051600160c060020a03199092168252519081900360200190f35b3480156106d357600080fd5b506106aa610bae565b3480156106e857600080fd5b506106fe600160c060020a031960043516610bba565b604080519889526020890197909752600160c060020a0319958616888801529390941660608701526080860191909152600160a060020a031660a085015260c084019190915260e083015251908190036101000190f35b34801561076157600080fd5b50610631610c20565b34801561077657600080fd5b50610631610c2b565b34801561078b57600080fd5b50610631610c31565b600160c060020a031916600090815260026020526040902054151590565b60015481565b60006107c2610c3e565b6107ca610c59565b849250600083158015906107dd57508415155b80156107f25750600160c060020a0319841615155b801561081b57503360009081526003602052604090205460c060020a02600160c060020a031916155b801561083e5750600160c060020a03198416600090815260026020526040902054155b801561085257506801158e460913d0000034145b151561085d57600080fd5b8583526020808401869052826080856000600b600019f1151561087f57600080fd5b508051600160a060020a038116151561089757600080fd5b836003600033600160a060020a0316600160a060020a0316815260200190815260200160002060006101000a81548167ffffffffffffffff021916908360c060020a900402179055506101006040519081016040528087600019168152602001866000191681526020016000809054906101000a900460c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a031916815260200133600160a060020a031681526020014381526020016000815260200160008152506002600086600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600060c060020a02600160c060020a0319166000809054906101000a900460c060020a02600160c060020a031916141515610acf5760008054600160c060020a031960c060020a918202168252600260208190526040909220909101805491860468010000000000000000026fffffffffffffffff0000000000000000199092169190911790555b6000805460c060020a860467ffffffffffffffff19918216811783556001805481019055600160a060020a03841680845260046020526040808520805490941690921790925551909190662386f26fc100009082818181858883f19350505050158015610b40573d6000803e3d6000fd5b5060408051600160c060020a03198616815233602082015281517ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389929181900390910190a1505050505050565b600160a060020a031660009081526003602052604090205460c060020a0290565b60005460c060020a0281565b600160c060020a03191660009081526002602081905260409091208054600182015492820154600483015460038401546005850154600690950154939660c060020a808502966801000000000000000090950402949293600160a060020a039092169290565b662386f26fc1000081565b61012c81565b6801158e460913d0000081565b60408051808201825290600290829080388339509192915050565b60206040519081016040528060019060208202803883395091929150505600a165627a7a7230582033d3d5f99c0734a675392bcaa3f4666c0ba838fc173a91f8e0ddeeaf4ce11bb00029"),
	}
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	alloc := decodePrealloc(mainnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.MainnetMasternodes)
	alloc[common.HexToAddress("0x6b7f544158e4dacf3247125a491241889829a436")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(1e+16), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x1a0FB32c69Eba29787222c8BcEd6eB34400c292F")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0xbC8bc5f5174b0ef35dCeEd2580adc55d60293A12")] = GenesisAccount{Balance: big.NewInt(1e+16)}
	alloc[common.HexToAddress("0x9685E1FC92B4e2F3CF0aa3F60d452f24Ee3183Ea")] = GenesisAccount{Balance: big.NewInt(1e+16)}

	return &Genesis{
		Config:     params.DevoteChainConfig,
		Nonce:      66,
		Timestamp:  1531551970,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1048576),
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
			common.BytesToAddress([]byte{1}):                               {Balance: big.NewInt(1)}, // ECRecover
			common.BytesToAddress([]byte{2}):                               {Balance: big.NewInt(1)}, // SHA256
			common.BytesToAddress([]byte{3}):                               {Balance: big.NewInt(1)}, // RIPEMD
			common.BytesToAddress([]byte{4}):                               {Balance: big.NewInt(1)}, // Identity
			common.BytesToAddress([]byte{5}):                               {Balance: big.NewInt(1)}, // ModExp
			common.BytesToAddress([]byte{6}):                               {Balance: big.NewInt(1)}, // ECAdd
			common.BytesToAddress([]byte{7}):                               {Balance: big.NewInt(1)}, // ECScalarMul
			common.BytesToAddress([]byte{8}):                               {Balance: big.NewInt(1)}, // ECPairing
			faucet:                                                         {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
			common.BytesToAddress(params.MasterndeContractAddress.Bytes()): masternodeContractAccount(params.TestnetMasternodes),
		},
	}
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	//for _, account := range p {
	//	ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	//}
	return ga
}

func initGenesisDevoteProtocol(g *Genesis, db ethdb.Database) *types.DevoteProtocol {

	dp, err := types.NewDevoteProtocolFromAtomic(db, &types.DevoteProtocolAtomic{})
	if err != nil {
		return nil
	}

	if g.Config != nil && g.Config.Devote != nil && g.Config.Devote.Witnesses != nil {
		dp.SetWitnesses(g.Config.Devote.Witnesses)
	}
	return dp
}
