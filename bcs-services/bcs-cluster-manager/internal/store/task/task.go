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

// Package task xxx
package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
	itypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

const (
	// TableName xxx
	TableName = "task"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey              = "taskid"
	defaultTaskListLength = 1000

	status = "status"
	start  = "start"
)

var (
	taskIndexes = []drivers.Index{
		{
			Name: TableName + "_idx",
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
		tableName: util.DataTableNamePrefix + TableName,
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
	[]*types.Task, error) {
	taskList := make([]*types.Task, 0)
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

// ListTaskMetrics list clusters task metrics
func (m *ModelTask) ListTaskMetrics(ctx context.Context, clusterId, startTime, endTime string,
	opt *options.ListOption) ([]*itypes.ClusterTaskMetrics, error) {
	taskList := make([]*itypes.ClusterTaskMetrics, 0)
	cond := genMetricsCondition(clusterId, startTime, endTime)
	subCond := m.genSubMetricsCondition()
	pipeline := append([]bson.M{cond}, subCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// DeleteFinishedTaskByDate delete finished task by date
func (m *ModelTask) DeleteFinishedTaskByDate(ctx context.Context, startTime, endTime string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	startCond := operator.NewLeafCondition(operator.Gte, operator.M{start: startTime})
	endCond := operator.NewLeafCondition(operator.Lte, operator.M{start: endTime})

	statusCond := operator.NewLeafCondition(operator.In, operator.M{
		status: []string{common.TaskStatusSuccess, common.TaskStatusFailure, common.TaskStatusTimeout},
	})

	cond := operator.NewBranchCondition(operator.And, statusCond, startCond, endCond)

	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// genSubMetricsCondition 生成子查询条件
func (m *ModelTask) genSubMetricsCondition() []bson.M {
	subCond := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"taskType":  "$tasktype",
					"clusterId": "$clusterid",
				},
				"totalCount": bson.M{
					"$sum": 1,
				}, // 统计总数
				"avgExecutionTime": bson.M{
					"$avg": "$executiontime",
				}, // 耗时平均值
				"successCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": 1,
							"else": 0,
						},
					}, // 成功次数
				},
			},
		},
		{
			"$lookup": bson.M{
				"from": m.tableName,
				"let": bson.M{
					"tt":  "$_id.taskType",
					"cid": "$_id.clusterId",
				},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$and": []bson.M{
									{
										"$eq": []string{"$tasktype", "$$tt"},
									},
									{
										"$eq": []string{"$clusterid", "$$cid"},
									},
									{
										"$eq": []string{"$status", "FAILURE"},
									},
								},
							},
						},
					},
					{
						"$group": bson.M{
							"_id": "$message",
							"failCount": bson.M{
								"$sum": 1,
							},
						},
					},
					{
						"$sort": bson.M{
							"failCount": -1,
						},
					},
					{
						"$limit": 1,
					},
				},
				"as": "topFail", // 失败top原因
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$topFail",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"taskType":  "$_id.taskType",
				"clusterId": "$_id.clusterId",
				"successRate": bson.M{
					"$divide": []string{"$successCount", "$totalCount"}, // 成功率
				},
				"topFailReason":    "$topFail._id",
				"avgExecutionTime": 1,
			},
		},
	}
	return subCond
}

// 通过条件筛选数据
func genMetricsCondition(clusterId, startTime, endTime string) bson.M {
	cond := bson.M{}
	// 通过集群id筛选数据
	if clusterId != "" && clusterId != "-" {
		cond["clusterid"] = clusterId
	}

	// 筛选时间
	if endTime != "" {
		cond["end"] = bson.M{
			"$lte": endTime,
		}
	}
	if startTime != "" {
		cond["start"] = bson.M{
			"$gte": startTime,
		}
	}

	if len(cond) != 0 {
		return bson.M{
			"$match": cond,
		}
	}

	return cond
}
