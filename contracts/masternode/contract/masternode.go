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
const ContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"proposalPeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"proposalAddr\",\"type\":\"address\"},{\"name\":\"voter\",\"type\":\"address\"}],\"name\":\"checkVote\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getId\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"initGovernanceAddress\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"governanceAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"proposalFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes8\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"account\",\"type\":\"address\"},{\"name\":\"blockOnlineAcc\",\"type\":\"uint256\"},{\"name\":\"blockLastPing\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzMin\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"proposalCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getVoteInfo\",\"outputs\":[{\"name\":\"voteCount\",\"type\":\"uint256\"},{\"name\":\"startBlock\",\"type\":\"uint256\"},{\"name\":\"stopBlock\",\"type\":\"uint256\"},{\"name\":\"creator\",\"type\":\"address\"},{\"name\":\"lastAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"blockPingTimeout\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastProposalAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"voteForGovernanceAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"createGovernanceAddressVote\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"etzPerNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"join\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"quit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"blockOnlineAcc\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"blockLastPing\",\"type\":\"uint256\"}],\"name\":\"ping\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"newVote\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"newProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"governanceAddressChange\",\"type\":\"event\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `608060405234801561001057600080fd5b5060007801000000000000000000000000000000000000000000000000026000806101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060006001819055506127048061007e6000396000f3006080604052600436106100e6576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306661abd14610dab57806316e7f17114610dd65780632c103c7914610e365780632f92673214610e6157806365f68c8914610e93578063795053d314610f20578063c1292cc314610f77578063c27cabb514610fd8578063c4e3ed9314611003578063c808021c14611138578063dc1e30da14611163578063e3596ce01461122e578063e7b895b614611259578063e8c74af2146112b0578063f834f524146112f3578063ff5ecad214611329575b6000806000806100f4612693565b6100fc6126b5565b6000806000803414151561010f57600080fd5b600460003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029850600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168977ffffffffffffffffffffffffffffffffffffffffffffffff1916141580156101dd57506101dc89611354565b5b1561041b57600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060060154975060008811156102fc578743039650610e108711156102a0576000600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501819055506102fb565b86600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501600082825401925050819055505b5b43600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601819055507fb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c289600260008c77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206005015443604051808477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001838152602001828152602001935050505060405180910390a1610da0565b600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029850600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060000154955060003414801561053357508877ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b801561054b5750856000191660006001026000191614155b80156105805750662386f26fc100006801158e460913d00000033073ffffffffffffffffffffffffffffffffffffffff163110155b801561058e57506000600154115b151561059957600080fd5b858560006002811015156105a957fe5b60200201906000191690816000191681525050600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206001015485600160028110151561061857fe5b602002019060001916908160001916815250506020846080876000600b600019f1151561064457600080fd5b83600060018110151561065357fe5b60200201516001900492506000780100000000000000000000000000000000000000000000000002600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a90047801000000000000000000000000000000000000000000000000029150600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a90047801000000000000000000000000000000000000000000000000029050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168277ffffffffffffffffffffffffffffffffffffffffffffffff19161415156108bb5780600260008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff19161415156109a05781600260008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506109db565b816000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b6101006040519081016040528060006001026000191681526020016000600102600019168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600073ffffffffffffffffffffffffffffffffffffffff16815260200160008152602001600081526020016000815250600260008b77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060a0820151816004015560c0820151816005015560e082015181600601559050506000780100000000000000000000000000000000000000000000000002600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff0219169083780100000000000000000000000000000000000000000000000090040217905550600180600082825403925050819055507f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd8933604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100006801158e460913d00000039081150290604051600060405180830381858888f19350505050158015610d9e573d6000803e3d6000fd5b505b505050505050505050005b348015610db757600080fd5b50610dc06113b8565b6040518082815260200191505060405180910390f35b348015610de257600080fd5b50610e1c600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050611354565b604051808215151515815260200191505060405180910390f35b348015610e4257600080fd5b50610e4b6113be565b6040518082815260200191505060405180910390f35b610e91600480360381019080803560001916906020019092919080356000191690602001909291905050506113c5565b005b348015610e9f57600080fd5b50610ed4600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611bde565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610f2c57600080fd5b50610f35611c4c565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b348015610f8357600080fd5b50610f8c611c72565b604051808277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b348015610fe457600080fd5b50610fed611c9c565b6040518082815260200191505060405180910390f35b34801561100f57600080fd5b50611049600480360381019080803577ffffffffffffffffffffffffffffffffffffffffffffffff19169060200190929190505050611ca8565b60405180896000191660001916815260200188600019166000191681526020018777ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018381526020018281526020019850505050505050505060405180910390f35b34801561114457600080fd5b5061114d611fa1565b6040518082815260200191505060405180910390f35b34801561116f57600080fd5b506111a4600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611fac565b604051808681526020018581526020018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019550505050505060405180910390f35b34801561123a57600080fd5b50611243612156565b6040518082815260200191505060405180910390f35b34801561126557600080fd5b5061126e61215c565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156112bc57600080fd5b506112f1600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050612182565b005b611327600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050612467565b005b34801561133557600080fd5b5061133e612686565b6040518082815260200191505060405180910390f35b60008060010260001916600260008477ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001546000191614159050919050565b60015481565b62124f8081565b60006113cf612693565b6113d76126b5565b60008593508560001916600060010260001916141580156114045750846000191660006001026000191614155b801561146257508377ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156115235750600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916145b80156115865750600260008577ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020019081526020016000206000015460001916600060010260001916145b801561159a57506801158e460913d0000034145b15156115a557600080fd5b858360006002811015156115b557fe5b60200201906000191690816000191681525050848360016002811015156115d857fe5b602002019060001916908160001916815250506020826080856000600b600019f1151561160457600080fd5b81600060018110151561161357fe5b6020020151600190049050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415151561165a57600080fd5b83600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506101006040519081016040528087600019168152602001866000191681526020016000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191681526020013373ffffffffffffffffffffffffffffffffffffffff168152602001438152602001600081526020016000815250600260008677ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600082015181600001906000191690556020820151816001019060001916905560408201518160020160006101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908378010000000000000000000000000000000000000000000000009004021790555060808201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060a0820151816004015560c0820151816005015560e08201518160060155905050600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19166000809054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff1916141515611a255783600260008060009054906101000a900478010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160086101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055505b836000806101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055506001806000828254019250508190555083600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548167ffffffffffffffff02191690837801000000000000000000000000000000000000000000000000900402179055508073ffffffffffffffffffffffffffffffffffffffff166108fc662386f26fc100009081150290604051600060405180830381858888f19350505050158015611b34573d6000803e3d6000fd5b507ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac3898433604051808377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff191681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a1505050505050565b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a90047801000000000000000000000000000000000000000000000000029050919050565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000809054906101000a900478010000000000000000000000000000000000000000000000000281565b678ac7230489e8000081565b600080600080600080600080600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600001549750600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600101549650600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160009054906101000a90047801000000000000000000000000000000000000000000000000029550600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060020160089054906101000a90047801000000000000000000000000000000000000000000000000029450600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600401549350600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060030160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169250600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600501549150600260008a77ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600601549050919395975091939597565b662386f26fc1000081565b6000806000806000600860008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001549450600860008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600101549350600860008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201549250600860008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060030160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169150600860008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060040160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905091939590929450565b610e1081565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600080600860008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002091506121cf33611bde565b9050600082600101541180156121e85750816001015443115b80156121f75750816002015443105b80156122555750600078010000000000000000000000000000000000000000000000000277ffffffffffffffffffffffffffffffffffffffffffffffff19168177ffffffffffffffffffffffffffffffffffffffffffffffff191614155b80156122b1575062124f80600260008377ffffffffffffffffffffffffffffffffffffffffffffffff191677ffffffffffffffffffffffffffffffffffffffffffffffff19168152602001908152602001600020600401544303115b801561234a575060001515600760008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515145b151561235557600080fd5b6001600760008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff02191690831515021790555060018260000160008282540192505081905550600260015481151561240b57fe5b04826000015411156124625743826002018190555082600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b505050565b6000600860008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001541480156124fb57506000600860008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060010154145b801561250e5750678ac7230489e8000034145b151561251957600080fd5b60a0604051908101604052806000815260200143815260200162124f80430181526020013373ffffffffffffffffffffffffffffffffffffffff168152602001600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815250600860008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008201518160000155602082015181600101556040820151816002015560608201518160030160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060808201518160040160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555090505050565b6801158e460913d0000081565b6040805190810160405280600290602082028038833980820191505090505090565b6020604051908101604052806001906020820280388339808201915050905050905600a165627a7a72305820404e2868f2203a2d8fc30cc0ca3d80fbd3b1c3ad8e4078fcf1dccbd36e744bbf0029`

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

