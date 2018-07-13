package state

import (
	"math/big"
	"math"
)

// EXP(−1÷(etz×50)×10000)×10000000+200000
// EXP(−1÷(etz×2)×1000)×200000+1000
func CalculatePower(prevBlock, newBlock, prevPower, balance *big.Int) *big.Int {
	if prevBlock.Cmp(newBlock) >= 0 || balance.Cmp(big.NewInt(1e+16)) < 0 {
		return prevPower
	}

	etz1 := new(big.Int).Div(balance, big.NewInt(1e+16))
	etz2 := float64(etz1.Uint64()) / 100.0

	max1 := math.Exp(-1/(etz2*50)*10000) * 10000000 + 200000
	max2 := new(big.Int).Mul(big.NewInt(int64(max1)), big.NewInt(36000e+9))

	blockGap := float64(new(big.Int).Sub(newBlock, prevBlock).Uint64())
	speed := math.Exp(-1/(etz2*2)*1000) * 200000 + 1000

	power1 := big.NewInt(int64(blockGap * speed))
	power1.Mul(power1, big.NewInt(18e+9))
	power2 := new(big.Int).Add(prevPower, power1)

	if power2.Cmp(max2) > 0 || prevPower.Cmp(power2) > 0 {
		power2 = max2
	}
	return power2
}

func MaxPower(balance *big.Int) *big.Int {
	etz1 := new(big.Int).Div(balance, big.NewInt(1e+16))
	etz2 := float64(etz1.Uint64()) / 100.0
	max := math.Exp(-1/(etz2*50)*10000) * 10000000 + 200000
	return new(big.Int).Mul(big.NewInt(int64(max)), big.NewInt(18e+9))
}

