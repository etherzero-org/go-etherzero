// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/etherzero/go-etherzero/accounts/abi"
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"name\":\"Enodes\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"masteraddr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getSingleEnode\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"ipport\",\"type\":\"uint64\"}],\"name\":\"register\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `608060405234801561001057600080fd5b50610c66806100206000396000f300608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd1461008857806320c14805146100b35780633b11c4c11461017e578063515e7e09146101d5578063a87d942c146102a0578063c0e64821146102cb578063c1292cc31461031e575b600080fd5b34801561009457600080fd5b5061009d61037f565b6040518082815260200191505060405180910390f35b3480156100bf57600080fd5b506100f9600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050610385565b60405180856000191660001916815260200184600019166000191681526020018377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018267ffffffffffffffff1667ffffffffffffffff16815260200194505050505060405180910390f35b34801561018a57600080fd5b506101936103ee565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156101e157600080fd5b5061021b600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff191690602001909291905050506103f3565b60405180856000191660001916815260200184600019166000191681526020018377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018267ffffffffffffffff1667ffffffffffffffff16815260200194505050505060405180910390f35b3480156102ac57600080fd5b506102b56105d2565b6040518082815260200191505060405180910390f35b3480156102d757600080fd5b5061031c60048036038101908080356000191690602001909291908035600019169060200190929190803567ffffffffffffffff1690602001909291905050506105dc565b005b34801561032a57600080fd5b50610333610bca565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b60015481565b60006020528060005260406000206000915090508060000154908060010154908060020160009054906101000a9004780100000000000000000000000000000000000000000000000002908060020160089054906101000a900467ffffffffffffffff16905084565b600a81565b600080600080600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168577ffffffffffffffffffffffffffffffffffffffffffffffff19161415151561045a57600080fd5b6000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015493506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206001015492506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a900478010000000000000000000000000000000000000000000000000291506000808677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a900467ffffffffffffffff1690509193509193565b6000600154905090565b6105e4610bf5565b6105ec610c17565b6000806000803414801561060c5750600060010260001916886000191614155b80156106245750600060010260001916876000191614155b80156106455750600067ffffffffffffffff168667ffffffffffffffff1614155b151561065057600080fd5b8785600060028110151561066057fe5b602002019060001916908160001916815250508685600160028110151561068357fe5b602002019060001916908160001916815250506020846080876000600b600019f115156106af57600080fd5b8360006001811015156106be57fe5b6020020151600190049250600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415801561073157508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b151561073c57600080fd5b879050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff1916141515156107a057600080fd5b600a73ffffffffffffffffffffffffffffffffffffffff166316e7f171826040518263ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001915050602060405180830381600087803b15801561084657600080fd5b505af115801561085a573d6000803e3d6000fd5b505050506040513d602081101561087057600080fd5b810190808051906020019092919050505091506001151582151514151561089657600080fd5b6000600102600019166000808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015460001916141561090657600180600082825401925050819055505b6080604051908101604052808960001916815260200188600019168152602001600260009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018767ffffffffffffffff168152506000808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908367ffffffffffffffff160217905550905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600260009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916141515610b855780600080600260009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b80600260006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505050505050505050565b600260009054906101000a900478010000000000000000000000000000000000000000000000000281565b6040805190810160405280600290602082028038833980820191505090505090565b6020604051908101604052806001906020820280388339808201915050905050905600a165627a7a72305820e4654bf1b507a3821b77fc765ba7882572c446980ece2c6b94a92bc1ae7990f70029`

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

// Masteraddr is a free data retrieval call binding the contract method 0x3b11c4c1.
//
// Solidity: function masteraddr() constant returns(address)
func (_Contract *ContractCaller) Masteraddr(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "masteraddr")
	return *ret0, err
}

// Masteraddr is a free data retrieval call binding the contract method 0x3b11c4c1.
//
// Solidity: function masteraddr() constant returns(address)
func (_Contract *ContractSession) Masteraddr() (common.Address, error) {
	return _Contract.Contract.Masteraddr(&_Contract.CallOpts)
}

// Masteraddr is a free data retrieval call binding the contract method 0x3b11c4c1.
//
// Solidity: function masteraddr() constant returns(address)
func (_Contract *ContractCallerSession) Masteraddr() (common.Address, error) {
	return _Contract.Contract.Masteraddr(&_Contract.CallOpts)
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
