/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package drivers

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// Index index for database table
type Index struct {
	Key        map[string]int32 `json:"key" bson:"key"`
	Name       string           `json:"name" bson:"name"`
	Unique     bool             `json:"unique" bson:"unique"`
	Background bool             `json:"background" bson:"background"`
}

// DB interface for database
type DB interface {
	// DataBase() get database name
	DataBase() string

	// Ping ping the database server
	Ping() error

	// Close close database client
	Close() error

	// HasTable if table exists
	HasTable(ctx context.Context, tableName string) (bool, error)

	// Table get table interface
	Table(tName string) Table

	// ListTableNames list table names
	ListTableNames(ctx context.Context) ([]string, error)

	// CreateTable create database table
	CreateTable(ctx context.Context, tableName string) error

	// DropTable drop database table
	DropTable(ctx context.Context, tableName string) error
}

// Find interface for find action
type Find interface {
	// WithProjection set returned fields
	WithProjection(projection map[string]int) Find

	// WithSort set sort order
	WithSort(sort map[string]interface{}) Find

	// WithStart set start offset
	WithStart(start int64) Find

	// WithLimit set limit of result
	WithLimit(limit int64) Find

	// One find one data by find option
	One(ctx context.Context, result interface{}) error

	// All find all data by find option
	All(ctx context.Context, result interface{}) error

	// Count count data number by find option
	Count(ctx context.Context) (int64, error)
}

// Timestamp timestamp for database
type Timestamp struct {
	Second uint32
	Index  uint32
}

// WatchEventType type of watch event
type WatchEventType string

const (
	// EventAdd add event
	EventAdd = "add"
	// EventUpdate update event
	EventUpdate = "update"
	// EventDelete delete event
	EventDelete = "delete"
	// EventError error event
	EventError = "error"
	// EventClose close event
	EventClose = "close"
)

// WatchEvent event returned by watch interface
type WatchEvent struct {
	Type           WatchEventType
	ClusterTime    time.Time
	DBName         string
	CollectionName string
	TxnNumber      int64
	Key            map[string]interface{}
	UpdatedFields  map[string]interface{}
	RemovedFields  []string
	Data           operator.M
}

// Watch interface for watch action
type Watch interface {

	// WithBatchSize set the maximum number of documents to be included in each batch returned by the server
	WithBatchSize(batch int32) Watch

	// WithFullContent set if watch action returned the full document
	WithFullContent(isFull bool) Watch

	// WithMaxAwaitTime set the maximum amount of time
	// that the server should wait for new documents to satisfy a tailable cursor query
	WithMaxAwaitTime(duration time.Duration) Watch

	// WithStartTimestamp set operation time that watch start
	// struct {
	// 	T   uint32
	//	I   uint32
	// }
	// the most significant 32 bits are a time_t value (seconds since the Unix epoch)
	// the least significant 32 bits are an incrementing ordinal for operations within a given second.
	WithStartTimestamp(uint32, uint32) Watch

	// DoWatch do watch action
	DoWatch(ctx context.Context) (chan *WatchEvent, error)
}

// Table interface for table in database
type Table interface {
	// CreateIndex create index
	CreateIndex(ctx context.Context, index Index) error

	// DropIndex
	DropIndex(ctx context.Context, indexName string) error

	// HasIndex
	HasIndex(ctx context.Context, indexName string) (bool, error)

	// Indexes
	Indexes(ctx context.Context) ([]Index, error)

	// Find get find object
	Find(condition *operator.Condition) Find

	// Aggregation aggregation operation
	Aggregation(ctx context.Context, pipeline interface{}, result interface{}) error

	// Insert insert many data
	Insert(ctx context.Context, docs []interface{}) (int, error)

	// Update update data by condition
	Update(ctx context.Context, condition *operator.Condition, data interface{}) error

	// UpdateMany update many data by condition
	UpdateMany(ctx context.Context, condition *operator.Condition, data interface{}) (int64, error)

	// Upsert update or insert data by condition
	Upsert(ctx context.Context, condition *operator.Condition, data interface{}) error

	// Delete delete data
	Delete(ctx context.Context, condition *operator.Condition) (int64, error)

	// Watch watch data
	Watch(condition []*operator.Condition) Watch
}
