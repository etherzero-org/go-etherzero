package eth

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"testing"

	"github.com/ethzero/go-ethzero/accounts/abi/bind"
	"github.com/ethzero/go-ethzero/accounts/abi/bind/backends"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/contracts/masternode/contract"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/core/types/masternode"
	"github.com/ethzero/go-ethzero/crypto"
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

func genKeys(N int) (keys []*ecdsa.PrivateKey) {
	for ; N > 0; N-- {
		key, _ := crypto.GenerateKey()
		keys = append(keys, key)
	}
	return keys
}

func genNodeID() discover.NodeID {
	key, _ := crypto.GenerateKey()
	return discover.PubkeyID(&key.PublicKey)
}

func newTestBackend() *backends.SimulatedBackend {
	val := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e+18))
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		addr0: {Balance: val},
		addr1: {Balance: val},
		addr2: {Balance: val},
	})
}

func newTestBackendAndKeys(N int) (*backends.SimulatedBackend, []*ecdsa.PrivateKey) {
	val := new(big.Int).Mul(big.NewInt(200000), big.NewInt(1e+18))
	genesis := core.GenesisAlloc{}
	keys := genKeys(N)
	for _, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		genesis[addr] = core.GenesisAccount{Balance: val}
	}
	return backends.NewSimulatedBackend(genesis), keys
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

//newMasternodeSet generate a new MasternodeSet
func newMasternodeSet(emptyFlag bool) (*masternode.MasternodeSet) {
	backend, keys := newTestBackendAndKeys(1)

	addr1, err := deploy(keys[0], big.NewInt(0), backend)
	if err != nil {
		fmt.Errorf("deploy contract: expected no error, got %v", err)
	}

	contract, err1 := contract.NewContract(addr1, backend)
	if err1 != nil {
		fmt.Errorf("expected no error, got %v", err1)
	}

	if emptyFlag {
		var (
			id1  [32]byte
			id2  [32]byte
			misc [32]byte
		)

		for i, key := range keys {
			ipString := fmt.Sprintf("127.0.0.%d", i)
			addr := net.TCPAddr{net.ParseIP(ipString), 2121 + i, ""}
			misc[0] = 1
			copy(misc[1:17], addr.IP)
			binary.BigEndian.PutUint16(misc[17:19], uint16(addr.Port))

			nodeID := genNodeID()
			copy(id1[:], nodeID[:32])
			copy(id2[:], nodeID[32:64])

			transactOpts := bind.NewKeyedTransactor(key)
			val := new(big.Int).Mul(big.NewInt(20), big.NewInt(1e+18))
			transactOpts.Value = val

			tx, err := contract.Register(transactOpts, id1, id2, misc)
			if err != nil {
				fmt.Println("Register Error:", tx, err)
			}

			backend.Commit()
		}

	}
	masternodes, _ := masternode.NewMasternodeSet(contract)

	count, err2 := contract.ContractCaller.Count(nil)
	fmt.Println("count", count.String(), err2)
	return masternodes
}
func TestMasternodeReg(t *testing.T) {
	ms := newMasternodeSet(true)
	ms.Show()
}
