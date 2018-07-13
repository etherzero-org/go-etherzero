// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/etherzero/go-etherzero"
	"github.com/etherzero/go-etherzero/accounts/abi"
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/event"
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getId\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"governanceAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes8\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"account\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"voteForGovernanceAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"createGovernanceAddressVote\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzPerNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"join\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"quit\",\"type\":\"event\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `0x608060405234801561001057600080fd5b506000805467ffffffffffffffff19168155600155610a53806100346000396000f3006080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306661abd811461038057806316e7f171146103a75780634da274fd146103dd57806365f68c89146103f0578063795053d31461042e578063c1292cc31461045f578063c4e3ed9314610474578063e8c74af2146104e2578063f834f52414610503578063ff5ecad214610517575b3360009081526005602090815260408083205460c060020a02600160c060020a0319811684526006909252822054909180341580156100eb5750600160c060020a0319841615155b80156100f657508215155b801561010c57506801158e460913d00000303110155b801561011a57506000600154115b151561012557600080fd5b5050600160c060020a031980831660009081526006602052604090206003015460c060020a808202926801000000000000000090920402908216156101ab57600160c060020a03198216600090815260066020526040902060030180546fffffffffffffffff000000000000000019166801000000000000000060c060020a8404021790555b600160c060020a03198116156101f357600160c060020a031981166000908152600660205260409020600301805467ffffffffffffffff191660c060020a840417905561020d565b6000805467ffffffffffffffff191660c060020a84041790555b6040805160e08101825260008082526020808301828152838501838152606085018481526080860185815260a0870186815260c08801878152600160c060020a03198e16808952600688528a892099518a55955160018a810191909155945160028a01559251600389018054935160c060020a9081900468010000000000000000026fffffffffffffffff0000000000000000199190930467ffffffffffffffff199586161716919091179055516004880155905160059687018054600160a060020a039290921673ffffffffffffffffffffffffffffffffffffffff1990921691909117905533808652958452938690208054909416909355825460001901909255835191825281019190915281517f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd929181900390910190a160405133906000906801158e460913d000009082818181858883f19350505050158015610379573d6000803e3d6000fd5b5050505050005b34801561038c57600080fd5b5061039561052c565b60408051918252519081900360200190f35b3480156103b357600080fd5b506103c9600160c060020a031960043516610532565b604080519115158252519081900360200190f35b6103ee600435602435604435610550565b005b3480156103fc57600080fd5b50610411600160a060020a03600435166107b8565b60408051600160c060020a03199092168252519081900360200190f35b34801561043a57600080fd5b506104436107d9565b60408051600160a060020a039092168252519081900360200190f35b34801561046b57600080fd5b506104116107e8565b34801561048057600080fd5b50610496600160c060020a0319600435166107f4565b60408051978852602088019690965286860194909452600160c060020a031992831660608701529116608085015260a0840152600160a060020a031660c0830152519081900360e00190f35b3480156104ee57600080fd5b506103ee600160a060020a0360043516610855565b6103ee600160a060020a036004351661095a565b34801561052357600080fd5b50610395610a1a565b60015481565b600160c060020a031916600090815260066020526040902054151590565b82801580159061055f57508215155b801561056a57508115155b801561057f5750600160c060020a0319811615155b80156105a857503360009081526005602052604090205460c060020a02600160c060020a031916155b80156105cb5750600160c060020a03198116600090815260066020526040902054155b80156105df57506801158e460913d0000034145b15156105ea57600080fd5b336000818152600560208181526040808420805460c060020a80890467ffffffffffffffff1992831617909255825160e0810184528b81528085018b81528185018b81528854600160c060020a0319908602811660608501908152608085018b81524360a0870190815260c087019d8e52838f168d526006909a52978b209451855592516001850155905160028401559051600383018054965186900468010000000000000000026fffffffffffffffff000000000000000019928790049790951696909617169290921790935592516004830155945192018054600160a060020a039390931673ffffffffffffffffffffffffffffffffffffffff19909316929092179091559054909102161561074d5760008054600160c060020a031960c060020a91820216825260066020526040909120600301805491830468010000000000000000026fffffffffffffffff0000000000000000199092169190911790555b6000805467ffffffffffffffff191660c060020a8304179055600180548101905560408051600160c060020a03198316815233602082015281517ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389929181900390910190a150505050565b600160a060020a031660009081526005602052604090205460c060020a0290565b600254600160a060020a031681565b60005460c060020a0281565b600160c060020a03191660009081526006602052604090208054600182015460028301546003840154600485015460059095015493959294919360c060020a8083029468010000000000000000909304029291600160a060020a0390911690565b600160a060020a03811660009081526004602052604081206001810154909110801561089457506000610887336107b8565b600160c060020a03191614155b80156108a257506002810154155b80156108d25750600160a060020a038216600090815260036020908152604080832033845290915290205460ff16155b15156108dd57600080fd5b600160a060020a03821660009081526003602090815260408083203384529091529020805460ff19166001908117909155815481018255546064906042028254919004116109565743600282810191909155805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0384161790555b5050565b600160a060020a0381166000908152600460205260409020541580156109995750600160a060020a038116600090815260046020526040902060010154155b15156109a457600080fd5b6040805160808101825260008082524360208084019182528385018381523360608601908152600160a060020a039788168552600490925294909220925183555160018301559151600282015590516003909101805473ffffffffffffffffffffffffffffffffffffffff191691909216179055565b6801158e460913d00000815600a165627a7a723058204799a0d81c2e172f93d560011009987aaaeaf4fbaa6af524d3fb02387e14c5050029`

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint256)
func (_Contract *ContractCaller) Count(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "count")
	return *ret0, err
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint256)
func (_Contract *ContractSession) Count() (*big.Int, error) {
	return _Contract.Contract.Count(&_Contract.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint256)
func (_Contract *ContractCallerSession) Count() (*big.Int, error) {
	return _Contract.Contract.Count(&_Contract.CallOpts)
}

// EtzPerNode is a free data retrieval call binding the contract method 0xff5ecad2.
//
// Solidity: function etzPerNode() constant returns(uint256)
func (_Contract *ContractCaller) EtzPerNode(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "etzPerNode")
	return *ret0, err
}

