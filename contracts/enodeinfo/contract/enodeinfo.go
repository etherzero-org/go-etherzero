// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	ethereum "github.com/etherzero/go-etherzero"
	"github.com/etherzero/go-etherzero/accounts/abi"
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"MasterAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"name\":\"Enodes\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getSingleEnode\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"name\":\"register\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `6080604052600a600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561005257600080fd5b50610ca9806100626000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806301ec4aed1461008857806306661abd146100df57806320c148051461010a578063515e7e09146101d5578063a87d942c146102a0578063c0e64821146102cb578063c1292cc31461031e575b600080fd5b34801561009457600080fd5b5061009d61037f565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156100eb57600080fd5b506100f46103a5565b6040518082815260200191505060405180910390f35b34801561011657600080fd5b50610150600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff191690602001909291905050506103ab565b60405180856000191660001916815260200184600019166000191681526020018377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018267ffffffffffffffff1667ffffffffffffffff16815260200194505050505060405180910390f35b3480156101e157600080fd5b5061021b600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050610414565b60405180856000191660001916815260200184600019166000191681526020018377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018267ffffffffffffffff1667ffffffffffffffff16815260200194505050505060405180910390f35b3480156102ac57600080fd5b506102b56105f3565b6040518082815260200191505060405180910390f35b3480156102d757600080fd5b5061031c60048036038101908080356000191690602001909291908035600019169060200190929190803567ffffffffffffffff1690602001909291905050506105fd565b005b34801561032a57600080fd5b50610333610c0d565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60025481565b60006020528060005260406000206000915090508060000154908060010154908060020160009054906101000a9004780100000000000000000000000000000000000000000000000002908060020160089054906101000a900467ffffffffffffffff16905084565b600080600080600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168577ffffffffffffffffffffffffffffffffffffffffffffffff19161415151561047b57600080fd5b6000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015493506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206001015492506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a900478010000000000000000000000000000000000000000000000000291506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a900467ffffffffffffffff1690509193509193565b6000600254905090565b610605610c38565b61060d610c5a565b6000806000803414801561062d5750600060010260001916886000191614155b80156106455750600060010260001916876000191614155b80156106665750600067ffffffffffffffff168667ffffffffffffffff1614155b151561067157600080fd5b8785600060028110151561068157fe5b60200201906000191690816000191681525050868560016002811015156106a457fe5b602002019060001916908160001916815250506020846080876000600b600019f115156106d057600080fd5b8360006001811015156106df57fe5b6020020151600190049250600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415801561075257508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b151561075d57600080fd5b879050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff1916141515156107c157600080fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166316e7f171826040518263ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001915050602060405180830381600087803b15801561088857600080fd5b505af115801561089c573d6000803e3d6000fd5b505050506040513d60208110156108b257600080fd5b81019080805190602001909291905050509150600115158215151415156108d857600080fd5b6000600102600019166000808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001546000191614156109495760016002600082825401925050819055505b6080604051908101604052808960001916815260200188600019168152602001600360009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018767ffffffffffffffff168152506000808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908367ffffffffffffffff160217905550905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600360009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916141515610bc85780600080600360009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b80600360006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505050505050505050565b600360009054906101000a900478010000000000000000000000000000000000000000000000000281565b6040805190810160405280600290602082028038833980820191505090505090565b6020604051908101604052806001906020820280388339808201915050905050905600a165627a7a723058205ffb14cb3e4ab04e50b1395180a28bab443224937449661534cfaf4d71f70c450029`

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

// Enodes is a free data retrieval call binding the contract method 0x20c14805.
//
// Solidity: function Enodes( bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractCaller) Enodes(opts *bind.CallOpts, arg0 [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	ret := new(struct {
		Id1    [32]byte
		Id2    [32]byte
		NextId [8]byte
		Ipport uint64
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "Enodes", arg0)
	return *ret, err
}

