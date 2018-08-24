package params

import (
	"testing"
	"github.com/etherzero/go-etherzero/p2p/discover"
	"fmt"
	"github.com/etherzero/go-etherzero/crypto"
	"crypto/ecdsa"
	"github.com/etherzero/go-etherzero/common"
)

func Test_MasterodeShortID(t *testing.T) {
	fmt.Println("id:")
	for _, n := range MainnetMasternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\"%x\",\n", node.ID[:8])
	}
}

func Test_MasterodeRegParams(t *testing.T) {
	for _, n := range MainnetMasternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\"0x%x\",\"0x%x\"\n", node.ID[:32], node.ID[32:])
	}
}

func Test_MasterodeRegParamsForTX(t *testing.T) {
	for i, n := range TestnetMasternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[%d] 0x2f926732%x\n", i, node.ID[:])
	}
}

func Test_PrintAllocCode(t *testing.T) {
	for _, n := range MainnetMasternodes {
		node, err := discover.ParseNode(n)
		if err != nil {
			panic(err)
		}
		pubkey, err := node.ID.Pubkey()
		addr := crypto.PubkeyToAddress(*pubkey)

		fmt.Printf("alloc[common.HexToAddress(\"%s\")] = GenesisAccount{Balance: big.NewInt(1e+16)}\n", addr.Hex())
	}
}

func newkey() *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("couldn't generate key: " + err.Error())
	}
	return key
}

func Test_MasternodeKey(t *testing.T){
	port := 20000
	for i := 1; i < 100; i++ {
		port++
		key := newkey()
		hex := common.Bytes2Hex(crypto.FromECDSA(key))
		pub := discover.PubkeyID(&key.PublicKey)
		fmt.Printf(	"\"enode://%s@0.0.0.0:%d\", // [%d] %s\n", common.Bytes2Hex(pub[:]), port, i, hex)
	}
}