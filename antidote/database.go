package antidote

import (
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/binary"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"os"
	"github.com/ethzero/go-ethzero/common"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/core"
	"fmt"
	"github.com/ethzero/go-ethzero/core/types"
	"github.com/ethzero/go-ethzero/common/hexutil"
)

var (
	AntidoteDBVersionKey = []byte("version")
	AntidoteDBItemPrefix = []byte("a:")      // Identifier to prefix node entries with

)

type AntidoteDB struct{
	lvl    *leveldb.DB   // Interface to the database itself
	quit   chan struct{} // Channel to signal the expiring thread to stop
}

func NewAntidoteDB(path string, version int, chain *core.BlockChain, chainDB core.DatabaseReader) (*AntidoteDB, error) {

	var db *AntidoteDB
	var err error
	if path == "" {
		db, err = newMemoryAntidoteDB()
	}else{
		db, err = newPersistentAntidoteDB(path, version)
	}

	max := chain.CurrentBlock().NumberU64()
	i := uint64(1)
	for ; i <= max; i++ {
		block := chain.GetBlockByNumber(i)
		block.NumberU64()
		txs := block.Transactions()
		receipts := core.GetBlockReceipts(chainDB, block.Hash(), i)
		for j, tx := range txs {
			var signer types.Signer = types.FrontierSigner{}
			if tx.Protected() {
				signer = types.NewEIP155Signer(tx.ChainId())
			}
			from, _ := types.Sender(signer, tx)
			fmt.Println("tx:", j, tx.Gas())
			receipt := receipts[j]
			fmt.Println("tx receipt:", j, hexutil.Uint64(receipt.CumulativeGasUsed), hexutil.Uint64(receipt.GasUsed), from.String())
		}

	}
	return db, err

}

// newMemoryAntidoteDB creates a new in-memory antidote database without a persistent
// backend.
func newMemoryAntidoteDB() (*AntidoteDB, error) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		return nil, err
	}
	return &AntidoteDB{
		lvl:  db,
		quit: make(chan struct{}),
	}, nil
}

// newPersistentAntidoteDB creates/opens a leveldb backed persistent antidote database,
// also flushing its contents in case of a version mismatch.
func newPersistentAntidoteDB(path string, version int) (*AntidoteDB, error) {
	opts := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(path, opts)
	if _, iscorrupted := err.(*errors.ErrCorrupted); iscorrupted {
		db, err = leveldb.RecoverFile(path, nil)
	}
	if err != nil {
		return nil, err
	}
	// The antidotes contained in the cache correspond to a certain protocol version.
	// Flush all antidotes if the version doesn't match.
	currentVer := make([]byte, binary.MaxVarintLen64)
	currentVer = currentVer[:binary.PutVarint(currentVer, int64(version))]

	blob, err := db.Get(AntidoteDBVersionKey, nil)
	switch err {
	case leveldb.ErrNotFound:
		// Version not found (i.e. empty cache), insert it
		if err := db.Put(AntidoteDBVersionKey, currentVer, nil); err != nil {
			db.Close()
			return nil, err
		}

	case nil:
		// Version present, flush if different
		if !bytes.Equal(blob, currentVer) {
			db.Close()
			if err = os.RemoveAll(path); err != nil {
				return nil, err
			}
			return newPersistentAntidoteDB(path, version)
		}
	}
	return &AntidoteDB{
		lvl:  db,
		quit: make(chan struct{}),
	}, nil
}

// makeKey generates the leveldb key-blob from a node id and its particular
// field of interest.
func makeKey(account common.Address) []byte {
	return append(AntidoteDBItemPrefix, account[:]...)
}

func (db *AntidoteDB) Put(account common.Address, txHash common.Hash, nonce, blockNumber uint64) {
	key := makeKey(account)
	blob := make([]byte, 32+8+8)
	copy(blob[:32], txHash[:])
	binary.BigEndian.PutUint64(blob[32:40], nonce)
	binary.BigEndian.PutUint64(blob[40:48], blockNumber)
	db.lvl.Put(key, blob, nil)
}

func (db *AntidoteDB) Get(account common.Address) (txHash common.Hash, nonce, blockNumber uint64) {
	key := makeKey(account)
	blob, err := db.lvl.Get(key, nil)
	if err != nil {
		log.Error("AntidoteDB Get", "error", err)
		return common.Hash{}, 0, 0
	}
	copy(txHash[:], blob[:32])
	nonce = binary.BigEndian.Uint64(blob[32:40])
	blockNumber = binary.BigEndian.Uint64(blob[40:48])
	return txHash, nonce, blockNumber
}

func (db *AntidoteDB) close() {
	close(db.quit)
	db.lvl.Close()
}