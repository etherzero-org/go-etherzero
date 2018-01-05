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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/ethzero/go-ethzero/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewMeter("etz/fetcher/prop/announces/in")
	propAnnounceOutTimer  = metrics.NewTimer("etz/fetcher/prop/announces/out")
	propAnnounceDropMeter = metrics.NewMeter("etz/fetcher/prop/announces/drop")
	propAnnounceDOSMeter  = metrics.NewMeter("etz/fetcher/prop/announces/dos")

	propBroadcastInMeter   = metrics.NewMeter("etz/fetcher/prop/broadcasts/in")
	propBroadcastOutTimer  = metrics.NewTimer("etz/fetcher/prop/broadcasts/out")
	propBroadcastDropMeter = metrics.NewMeter("etz/fetcher/prop/broadcasts/drop")
	propBroadcastDOSMeter  = metrics.NewMeter("etz/fetcher/prop/broadcasts/dos")

	headerFetchMeter = metrics.NewMeter("etz/fetcher/fetch/headers")
	bodyFetchMeter   = metrics.NewMeter("etz/fetcher/fetch/bodies")

	headerFilterInMeter  = metrics.NewMeter("etz/fetcher/filter/headers/in")
	headerFilterOutMeter = metrics.NewMeter("etz/fetcher/filter/headers/out")
	bodyFilterInMeter    = metrics.NewMeter("etz/fetcher/filter/bodies/in")
	bodyFilterOutMeter   = metrics.NewMeter("etz/fetcher/filter/bodies/out")
)