// BlockPingTimeout is a free data retrieval call binding the contract method 0xe3596ce0.
//
// Solidity: function blockPingTimeout() constant returns(uint256)
func (_Contract *ContractCaller) BlockPingTimeout(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "blockPingTimeout")
	return *ret0, err
}

// BlockPingTimeout is a free data retrieval call binding the contract method 0xe3596ce0.
//
// Solidity: function blockPingTimeout() constant returns(uint256)
func (_Contract *ContractSession) BlockPingTimeout() (*big.Int, error) {
	return _Contract.Contract.BlockPingTimeout(&_Contract.CallOpts)
}

// BlockPingTimeout is a free data retrieval call binding the contract method 0xe3596ce0.
//
// Solidity: function blockPingTimeout() constant returns(uint256)
func (_Contract *ContractCallerSession) BlockPingTimeout() (*big.Int, error) {
	return _Contract.Contract.BlockPingTimeout(&_Contract.CallOpts)
}

// CheckVote is a free data retrieval call binding the contract method 0x6069e56e.
//
// Solidity: function checkVote(proposalAddr address, voter address) constant returns(bool)
func (_Contract *ContractCaller) CheckVote(opts *bind.CallOpts, proposalAddr common.Address, voter common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "checkVote", proposalAddr, voter)
	return *ret0, err
}

