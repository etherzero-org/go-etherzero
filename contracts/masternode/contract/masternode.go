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
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\"},{\"name\":\"subId\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getId\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"subId\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes32\"},{\"name\":\"nextId\",\"type\":\"bytes32\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"account\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"MasterNode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzPerNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"join\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"quit\",\"type\":\"event\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `0x608060405234801561001057600080fd5b50610adc806100206000396000f30060806040526004361061008e576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd146103f15780634da274fd1461041c5780637a02dc061461045c57806399f826a514610510578063a9edf68e1461056f578063b7dc560f14610586578063e31a9d92146105cf578063ff5ecad214610602575b6000806000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205492506000341480156100f15750826000191660006001026000191614155b80156101275750600360008460001916600019168152602001908152602001600020600001546000191660006001026000191614155b801561015357506801158e460913d000003073ffffffffffffffffffffffffffffffffffffffff163110155b801561016157506000600154115b151561016c57600080fd5b60036000846000191660001916815260200190815260200160002060020154915060036000846000191660001916815260200190815260200160002060030154905060006001026000191682600019161415156101eb578060036000846000191660001916815260200190815260200160002060030181600019169055505b600060010260001916816000191614151561022c57816003600083600019166000191681526020019081526020016000206002018160001916905550610238565b81600081600019169055505b60c060405190810160405280600060010260001916815260200160006001026000191681526020016000600102600019168152602001600060010260001916815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff16815250600360008560001916600019168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020190600019169055606082015181600301906000191690556080820151816004015560a08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550905050600180600082825403925050819055507f23063f0f5a69261f0edb82655de7f0eb5fe0fe458195c7ed1f510575d3547d1d8360405180826000191660001916815260200191505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc6801158e460913d000009081150290604051600060405180830381858888f193505050501580156103eb573d6000803e3d6000fd5b50505050005b3480156103fd57600080fd5b5061040661062d565b6040518082815260200191505060405180910390f35b61045a600480360381019080803560001916906020019092919080356000191690602001909291908035600019169060200190929190505050610633565b005b34801561046857600080fd5b5061048b600480360381019080803560001916906020019092919050505061090d565b6040518087600019166000191681526020018660001916600019168152602001856000191660001916815260200184600019166000191681526020018381526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001965050505050505060405180910390f35b34801561051c57600080fd5b50610551600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610a05565b60405180826000191660001916815260200191505060405180910390f35b34801561057b57600080fd5b50610584610a4e565b005b34801561059257600080fd5b506105b56004803603810190808035600019169060200190929190505050610a67565b604051808215151515815260200191505060405180910390f35b3480156105db57600080fd5b506105e4610a9d565b60405180826000191660001916815260200191505060405180910390f35b34801561060e57600080fd5b50610617610aa3565b6040518082815260200191505060405180910390f35b60015481565b82600019166000600102600019161415801561065b5750816000191660006001026000191614155b80156106735750806000191660006001026000191614155b80156106c95750600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205460001916600060010260001916145b80156106fe57506003600084600019166000191681526020019081526020016000206000015460001916600060010260001916145b801561071257506801158e460913d0000034145b151561071d57600080fd5b82600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020816000191690555060c060405190810160405280836000191681526020018260001916815260200160005460001916815260200160006001026000191681526020014381526020013373ffffffffffffffffffffffffffffffffffffffff16815250600360008560001916600019168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020190600019169055606082015181600301906000191690556080820151816004015560a08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550905050600060010260001916600054600019161415156108ae57826003600080546000191660001916815260200190815260200160002060030181600019169055505b8260008160001916905550600180600082825401925050819055507fad677d0bdb16d7a303768a7f4dbe2a5cb5583e29935c11dac729f7e2eb91ef9f8360405180826000191660001916815260200191505060405180910390a1505050565b60008060008060008060036000886000191660001916815260200190815260200160002060000154955060036000886000191660001916815260200190815260200160002060010154945060036000886000191660001916815260200190815260200160002060020154935060036000886000191660001916815260200190815260200160002060030154925060036000886000191660001916815260200190815260200160002060040154915060036000886000191660001916815260200190815260200160002060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905091939550919395565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000600102600081600019169055506000600181905550565b60008060010260001916600360008460001916600019168152602001908152602001600020600001546000191614159050919050565b60005481565b6801158e460913d00000815600a165627a7a72305820b7e128ac3d30742e14daca78fcf1426e35dc94dce82bee36ddd6aef307e67a2e0029`

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
// Solidity: function getId(addr address) constant returns(id bytes32)
func (_Contract *ContractCaller) GetId(opts *bind.CallOpts, addr common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getId", addr)
	return *ret0, err
}

// GetId is a free data retrieval call binding the contract method 0x65f68c89.
//
// Solidity: function getId(addr address) constant returns(id bytes32)
func (_Contract *ContractSession) GetId(addr common.Address) ([32]byte, error) {
	return _Contract.Contract.GetId(&_Contract.CallOpts, addr)
}

