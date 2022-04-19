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

package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/watchbus"
)

const (
	storeActionDefaultLimit = 3000

	// database key for delete flag
	// Because mongodb does not support returning deleted data in changestream, we will do soft-deletion the data here
	databaseFieldNameForDeletionFlag = "_isBcsObjectDeleted"
	databaseIndexNameForDeletionFlag = "bcs_object_deletion_flag_idx"

	// set span dbType
	dbType = "mongo"
)

// StoreGetOption option for get action
type StoreGetOption struct {
	Fields         []string
	Sort           map[string]int
	Cond           *operator.Condition
	Offset         int64
	Limit          int64
	IsAllDocuments bool
}

// StorePutOption option for put action
type StorePutOption struct {
	UniqueKey     []string
	Cond          *operator.Condition
	CreateTimeKey string
	UpdateTimeKey string
}

// StoreRemoveOption option for remove action
type StoreRemoveOption struct {
	Cond *operator.Condition
	// IgnoreNotFound if return err when data not found
	IgnoreNotFound bool
}

// Store action for rest request
type Store struct {
	mDriver         drivers.DB
	eventBus        *watchbus.EventBus
	tableCache      mapset.Set
	tableIndexCache map[string]*drivers.Index
	defaultLimit    int64
	doSoftDelete    bool
}

// NewStore create store action
func NewStore(mDriver drivers.DB, eb *watchbus.EventBus) *Store {
	return &Store{
		mDriver:         mDriver,
		eventBus:        eb,
		tableCache:      mapset.NewSet(),
		tableIndexCache: make(map[string]*drivers.Index),
		defaultLimit:    storeActionDefaultLimit,
	}
}

func fieldsToProjection(fields []string) map[string]int {
	projectionMap := make(map[string]int)
	for _, field := range fields {
		projectionMap[field] = 1
	}
	return projectionMap
}

// SetSoftDeletion set store to do soft delet
func (a *Store) SetSoftDeletion(flag bool) {
	a.doSoftDelete = flag
}

// GetDB return db interface
func (a *Store) GetDB() drivers.DB {
	return a.mDriver
}

// Get get something from db according to request
func (a *Store) Get(ctx context.Context, resourceType string, opt *StoreGetOption) ([]operator.M, error) {
	if opt == nil {
		return nil, fmt.Errorf("StoreGetOption cannot be empty")
	}
	if opt.Cond == nil {
		return nil, fmt.Errorf("Cond in StoreGetOption cannot be empty")
	}
	const (
		OperationName   = "storage-Get"
		OperationMethod = "Get"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)

	projection := fieldsToProjection(opt.Fields)
	mList := make([]operator.M, 0)

	findCond := opt.Cond
	if a.doSoftDelete {
		// search for data which is not marked deleted
		delFlagCond := operator.NewLeafCondition(operator.Ne, operator.M{databaseFieldNameForDeletionFlag: true})
		findCond = operator.NewBranchCondition(operator.And, opt.Cond, delFlagCond)
	} else {
		findCond = operator.NewBranchCondition(operator.And, opt.Cond)
	}
	finder := a.mDriver.Table(resourceType).Find(findCond)
	if len(projection) != 0 {
		finder = finder.WithProjection(projection)
	}
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(mapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit != 0 {
		finder = finder.WithLimit(opt.Limit)
	} else {
		if !opt.IsAllDocuments {
			finder = finder.WithLimit(storeActionDefaultLimit)
		}
	}

	if err := finder.All(ctx, &mList); err != nil {
		blog.Errorf("failed to query, err %s", err.Error())
		utils.SetSpanLogTagError(span, err)
		return nil, fmt.Errorf("failed to query, err %s", err.Error())
	}
	retList := make([]operator.M, 0)
	for _, m := range mList {
		tmpM := dollarRecover(m)
		// remove delete flag in returned result
		if a.doSoftDelete {
			if _, found := tmpM[databaseFieldNameForDeletionFlag]; found {
				delete(tmpM, databaseFieldNameForDeletionFlag)
			}
		}
		retList = append(retList, tmpM)
	}
	return retList, nil
}

// GetIndex get indexes of table
func (a *Store) GetIndex(ctx context.Context, resourceType string) (*drivers.Index, error) {
	const (
		OperationName   = "storage-GetIndex"
		OperationMethod = "GetIndex"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)

	err := a.ensureTable(ctx, resourceType, drivers.Index{})
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return nil, err
	}
	// check index cache
	if index, ok := a.tableIndexCache[resourceType]; ok {
		return index, nil
	}

	// did not hit cache
	indexes, err := a.mDriver.Table(resourceType).Indexes(ctx)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return nil, err
	}
	// build index
	if len(indexes) != 0 {
		for i, index := range indexes {
			if index.Unique {
				a.tableIndexCache[resourceType] = &indexes[i]
				break
			}
		}
	}
	// check index cache
	if index, ok := a.tableIndexCache[resourceType]; ok {
		return index, nil
	}
	// did not have unique index
	return nil, nil
}

