package masternode

import (
	"fmt"
	"testing"

	"github.com/ethzero/go-ethzero/common"
)

func TestTxLockVote_IsValid(t *testing.T) {
	//var _startDate int64 = time.Now().Unix()             //+ int64(time.Second*10)
	//var createdTime time.Time = time.Unix(_startDate, 0) //.Format("2006-01-02 15:04:05")
	//fmt.Printf("%v\n", uint64(time.Now().Sub(createdTime)))
	//fmt.Printf("%v\n", time.Now().Sub(createdTime))
	var hash common.Hash
	for i := range hash {
		hash[i] = byte(i)
	}
	txLockVote := NewTxLockVote(
		hash,
		fmt.Sprintf("aaa"),
	)
	fmt.Printf("timeout ret is %v", txLockVote.IsTimeOut())
}
