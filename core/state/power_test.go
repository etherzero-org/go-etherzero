package state

import (
	"testing"
	"math/big"
	"fmt"
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