// CreateIndex create single index for table
func (a *Store) CreateIndex(ctx context.Context, resourceType string, index drivers.Index) error {
	const (
		OperationName   = "storage-CreateIndex"
		OperationMethod = "CreateIndex"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)
	err := a.ensureTable(ctx, resourceType, index)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}

	return nil
}

// DeleteIndex delete single index for table
func (a *Store) DeleteIndex(ctx context.Context, resourceType string, indexName string) error {
	const (
		OperationName   = "storage-DeleteIndex"
		OperationMethod = "DeleteIndex"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)
	err := a.mDriver.Table(resourceType).DropIndex(ctx, indexName)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}
	return nil
}

// Get get something from db according to request
func (a *Store) Count(ctx context.Context, resourceType string, opt *StoreGetOption) (int64, error) {
	const (
		OperationName   = "storage-Count"
		OperationMethod = "Count"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)

	if opt == nil {
		err := fmt.Errorf("StoreGetOption cannot be empty")
		utils.SetSpanLogTagError(span, err)
		return 0, err
	}
	if opt.Cond == nil {
		err := fmt.Errorf("Cond in StoreGetOption cannot be empty")
		utils.SetSpanLogTagError(span, err)
		return 0, err
	}
	var countCond *operator.Condition
	if a.doSoftDelete {
		// search for data which is not marked deleted
		delFlagCond := operator.NewLeafCondition(operator.Ne, operator.M{databaseFieldNameForDeletionFlag: true})
		countCond = operator.NewBranchCondition(operator.And, opt.Cond, delFlagCond)
	} else {
		countCond = operator.NewBranchCondition(operator.And, opt.Cond)
	}
	projection := fieldsToProjection(opt.Fields)
	finder := a.mDriver.Table(resourceType).Find(countCond)
	if len(projection) != 0 {
		finder = finder.WithProjection(projection)
	}
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(mapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit != 0 {
		finder = finder.WithLimit(opt.Limit)
	} else {
		if !opt.IsAllDocuments {
			finder = finder.WithLimit(storeActionDefaultLimit)
		}
	}
	count, err := finder.Count(ctx)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("failed to query, err %s", err.Error())
		return 0, fmt.Errorf("failed to query, err %s", err.Error())
	}
	return count, nil
}

func (a *Store) ensureTable(ctx context.Context, tableName string, index drivers.Index) error {
	// find from cache
	if a.tableCache.Contains(tableName) {
		return nil
	}
	// find table from db
	hasTable, err := a.mDriver.HasTable(ctx, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := a.mDriver.CreateTable(ctx, tableName)
		if tErr != nil {
			return tErr
		}
	}
	// only ensure index when index name is not empty
	if len(index.Name) != 0 {
		// find index from db
		hasIndex, err := a.mDriver.Table(tableName).HasIndex(ctx, index.Name)
		if err != nil {
			return err
		}
		if !hasIndex {
			if iErr := a.mDriver.Table(tableName).CreateIndex(ctx, index); iErr != nil {
				return iErr
			}
		}
	}

	// if soft delete flag is true, add index for delete flag field
	if a.doSoftDelete {
		// ensure deletionflag index
		bcsDeletionFlagIndex := drivers.Index{
			Name:   databaseIndexNameForDeletionFlag,
			Unique: false,
			Key: bson.D{
				bson.E{Key: databaseFieldNameForDeletionFlag, Value: 1},
			},
		}
		hasDelIndex, err := a.mDriver.Table(tableName).HasIndex(ctx, bcsDeletionFlagIndex.Name)
		if err != nil {
			return err
		}
		if !hasDelIndex {
			if iErr := a.mDriver.Table(tableName).CreateIndex(ctx, bcsDeletionFlagIndex); iErr != nil {
				return iErr
			}
		}
	}

	a.tableCache.Add(tableName)
	return nil
}