// CheckVote is a free data retrieval call binding the contract method 0x6069e56e.
//
// Solidity: function checkVote(proposalAddr address, voter address) constant returns(bool)
func (_Contract *ContractSession) CheckVote(proposalAddr common.Address, voter common.Address) (bool, error) {
	return _Contract.Contract.CheckVote(&_Contract.CallOpts, proposalAddr, voter)
}

// CheckVote is a free data retrieval call binding the contract method 0x6069e56e.
//
// Solidity: function checkVote(proposalAddr address, voter address) constant returns(bool)
func (_Contract *ContractCallerSession) CheckVote(proposalAddr common.Address, voter common.Address) (bool, error) {
	return _Contract.Contract.CheckVote(&_Contract.CallOpts, proposalAddr, voter)
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

// EtzMin is a free data retrieval call binding the contract method 0xc808021c.
//
// Solidity: function etzMin() constant returns(uint256)
func (_Contract *ContractCaller) EtzMin(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "etzMin")
	return *ret0, err
}

// EtzMin is a free data retrieval call binding the contract method 0xc808021c.
//
// Solidity: function etzMin() constant returns(uint256)
func (_Contract *ContractSession) EtzMin() (*big.Int, error) {
	return _Contract.Contract.EtzMin(&_Contract.CallOpts)
}

