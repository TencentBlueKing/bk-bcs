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

// Package operationlog xxx
package operationlog

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

const (
	taskStepTableName = "tasksteplog"
)

// ModelTaskStepLog database operation for task_step_log
type ModelTaskStepLog struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

var (
	taskStepLogIndexes = []drivers.Index{
		{
			Name: taskStepTableName + "_idx",
			Key: bson.D{
				bson.E{Key: taskID, Value: 1},
			},
			Unique: false,
		},
	}
)

// NewTaskStepLog create taskStepLog model
func NewTaskStepLog(db drivers.DB) *ModelTaskStepLog {
	return &ModelTaskStepLog{
		tableName: util.DataTableNamePrefix + taskStepTableName,
		db:        db,
		indexes:   taskStepLogIndexes,
	}
}

// ensureTable ensure table
func (m *ModelTaskStepLog) ensureTable(ctx context.Context) error {
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

// CreateTaskStepLogInfo create task step info log
func (m *ModelTaskStepLog) CreateTaskStepLogInfo(ctx context.Context, taskID, stepName, message string) {
	if taskID == "" || stepName == "" || message == "" {
		blog.Error("parameter cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		blog.Error(err.Error())
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{types.TaskStepLog{
		TaskID:     taskID,
		StepName:   stepName,
		Level:      "INFO",
		Message:    message,
		CreateTime: time.Now().Format(time.RFC3339Nano),
	}}); err != nil {
		blog.Error(err.Error())
	}
}

// CreateTaskStepLogWarn create task step warn log
func (m *ModelTaskStepLog) CreateTaskStepLogWarn(ctx context.Context, taskID, stepName, message string) {
	if taskID == "" || stepName == "" || message == "" {
		blog.Error("parameter cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		blog.Error(err.Error())
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{types.TaskStepLog{
		TaskID:     taskID,
		StepName:   stepName,
		Level:      "WARN",
		Message:    message,
		CreateTime: time.Now().Format(time.RFC3339Nano),
	}}); err != nil {
		blog.Error(err.Error())
	}
}

// CreateTaskStepLogError create task step error log
func (m *ModelTaskStepLog) CreateTaskStepLogError(ctx context.Context, taskID, stepName, message string) {
	if taskID == "" || stepName == "" || message == "" {
		blog.Error("parameter cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		blog.Error(err.Error())
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{types.TaskStepLog{
		TaskID:     taskID,
		StepName:   stepName,
		Level:      "ERROR",
		Message:    message,
		CreateTime: time.Now().Format(time.RFC3339Nano),
	}}); err != nil {
		blog.Error(err.Error())
	}
}

// DeleteTaskStepLogByTaskID delete taskStepLog
func (m *ModelTaskStepLog) DeleteTaskStepLogByTaskID(ctx context.Context, taskID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskID: taskID,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		blog.Warnf("no taskStepLog delete with taskID %s", taskID)
	}
	return nil
}

// CountTaskStepLog count taskStepLog
func (m *ModelTaskStepLog) CountTaskStepLog(ctx context.Context, cond *operator.Condition) (
	int64, error) {
	return m.db.Table(m.tableName).Find(cond).Count(ctx)
}

// ListTaskStepLog list logs
func (m *ModelTaskStepLog) ListTaskStepLog(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]*types.TaskStepLog, error) {

	logList := make([]*types.TaskStepLog, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultLogListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if opt.All {
		finder = finder.WithLimit(0)
	}

	if err := finder.All(ctx, &logList); err != nil {
		return nil, err
	}

	return logList, nil
}
