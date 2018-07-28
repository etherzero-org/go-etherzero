package devote

import (
	"testing"
	"github.com/etherzero/go-etherzero/ethdb"
	"fmt"
)

func newDevoteProtocol() *DevoteProtocol {
	db := ethdb.NewMemDatabase()
	ctxAtomic := &DevoteProtocolAtomic{}
	dp, err := NewDevoteProtocolFromAtomic(db, ctxAtomic)
	if err != nil {
		fmt.Println("err ", err)
	}
	return dp
}
func TestDevoteProtocol_ProtocolAtomic(t *testing.T) {

}
