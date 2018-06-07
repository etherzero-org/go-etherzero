package antidote

import (
	"testing"
	"fmt"
	"github.com/ethzero/go-ethzero/common"
)

func TestAntidoteDBInt64(t *testing.T) {
	db, _ := NewAntidoteDB("/Users/rolong/Library/Ethzero/testnet/geth/antidote", 1, nil)
	defer db.close()

	account := common.HexToAddress("0x01")
	fmt.Println(common.HexToHash("0x0202").String(), 1, 2)

	db.Put(account, common.HexToHash("0x0202"), 1, 2)
	txHash, nonce, blockNumber := db.Get(account)

	fmt.Println(txHash.String(), nonce, blockNumber)
}
