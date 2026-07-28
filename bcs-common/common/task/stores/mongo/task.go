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

// Package mongo implements task storage
package mongo

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

const (
	// TableName task table name
	TableName = "task"
	// TableUniqueKey task unique key
	TableUniqueKey = "taskID"
	// TableCustomIndex task custom index
	TableCustomIndex = "index"
	// DefaultTaskListLength default task list length
	DefaultTaskListLength = 1000
)

var (
	taskIndexes = []drivers.Index{
		{
			Name: TableName + "_idx",
			Key: bson.D{
				bson.E{Key: TableUniqueKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: TableName + "_custom_idx",
			Key: bson.D{
				bson.E{Key: TableCustomIndex, Value: 1},
			},
		},
	}
)

// ModelTask database operation for Task
type ModelTask struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create Task model
func New(db drivers.DB, tablePrefix string) iface.Store {
	return &ModelTask{
		tableName: tablePrefix + "_" + TableName,
		indexes:   taskIndexes,
		db:        db,
	}
}

// EnsureTable ensure table
func (m *ModelTask) EnsureTable(ctx context.Context, dst ...any) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := ensureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateTask create Task
func (m *ModelTask) CreateTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task to be created cannot be empty")
	}
	if err := m.EnsureTable(ctx); err != nil {
		return err
	}

	record := toMongoTask(task)
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{record}); err != nil {
		return err
	}
	return nil
}

// UpdateTask update task
func (m *ModelTask) UpdateTask(ctx context.Context, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be empty")
	}

	if err := m.EnsureTable(ctx); err != nil {
		return err
	}

	record := toMongoTask(task)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		TableUniqueKey: record.TaskID,
	})

	//! object all field update, make sure that task
	//! all fields are setting, otherwise some fields
	//! will be override with nil value
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": record})
}

// DeleteTask delete task
func (m *ModelTask) DeleteTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task id cannot be empty")
	}

	if err := m.EnsureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		TableUniqueKey: taskID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetTask get task
func (m *ModelTask) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id cannot be empty")
	}
	if err := m.EnsureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		TableUniqueKey: taskID,
	})
	record := &Task{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, record); err != nil {
		return nil, err
	}
	return toTask(record), nil
}

// ListTask list task
func (m *ModelTask) ListTask(ctx context.Context, opt *iface.ListOption) (*iface.Pagination[types.Task], error) {
	if err := m.EnsureTable(ctx); err != nil {
		return nil, err
	}

	conditions := []*operator.Condition{operator.EmptyCondition}

	if opt.TaskID != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"taskID": opt.TaskID,
		}))
	}
	if opt.TaskType != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"taskType": opt.TaskType,
		}))
	}
	if opt.TaskName != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"taskName": opt.TaskName,
		}))
	}
	if opt.Status != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"status": opt.Status,
		}))
	}
	if opt.Creator != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"creator": opt.Creator,
		}))
	}

	if opt.CreatedGte != nil || opt.CreatedLte != nil {
		timeRange := make(operator.M)
		if opt.CreatedGte != nil {
			timeRange["$gte"] = *opt.CreatedGte
		}
		if opt.CreatedLte != nil {
			timeRange["$lte"] = *opt.CreatedLte
		}
		conditions = append(conditions, operator.NewLeafCondition(operator.And, operator.M{
			"start": timeRange,
		}))
	}

	condition := operator.NewBranchCondition(operator.And, conditions...)

	count, err := m.db.Table(m.tableName).Find(condition).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count tasks failed: %v", err)
	}

	finder := m.db.Table(m.tableName).Find(condition)

	if len(opt.Sort) != 0 {
		finder = finder.WithSort(MapInt2MapIf(opt.Sort))
	} else {
		finder = finder.WithSort(operator.M{"start": -1})
	}

	if opt.Offset > 0 {
		finder = finder.WithStart(opt.Offset)
	}

	if opt.Limit > 0 && opt.Limit <= DefaultTaskListLength {
		finder = finder.WithLimit(opt.Limit)
	} else {
		finder = finder.WithLimit(DefaultTaskListLength)
	}

	records := make([]*Task, 0)
	if err := finder.All(ctx, &records); err != nil {
		return nil, fmt.Errorf("find tasks failed: %v", err)
	}

	tasks := make([]*types.Task, 0, len(records))
	for _, record := range records {
		if task := toTask(record); task != nil {
			tasks = append(tasks, task)
		}
	}

	return &iface.Pagination[types.Task]{
		Count: count,
		Items: tasks,
	}, nil
}
