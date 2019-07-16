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
 *
 */

package operator

import (
	"context"
	"time"
)

type DBInfo struct {
	Addr           []string
	ConnectTimeout time.Duration
	Username       string
	Password       string

	Database     string
	Mode         int
	ListenerName string
	MaxOpenConn  int
	MaxIdleConn  int
}

type ChangeInfo struct {
	Updated int
	Removed int
	Matched int
}

type OperationType string

const (
	None      OperationType = "none"
	Query     OperationType = "query"
	Insert    OperationType = "insert"
	Upsert    OperationType = "upsert"
	Update    OperationType = "update"
	UpdateAll OperationType = "updateAll"
	Remove    OperationType = "remove"
	RemoveAll OperationType = "removeAll"
	Count     OperationType = "count"
	Tables    OperationType = "tables"
	Databases OperationType = "databases"
	SetTableV OperationType = "setTableV"
	GetTableV OperationType = "getTableV"
	Tail      OperationType = "tail"
)

type M map[string]interface{}

func (m M) Update(key string, value interface{}) M {
	m[key] = value
	return m
}

// Tank defined a basic operating unit. It can be called by making a chain of Tank.
// Every block of the chain should be a new struct which is cloned from last one.
//
//  Tank1(filter1)----------> Tank2(filter2)-----------------------------> Tank3(operating)
//  conf1(filter1)----------> conf2(cloned from conf1 and add filter2)---> conf3(cloned from conf2)
//
// So that it can be easy to extend
type Tank interface {
	// Close the connections
	Close()

	// Get the operation result value
	GetValue() []interface{}

	// Get the value length or count num
	GetLen() int

	// Get the changeInfo of update/remove
	GetChangeInfo() *ChangeInfo

	// Get the error if existed
	GetError() error

	// List databases
	Databases() Tank

	// Switch to database, in zk it will be the first layer of tree
	Using(name string) Tank

	// List tables
	Tables() Tank

	// Set a value to table, like a key-value option
	// In tree-like database such as zookeeper, it can be use to set value to a provided path,
	// in others it should be ignored
	SetTableV(data interface{}) Tank

	// Get the value of table, set by SetTableV()
	// in other databases which not support SetTable(), it should be ignored
	GetTableV() Tank

	// From tables, in mongodb it will be the collection, in zk it will be the
	From(name string) Tank

	// Set distinct key
	Distinct(key string) Tank

	// Make the returned value order by key1, key2, key3... and will be reversed if -key1 is given
	OrderBy(key ...string) Tank

	// Set select key
	Select(key ...string) Tank

	// Set offset
	Offset(n int) Tank

	// Set limit
	Limit(n int) Tank

	// Set unique index key
	Index(key ...string) Tank

	// Add filter by *Condition, multi-liner-filter will be combine with "AND"
	Filter(cond *Condition, args ...interface{}) Tank

	// Do the count query
	Count() Tank

	// Do the query according to the filter chain before
	Query(args ...interface{}) Tank

	// Do the insert with data
	Insert(data ...M) Tank

	// Do the update or insert with data according to the filter chain before
	Upsert(data M, args ...interface{}) Tank

	// Do the update with data according to the filter chain before, update the first one
	Update(data M, args ...interface{}) Tank

	// Do the update and update all matched thing
	UpdateAll(data M, args ...interface{}) Tank

	// Do the remove according to the filter chain before, remove the first one
	Remove(args ...interface{}) Tank

	// Do the remove and remove all matched thing
	RemoveAll(args ...interface{}) Tank

	// Watch table then return a chan Event.
	Watch(opts *WatchOptions) (chan *Event, context.CancelFunc)
}

// the method type for getting Tank by providing config name
type GetNewTank func() Tank