// GetId is a free data retrieval call binding the contract method 0x65f68c89.
//
// Solidity: function getId(addr address) constant returns(id bytes32)
func (_Contract *ContractCallerSession) GetId(addr common.Address) ([32]byte, error) {
	return _Contract.Contract.GetId(&_Contract.CallOpts, addr)
}

// GetInfo is a free data retrieval call binding the contract method 0x7a02dc06.
//
// Solidity: function getInfo(id bytes32) constant returns(subId bytes32, misc bytes32, preId bytes32, nextId bytes32, blockNumber uint256, account address)
func (_Contract *ContractCaller) GetInfo(opts *bind.CallOpts, id [32]byte) (struct {
	SubId       [32]byte
	Misc        [32]byte
	PreId       [32]byte
	NextId      [32]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	ret := new(struct {
		SubId       [32]byte
		Misc        [32]byte
		PreId       [32]byte
		NextId      [32]byte
		BlockNumber *big.Int
		Account     common.Address
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getInfo", id)
	return *ret, err
}

// GetInfo is a free data retrieval call binding the contract method 0x7a02dc06.
//
// Solidity: function getInfo(id bytes32) constant returns(subId bytes32, misc bytes32, preId bytes32, nextId bytes32, blockNumber uint256, account address)
func (_Contract *ContractSession) GetInfo(id [32]byte) (struct {
	SubId       [32]byte
	Misc        [32]byte
	PreId       [32]byte
	NextId      [32]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// GetInfo is a free data retrieval call binding the contract method 0x7a02dc06.
//
// Solidity: function getInfo(id bytes32) constant returns(subId bytes32, misc bytes32, preId bytes32, nextId bytes32, blockNumber uint256, account address)
func (_Contract *ContractCallerSession) GetInfo(id [32]byte) (struct {
	SubId       [32]byte
	Misc        [32]byte
	PreId       [32]byte
	NextId      [32]byte
	BlockNumber *big.Int
	Account     common.Address
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// Has is a free data retrieval call binding the contract method 0xb7dc560f.
//
// Solidity: function has(id bytes32) constant returns(bool)
func (_Contract *ContractCaller) Has(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "has", id)
	return *ret0, err
}

// Has is a free data retrieval call binding the contract method 0xb7dc560f.
//
// Solidity: function has(id bytes32) constant returns(bool)
func (_Contract *ContractSession) Has(id [32]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// Has is a free data retrieval call binding the contract method 0xb7dc560f.
//
// Solidity: function has(id bytes32) constant returns(bool)
func (_Contract *ContractCallerSession) Has(id [32]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes32)
func (_Contract *ContractCaller) LastId(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastId")
	return *ret0, err
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes32)
func (_Contract *ContractSession) LastId() ([32]byte, error) {
	return _Contract.Contract.LastId(&_Contract.CallOpts)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes32)
func (_Contract *ContractCallerSession) LastId() ([32]byte, error) {
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
// Solidity: function register(id bytes32, subId bytes32, misc bytes32) returns()
func (_Contract *ContractTransactor) Register(opts *bind.TransactOpts, id [32]byte, subId [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register", id, subId, misc)
}

// Register is a paid mutator transaction binding the contract method 0x4da274fd.
//
// Solidity: function register(id bytes32, subId bytes32, misc bytes32) returns()
func (_Contract *ContractSession) Register(id [32]byte, subId [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id, subId, misc)
}

// Register is a paid mutator transaction binding the contract method 0x4da274fd.
//
// Solidity: function register(id bytes32, subId bytes32, misc bytes32) returns()
func (_Contract *ContractTransactorSession) Register(id [32]byte, subId [32]byte, misc [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id, subId, misc)
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
	Id  [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterJoin is a free log retrieval operation binding the contract event 0xad677d0bdb16d7a303768a7f4dbe2a5cb5583e29935c11dac729f7e2eb91ef9f.
//
// Solidity: event join(id bytes32)
func (_Contract *ContractFilterer) FilterJoin(opts *bind.FilterOpts) (*ContractJoinIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return &ContractJoinIterator{contract: _Contract.contract, event: "join", logs: logs, sub: sub}, nil
}

// WatchJoin is a free log subscription operation binding the contract event 0xad677d0bdb16d7a303768a7f4dbe2a5cb5583e29935c11dac729f7e2eb91ef9f.
//
// Solidity: event join(id bytes32)
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
				if err := _Contract.contract.UnpackLog(&event.Id, "join", log); err != nil {
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
	Id  [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterQuit is a free log retrieval operation binding the contract event 0x23063f0f5a69261f0edb82655de7f0eb5fe0fe458195c7ed1f510575d3547d1d.
//
// Solidity: event quit(id bytes32)
func (_Contract *ContractFilterer) FilterQuit(opts *bind.FilterOpts) (*ContractQuitIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return &ContractQuitIterator{contract: _Contract.contract, event: "quit", logs: logs, sub: sub}, nil
}

// WatchQuit is a free log subscription operation binding the contract event 0x23063f0f5a69261f0edb82655de7f0eb5fe0fe458195c7ed1f510575d3547d1d.
//
// Solidity: event quit(id bytes32)
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
				if err := _Contract.contract.UnpackLog(&event.Id, "quit", log); err != nil {
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
