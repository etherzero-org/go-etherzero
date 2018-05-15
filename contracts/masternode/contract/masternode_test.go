// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package contract

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethzero/go-ethzero/accounts/abi/bind"
	"github.com/ethzero/go-ethzero/accounts/abi/bind/backends"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/core"
	"github.com/ethzero/go-ethzero/crypto"
	"fmt"
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
	return backends.NewSimulatedBackend(core.GenesisAlloc{
		addr0: {Balance: big.NewInt(1000000000000000000)},
		addr1: {Balance: big.NewInt(1000000000000000000)},
		addr2: {Balance: big.NewInt(1000000000000000000)},
	})
}

func deploy(prvKey *ecdsa.PrivateKey, amount *big.Int, backend *backends.SimulatedBackend) (common.Address, error) {
	deployTransactor := bind.NewKeyedTransactor(prvKey)
	deployTransactor.Value = amount
	addr, _, _, err := DeployContract(deployTransactor, backend)
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
	contract, err1 := NewContract(addr0, backend)
	if err1 != nil {
		t.Fatalf("expected no error, got %v", err1)
	}
	fmt.Println(addr0.Hex())

	count, err2 := contract.ContractCaller.Count(nil)
	fmt.Println("count", count.String(), err2)

}
