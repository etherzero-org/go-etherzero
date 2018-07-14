package eth

import (
	"fmt"
	"io/ioutil"
	"github.com/etherzero/go-etherzero/common"
	"github.com/etherzero/go-etherzero/consensus/ethash"
	"github.com/etherzero/go-etherzero/core"
	"github.com/etherzero/go-etherzero/core/types/masternode"
	"github.com/etherzero/go-etherzero/node"
	"github.com/etherzero/go-etherzero/ethdb"
	"github.com/etherzero/go-etherzero/params"
	"github.com/etherzero/go-etherzero/core/vm"
	"github.com/etherzero/go-etherzero/crypto"
	"time"
	"math/big"
)

const (
	testInstance = "console-tester"
	testAddress1 = "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
)

var (
	key0, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	key1, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	key2, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	addr0   = crypto.PubkeyToAddress(key0.PublicKey)
	addr1   = crypto.PubkeyToAddress(key1.PublicKey)
	addr2   = crypto.PubkeyToAddress(key2.PublicKey)
)

func newEtherrum() *Ethereum {
	// Create a temporary storage for the node keys and initialize it
	workspace, err := ioutil.TempDir("", "console-tester-")
	if err != nil {
		fmt.Printf("failed to create temporary keystore: %v", err)
	}

	// Create a networkless protocol stack and start an Ethereum service within
	stack, err := node.New(&node.Config{DataDir: workspace, UseLightweightKDF: true, Name: testInstance})
	if err != nil {
		fmt.Printf("failed to create node: %v", err)
	}
	ethConf := &Config{
		Genesis:   core.DeveloperGenesisBlock(15, common.Address{}),
		Etherbase: common.HexToAddress(testAddress1),
		Ethash: ethash.Config{
			PowMode: ethash.ModeTest,
		},
	}

	if err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) { return New(ctx, ethConf) }); err != nil {
		fmt.Printf("failed to register Ethereum protocol: %v", err)
	}
	// Start the node and assemble the JavaScript console around it
	if err = stack.Start(); err != nil {
		fmt.Printf("failed to start test stack: %v", err)
	}
	_, err = stack.Attach()
	if err != nil {
		fmt.Printf("failed to attach to node: %v", err)
	}

	// Create the final tester and return
	var ethereum *Ethereum
	err = stack.Service(&ethereum)
	if err != nil {
		fmt.Printf("failed to as a service: %v", err)
	}

	ethereum.blockchain = newBlockChain()
	return ethereum
}
func genNewgspec() core.Genesis {
	return core.Genesis{
		Config: params.TestChainConfig,
		Alloc:  core.GenesisAlloc{addr1: {Balance: big.NewInt(10000000000000)}},
	}
}

func newBlockChain() *core.BlockChain {
	db := ethdb.NewMemDatabase()
	fmt.Println("etherzero", genNewgspec().Config)
	//bc, _ := core.NewBlockChain(db, nil, oldcustomg.Config, ethash.NewFullFaker(), vm.Config{})
	cacheConfig := &core.CacheConfig{
		Disabled:      true, // Whether to disable trie write caching (archive node)
		TrieNodeLimit: 1,    // Memory limit (MB) at which to flush the current in-memory trie to disk
		TrieTimeLimit: time.Duration(10),
	}
	vmConfig := vm.Config{
		Debug:                   true,
		NoRecursion:             true,
		EnablePreimageRecording: true,
	}
	a := genNewgspec()
	core.SetupGenesisBlock(db, &a)
	chainman, _ := core.NewBlockChain(db, cacheConfig, a.Config, ethash.NewFaker(), vmConfig)
	hash, number := common.Hash{0: 0xff}, uint64(314)
	core.WriteCanonicalHash(db, hash, number)
	return chainman
}

func returnMasternodeManager() *MasternodeManager {
	//// initial the parameter may needed during this test function
	//eth := newEtherrum()
	return &MasternodeManager{
		newPeerCh:   make(chan *peer),
		masternodes: &masternode.MasternodeSet{},
	}
}

