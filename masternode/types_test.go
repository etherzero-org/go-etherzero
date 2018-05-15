package masternode

import (
	"testing"
	"github.com/ethzero/go-ethzero/crypto"
	"encoding/hex"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"fmt"
)

var (
	testAccountKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAccount       = crypto.PubkeyToAddress(testAccountKey.PublicKey)
	testTxHash        = "0x612f8d37bd75944683d4d58386e6dc7c113e3cd7e87253435cab017ca0e7a34b"
	testAccount2      = "0x289e14fc8eaf1ba80224eb9ea07c92a938f27d16"
)

func TestSign(t *testing.T) {
	t.Log("testAccount: ", testAccount.Hex())
	id := discover.PubkeyID(&testAccountKey.PublicKey)
	t.Log("id: ", id)

	//time := time.Now().Unix()

	msg := PingMsg{
		ID: id,
		Port: 1,
	}
	err := msg.Update(testAccountKey)

	t.Log(hex.EncodeToString(msg.Sig), err)

	myid, err := msg.Check()
	if err != nil {
		fmt.Println(err)
	}
	t.Log("ID: ", myid)


}

