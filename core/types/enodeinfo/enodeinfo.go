package enodeinfo

import (
	"github.com/etherzero/go-etherzero/accounts/abi/bind"
	"github.com/etherzero/go-etherzero/contracts/enodeinfo/contract"
	"math/big"
	"fmt"
	"github.com/etherzero/go-etherzero/aux"
	"github.com/etherzero/go-etherzero/p2p/enode"
)

func GetNodesByBlockNumber(contract *contract.Contract, blockNumber *big.Int, nodeid [8]byte) (node *enode.Node, err error) {
	if blockNumber == nil {
		blockNumber = new(big.Int)
	}
	opts := new(bind.CallOpts)
	opts.BlockNumber = blockNumber

	data, err := contract.ContractCaller.GetSingleEnode(opts, nodeid)
	if err != nil {
		fmt.Printf("GetNodesByBlockNumber failed,err: %v\n", err)
		return
	}
	if data.Id1 == [32]byte{} ||
		data.Id2 == [32]byte{} ||
		len(data.Id2) != 32 ||
		len(data.Id1) != 32 ||
		data.Ipport == uint64(0) {
		return
	}
	node = aux.NewDiscoverNode(data.Id1, data.Id2, data.Ipport)
	fmt.Printf("node is %v", node.String())
	return
}
