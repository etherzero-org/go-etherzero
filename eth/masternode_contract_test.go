package eth

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethzero/go-ethzero/accounts/abi/bind"
	"github.com/ethzero/go-ethzero/accounts/abi/bind/backends"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
	"fmt"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"encoding/binary"
	"net"
	"github.com/ethzero/go-ethzero/p2p/discover"
)

var (
	key0, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	key1, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	key2, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	addr0   = crypto.PubkeyToAddress(key0.PublicKey)
	addr1   = crypto.PubkeyToAddress(key1.PublicKey)
	addr2   = crypto.PubkeyToAddress(key2.PublicKey)
)

func newTestBackend() *backends.SimulatedBackend {
	val, _ := new(big.Int).SetString("20000000000000000000000", 10)
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		addr0: {Balance: val},
		addr1: {Balance: val},
		addr2: {Balance: val},
	})
}

func deploy(prvKey *ecdsa.PrivateKey, amount *big.Int, backend *backends.SimulatedBackend) (common.Address, error) {
	deployTransactor := bind.NewKeyedTransactor(prvKey)
	deployTransactor.Value = amount
	addr, _, _, err := contract.DeployContract(deployTransactor, backend)
	if err != nil {
		return common.Address{}, err
	}
	backend.Commit()
	return addr, nil
}

func TestIssueAndReceive(t *testing.T) {
	backend := newTestBackend()

	addr0, err := deploy(key0, big.NewInt(0), backend)
	if err != nil {
		t.Fatalf("deploy contract: expected no error, got %v", err)
	}

	contract, err1 := contract.NewContract(addr0, backend)
	if err1 != nil {
		t.Fatalf("expected no error, got %v", err1)
	}

	var (
		id1 [32]byte
		id2 [32]byte
		misc [32]byte
	)

	addr := net.TCPAddr{net.ParseIP("127.0.0.88"), 21212, ""}

	misc[0] = 1
	copy(misc[1:17], addr.IP)
	binary.BigEndian.PutUint16(misc[17:19], uint16(addr.Port))

	nodeID, _ := discover.HexID("0x2cb5063f3fe98370023ecbf05a5f61534ac724e8bfc52e72e2f33dc57e6328a15bb6c09ce296c546a35c1469b6d2a013d6fc1f2a123ee867764e8c5e184e46ce")

	copy(id1[:], nodeID[:32])
	copy(id2[:], nodeID[32:64])

	transactOpts := bind.NewKeyedTransactor(key0)
	val, _ := new(big.Int).SetString("20000000000000000000", 10)
	transactOpts.Value = val

	tx, err := contract.Register(transactOpts, id1, id2, misc)
	fmt.Println("Register", tx, err)
	backend.Commit()

	masternodes, _ := masternode.NewMasternodeSet(contract)
	masternodes.Show()

	count, err2 := contract.ContractCaller.Count(nil)
	fmt.Println("count", count.String(), err2)
}