// EtzMin is a free data retrieval call binding the contract method 0xc808021c.
//
// Solidity: function etzMin() constant returns(uint256)
func (_Contract *ContractCallerSession) EtzMin() (*big.Int, error) {
	return _Contract.Contract.EtzMin(&_Contract.CallOpts)
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
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address, blockOnlineAcc uint256, blockLastPing uint256)
func (_Contract *ContractCaller) GetInfo(opts *bind.CallOpts, id [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	BlockNumber    *big.Int
	Account        common.Address
	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
}, error) {
	ret := new(struct {
		Id1            [32]byte
		Id2            [32]byte
		PreId          [8]byte
		NextId         [8]byte
		BlockNumber    *big.Int
		Account        common.Address
		BlockOnlineAcc *big.Int
		BlockLastPing  *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getInfo", id)
	return *ret, err
}

// GetInfo is a free data retrieval call binding the contract method 0xc4e3ed93.
//
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address, blockOnlineAcc uint256, blockLastPing uint256)
func (_Contract *ContractSession) GetInfo(id [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	BlockNumber    *big.Int
	Account        common.Address
	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// GetInfo is a free data retrieval call binding the contract method 0xc4e3ed93.
//
// Solidity: function getInfo(id bytes8) constant returns(id1 bytes32, id2 bytes32, preId bytes8, nextId bytes8, blockNumber uint256, account address, blockOnlineAcc uint256, blockLastPing uint256)
func (_Contract *ContractCallerSession) GetInfo(id [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	BlockNumber    *big.Int
	Account        common.Address
	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, id)
}

// GetVoteInfo is a free data retrieval call binding the contract method 0xdc1e30da.
//
// Solidity: function getVoteInfo(addr address) constant returns(voteCount uint256, startBlock uint256, stopBlock uint256, creator address, lastAddress address)
func (_Contract *ContractCaller) GetVoteInfo(opts *bind.CallOpts, addr common.Address) (struct {
	VoteCount   *big.Int
	StartBlock  *big.Int
	StopBlock   *big.Int
	Creator     common.Address
	LastAddress common.Address
}, error) {
	ret := new(struct {
		VoteCount   *big.Int
		StartBlock  *big.Int
		StopBlock   *big.Int
		Creator     common.Address
		LastAddress common.Address
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getVoteInfo", addr)
	return *ret, err
}

// GetVoteInfo is a free data retrieval call binding the contract method 0xdc1e30da.
//
// Solidity: function getVoteInfo(addr address) constant returns(voteCount uint256, startBlock uint256, stopBlock uint256, creator address, lastAddress address)
func (_Contract *ContractSession) GetVoteInfo(addr common.Address) (struct {
	VoteCount   *big.Int
	StartBlock  *big.Int
	StopBlock   *big.Int
	Creator     common.Address
	LastAddress common.Address
}, error) {
	return _Contract.Contract.GetVoteInfo(&_Contract.CallOpts, addr)
}

// GetVoteInfo is a free data retrieval call binding the contract method 0xdc1e30da.
//
// Solidity: function getVoteInfo(addr address) constant returns(voteCount uint256, startBlock uint256, stopBlock uint256, creator address, lastAddress address)
func (_Contract *ContractCallerSession) GetVoteInfo(addr common.Address) (struct {
	VoteCount   *big.Int
	StartBlock  *big.Int
	StopBlock   *big.Int
	Creator     common.Address
	LastAddress common.Address
}, error) {
	return _Contract.Contract.GetVoteInfo(&_Contract.CallOpts, addr)
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

// LastProposalAddress is a free data retrieval call binding the contract method 0xe7b895b6.
//
// Solidity: function lastProposalAddress() constant returns(address)
func (_Contract *ContractCaller) LastProposalAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastProposalAddress")
	return *ret0, err
}

// LastProposalAddress is a free data retrieval call binding the contract method 0xe7b895b6.
//
// Solidity: function lastProposalAddress() constant returns(address)
func (_Contract *ContractSession) LastProposalAddress() (common.Address, error) {
	return _Contract.Contract.LastProposalAddress(&_Contract.CallOpts)
}

// LastProposalAddress is a free data retrieval call binding the contract method 0xe7b895b6.
//
// Solidity: function lastProposalAddress() constant returns(address)
func (_Contract *ContractCallerSession) LastProposalAddress() (common.Address, error) {
	return _Contract.Contract.LastProposalAddress(&_Contract.CallOpts)
}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() constant returns(uint256)
func (_Contract *ContractCaller) ProposalCount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "proposalCount")
	return *ret0, err
}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() constant returns(uint256)
func (_Contract *ContractSession) ProposalCount() (*big.Int, error) {
	return _Contract.Contract.ProposalCount(&_Contract.CallOpts)
}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() constant returns(uint256)
func (_Contract *ContractCallerSession) ProposalCount() (*big.Int, error) {
	return _Contract.Contract.ProposalCount(&_Contract.CallOpts)
}

// ProposalFee is a free data retrieval call binding the contract method 0xc27cabb5.
//
// Solidity: function proposalFee() constant returns(uint256)
func (_Contract *ContractCaller) ProposalFee(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "proposalFee")
	return *ret0, err
}

// ProposalFee is a free data retrieval call binding the contract method 0xc27cabb5.
//
// Solidity: function proposalFee() constant returns(uint256)
func (_Contract *ContractSession) ProposalFee() (*big.Int, error) {
	return _Contract.Contract.ProposalFee(&_Contract.CallOpts)
}

// ProposalFee is a free data retrieval call binding the contract method 0xc27cabb5.
//
// Solidity: function proposalFee() constant returns(uint256)
func (_Contract *ContractCallerSession) ProposalFee() (*big.Int, error) {
	return _Contract.Contract.ProposalFee(&_Contract.CallOpts)
}

// ProposalPeriod is a free data retrieval call binding the contract method 0x2c103c79.
//
// Solidity: function proposalPeriod() constant returns(uint256)
func (_Contract *ContractCaller) ProposalPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "proposalPeriod")
	return *ret0, err
}

