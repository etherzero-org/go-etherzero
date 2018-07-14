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
	"encoding/binary"
	"github.com/etherzero/go-etherzero/crypto"
	"github.com/etherzero/go-etherzero/p2p/discover"
	"net"
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
	block := types.NewBlock(head, nil, nil, nil, nil)
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
	tempBytes := common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006")
	data := make(map[common.Hash]common.Hash)
	var (
		lastKey4 common.Hash
	)
	for _, n := range masternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}
		var contextId common.Hash

		copy(contextId[24:32], tempBytes[:8])

		id1 := common.BytesToHash(node.ID[:32])
		id2 := common.BytesToHash(node.ID[32:])
		var misc [32]byte
		misc[0] = 1
		var ip net.IP
		if len(node.IP) == 4 {
			ip = net.IPv4(node.IP[0], node.IP[1], node.IP[2], node.IP[3])
		} else {
			ip = node.IP
		}
		copy(misc[1:17], ip[:16])
		binary.BigEndian.PutUint16(misc[17:19], uint16(node.TCP))

		if lastContextId, ok := data[lastKey4]; ok {
			copy(lastContextId[16:24], id1[:8])
			data[lastKey4] = lastContextId
		}

		copy(tempBytes[:8], id1[:8])
		key := new(big.Int).SetBytes(crypto.Keccak256(tempBytes))
		key1 := common.BytesToHash(key.Bytes())
		key2 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes())
		key3 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes())
		key4 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes())
		key5 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes())
		key6 := common.BytesToHash(key.Add(key, big.NewInt(1)).Bytes())
		lastKey4 = key4
		data[key1] = id1
		data[key2] = id2
		data[key3] = misc
		data[key4] = contextId
		data[key5] = common.HexToHash("01")
		data[key6] = common.HexToHash("ff")
	}
	data[common.HexToHash("00")] = common.BytesToHash(tempBytes[:8])
	data[common.HexToHash("01")] = common.BytesToHash(big.NewInt(int64(len(masternodes))).Bytes())

	return GenesisAccount{
		Balance: big.NewInt(2),
		Nonce:   1,
		Storage: data,
		Code:    common.Hex2Bytes("6080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306661abd811461038057806316e7f171146103a75780634da274fd146103dd57806365f68c89146103f0578063795053d31461042e578063c1292cc31461045f578063c4e3ed9314610474578063e8c74af2146104e2578063f834f52414610503578063ff5ecad214610517575b3360009081526005602090815260408083205460c060020a02600160c060020a0319811684526006909252822054909180341580156100eb5750600160c060020a0319841615155b80156100f657508215155b801561010c57506801158e460913d00000303110155b801561011a57506000600154115b151561012557600080fd5b5050600160c060020a031980831660009081526006602052604090206003015460c060020a808202926801000000000000000090920402908216156101ab57600160c060020a03198216600090815260066020526040902060030180546fffffffffffffffff000000000000000019166801000000000000000060c060020a8404021790555b600160c060020a03198116156101f357600160c060020a031981166000908152600660205260409020600301805467ffffffffffffffff191660c060020a840417905561020d565b6000805467ffffffffffffffff191660c060020a84041790555b6040805160e08101825260008082526020808301828152838501838152606085018481526080860185815260a0870186815260c08801878152600160c060020a03198e16808952600688528a892099518a55955160018a810191909155945160028a01559251600389018054935160c060020a9081900468010000000000000000026fffffffffffffffff0000000000000000199190930467ffffffffffffffff199586161716919091179055516004880155905160059687018054600160a060020a039290921673ffffffffffffffffffffffffffffffffffffffff1990921691909117905533808652958452938690208054909416909355825460001901909255835191825281019190915281517f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd929181900390910190a160405133906000906801158e460913d000009082818181858883f19350505050158015610379573d6000803e3d6000fd5b5050505050005b34801561038c57600080fd5b5061039561052c565b60408051918252519081900360200190f35b3480156103b357600080fd5b506103c9600160c060020a031960043516610532565b604080519115158252519081900360200190f35b6103ee600435602435604435610550565b005b3480156103fc57600080fd5b50610411600160a060020a03600435166107b8565b60408051600160c060020a03199092168252519081900360200190f35b34801561043a57600080fd5b506104436107d9565b60408051600160a060020a039092168252519081900360200190f35b34801561046b57600080fd5b506104116107e8565b34801561048057600080fd5b50610496600160c060020a0319600435166107f4565b60408051978852602088019690965286860194909452600160c060020a031992831660608701529116608085015260a0840152600160a060020a031660c0830152519081900360e00190f35b3480156104ee57600080fd5b506103ee600160a060020a0360043516610855565b6103ee600160a060020a036004351661095a565b34801561052357600080fd5b50610395610a1a565b60015481565b600160c060020a031916600090815260066020526040902054151590565b82801580159061055f57508215155b801561056a57508115155b801561057f5750600160c060020a0319811615155b80156105a857503360009081526005602052604090205460c060020a02600160c060020a031916155b80156105cb5750600160c060020a03198116600090815260066020526040902054155b80156105df57506801158e460913d0000034145b15156105ea57600080fd5b336000818152600560208181526040808420805460c060020a80890467ffffffffffffffff1992831617909255825160e0810184528b81528085018b81528185018b81528854600160c060020a0319908602811660608501908152608085018b81524360a0870190815260c087019d8e52838f168d526006909a52978b209451855592516001850155905160028401559051600383018054965186900468010000000000000000026fffffffffffffffff000000000000000019928790049790951696909617169290921790935592516004830155945192018054600160a060020a039390931673ffffffffffffffffffffffffffffffffffffffff19909316929092179091559054909102161561074d5760008054600160c060020a031960c060020a91820216825260066020526040909120600301805491830468010000000000000000026fffffffffffffffff0000000000000000199092169190911790555b6000805467ffffffffffffffff191660c060020a8304179055600180548101905560408051600160c060020a03198316815233602082015281517ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389929181900390910190a150505050565b600160a060020a031660009081526005602052604090205460c060020a0290565b600254600160a060020a031681565b60005460c060020a0281565b600160c060020a03191660009081526006602052604090208054600182015460028301546003840154600485015460059095015493959294919360c060020a8083029468010000000000000000909304029291600160a060020a0390911690565b600160a060020a03811660009081526004602052604081206001810154909110801561089457506000610887336107b8565b600160c060020a03191614155b80156108a257506002810154155b80156108d25750600160a060020a038216600090815260036020908152604080832033845290915290205460ff16155b15156108dd57600080fd5b600160a060020a03821660009081526003602090815260408083203384529091529020805460ff19166001908117909155815481018255546064906042028254919004116109565743600282810191909155805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0384161790555b5050565b600160a060020a0381166000908152600460205260409020541580156109995750600160a060020a038116600090815260046020526040902060010154155b15156109a457600080fd5b6040805160808101825260008082524360208084019182528385018381523360608601908152600160a060020a039788168552600490925294909220925183555160018301559151600282015590516003909101805473ffffffffffffffffffffffffffffffffffffffff191691909216179055565b6801158e460913d00000815600a165627a7a723058204799a0d81c2e172f93d560011009987aaaeaf4fbaa6af524d3fb02387e14c5050029"),
	}
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	alloc := decodePrealloc(mainnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.MainnetMasternodes)
	return &Genesis{
		Config:     params.DevoteChainConfig,
		Nonce:      66,
		Timestamp:  1531551970,
		ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
		GasLimit:   5000,
		Difficulty: big.NewInt(17179869184),
		Alloc:      alloc,
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	alloc := decodePrealloc(testnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.TestnetMasternodes)
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
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}

func initGenesisDevoteProtocol(g *Genesis, db ethdb.Database) *types.DevoteProtocol {

	dp, err := types.NewDevoteProtocolFromAtomic(db, &types.DevoteProtocolAtomic{})
	if err != nil {
		return nil
	}

	if g.Config != nil && g.Config.Devote != nil && g.Config.Devote.Witnesses != nil {
		dp.SetWitnesses(g.Config.Devote.Witnesses)
		for _, witness := range g.Config.Devote.Witnesses {
			dp.MasternodeTrie().TryUpdate([]byte(witness), common.Address{}.Bytes())
		}
	}
	return dp
}
