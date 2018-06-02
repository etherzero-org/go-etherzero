package masternode

import (
	"fmt"
	"testing"

	"github.com/ethzero/go-ethzero/common"
)

func TestTxLockVote_IsValid(t *testing.T) {
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	txLockVote := NewTxLockVote(
		hash,
		fmt.Sprintf("0x65bc97ef01b35f86a45c319675a05699e0947743c59ed53d0a918fb215c5ee5f"),
	)
	fmt.Printf("timeout ret is %v", txLockVote.IsTimeOut())
}
