// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	"github.com/ethzero/go-ethzero"
	"github.com/ethzero/go-ethzero/accounts/abi"
	"github.com/ethzero/go-ethzero/accounts/abi/bind"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/event"
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"quit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"join\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[],\"name\":\"MasterNode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzPerNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getId\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes8\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"account\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `0x608060405234801561001057600080fd5b5061165e806100206000396000f30060806040526004361061008e576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd146108ae57806316e7f171146108d95780634da274fd1461093957806365f68c8914610979578063a9edf68e14610a06578063c1292cc314610a1d578063c4e3ed9314610a7e578063ff5ecad214610bb4575b600080600080600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029350600360008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015492506000341480156101ac57508377ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156101c45750826000191660006001026000191614155b80156101f057506801158e460913d000003073ffffffffffffffffffffffffffffffffffffffff163110155b80156101fe57506000600154115b151561020957600080fd5b600360008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a90047801000000000000000000000000000000000000000000000000029150600360008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160089054906101000a90047801000000000000000000000000000000000000000000000000029050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168277ffffffffffffffffffffffffffffffffffffffffffffffff19161415156103d25780600360008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff19161415156104b75781600360008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506104f2565b816000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b60e060405190810160405280600060010260001916815260200160006001026000191681526020016000600102600019168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815250600360008677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060008201518160000190600019169055602082015181600101906000191690556040820151816002019060001916905560608201518160030160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060a0820151816004015560c08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055509050506000780100000000000000000000000000000000000000000000000002600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd8433604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc6801158e460913d000009081150290604051600060405180830381858888f193505050501580156108a7573d6000803e3d6000fd5b5050505050005b3480156108ba57600080fd5b506108c3610bdf565b6040518082815260200191505060405180910390f35b3480156108e557600080fd5b5061091f600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050610be5565b604051808215151515815260200191505060405180910390f35b610977600480360381019080803560001916906020019092919080356000191690602001909291908035600019169060200190929190505050610c49565b005b34801561098557600080fd5b506109ba600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611284565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610a1257600080fd5b50610a1b6112f2565b005b348015610a2957600080fd5b50610a32611352565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610a8a57600080fd5b50610ac4600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff1916906020019092919050505061137c565b604051808860001916600019168152602001876000191660001916815260200186600019166000191681526020018577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018381526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200197505050505050505060405180910390f35b348015610bc057600080fd5b50610bc9611625565b6040518082815260200191505060405180910390f35b60015481565b60008060010260001916600360008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001546000191614159050919050565b6000839050836000191660006001026000191614158015610c765750826000191660006001026000191614155b8015610c8e5750816000191660006001026000191614155b8015610d4f5750600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916145b8015610db25750600360008277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015460001916600060010260001916145b8015610dc657506801158e460913d0000034145b1515610dd157600080fd5b80600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060e0604051908101604052808560001916815260200184600019168152602001836000191681526020016000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020014381526020013373ffffffffffffffffffffffffffffffffffffffff16815250600360008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060008201518160000190600019169055602082015181600101906000191690556040820151816002019060001916905560608201518160030160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060a0820151816004015560c08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19166000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19161415156111935780600360008060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b806000806101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825401925050819055507ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac3898133604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a150505050565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029050919050565b60007801000000000000000000000000000000000000000000000000026000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506000600181905550565b6000809054906101000a900478010000000000000000000000000000000000000000000000000281565b6000806000806000806000600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001549650600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600101549550600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600201549450600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a90047801000000000000000000000000000000000000000000000000029350600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160089054906101000a90047801000000000000000000000000000000000000000000000000029250600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600401549150600360008977ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050919395979092949650565b6801158e460913d00000815600a165627a7a723058200003d6d3d350f74c3d808417941e122e51cfe57b937e66b1b9502889830b04990029`

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

// MasterNode is a paid mutator transaction binding the contract method 0xa9edf68e.
//
// Solidity: function MasterNode() returns()
func (_Contract *ContractTransactor) MasterNode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "MasterNode")
}

// MasterNode is a paid mutator transaction binding the contract method 0xa9edf68e.
//
// Solidity: function MasterNode() returns()
func (_Contract *ContractSession) MasterNode() (*types.Transaction, error) {
	return _Contract.Contract.MasterNode(&_Contract.TransactOpts)
}

// MasterNode is a paid mutator transaction binding the contract method 0xa9edf68e.
//
// Solidity: function MasterNode() returns()
func (_Contract *ContractTransactorSession) MasterNode() (*types.Transaction, error) {
	return _Contract.Contract.MasterNode(&_Contract.TransactOpts)
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
// Solidity: event join(id bytes8, addr address)
func (_Contract *ContractFilterer) FilterJoin(opts *bind.FilterOpts) (*ContractJoinIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return &ContractJoinIterator{contract: _Contract.contract, event: "join", logs: logs, sub: sub}, nil
}

// WatchJoin is a free log subscription operation binding the contract event 0xf19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389.
//
// Solidity: event join(id bytes8, addr address)
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
// Solidity: event quit(id bytes8, addr address)
func (_Contract *ContractFilterer) FilterQuit(opts *bind.FilterOpts) (*ContractQuitIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return &ContractQuitIterator{contract: _Contract.contract, event: "quit", logs: logs, sub: sub}, nil
}

// WatchQuit is a free log subscription operation binding the contract event 0x86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd.
//
// Solidity: event quit(id bytes8, addr address)
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
