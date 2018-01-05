// Copyright 2016 The go-ethzero Authors
// This file is part of the go-ethzero library.
//
// The go-ethzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethzero library. If not, see <http://www.gnu.org/licenses/>.

package les

import (
	"github.com/ethzero/go-ethzero/metrics"
	"github.com/ethzero/go-ethzero/p2p"
)

var (
	/*	propTxnInPacketsMeter     = metrics.NewMeter("etz/prop/txns/in/packets")
		propTxnInTrafficMeter     = metrics.NewMeter("etz/prop/txns/in/traffic")
		propTxnOutPacketsMeter    = metrics.NewMeter("etz/prop/txns/out/packets")
		propTxnOutTrafficMeter    = metrics.NewMeter("etz/prop/txns/out/traffic")
		propHashInPacketsMeter    = metrics.NewMeter("etz/prop/hashes/in/packets")
		propHashInTrafficMeter    = metrics.NewMeter("etz/prop/hashes/in/traffic")
		propHashOutPacketsMeter   = metrics.NewMeter("etz/prop/hashes/out/packets")
		propHashOutTrafficMeter   = metrics.NewMeter("etz/prop/hashes/out/traffic")
		propBlockInPacketsMeter   = metrics.NewMeter("etz/prop/blocks/in/packets")
		propBlockInTrafficMeter   = metrics.NewMeter("etz/prop/blocks/in/traffic")
		propBlockOutPacketsMeter  = metrics.NewMeter("etz/prop/blocks/out/packets")
		propBlockOutTrafficMeter  = metrics.NewMeter("etz/prop/blocks/out/traffic")
		reqHashInPacketsMeter     = metrics.NewMeter("etz/req/hashes/in/packets")
		reqHashInTrafficMeter     = metrics.NewMeter("etz/req/hashes/in/traffic")
		reqHashOutPacketsMeter    = metrics.NewMeter("etz/req/hashes/out/packets")
		reqHashOutTrafficMeter    = metrics.NewMeter("etz/req/hashes/out/traffic")
		reqBlockInPacketsMeter    = metrics.NewMeter("etz/req/blocks/in/packets")
		reqBlockInTrafficMeter    = metrics.NewMeter("etz/req/blocks/in/traffic")
		reqBlockOutPacketsMeter   = metrics.NewMeter("etz/req/blocks/out/packets")
		reqBlockOutTrafficMeter   = metrics.NewMeter("etz/req/blocks/out/traffic")
		reqHeaderInPacketsMeter   = metrics.NewMeter("etz/req/headers/in/packets")
		reqHeaderInTrafficMeter   = metrics.NewMeter("etz/req/headers/in/traffic")
		reqHeaderOutPacketsMeter  = metrics.NewMeter("etz/req/headers/out/packets")
		reqHeaderOutTrafficMeter  = metrics.NewMeter("etz/req/headers/out/traffic")
		reqBodyInPacketsMeter     = metrics.NewMeter("etz/req/bodies/in/packets")
		reqBodyInTrafficMeter     = metrics.NewMeter("etz/req/bodies/in/traffic")
		reqBodyOutPacketsMeter    = metrics.NewMeter("etz/req/bodies/out/packets")
		reqBodyOutTrafficMeter    = metrics.NewMeter("etz/req/bodies/out/traffic")
		reqStateInPacketsMeter    = metrics.NewMeter("etz/req/states/in/packets")
		reqStateInTrafficMeter    = metrics.NewMeter("etz/req/states/in/traffic")
		reqStateOutPacketsMeter   = metrics.NewMeter("etz/req/states/out/packets")
		reqStateOutTrafficMeter   = metrics.NewMeter("etz/req/states/out/traffic")
		reqReceiptInPacketsMeter  = metrics.NewMeter("etz/req/receipts/in/packets")
		reqReceiptInTrafficMeter  = metrics.NewMeter("etz/req/receipts/in/traffic")
		reqReceiptOutPacketsMeter = metrics.NewMeter("etz/req/receipts/out/packets")
		reqReceiptOutTrafficMeter = metrics.NewMeter("etz/req/receipts/out/traffic")*/
	miscInPacketsMeter  = metrics.NewMeter("les/misc/in/packets")
	miscInTrafficMeter  = metrics.NewMeter("les/misc/in/traffic")
	miscOutPacketsMeter = metrics.NewMeter("les/misc/out/packets")
	miscOutTrafficMeter = metrics.NewMeter("les/misc/out/traffic")
)

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsMeter, miscInTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsMeter, miscOutTrafficMeter
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
