// Copyright 2017 The go-ethzero Authors
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

package core

import(
	"github.com/ethzero/go-ethzero/core/types"
)
// RuleConfig are the configuration parameters of the transaction rule.
type RuleConfig struct {

	AccountTransactionLimit uint64

	AccountTransactionData uint64

	AccountTransationContract uint64

}

//DefaultRuleConfig contains the default configurations for the transaction rule
var DefaultRuleConfig = RuleConfig{

	AccountTransactionLimit:200,

	AccountTransactionData:200,

	AccountTransationContract:300,


}


type Rule interface{

	Validate(tx *types.Transaction, local bool) error
}

type RuleFunc func(tx *types.Transaction, local bool) error

func (rule RuleFunc) Validate(tx *types.Transaction, local bool) error{

	return rule(tx,local)
}

