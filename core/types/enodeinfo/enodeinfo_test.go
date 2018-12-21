package enodeinfo

import (
	"testing"
	"github.com/etherzero/go-etherzero/common"
	"fmt"
)

func Test_AAAAA(t *testing.T) {
	id1Tmp := "f29554ab856f4a84657b006a9b47691d5f93ff5bedb46d082ad0a9208ff0be62"
	id2Tmp := "35fec1fe5f1043cb9c3d20285549b0896bec7871032667a33167911bc806b7a3"

	id1tmp := common.Hex2Bytes(id1Tmp)
	id2tmp := common.Hex2Bytes(id2Tmp)

	port := uint64(9151314447111835180)
	var id1, id2 [32]byte
	copy(id1[:], id1tmp)
	copy(id2[:], id2tmp)

	node := newDiscoverNode(id1, id2, port)
	fmt.Printf("id1  %+v id2 %+v  node %+v", string(id1[:]), string(id2[:]), node.String())
}
