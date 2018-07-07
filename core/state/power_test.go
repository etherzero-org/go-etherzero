package state

import (
	"testing"
	"math/big"
	"fmt"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/rlp"
)

func TestPower(t *testing.T){
	prevBlock := big.NewInt(100)
	newBlock := big.NewInt(110)
	prevPower := big.NewInt(1234)
	for i := 0; i < 10; i++ {
		balance := new(big.Int).Mul(big.NewInt(1e+15), big.NewInt(int64(i*20000)))
		Power := CalculatePower(prevBlock, newBlock, prevPower, balance)
		fmt.Println(Power.String())
	}
}

func TestTx(t *testing.T) {
	encodedTx := common.Hex2Bytes("f8657a85098bca5a0082520894bb512c9b0b99f1ff7b662f8a1790ae769165438f808081d3a057226425b5a1b8f81918cef3e43caf2dd8aa4e00bb730f73396e48147b4a01e9a01d881e2f166c9233ea6b550364c08b365191caa85cf647ca2e7016197af2439a")
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		fmt.Println("error", err)
	}
	fmt.Println(tx.String(), tx.ChainId())
}
