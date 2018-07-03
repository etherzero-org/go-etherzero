// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"github.com/etherzero/go-etherzero/accounts/abi"
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/event"
	"math/big"
	"strings"

	"github.com/etherzero/go-etherzero"
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"quit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"join\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzPerNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getId\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"misc\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes8\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"account\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `0x608060405234801561001057600080fd5b5060007801000000000000000000000000000000000000000000000000026000806101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600060018190555061217b8061007e6000396000f3006080604052600436106100c5576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd1461091b57806316e7f171146109465780633aa8cd8b146109a657806365f68c8914610a06578063795053d314610a93578063c1292cc314610aea578063c4e3ed9314610b4b578063c808021c14610c8f578063d81a655b14610cba578063e3596ce014610d29578063e8c74af214610d54578063f834f52414610d97578063ff5ecad214610dcd575b600080600080600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029350600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015492506000341480156101e357508377ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156101fb5750826000191660006001026000191614155b80156102305750662386f26fc100006801158e460913d00000033073ffffffffffffffffffffffffffffffffffffffff163110155b801561023e57506000600154115b151561024957600080fd5b600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a90047801000000000000000000000000000000000000000000000000029150600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160089054906101000a90047801000000000000000000000000000000000000000000000000029050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168277ffffffffffffffffffffffffffffffffffffffffffffffff19161415156104125780600660008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff19161415156104f75781600660008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550610532565b816000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b61012060405190810160405280600060010260001916815260200160006001026000191681526020016000600102600019168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200160008152602001600073ffffffffffffffffffffffffffffffffffffffff168152602001600081526020016000815250600660008677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060008201518160000190600019169055602082015181600101906000191690556040820151816002019060001916905560608201518160030160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060a0820151816004015560c08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060e0820151816006015561010082015181600701559050506000780100000000000000000000000000000000000000000000000002600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd8433604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100006801158e460913d00000039081150290604051600060405180830381858888f19350505050158015610914573d6000803e3d6000fd5b5050505050005b34801561092757600080fd5b50610930610df8565b6040518082815260200191505060405180910390f35b34801561095257600080fd5b5061098c600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050610dfe565b604051808215151515815260200191505060405180910390f35b610a04600480360381019080803560001916906020019092919080356000191690602001909291908035600019169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610e62565b005b348015610a1257600080fd5b50610a47600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061156e565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610a9f57600080fd5b50610aa86115dc565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b348015610af657600080fd5b50610aff611602565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610b5757600080fd5b50610b91600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff1916906020019092919050505061162c565b604051808a60001916600019168152602001896000191660001916815260200188600019166000191681526020018777ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001838152602001828152602001995050505050505050505060405180910390f35b348015610c9b57600080fd5b50610ca4611978565b6040518082815260200191505060405180910390f35b348015610cc657600080fd5b50610d0f60048036038101908080359060200190929190803560001916906020019092919080356000191690602001909291908035600019169060200190929190505050611983565b604051808215151515815260200191505060405180910390f35b348015610d3557600080fd5b50610d3e611cf9565b6040518082815260200191505060405180910390f35b348015610d6057600080fd5b50610d95600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611cff565b005b610dcb600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611f79565b005b348015610dd957600080fd5b50610de26120fc565b6040518082815260200191505060405180910390f35b60015481565b60008060010260001916600660008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001546000191614159050919050565b6000849050846000191660006001026000191614158015610e8f5750836000191660006001026000191614155b8015610ea75750826000191660006001026000191614155b8015610f0557508077ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b8015610fc65750600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916145b80156110295750600660008277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015460001916600060010260001916145b801561103d57506801158e460913d0000034145b151561104857600080fd5b80600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550610120604051908101604052808660001916815260200185600019168152602001846000191681526020016000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020014381526020013373ffffffffffffffffffffffffffffffffffffffff168152602001600081526020016000815250600660008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060008201518160000190600019169055602082015181600101906000191690556040820151816002019060001916905560608201518160030160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060a0820151816004015560c08201518160050160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060e082015181600601556101008201518160070155905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19166000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614151561142e5780600660008060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b806000806101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825401925050819055508173ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100009081150290604051600060405180830381858888f193505050501580156114c5573d6000803e3d6000fd5b507ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac3898133604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a15050505050565b6000600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029050919050565b600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900478010000000000000000000000000000000000000000000000000281565b6000806000806000806000806000600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001549850600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600101549750600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600201549650600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a90047801000000000000000000000000000000000000000000000000029550600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160089054906101000a90047801000000000000000000000000000000000000000000000000029450600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600401549350600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060050160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169250600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601549150600660008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206007015490509193959799909294969850565b662386f26fc1000081565b600061198d612109565b61199561212c565b60008060008943101580156119b9575060026101688115156119b357fe5b048a4303105b15156119c457600080fd5b89408560006004811015156119d557fe5b60200201906000191690816000191681525050888560016004811015156119f857fe5b6020020190600019169081600019168152505087856002600481101515611a1b57fe5b6020020190600019169081600019168152505086856003600481101515611a3e57fe5b6020020190600019169081600019168152505060208460808760006009600019f11515611a6a57600080fd5b836000600181101515611a7957fe5b60200201519250611a8983610dfe565b1515611a9457600080fd5b600660008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060070154915043600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600701819055506000821115611c1e578143039050610168811115611ba3576000600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060060181905550611c1d565b6002610168811515611bb157fe5b04811015611bc25760009550611cec565b80600660008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601600082825401925050819055505b5b7f117a9c2fecedc1787965b992eb8230aac559e7add86d4d9e1897540dd4ee037a83600660008677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206006015443604051808477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001838152602001828152602001935050505060405180910390a1600195505b5050505050949350505050565b61016881565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905060008160010154118015611db05750600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916611d923361156e565b77ffffffffffffffffffffffffffffffffffffffffffffffff191614155b8015611dc0575060008160020154145b8015611e59575060001515600360008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515145b1515611e6457600080fd5b6001600360008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600181600001600082825401925050819055506064604260015402811515611f1d57fe5b048160000154101515611f755743816002018190555081600260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b5050565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000015414801561200d57506000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060010154145b151561201857600080fd5b60806040519081016040528060008152602001438152602001600081526020013373ffffffffffffffffffffffffffffffffffffffff16815250600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008201518160000155602082015181600101556040820151816002015560608201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555090505050565b6801158e460913d0000081565b608060405190810160405280600490602082028038833980820191505090505090565b6020604051908101604052806001906020820280388339808201915050905050905600a165627a7a7230582043887d46514194ca57996f3aeee442cadd0d6972750edde133140f40671b0df00029`

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
