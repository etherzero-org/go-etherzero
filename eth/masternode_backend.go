// Copyright 2018 The go-etherzero Authors
// This file is part of the go-etherzero library.
//
// The go-etherzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-etherzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-etherzero library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"sync"

	"github.com/etherzero/go-etherzero/core/types"
	"github.com/etherzero/go-etherzero/common"
	"github.com/ethzero/go-ethzero/event"
)

type MasternodeManager struct {

	votes map[common.Hash]*types.Vote   // vote hash -> vote
	voteFeed event.Feed
	mu         sync.Mutex
}

func NewMasternodeManager() *MasternodeManager{

	// Create the masternode manager with its initial settings
	manager:=&MasternodeManager{
		votes:make(map[common.Hash]*types.Vote),

	}

	return manager
}



func (self *MasternodeManager) AddVote(){}


func (self *MasternodeManager) process(vote *types.Vote){


}

func (self *MasternodeManager) Clear(){
	self.mu.Lock()
	defer self.mu.Unlock()

	self.votes = make(map[common.Hash]*types.Vote)
}

func (self *MasternodeManager) Has(hash common.Hash) bool {
	self.mu.Lock()
	defer self.mu.Unlock()

	if vote := self.votes[hash]; vote != nil {
		return true
	}
	return false
}



