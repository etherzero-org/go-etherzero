package trie

import (
	"github.com/etherzero/go-etherzero/common"
	"os"
	"fmt"
	"bufio"
	"math/big"
	"github.com/etherzero/go-etherzero/rlp"
)

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash
	CodeHash []byte
}

func Export(root common.Hash, srcDb *Database, dstPath string) {
	f, err := os.Create(dstPath)
	if err != nil {
		fmt.Printf("create map file error: %v\n", err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	fmt.Println("root:", root.Hex())

	srcTrie, err := New(root, srcDb)
	if err != nil {
		panic(err)
	}

	accountCount := 0
	accountTotal := 0
	contractCount := 0

	contractBalance := big.NewInt(0)
	accountBalance := big.NewInt(0)

	it := NewIterator(srcTrie.NodeIterator(nil))
	for it.Next() {
		var data Account
		if err := rlp.DecodeBytes(it.Value, &data); err != nil {
			panic(err)
		}
		if data.Balance.Cmp(big.NewInt(1e+16)) > 0 {
			if data.Root == emptyRoot {
				accountCount++
				accountBalance.Add(accountBalance, data.Balance)
				bin := append(it.Key, common.LeftPadBytes(data.Balance.Bytes(), 11)...)
				if len(bin) != 43 {
					panic("error len")
				}
				w.Write(bin)
			} else {
				contractBalance.Add(contractBalance, data.Balance)
				contractCount++
			}
		}
		accountTotal++
	}

	fmt.Println("Account Total Count :", accountTotal)
	fmt.Println("Accounts Count      :", accountCount)
	fmt.Println("Contract Count      :", contractCount)
	fmt.Println("Contract Balance Sum:", contractBalance.String())
	fmt.Println("Accounts Balance Sum:", accountBalance.String())

	w.Flush()
}
