// Copyright 2015 The go-ethzero Authors
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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/ethzero/go-ethzero/metrics"
)

var (
	headerInMeter      = metrics.NewMeter("etz/downloader/headers/in")
	headerReqTimer     = metrics.NewTimer("etz/downloader/headers/req")
	headerDropMeter    = metrics.NewMeter("etz/downloader/headers/drop")
	headerTimeoutMeter = metrics.NewMeter("etz/downloader/headers/timeout")

	bodyInMeter      = metrics.NewMeter("etz/downloader/bodies/in")
	bodyReqTimer     = metrics.NewTimer("etz/downloader/bodies/req")
	bodyDropMeter    = metrics.NewMeter("etz/downloader/bodies/drop")
	bodyTimeoutMeter = metrics.NewMeter("etz/downloader/bodies/timeout")

	receiptInMeter      = metrics.NewMeter("etz/downloader/receipts/in")
	receiptReqTimer     = metrics.NewTimer("etz/downloader/receipts/req")
	receiptDropMeter    = metrics.NewMeter("etz/downloader/receipts/drop")
	receiptTimeoutMeter = metrics.NewMeter("etz/downloader/receipts/timeout")

	stateInMeter   = metrics.NewMeter("etz/downloader/states/in")
	stateDropMeter = metrics.NewMeter("etz/downloader/states/drop")
)