// EtzPerNode is a free data retrieval call binding the contract method 0xff5ecad2.
//
// Solidity: function etzPerNode() constant returns(uint256)
func (_Contract *ContractSession) EtzPerNode() (*big.Int, error) {
	return _Contract.Contract.EtzPerNode(&_Contract.CallOpts)
}

// EtzPerNode is a free data retrieval call binding the contract method 0xff5ecad2.
//
// Solidity: function etzPerNode() constant returns(uint256)
func (_Contract *ContractCallerSession) EtzPerNode() (*big.Int, error) {
	return _Contract.Contract.EtzPerNode(&_Contract.CallOpts)
}

// GetId is a free data retrieval call binding the contract method 0x65f68c89.
//
// Solidity: function getId(addr address) constant returns(id bytes8)
func (_Contract *ContractCaller) GetId(opts *bind.CallOpts, addr common.Address) ([8]byte, error) {
	var (
		ret0 = new([8]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getId", addr)
	return *ret0, err
}

// GetId is a free data retrieval call binding the contract method 0x65f68c89.
//
// Solidity: function getId(addr address) constant returns(id bytes8)
func (_Contract *ContractSession) GetId(addr common.Address) ([8]byte, error) {
	return _Contract.Contract.GetId(&_Contract.CallOpts, addr)
}

// GetId is a free data retrieval call binding the contract method 0x65f68c89.
//
// Solidity: function getId(addr address) constant returns(id bytes8)
func (_Contract *ContractCallerSession) GetId(addr common.Address) ([8]byte, error) {
	return _Contract.Contract.GetId(&_Contract.CallOpts, addr)
}

// GetInfo is a free data retrieval call binding the contract method 0xc4e3ed93.
//
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, misc bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address)
func (_Contract *ContractCaller) GetInfo(opts *bind.CallOpts, id [8]byte) (struct {
	Id1         [32]byte
	Id2         [32]byte
	Misc        [32]byte
	PreId       [8]byte
	NextId      [8]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	ret := new(struct {
		Id1         [32]byte
		Id2         [32]byte
		Misc        [32]byte
		PreId       [8]byte
		NextId      [8]byte
		BlockNumber *big.Int
		Account     common.Address
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getInfo", id)
	return *ret, err
}

// GetInfo is a free data retrieval call binding the contract method 0xc4e3ed93.
//
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, misc bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address)
func (_Contract *ContractSession) GetInfo(id [8]byte) (struct {
	Id1         [32]byte
	Id2         [32]byte
	Misc        [32]byte
	PreId       [8]byte
	NextId      [8]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// GetInfo is a free data retrieval call binding the contract method 0xc4e3ed93.
//
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, misc bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address)
func (_Contract *ContractCallerSession) GetInfo(id [8]byte) (struct {
	Id1         [32]byte
	Id2         [32]byte
	Misc        [32]byte
	PreId       [8]byte
	NextId      [8]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() constant returns(address)
func (_Contract *ContractCaller) GovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "governanceAddress")
	return *ret0, err
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() constant returns(address)
func (_Contract *ContractSession) GovernanceAddress() (common.Address, error) {
	return _Contract.Contract.GovernanceAddress(&_Contract.CallOpts)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() constant returns(address)
func (_Contract *ContractCallerSession) GovernanceAddress() (common.Address, error) {
	return _Contract.Contract.GovernanceAddress(&_Contract.CallOpts)
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(id bytes8) constant returns(bool)
func (_Contract *ContractCaller) Has(opts *bind.CallOpts, id [8]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "has", id)
	return *ret0, err
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(id bytes8) constant returns(bool)
func (_Contract *ContractSession) Has(id [8]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(id bytes8) constant returns(bool)
func (_Contract *ContractCallerSession) Has(id [8]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractCaller) LastId(opts *bind.CallOpts) ([8]byte, error) {
	var (
		ret0 = new([8]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastId")
	return *ret0, err
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractSession) LastId() ([8]byte, error) {
	return _Contract.Contract.LastId(&_Contract.CallOpts)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractCallerSession) LastId() ([8]byte, error) {
	return _Contract.Contract.LastId(&_Contract.CallOpts)
}

// CreateGovernanceAddressVote is a paid mutator transaction binding the contract method 0xf834f524.
//
// Solidity: function createGovernanceAddressVote(addr address) returns()
func (_Contract *ContractTransactor) CreateGovernanceAddressVote(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "createGovernanceAddressVote", addr)
}

// CreateGovernanceAddressVote is a paid mutator transaction binding the contract method 0xf834f524.
//
// Solidity: function createGovernanceAddressVote(addr address) returns()
func (_Contract *ContractSession) CreateGovernanceAddressVote(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.CreateGovernanceAddressVote(&_Contract.TransactOpts, addr)
}

// CreateGovernanceAddressVote is a paid mutator transaction binding the contract method 0xf834f524.
//
// Solidity: function createGovernanceAddressVote(addr address) returns()
func (_Contract *ContractTransactorSession) CreateGovernanceAddressVote(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.CreateGovernanceAddressVote(&_Contract.TransactOpts, addr)
}

// Register is a paid mutator transaction binding the contract method 0x4da274fd.
//
// Solidity: function register(id1 bytes32, id2 bytes32, misc bytes32) returns()
func (_Contract *ContractTransactor) Register(opts *bind.TransactOpts, id1 [32]byte, id2 [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register", id1, id2, misc)
}

// Register is a paid mutator transaction binding the contract method 0x4da274fd.
//
// Solidity: function register(id1 bytes32, id2 bytes32, misc bytes32) returns()
func (_Contract *ContractSession) Register(id1 [32]byte, id2 [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2, misc)
}

// Register is a paid mutator transaction binding the contract method 0x4da274fd.
//
// Solidity: function register(id1 bytes32, id2 bytes32, misc bytes32) returns()
func (_Contract *ContractTransactorSession) Register(id1 [32]byte, id2 [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2, misc)
}

// VoteForGovernanceAddress is a paid mutator transaction binding the contract method 0xe8c74af2.
//
// Solidity: function voteForGovernanceAddress(addr address) returns()
func (_Contract *ContractTransactor) VoteForGovernanceAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "voteForGovernanceAddress", addr)
}

// VoteForGovernanceAddress is a paid mutator transaction binding the contract method 0xe8c74af2.
//
// Solidity: function voteForGovernanceAddress(addr address) returns()
func (_Contract *ContractSession) VoteForGovernanceAddress(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.VoteForGovernanceAddress(&_Contract.TransactOpts, addr)
}

// VoteForGovernanceAddress is a paid mutator transaction binding the contract method 0xe8c74af2.
//
// Solidity: function voteForGovernanceAddress(addr address) returns()
func (_Contract *ContractTransactorSession) VoteForGovernanceAddress(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.VoteForGovernanceAddress(&_Contract.TransactOpts, addr)
}

// ContractJoinIterator is returned from FilterJoin and is used to iterate over the raw logs and unpacked data for Join events raised by the Contract contract.
type ContractJoinIterator struct {
	Event *ContractJoin // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractJoinIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractJoin)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractJoin)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractJoinIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractJoinIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractJoin represents a Join event raised by the Contract contract.
type ContractJoin struct {
	Id   [8]byte
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterJoin is a free log retrieval operation binding the contract event 0xf19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389.
//
// Solidity: e join(id bytes8, addr address)
func (_Contract *ContractFilterer) FilterJoin(opts *bind.FilterOpts) (*ContractJoinIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return &ContractJoinIterator{contract: _Contract.contract, event: "join", logs: logs, sub: sub}, nil
}

// WatchJoin is a free log subscription operation binding the contract event 0xf19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389.
//
// Solidity: e join(id bytes8, addr address)
func (_Contract *ContractFilterer) WatchJoin(opts *bind.WatchOpts, sink chan<- *ContractJoin) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractJoin)
				if err := _Contract.contract.UnpackLog(event, "join", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ContractQuitIterator is returned from FilterQuit and is used to iterate over the raw logs and unpacked data for Quit events raised by the Contract contract.
type ContractQuitIterator struct {
	Event *ContractQuit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractQuitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractQuit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractQuit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractQuitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractQuitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractQuit represents a Quit event raised by the Contract contract.
type ContractQuit struct {
	Id   [8]byte
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterQuit is a free log retrieval operation binding the contract event 0x86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd.
//
// Solidity: e quit(id bytes8, addr address)
func (_Contract *ContractFilterer) FilterQuit(opts *bind.FilterOpts) (*ContractQuitIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return &ContractQuitIterator{contract: _Contract.contract, event: "quit", logs: logs, sub: sub}, nil
}

// WatchQuit is a free log subscription operation binding the contract event 0x86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd.
//
// Solidity: e quit(id bytes8, addr address)
func (_Contract *ContractFilterer) WatchQuit(opts *bind.WatchOpts, sink chan<- *ContractQuit) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractQuit)
				if err := _Contract.contract.UnpackLog(event, "quit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
