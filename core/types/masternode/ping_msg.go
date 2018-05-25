package masternode

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethzero/go-ethzero/crypto"
	"github.com/ethzero/go-ethzero/crypto/secp256k1"
	"github.com/ethzero/go-ethzero/log"
	"github.com/ethzero/go-ethzero/p2p/discover"
	"github.com/ethzero/go-ethzero/rlp"
	"math/big"
	"net"
	"time"
)

type PingMsg struct {
	ID      discover.NodeID
	IP      net.IP
	Port    uint16
	SigTime big.Int
	Sig     []byte
}

func (pm *PingMsg) Update(priv *ecdsa.PrivateKey) error {
	b := new(bytes.Buffer)
	pm.Sig = nil
	pm.SigTime.SetInt64(time.Now().Unix())
	if err := rlp.Encode(b, pm); err != nil {
		log.Error("Can't encode PingMsg packet", "err", err)
		return err
	}
	sig, err := crypto.Sign(crypto.Keccak256(b.Bytes()), priv)
	if err != nil {
		log.Error("Can't sign PingMsg packet", "err", err)
		return err
	}
	pm.Sig = sig
	return err
}

func (pm *PingMsg) Check() (id discover.NodeID, err error) {
	b := new(bytes.Buffer)
	sig := pm.Sig
	pm.Sig = nil
	if err := rlp.Encode(b, pm); err != nil {
		return id, err
	}
	hash := crypto.Keccak256(b.Bytes())

	pubkey, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return id, err
	}
	if len(pubkey)-1 != len(id) {
		return id, fmt.Errorf("recovered pubkey has %d bits, want %d bits", len(pubkey)*8, (len(id)+1)*8)
	}
	for i := range id {
		id[i] = pubkey[i+1]
	}
	if pm.ID != id {
		return id, fmt.Errorf("pm.ID != id, %s, %s", pm.ID, id)
	}
	return id, nil
}
