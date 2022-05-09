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

package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "task"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey              = "taskid"
	defaultTaskListLength = 1000
)

var (
	taskIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: tableKey, Value: 1},
			},
			Unique: true,
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
func New(db drivers.DB) *ModelTask {
	return &ModelTask{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   taskIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelTask) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := util.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
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
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{task}); err != nil {
		return err
	}
	return nil
}

// UpdateTask update task
func (m *ModelTask) UpdateTask(ctx context.Context, task *types.Task) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: task.TaskID,
	})
	//! object all field update, make sure that task
	//! all fields are setting, otherwise some fields
	//! will be override with nil value
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": task})
}

// PatchTask update task partially
func (m *ModelTask) PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: taskID,
	})
	//! we patch fields that need to be updated
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": patchs})
}

// DeleteTask delete task
func (m *ModelTask) DeleteTask(ctx context.Context, taskID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: taskID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetTask get task
func (m *ModelTask) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: taskID,
	})
	task := &types.Task{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// ListTask list clusters
func (m *ModelTask) ListTask(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.Task, error) {
	taskList := make([]types.Task, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultTaskListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &taskList); err != nil {
		return nil, err
	}
	return taskList, nil
}