// Enodes is a free data retrieval call binding the contract method 0x20c14805.
//
// Solidity: function Enodes( bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractSession) Enodes(arg0 [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	return _Contract.Contract.Enodes(&_Contract.CallOpts, arg0)
}

// Enodes is a free data retrieval call binding the contract method 0x20c14805.
//
// Solidity: function Enodes( bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractCallerSession) Enodes(arg0 [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	return _Contract.Contract.Enodes(&_Contract.CallOpts, arg0)
}

// MasterAddr is a free data retrieval call binding the contract method 0x01ec4aed.
//
// Solidity: function MasterAddr() constant returns(address)
func (_Contract *ContractCaller) MasterAddr(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "MasterAddr")
	return *ret0, err
}

// MasterAddr is a free data retrieval call binding the contract method 0x01ec4aed.
//
// Solidity: function MasterAddr() constant returns(address)
func (_Contract *ContractSession) MasterAddr() (common.Address, error) {
	return _Contract.Contract.MasterAddr(&_Contract.CallOpts)
}

// MasterAddr is a free data retrieval call binding the contract method 0x01ec4aed.
//
// Solidity: function MasterAddr() constant returns(address)
func (_Contract *ContractCallerSession) MasterAddr() (common.Address, error) {
	return _Contract.Contract.MasterAddr(&_Contract.CallOpts)
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

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() constant returns(uint256)
func (_Contract *ContractCaller) GetCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getCount")
	return *ret0, err
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() constant returns(uint256)
func (_Contract *ContractSession) GetCount() (*big.Int, error) {
	return _Contract.Contract.GetCount(&_Contract.CallOpts)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() constant returns(uint256)
func (_Contract *ContractCallerSession) GetCount() (*big.Int, error) {
	return _Contract.Contract.GetCount(&_Contract.CallOpts)
}

// GetSingleEnode is a free data retrieval call binding the contract method 0x515e7e09.
//
// Solidity: function getSingleEnode(id bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractCaller) GetSingleEnode(opts *bind.CallOpts, id [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	ret := new(struct {
		Id1    [32]byte
		Id2    [32]byte
		NextId [8]byte
		Ipport uint64
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getSingleEnode", id)
	return *ret, err
}

// GetSingleEnode is a free data retrieval call binding the contract method 0x515e7e09.
//
// Solidity: function getSingleEnode(id bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractSession) GetSingleEnode(id [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	return _Contract.Contract.GetSingleEnode(&_Contract.CallOpts, id)
}

// GetSingleEnode is a free data retrieval call binding the contract method 0x515e7e09.
//
// Solidity: function getSingleEnode(id bytes8) constant returns(id1 bytes32, id2 bytes32, nextId bytes8, ipport uint64)
func (_Contract *ContractCallerSession) GetSingleEnode(id [8]byte) (struct {
	Id1    [32]byte
	Id2    [32]byte
	NextId [8]byte
	Ipport uint64
}, error) {
	return _Contract.Contract.GetSingleEnode(&_Contract.CallOpts, id)
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

// Register is a paid mutator transaction binding the contract method 0xc0e64821.
//
// Solidity: function register(id1 bytes32, id2 bytes32, ipport uint64) returns()
func (_Contract *ContractTransactor) Register(opts *bind.TransactOpts, id1 [32]byte, id2 [32]byte, ipport uint64) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register", id1, id2, ipport)
}

// Register is a paid mutator transaction binding the contract method 0xc0e64821.
//
// Solidity: function register(id1 bytes32, id2 bytes32, ipport uint64) returns()
func (_Contract *ContractSession) Register(id1 [32]byte, id2 [32]byte, ipport uint64) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2, ipport)
}

// Register is a paid mutator transaction binding the contract method 0xc0e64821.
//
// Solidity: function register(id1 bytes32, id2 bytes32, ipport uint64) returns()
func (_Contract *ContractTransactorSession) Register(id1 [32]byte, id2 [32]byte, ipport uint64) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2, ipport)
}
