/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	modelTaskIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: taskIDKey, Value: 1},
				bson.E{Key: isDeletedKey, Value: 1},
			},
			Unique: true,
			Name:   taskTableName + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: strategyKey, Value: 1},
			},
			Name: strategyKey + "_1",
		},
	}
)

// ModelTask defines strategy
type ModelTask struct {
	Public
}

// NewModelTask returns a new ModelTask
func NewModelTask(db drivers.DB) *ModelTask {
	return &ModelTask{Public{
		TableName: tableNamePrefix + taskTableName,
		Indexes:   modelTaskIndexes,
		DB:        db,
	}}
}

// CreateTask create task
func (m *ModelTask) CreateTask(task *storage.ScaleDownTask, opt *storage.CreateOptions) error {
	if opt == nil {
		return fmt.Errorf("CreateOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskIDKey:    task.TaskID,
		isDeletedKey: false,
	})
	retTask := &storage.ScaleDownTask{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retTask); err != nil {
		// 如果查不到，创建
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			task.CreatedTime = time.Now()
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{task})
			if err != nil {
				return fmt.Errorf("task does not exist, insert error: %v", err)
			}
			return nil
		}
		return fmt.Errorf("find task error: %v", err)
	}
	// 如果查到，且opt.OverWriteIfExist为true，更新
	if !opt.OverWriteIfExist {
		return fmt.Errorf("task exists")
	}
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": task}); err != nil {
		return fmt.Errorf("update task error: %v", err)
	}
	return nil
}

// UpdateTask update task
func (m *ModelTask) UpdateTask(task *storage.ScaleDownTask,
	opt *storage.UpdateOptions) (*storage.ScaleDownTask, error) {
	if opt == nil {
		return nil, fmt.Errorf("UpdateOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskIDKey:    task.TaskID,
		isDeletedKey: false,
	})
	retTask := &storage.ScaleDownTask{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retTask); err != nil {
		// 如果找不到且opt.CreateIfNotExist，创建新的
		if errors.Is(err, drivers.ErrTableRecordNotFound) && opt.CreateIfNotExist {
			task.CreatedTime = time.Now()
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{task})
			if err != nil {
				return nil, fmt.Errorf("task does not exist, insert error: %v", err)
			}
			return task, nil
		}
		return nil, fmt.Errorf("find nodeGroupMgrStrategy error: %v", err)
	}
	mergeByte, err := MergePatch(retTask, task, opt.OverwriteZeroOrEmptyStr)
	if err != nil {
		return nil, fmt.Errorf("merge task error:%v", err)
	}
	mergeTask := &storage.ScaleDownTask{}
	err = json.Unmarshal(mergeByte, mergeTask)
	if err != nil {
		return nil, fmt.Errorf("unmarshal mergeTask error:%v", err)
	}
	// 如果查到,更新
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": mergeTask}); err != nil {
		return nil, fmt.Errorf("update task error: %v", err)
	}
	return mergeTask, nil
}

// GetTask get task
func (m *ModelTask) GetTask(taskID string, opt *storage.GetOptions) (*storage.ScaleDownTask, error) {
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskIDKey:    taskID,
		isDeletedKey: opt.GetSoftDeleted,
	})
	retTask := &storage.ScaleDownTask{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retTask); err != nil {
		blog.Infof("task[%s] not exist", taskID)
		// 如果查不到，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) && !opt.ErrIfNotExist {
			return nil, nil
		}
		return nil, fmt.Errorf("find task error: %v", err)
	}
	return retTask, nil
}

// ListTasks list tasks
func (m *ModelTask) ListTasks(opt *storage.ListOptions) ([]*storage.ScaleDownTask, error) {
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	page := opt.Page
	limit := opt.Limit
	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	// 只有不设置分页，且limit为0时，才查询全量，否则依然按照limit和page分页查询
	if !opt.DoPagination && opt.Limit == 0 {
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get task count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	taskList := make([]*storage.ScaleDownTask, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{taskIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &taskList)
	if err != nil {
		return nil, fmt.Errorf("list tasks err:%v", err)
	}
	return taskList, nil
}

// DeleteTask soft delete task
func (m *ModelTask) DeleteTask(taskID string, opt *storage.DeleteOptions) (*storage.ScaleDownTask, error) {
	if opt == nil {
		return nil, fmt.Errorf("DeleteOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskIDKey:    taskID,
		isDeletedKey: false,
	})
	retTask := &storage.ScaleDownTask{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retTask); err != nil {
		// 如果找不到且ErrIfNotExist为true，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			if opt.ErrIfNotExist {
				return nil, fmt.Errorf("task does not exist")
			}
			// 返回nil
			return nil, nil
		}
		// 返回error
		return nil, fmt.Errorf("find task error: %v", err)
	}
	// 如果查到，删除
	retTask.IsDeleted = true
	retTask.UpdatedTime = time.Now()
	if err := m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retTask}); err != nil {
		return nil, fmt.Errorf("soft delete nodeGroupStrategy error: %v", err)
	}
	return retTask, nil
}

// ListTasksByStrategy list tasks by strategy name
func (m *ModelTask) ListTasksByStrategy(strategy string, opt *storage.ListOptions) ([]*storage.ScaleDownTask, error) {
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	page := opt.Page
	limit := opt.Limit
	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		strategyKey:  strategy,
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	// 只有不设置分页，且limit为0时，才查询全量，否则依然按照limit和page分页查询
	if !opt.DoPagination && opt.Limit == 0 {
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get task count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	taskList := make([]*storage.ScaleDownTask, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{taskIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &taskList)
	if err != nil {
		return nil, fmt.Errorf("list tasks err:%v", err)
	}
	return taskList, nil
}
