/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iptables

// Interface define iptables operation
type Interface interface {
	Exists(table, chain string, rulespec ...string) (bool, error)
	Insert(table, chain string, pos int, rulespec ...string) error

	//InsertUnique(table string, chain string, pos int, args ...string) error

	Append(table, chain string, rulespec ...string) error
	AppendUnique(table, chain string, rulespec ...string) error
	Delete(table, chain string, rulespec ...string) error
	List(table, chain string) ([]string, error)
	ListWithCounters(table, chain string) ([]string, error)
	ListChains(table string) ([]string, error)
	Stats(table, chain string) ([][]string, error)
	NewChain(table, chain string) error
	ClearChain(table, chain string) error
	RenameChain(table, oldChain, newChain string) error
	DeleteChain(table, chain string) error
	ChangePolicy(table, chain, target string) error
	HasRandomFully() bool
	GetIptablesVersion() (int, int, int)
}