// ProposalPeriod is a free data retrieval call binding the contract method 0x2c103c79.
//
// Solidity: function proposalPeriod() constant returns(uint256)
func (_Contract *ContractSession) ProposalPeriod() (*big.Int, error) {
	return _Contract.Contract.ProposalPeriod(&_Contract.CallOpts)
}

// ProposalPeriod is a free data retrieval call binding the contract method 0x2c103c79.
//
// Solidity: function proposalPeriod() constant returns(uint256)
func (_Contract *ContractCallerSession) ProposalPeriod() (*big.Int, error) {
	return _Contract.Contract.ProposalPeriod(&_Contract.CallOpts)
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

// InitGovernanceAddress is a paid mutator transaction binding the contract method 0x691444c1.
//
// Solidity: function initGovernanceAddress(addr address) returns()
func (_Contract *ContractTransactor) InitGovernanceAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initGovernanceAddress", addr)
}

// InitGovernanceAddress is a paid mutator transaction binding the contract method 0x691444c1.
//
// Solidity: function initGovernanceAddress(addr address) returns()
func (_Contract *ContractSession) InitGovernanceAddress(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.InitGovernanceAddress(&_Contract.TransactOpts, addr)
}

// InitGovernanceAddress is a paid mutator transaction binding the contract method 0x691444c1.
//
// Solidity: function initGovernanceAddress(addr address) returns()
func (_Contract *ContractTransactorSession) InitGovernanceAddress(addr common.Address) (*types.Transaction, error) {
	return _Contract.Contract.InitGovernanceAddress(&_Contract.TransactOpts, addr)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(id1 bytes32, id2 bytes32) returns()
func (_Contract *ContractTransactor) Register(opts *bind.TransactOpts, id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register", id1, id2)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(id1 bytes32, id2 bytes32) returns()
func (_Contract *ContractSession) Register(id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(id1 bytes32, id2 bytes32) returns()
func (_Contract *ContractTransactorSession) Register(id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2)
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

// ContractGovernanceAddressChangeIterator is returned from FilterGovernanceAddressChange and is used to iterate over the raw logs and unpacked data for GovernanceAddressChange events raised by the Contract contract.
type ContractGovernanceAddressChangeIterator struct {
	Event *ContractGovernanceAddressChange // Event containing the contract specifics and raw log

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
func (it *ContractGovernanceAddressChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractGovernanceAddressChange)
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
		it.Event = new(ContractGovernanceAddressChange)
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
func (it *ContractGovernanceAddressChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractGovernanceAddressChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractGovernanceAddressChange represents a GovernanceAddressChange event raised by the Contract contract.
type ContractGovernanceAddressChange struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterGovernanceAddressChange is a free log retrieval operation binding the contract event 0x2afa9f59c781db7a7ab5d83a590c0869db90657d3d51d7afe1c8ec41e088a22c.
//
// Solidity: e governanceAddressChange(from address, to address)
func (_Contract *ContractFilterer) FilterGovernanceAddressChange(opts *bind.FilterOpts) (*ContractGovernanceAddressChangeIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "governanceAddressChange")
	if err != nil {
		return nil, err
	}
	return &ContractGovernanceAddressChangeIterator{contract: _Contract.contract, event: "governanceAddressChange", logs: logs, sub: sub}, nil
}

// WatchGovernanceAddressChange is a free log subscription operation binding the contract event 0x2afa9f59c781db7a7ab5d83a590c0869db90657d3d51d7afe1c8ec41e088a22c.
//
// Solidity: e governanceAddressChange(from address, to address)
func (_Contract *ContractFilterer) WatchGovernanceAddressChange(opts *bind.WatchOpts, sink chan<- *ContractGovernanceAddressChange) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "governanceAddressChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractGovernanceAddressChange)
				if err := _Contract.contract.UnpackLog(event, "governanceAddressChange", log); err != nil {
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

// ContractNewProposalIterator is returned from FilterNewProposal and is used to iterate over the raw logs and unpacked data for NewProposal events raised by the Contract contract.
type ContractNewProposalIterator struct {
	Event *ContractNewProposal // Event containing the contract specifics and raw log

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
func (it *ContractNewProposalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractNewProposal)
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
		it.Event = new(ContractNewProposal)
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
func (it *ContractNewProposalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractNewProposalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractNewProposal represents a NewProposal event raised by the Contract contract.
type ContractNewProposal struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNewProposal is a free log retrieval operation binding the contract event 0xfcb77511a4d50d7ad5235ca4e1d7054d65140fe505eec9d700a69622a813485c.
//
// Solidity: e newProposal(from address, to address)
func (_Contract *ContractFilterer) FilterNewProposal(opts *bind.FilterOpts) (*ContractNewProposalIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "newProposal")
	if err != nil {
		return nil, err
	}
	return &ContractNewProposalIterator{contract: _Contract.contract, event: "newProposal", logs: logs, sub: sub}, nil
}

// WatchNewProposal is a free log subscription operation binding the contract event 0xfcb77511a4d50d7ad5235ca4e1d7054d65140fe505eec9d700a69622a813485c.
//
// Solidity: e newProposal(from address, to address)
func (_Contract *ContractFilterer) WatchNewProposal(opts *bind.WatchOpts, sink chan<- *ContractNewProposal) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "newProposal")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractNewProposal)
				if err := _Contract.contract.UnpackLog(event, "newProposal", log); err != nil {
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

// ContractNewVoteIterator is returned from FilterNewVote and is used to iterate over the raw logs and unpacked data for NewVote events raised by the Contract contract.
type ContractNewVoteIterator struct {
	Event *ContractNewVote // Event containing the contract specifics and raw log

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
func (it *ContractNewVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractNewVote)
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
		it.Event = new(ContractNewVote)
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
func (it *ContractNewVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractNewVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractNewVote represents a NewVote event raised by the Contract contract.
type ContractNewVote struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNewVote is a free log retrieval operation binding the contract event 0x0b16242fe09b9cf36e327548ad3c0c195442ee19f92b8b57fcf2d8cd765e9c7c.
//
// Solidity: e newVote(from address, to address)
func (_Contract *ContractFilterer) FilterNewVote(opts *bind.FilterOpts) (*ContractNewVoteIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "newVote")
	if err != nil {
		return nil, err
	}
	return &ContractNewVoteIterator{contract: _Contract.contract, event: "newVote", logs: logs, sub: sub}, nil
}

// WatchNewVote is a free log subscription operation binding the contract event 0x0b16242fe09b9cf36e327548ad3c0c195442ee19f92b8b57fcf2d8cd765e9c7c.
//
// Solidity: e newVote(from address, to address)
func (_Contract *ContractFilterer) WatchNewVote(opts *bind.WatchOpts, sink chan<- *ContractNewVote) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "newVote")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractNewVote)
				if err := _Contract.contract.UnpackLog(event, "newVote", log); err != nil {
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

// ContractPingIterator is returned from FilterPing and is used to iterate over the raw logs and unpacked data for Ping events raised by the Contract contract.
type ContractPingIterator struct {
	Event *ContractPing // Event containing the contract specifics and raw log

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
func (it *ContractPingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractPing)
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
		it.Event = new(ContractPing)
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
func (it *ContractPingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractPingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractPing represents a Ping event raised by the Contract contract.
type ContractPing struct {
	Id             [8]byte
	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPing is a free log retrieval operation binding the contract event 0xb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c2.
//
// Solidity: e ping(id bytes8, blockOnlineAcc uint256, blockLastPing uint256)
func (_Contract *ContractFilterer) FilterPing(opts *bind.FilterOpts) (*ContractPingIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ping")
	if err != nil {
		return nil, err
	}
	return &ContractPingIterator{contract: _Contract.contract, event: "ping", logs: logs, sub: sub}, nil
}

// WatchPing is a free log subscription operation binding the contract event 0xb620b17a993c1ab2769ca9e6d72d178499b0cd9b800d62e9b3d502e01bca76c2.
//
// Solidity: e ping(id bytes8, blockOnlineAcc uint256, blockLastPing uint256)
func (_Contract *ContractFilterer) WatchPing(opts *bind.WatchOpts, sink chan<- *ContractPing) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ping")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractPing)
				if err := _Contract.contract.UnpackLog(event, "ping", log); err != nil {
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
