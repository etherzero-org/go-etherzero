package masternode

import (
	"encoding/binary"
	"fmt"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"net"
	"net/url"
)

type halfID [32]byte

type ContractData struct {
	Node    *Node
	Version int
	Pre     halfID
	Next    halfID
	block   uint64
}

func (c *ContractData) String() string {
	return fmt.Sprintf(" node:%s\n v:%d\n pre:%x\n next:%x", c.Version, c.Node, c.Pre, c.Next)
}

func GetContractData(contract *contract.Contract, id halfID) (*ContractData, error) {
	data, err := contract.ContractCaller.GetInfo(nil, id)
	if err != nil {
		return &ContractData{}, err
	}
	version := int(data.Misc[0])
	var ip net.IP = data.Misc[1:17]
	port := binary.BigEndian.Uint16(data.Misc[17:19])
	nodeId, _ := discover.BytesID(append(id[:], data.SubId[:]...))
	node := NewNode(nodeId, ip, port)
	contractData := &ContractData{
		Node:    node,
		Version: version,
		Pre:     data.PreId,
		Next:    data.NextId,
		block:   data.BlockNumber.Uint64(),
	}
	// fmt.Println(contractData)
	return contractData, nil
}

type Node struct {
	ID    discover.NodeID
	IP    net.IP
	Port  uint16
	State int
}

func NewNode(id discover.NodeID, ip net.IP, port uint16) *Node {
	return &Node{
		ID:    id,
		IP:    ip,
		Port:  port,
		State: 0,
	}
}

func (n *Node) String() string {
	u := url.URL{Scheme: "enode"}
	addr := net.TCPAddr{IP: n.IP, Port: int(n.Port)}
	u.User = url.User(fmt.Sprintf("%x", n.ID[:]))
	u.Host = addr.String()
	return u.String()
}

type NodeList map[halfID]*Node

func NewNodeList() *NodeList {
	return &NodeList{}
}

func (l *NodeList) Put(m *Node) {
	hid := NodeID2HalfID(m.ID)
	(*l)[hid] = m
}

func (l *NodeList) NodeJoin(contract *contract.Contract, id halfID) {
	if data, err := GetContractData(contract, id); err == nil {
		l.Put(data.Node)
	}
}

func (l *NodeList) NodeQuit(contract *contract.Contract, id halfID) {
	delete((*l), id)
}

func (l *NodeList) Show() {
	for _, n := range *l {
		fmt.Println(n.String())
	}
}

func (l *NodeList) Len() int {
	return len(*l)
}

func (l *NodeList) Init(contract *contract.Contract) {
	var (
		lastId halfID
		data   *ContractData
	)
	lastId, err := contract.LastId(nil)
	if err != nil {
		return
	}
	for lastId != (halfID{}) {
		data, err = GetContractData(contract, lastId)
		if err != nil {
			log.Error("Init GetContractData", "error", err)
			break
		}
		l.Put(data.Node)
		lastId = data.Pre
	}
}

func NodeID2HalfID(ID discover.NodeID) halfID {
	var hid halfID
	hidS := hid[0:32]
	copy(hidS, ID[0:32])
	return hid
}
