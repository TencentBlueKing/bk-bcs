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

package memory

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// DBTableWatcher db table watcher
type DBTableWatcher struct {
	t                *DBTable
	projection       map[string]int
	batchSize        int32
	isFull           bool
	maxAwaitDuration time.Duration
	startTimestamp   *primitive.Timestamp
	conditions       []*operator.Condition
}

// DBTableFinder db table finder
type DBTableFinder struct {
	t          *DBTable
	condition  *operator.Condition
	projection map[string]int
	start      int64
	limit      int64
}

// DBTable table in memory
type DBTable struct {
	counter int64
	// fake indexes
	indexes map[string]drivers.Index
	// data
	data map[int64]bson.M
}

func newDBTable() *DBTable {
	return &DBTable{
		counter: 0,
		indexes: make(map[string]drivers.Index),
		data:    make(map[int64]bson.M),
	}
}

// DB db in memory
type DB struct {
	dbName string
	data   map[string]*DBTable
}

// NewDB create db
func NewDB(dbName string) *DB {
	return &DB{
		dbName: dbName,
		data:   make(map[string]*DBTable),
	}
}

// DataBase get database
func (db *DB) DataBase() string {
	return db.dbName
}

// Close close
func (db *DB) Close() error {
	return nil
}

// Ping ping
func (db *DB) Ping() error {
	return nil
}

// CreateTable create table
func (db *DB) CreateTable(ctx context.Context, tableName string) error {
	_, ok := db.data[tableName]
	if ok {
		return fmt.Errorf("table %s already exists", tableName)
	}
	db.data[tableName] = newDBTable()
	return nil
}

// HasTable if table exists
func (db *DB) HasTable(ctx context.Context, tableName string) (bool, error) {
	_, ok := db.data[tableName]
	return ok, nil
}

// ListTableNames list table names
func (db *DB) ListTableNames(ctx context.Context) ([]string, error) {
	var retList []string
	for key := range db.data {
		retList = append(retList, key)
	}
	return retList, nil
}

// DropTable drop table
func (db *DB) DropTable(ctx context.Context, tableName string) error {
	_, ok := db.data[tableName]
	if !ok {
		return fmt.Errorf("table %s not found", tableName)
	}
	delete(db.data, tableName)
	return nil
}

// Table get table object
func (db *DB) Table(tableName string) drivers.Table {
	dbTable, ok := db.data[tableName]
	if !ok {
		dbTable = newDBTable()
		db.data[tableName] = dbTable
	}
	return dbTable
}

// CreateIndex create index for collectin
func (t *DBTable) CreateIndex(ctx context.Context, idx drivers.Index) error {
	_, ok := t.indexes[idx.Name]
	if ok {
		return fmt.Errorf("index %s already exists", idx.Name)
	}
	t.indexes[idx.Name] = idx
	return nil
}

// DropIndex drop index
func (t *DBTable) DropIndex(ctx context.Context, indexName string) error {
	_, ok := t.indexes[indexName]
	if !ok {
		return fmt.Errorf("index %s does not exists", indexName)
	}
	delete(t.indexes, indexName)
	return nil
}

// HasIndex has index
func (t *DBTable) HasIndex(ctx context.Context, indexName string) (bool, error) {
	_, ok := t.indexes[indexName]
	return ok, nil
}

// Indexes list indexes of table
func (t *DBTable) Indexes(ctx context.Context) ([]drivers.Index, error) {
	var idxArr []drivers.Index
	for _, idx := range t.indexes {
		idxArr = append(idxArr, idx)
	}
	return idxArr, nil
}

// Find return finder
func (t *DBTable) Find(condition *operator.Condition) drivers.Find {
	return &DBTableFinder{
		t:         t,
		condition: condition,
	}
}

// Insert do insert
func (t *DBTable) Insert(ctx context.Context, docs []interface{}) (int, error) {
	for _, doc := range docs {
		bytes, err := bson.Marshal(doc)
		if err != nil {
			return 0, err
		}
		tmpBson := make(bson.M)
		err = bson.Unmarshal(bytes, &tmpBson)
		if err != nil {
			return 0, err
		}
		time.Now().UnixNano()
		t.data[t.counter] = tmpBson
		t.counter++
	}
	return 0, nil
}

// Update do update
func (t *DBTable) Update(ctx context.Context, condition *operator.Condition, data interface{}) error {
	return nil
}

// UpdateMany do update manay
func (t *DBTable) UpdateMany(ctx context.Context, condition *operator.Condition, data interface{}) (int64, error) {
	return 0, nil
}

// Upsert do upsert
func (t *DBTable) Upsert(ctx context.Context, condition *operator.Condition, data interface{}) error {
	return nil
}

// Delete do delete
func (t *DBTable) Delete(ctx context.Context, condition *operator.Condition) (int64, error) {
	return 0, nil
}

// Watch watch data
func (t *DBTable) Watch(conditions []*operator.Condition) drivers.Watch {
	return &DBTableWatcher{}
}

// WithProjection with projection
func (f *DBTableFinder) WithProjection(project map[string]int) drivers.Find {
	return f
}

// WithSort set sort order
func (f *DBTableFinder) WithSort(sort map[string]interface{}) drivers.Find {
	return f
}

// WithStart set start offset
func (f *DBTableFinder) WithStart(start int64) drivers.Find {
	return f
}

// WithLimit set start offset
func (f *DBTableFinder) WithLimit(start int64) drivers.Find {
	return f
}

// One find one
func (f *DBTableFinder) One(ctx context.Context, result interface{}) error {
	return nil
}

// All find all
func (f *DBTableFinder) All(ctx context.Context, result interface{}) error {
	return nil
}

// Count count data
func (f *DBTableFinder) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

// WithBatchSize set batch size
func (w *DBTableWatcher) WithBatchSize(batch int32) drivers.Watch {
	return w
}

// WithFullContent set if watch action returned the full document
func (w *DBTableWatcher) WithFullContent(isFull bool) drivers.Watch {
	return w
}

// WithMaxAwaitTime set the maximum amount of time
func (w *DBTableWatcher) WithMaxAwaitTime(duration time.Duration) drivers.Watch {
	return w
}

// WithStartTimestamp set operation time that watch start
func (w *DBTableWatcher) WithStartTimestamp(timeSec uint32, index uint32) drivers.Watch {
	return w
}

// DoWatch do watch action
func (w *DBTableWatcher) DoWatch(ctx context.Context) (chan *drivers.WatchEvent, error) {
	return nil, nil
}