// Put put something into db according to request
func (a *Store) Put(ctx context.Context, resourceType string, data operator.M, opt *StorePutOption) error {
	const (
		OperationName   = "storage-Put"
		OperationMethod = "Put"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)

	if opt == nil {
		err := fmt.Errorf("StorePutOption cannot be empty")
		utils.SetSpanLogTagError(span, err)
		return err
	}

	var index drivers.Index
	if len(opt.UniqueKey) != 0 {
		index.Name = resourceType + "_idx"
		index.Unique = true
		index.Key = bson.D{}
		for _, key := range opt.UniqueKey {
			index.Key = append(index.Key, bson.E{Key: key, Value: 1})
		}
	}

	// ensure table index
	if err := a.ensureTable(ctx, resourceType, index); err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}

	data = dollarHandler(data)

	timeNow := time.Now()
	if opt.Cond == nil {
		data[opt.CreateTimeKey] = timeNow
		data[databaseFieldNameForDeletionFlag] = false
		if _, err := a.mDriver.Table(resourceType).Insert(ctx, []interface{}{data}); err != nil {
			utils.SetSpanLogTagError(span, err)
			return err
		}
		return nil
	}

	countCond := opt.Cond
	if a.doSoftDelete {
		// search for data which is not marked deleted
		delFlagCond := operator.NewLeafCondition(operator.Ne, operator.M{databaseFieldNameForDeletionFlag: true})
		countCond = operator.NewBranchCondition(operator.And, opt.Cond, delFlagCond)
		// overriding the deletion flag value
		data[databaseFieldNameForDeletionFlag] = false
	} else {
		countCond = operator.NewBranchCondition(operator.And, opt.Cond)
	}
	counter, err := a.mDriver.Table(resourceType).Find(countCond).Count(ctx)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}
	if counter == 0 && len(opt.CreateTimeKey) != 0 {
		data[opt.CreateTimeKey] = timeNow
	}
	if len(opt.UpdateTimeKey) != 0 {
		data[opt.UpdateTimeKey] = timeNow
	}
	if err := a.mDriver.Table(resourceType).Upsert(ctx, opt.Cond, operator.M{"$set": data}); err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}
	return nil
}

// Remove remove something from db according to request
func (a *Store) Remove(ctx context.Context, resourceType string, opt *StoreRemoveOption) error {
	const (
		OperationName   = "storage-Remove"
		OperationMethod = "Remove"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	defer span.Finish()
	setDBSpanTags(a, span, resourceType, OperationMethod)

	if opt == nil {
		err := fmt.Errorf("StoreRemoveOption cannot be empty")
		utils.SetSpanLogTagError(span, err)
		return err
	}
	if a.doSoftDelete {
		err := a.doDeleteSoft(ctx, resourceType, opt)
		if err != nil {
			utils.SetSpanLogTagError(span, err)
			return err
		}
		return nil
	}

	err := a.doDelete(ctx, resourceType, opt)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return err
	}
	return nil
}

