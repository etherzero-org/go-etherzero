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

package etzclient

import "github.com/ethzero/go-ethzero"

// Verify that Client implements the ethzero interfaces.
var (
	_ = ethzero.ChainReader(&Client{})
	_ = ethzero.TransactionReader(&Client{})
	_ = ethzero.ChainStateReader(&Client{})
	_ = ethzero.ChainSyncReader(&Client{})
	_ = ethzero.ContractCaller(&Client{})
	_ = ethzero.GasEstimator(&Client{})
	_ = ethzero.GasPricer(&Client{})
	_ = ethzero.LogFilterer(&Client{})
	_ = ethzero.PendingStateReader(&Client{})
	// _ = ethzero.PendingStateEventer(&Client{})
	_ = ethzero.PendingContractCaller(&Client{})
)
