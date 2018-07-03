// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package masternode

import (
	"fmt"
	"testing"
	"time"
)

func Test_rlphash(t *testing.T) {
	startDate_ := time.Now().Unix() + int64(time.Second*10)
	startDate := time.Unix(startDate_, 0).Format("2006-01-02 15:04:05")
	createdTime, _ := time.Parse(startDate, "2006-01-02 15:04:05")

	fmt.Printf("%v", uint64(time.Now().Sub(createdTime)))
}
