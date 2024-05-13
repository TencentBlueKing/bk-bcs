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

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/iptables/third_parties/iptables"
)

// nsIPTable used to handler iptables in container. The pid is
// container's PID, it used to do nsenter to container's network namespace.
type nsIPTable struct {
	pid string

	iptablesClient *iptables.IPTables
}

// NewNSIPTable create the ipTable handler for container
func NewNSIPTable(pid string) (*nsIPTable, error) {
	iptablesClient, err := iptables.New(pid)
	if err != nil {
		return nil, err
	}
	return &nsIPTable{
		pid:            pid,
		iptablesClient: iptablesClient,
	}, nil
}

// Exists checks if given rulespec in specified table/chain exists
func (n *nsIPTable) Exists(table, chain string, rulespec ...string) (bool, error) {
	return n.iptablesClient.Exists(table, chain, rulespec...)
}

// Insert inserts rulespec to specified table/chain (in specified pos)
func (n *nsIPTable) Insert(table, chain string, pos int, rulespec ...string) error {
	return n.iptablesClient.Insert(table, chain, pos, rulespec...)
}

// Append appends rulespec to specified table/chain
func (n *nsIPTable) Append(table, chain string, rulespec ...string) error {
	return n.iptablesClient.Append(table, chain, rulespec...)
}

// AppendUnique acts like Append except that it won't add a duplicate
func (n *nsIPTable) AppendUnique(table, chain string, rulespec ...string) error {
	return n.iptablesClient.AppendUnique(table, chain, rulespec...)
}

// Delete removes rulespec in specified table/chain
func (n *nsIPTable) Delete(table, chain string, rulespec ...string) error {
	return n.iptablesClient.Delete(table, chain, rulespec...)
}

// List rules in specified table/chain
func (n *nsIPTable) List(table, chain string) ([]string, error) {
	return n.iptablesClient.List(table, chain)
}

// ListWithCounters rules (with counters) in specified table/chain
func (n *nsIPTable) ListWithCounters(table, chain string) ([]string, error) {
	return n.iptablesClient.ListWithCounters(table, chain)
}

// ListChains returns a slice containing the name of each chain in the specified table.
func (n *nsIPTable) ListChains(table string) ([]string, error) {
	return n.iptablesClient.ListChains(table)
}

// Stats lists rules including the byte and packet counts
func (n *nsIPTable) Stats(table, chain string) ([][]string, error) {
	return n.iptablesClient.Stats(table, chain)
}

// NewChain creates a new chain in the specified table.
// If the chain already exists, it will result in an error.
func (n *nsIPTable) NewChain(table, chain string) error {
	return n.iptablesClient.NewChain(table, chain)
}

// ClearChain flushed (deletes all rules) in the specified table/chain.
// If the chain does not exist, a new one will be created
func (n *nsIPTable) ClearChain(table, chain string) error {
	return n.iptablesClient.ClearChain(table, chain)
}

// RenameChain renames the old chain to the new one.
func (n *nsIPTable) RenameChain(table, oldChain, newChain string) error {
	return n.iptablesClient.RenameChain(table, oldChain, newChain)
}

// DeleteChain deletes the chain in the specified table.
// The chain must be empty
func (n *nsIPTable) DeleteChain(table, chain string) error {
	return n.iptablesClient.DeleteChain(table, chain)
}

// ChangePolicy changes policy on chain to target
func (n *nsIPTable) ChangePolicy(table, chain, target string) error {
	return n.iptablesClient.ChangePolicy(table, chain, target)
}

// HasRandomFully Check if the underlying iptables command supports the --random-fully flag
func (n *nsIPTable) HasRandomFully() bool {
	return n.iptablesClient.HasRandomFully()
}

// GetIptablesVersion Return version components of the underlying iptables command
func (n *nsIPTable) GetIptablesVersion() (int, int, int) {
	return n.iptablesClient.GetIptablesVersion()
}
