package eth

import (
	"github.com/ethzero/go-ethzero/p2p/discover"
	"github.com/ethzero/go-ethzero/common"
	"math/big"
	"net"
	"sync"
	"fmt"
	"github.com/ethzero/go-ethzero/log"
	"encoding/binary"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
)

type ContractNode struct {
	id         string
	Node        *discover.Node
	Account     common.Address
	OriginBlock uint64
	Height      *big.Int
	State       int
}

func newContractNode(nodeId discover.NodeID, ip net.IP, port uint16, account common.Address, block uint64) *ContractNode {
	id := GetContractNodeID(nodeId)
	n := discover.NewNode(nodeId, ip, 0, port)
	return &ContractNode{
		id:         id,
		Node:        n,
		Account:     account,
		OriginBlock: block,
		State:       0,
	}
}

func (n *ContractNode) String() string {
	return n.Node.String()
}

type ContractNodeSet struct {
	nodes       map[string]*ContractNode
	lock        sync.RWMutex
	closed      bool
	contract    *contract.Contract
	initialized bool
}

func NewContractNodeSet() *ContractNodeSet {
	return &ContractNodeSet{
		nodes: make(map[string]*ContractNode),
	}
}

func (ns *ContractNodeSet) Initialized() bool {
	return ns.initialized
}

// Register injects a new node into the working set, or returns an error if the
// node is already known.
func (ns *ContractNodeSet) Register(n *ContractNode) error {
	ns.lock.Lock()
	defer ns.lock.Unlock()

	if ns.closed {
		return errClosed
	}
	if _, ok := ns.nodes[n.id]; ok {
		return errAlreadyRegistered
	}
	ns.nodes[n.id] = n
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ns *ContractNodeSet) Unregister(id string) error {
	ns.lock.Lock()
	defer ns.lock.Unlock()

	if _, ok := ns.nodes[id]; !ok {
		return errNotRegistered
	}
	delete(ns.nodes, id)
	return nil
}

// Peer retrieves the registered peer with the given id.
func (ns *ContractNodeSet) Node(id string) *ContractNode {
	ns.lock.RLock()
	defer ns.lock.RUnlock()

	return ns.nodes[id]
}

// Close disconnects all nodes.
// No new nodes can be registered after Close has returned.
func (ns *ContractNodeSet) Close() {
	ns.lock.Lock()
	defer ns.lock.Unlock()

	//for _, p := range ns.nodes {
	//	p.Disconnect(p2p.DiscQuitting)
	//}
	ns.closed = true
}

func (ns *ContractNodeSet) NodeJoin(id [8]byte) (*ContractNode, error) {
	ctx, err := GetContractNodeContext(ns.contract, id)
	if err != nil {
		return &ContractNode{}, err
	}
	err = ns.Register(ctx.Node)
	if err != nil {
		return &ContractNode{}, err
	}
	return ctx.Node, nil
}

func (ns *ContractNodeSet) NodeQuit(id [8]byte) {
	ns.Unregister(fmt.Sprintf("%x", id[:8]))
}

func (ns *ContractNodeSet) Show() {
	for _, n := range ns.nodes {
		fmt.Println(n.String())
	}
}

func (ns *ContractNodeSet) GetNodes() *map[string]*ContractNode {
	return &ns.nodes
}

func (ns *ContractNodeSet) Len() int {
	return len(ns.nodes)
}

func (ns *ContractNodeSet) Init(contract *contract.Contract) {
	var (
		lastId [8]byte
		ctx    *contractNodeContext
	)
	lastId, err := contract.LastId(nil)
	if err != nil {
		return
	}
	for lastId != ([8]byte{}) {
		ctx, err = GetContractNodeContext(contract, lastId)
		if err != nil {
			log.Error("Init NodeSet", "error", err)
			break
		}
		ns.Register(ctx.Node)
		lastId = ctx.pre
	}
	ns.contract = contract
	ns.initialized = true
}

func GetContractNodeID(ID discover.NodeID) string {
	return fmt.Sprintf("%x", ID[:8])
}

type contractNodeContext struct {
	Node *ContractNode
	pre  [8]byte
	next [8]byte
}

func GetContractNodeContext(contract *contract.Contract, id [8]byte) (*contractNodeContext, error) {
	data, err := contract.ContractCaller.GetInfo(nil, id)
	if err != nil {
		return &contractNodeContext{}, err
	}
	// version := int(data.Misc[0])
	var ip net.IP = data.Misc[1:17]
	port := binary.BigEndian.Uint16(data.Misc[17:19])
	nodeId, _ := discover.BytesID(append(data.Id1[:], data.Id2[:]...))
	node := newContractNode(nodeId, ip, port, data.Account, data.BlockNumber.Uint64())

	return &contractNodeContext{
		Node: node,
		pre:  data.PreId,
		next: data.NextId,
	}, nil
}