func (a *Store) doDelete(ctx context.Context, resourceType string, opt *StoreRemoveOption) error {
	deleteCounter, err := a.mDriver.Table(resourceType).Delete(ctx, opt.Cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 && !opt.IgnoreNotFound {
		return storageErr.ResourceDoesNotExist
	}
	return nil
}

func (a *Store) doDeleteSoft(ctx context.Context, resourceType string, opt *StoreRemoveOption) error {
	delFlagCond := operator.NewLeafCondition(operator.Ne, operator.M{databaseFieldNameForDeletionFlag: true})
	delCond := operator.NewBranchCondition(operator.And, opt.Cond, delFlagCond)
	deleteCounter, err := a.mDriver.Table(resourceType).UpdateMany(ctx, delCond, operator.M{"$set": operator.M{
		databaseFieldNameForDeletionFlag: true,
	}})
	if err != nil {
		return err
	}
	if deleteCounter == 0 && !opt.IgnoreNotFound {
		return storageErr.ResourceDoesNotExist
	}
	return nil
}

func mapInt2MapIf(m map[string]int) map[string]interface{} {
	newM := make(map[string]interface{})
	for k, v := range m {
		newM[k] = v
	}
	return newM
}

// EventType event type
type EventType int32

const (
	// Nop no operation event
	Nop EventType = iota
	// Add add event
	Add
	// Del delete event
	Del
	// Chg change event
	Chg
	// SChg self change event
	SChg
	// Brk event
	Brk EventType = -1
)

var (
	eventTypeNames = map[EventType]string{
		Nop:  "EventNop",
		Add:  "EventAdd",
		Del:  "EventDelete",
		Chg:  "EventChange",
		SChg: "EventSelfChange",
		Brk:  "EventWatchBreak",
	}
)

var (
	// EventWatchBreak watch break event
	EventWatchBreak = &Event{Type: Brk, Value: nil}
	// EventWatchBreakBytes watch break event content
	EventWatchBreakBytes, _ = json.Marshal(EventWatchBreak)
)

// Event event of watch
type Event struct {
	Type  EventType  `json:"type"`
	Value operator.M `json:"value"`
}

// StoreWatchOption option for watch action
type StoreWatchOption struct {
	Cond      operator.M
	SelfOnly  bool
	MaxEvents uint
	Timeout   time.Duration
	MustDiff  string
}

func watchMatch(data, cond operator.M) bool {
	for k, v := range cond {
		dataValue, ok := data[k]
		if !ok {
			return false
		}
		if !reflect.DeepEqual(v, dataValue) {
			return false
		}
	}
	return true
}

// Watch watch some resource type
func (a *Store) Watch(ctx context.Context, resourceType string, opt *StoreWatchOption) (chan *Event, error) {
	const (
		OperationName   = "storage-Watch"
		OperationMethod = "Watch"
	)
	span, ctx := utils.StartSpanFromContext(ctx, OperationName)
	setDBSpanTags(a, span, resourceType, OperationMethod)

	id := uuid.New().String()
	utils.SetSpanCommonTag(span, "uuid", id)

	dbEvent := make(chan *drivers.WatchEvent, 100)
	err := a.eventBus.Subscribe(resourceType, id, dbEvent)
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		return nil, err
	}

	retEvent := make(chan *Event, 100)
	go func() {
		defer span.Finish()
		defer a.eventBus.Unsubscribe(resourceType, id)
		eventCounter := 0
		for {
			select {
			case e := <-dbEvent:
				if len(opt.MustDiff) != 0 {
					if e.Type == drivers.EventUpdate && len(e.UpdatedFields) == 0 && len(e.RemovedFields) == 0 {
						blog.V(5).Infof("watcher %s of topic %s ignore no-diff update event %+v",
							id, resourceType, e)
						continue
					}
				}
				if len(opt.Cond) != 0 {
					if e.Type == drivers.EventAdd || e.Type == drivers.EventUpdate || e.Type == drivers.EventDelete {
						if !watchMatch(e.Data, opt.Cond) {
							continue
						}
					}
				}
				switch e.Type {
				case drivers.EventAdd:
					retEvent <- &Event{
						Type:  Add,
						Value: e.Data,
					}
				case drivers.EventUpdate:
					// send delete event when doing soft delete and meeting delete flag
					if a.doSoftDelete {
						if deleteFlagValue, ok := e.UpdatedFields[databaseFieldNameForDeletionFlag]; ok {
							deleteFlag, assertOk := deleteFlagValue.(bool)
							if assertOk && deleteFlag {
								retEvent <- &Event{
									Type:  Del,
									Value: e.Data,
								}
							}
						}
					}
					retEvent <- &Event{
						Type:  Chg,
						Value: e.Data,
					}
				case drivers.EventDelete:
					// ignore delete event when doing soft delete
					if a.doSoftDelete {
						continue
					}
					retEvent <- &Event{
						Type:  Del,
						Value: e.Data,
					}
				case drivers.EventError, drivers.EventClose:
					retEvent <- &Event{
						Type:  Brk,
						Value: e.Data,
					}
					return
				default:
					retEvent <- &Event{
						Type:  Nop,
						Value: e.Data,
					}
				}
				eventCounter++
				if opt.MaxEvents != 0 {
					if uint(eventCounter) >= opt.MaxEvents {
						blog.Infof("watcher %s for topic %s exceeds max event %d", id, resourceType, opt.MaxEvents)
						retEvent <- &Event{
							Type: Brk,
						}
						return
					}
				}
			}
		}
	}()
	return retEvent, nil
}

func setDBSpanTags(s *Store, span opentracing.Span, table string, operation string) {
	utils.SetSpanDBTag(span, utils.DBType, dbType)
	utils.SetSpanDBTag(span, utils.DBDatabase, s.mDriver.DataBase())
	utils.SetSpanDBTag(span, utils.DBTable, table)
	utils.SetSpanDBTag(span, utils.DbOperation, operation)
}
