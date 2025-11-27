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

// ListClusterWholeTaskMetrics list clusters whole success task metrics
func (m *ModelTask) ListClusterWholeTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.ClusterWholeTaskMetrics, error) {

	taskList := make([]*itypes.ClusterWholeTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	successCond := genClusterWholeSuccessMetrics()
	pipeline := append([]bson.M{cond}, successCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// ListClusterSubSuccessTaskMetrics list clusters sub success task metrics
func (m *ModelTask) ListClusterSubSuccessTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.ClusterSubSuccessTaskMetrics, error) {

	taskList := make([]*itypes.ClusterSubSuccessTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	successCond := genClusterSubSuccessTaskMetrics()
	pipeline := append([]bson.M{cond}, successCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// ListClusterSubFailTaskMetrics list clusters sub fail task metrics
func (m *ModelTask) ListClusterSubFailTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.ClusterSubFailTaskMetrics, error) {

	taskList := make([]*itypes.ClusterSubFailTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	failCond := genClusterSubFailTaskMetrics()
	pipeline := append([]bson.M{cond}, failCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// ListBusinessWholeTaskMetrics list business whole success task metrics
func (m *ModelTask) ListBusinessWholeTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.BusinessWholeTaskMetrics, error) {

	taskList := make([]*itypes.BusinessWholeTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	successCond := genBusinessWholeSuccessMetrics()
	pipeline := append([]bson.M{cond}, successCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// ListBusinessSubSuccessTaskMetrics list business sub success task metrics
func (m *ModelTask) ListBusinessSubSuccessTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.BusinessSubSuccessTaskMetrics, error) {

	taskList := make([]*itypes.BusinessSubSuccessTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	successCond := genBusinessSubSuccessTaskMetrics()
	pipeline := append([]bson.M{cond}, successCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// ListBusinessSubFailTaskMetrics list business sub fail task metrics
func (m *ModelTask) ListBusinessSubFailTaskMetrics(
	ctx context.Context, startTime, endTime string) ([]*itypes.BusinessSubFailTaskMetrics, error) {

	taskList := make([]*itypes.BusinessSubFailTaskMetrics, 0)
	cond := genTimeCondition(startTime, endTime)
	failCond := genBusinessSubFailTaskMetrics()
	pipeline := append([]bson.M{cond}, failCond...)
	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}

// 通过时间条件筛选数据
func genTimeCondition(startTime, endTime string) bson.M {
	cond := bson.M{}

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

// genClusterWholeSuccessMetrics 生成集群维度整体成功数据的子查询条件
func genClusterWholeSuccessMetrics() []bson.M {
	successCond := []bson.M{
		{
			"$group": bson.M{
				"_id": "$clusterid",
				"totalTasks": bson.M{
					"$sum": 1,
				}, // 总任务数
				"successTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": 1,
							"else": 0,
						},
					}, // 成功任务数
				},
				"avgExecutionTime": bson.M{
					"$avg": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": "$executiontime",
							"else": nil,
						},
					}, // 成功任务平均耗时
				},
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"clusterId": "$_id",
				"successRate": bson.M{
					"$divide": []string{"$successTasks", "$totalTasks"},
				}, // 成功率
				"avgExecutionTime": 1,
			},
		},
	}
	return successCond
}

// genBusinessWholeSuccessMetrics 生成查询业务维度整体成功数据的子查询条件
func genBusinessWholeSuccessMetrics() []bson.M {
	successCond := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "bcsclustermanagerv2_cluster",
				"localField":   "clusterid",
				"foreignField": "clusterid",
				"as":           "cluster",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$cluster", // 拆开数组
				"preserveNullAndEmptyArrays": true,       // 保留左表中没有匹配的行
			},
		},
		{
			"$group": bson.M{
				"_id": "$cluster.businessid",
				"totalTasks": bson.M{
					"$sum": 1,
				}, // 总任务数
				"successTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": 1,
							"else": 0,
						},
					}, // 成功任务数
				},
				"avgExecutionTime": bson.M{
					"$avg": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": "$executiontime",
							"else": nil,
						},
					}, // 成功任务平均耗时
				},
			},
		},
		{
			"$project": bson.M{
				"_id":        0,
				"businessId": "$_id",
				"successRate": bson.M{
					"$divide": []string{"$successTasks", "$totalTasks"},
				}, // 成功率
				"avgExecutionTime": 1,
			},
		},
	}
	return successCond
}

// genClusterSubSuccessTaskMetrics 生成集群维度子任务成功数据的查询条件
func genClusterSubSuccessTaskMetrics() []bson.M {
	successCond := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"taskType":  "$tasktype",
					"clusterId": "$clusterid",
				},
				"totalTasks": bson.M{
					"$sum": 1,
				}, // 总任务数
				"successTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": 1,
							"else": 0,
						},
					}, // 成功任务数
				},
				"failTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "FAILURE"}},
							"then": 1,
							"else": 0,
						},
					}, // 失败任务数
				},
				"avgExecutionTime": bson.M{
					"$avg": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": "$executiontime",
							"else": nil,
						},
					}, // 成功任务平均耗时
				},
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"clusterId": "$_id.clusterId",
				"taskType":  "$_id.taskType",
				"successRate": bson.M{
					"$divide": []string{"$successTasks", "$totalTasks"},
				}, // 成功率
				"failTasks":        "$failTasks", // 失败任务数
				"avgExecutionTime": 1,
			},
		},
	}
	return successCond
}

// genBusinessSubSuccessTaskMetrics 生成业务维度子任务成功数据的查询条件
func genBusinessSubSuccessTaskMetrics() []bson.M {
	successCond := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "bcsclustermanagerv2_cluster",
				"localField":   "clusterid",
				"foreignField": "clusterid",
				"as":           "cluster",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$cluster", // 拆开数组
				"preserveNullAndEmptyArrays": true,       // 保留左表中没有匹配的行
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"taskType":   "$tasktype",
					"businessId": "$cluster.businessid",
				},
				"totalTasks": bson.M{
					"$sum": 1,
				}, // 总任务数
				"successTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": 1,
							"else": 0,
						},
					}, // 成功任务数
				},
				"failTasks": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "FAILURE"}},
							"then": 1,
							"else": 0,
						},
					}, // 失败任务数
				},
				"avgExecutionTime": bson.M{
					"$avg": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []string{"$status", "SUCCESS"}},
							"then": "$executiontime",
							"else": nil,
						},
					}, // 成功任务平均耗时
				},
			},
		},
		{
			"$project": bson.M{
				"_id":        0,
				"businessId": "$_id.businessId",
				"taskType":   "$_id.taskType",
				"successRate": bson.M{
					"$divide": []string{"$successTasks", "$totalTasks"},
				}, // 成功率
				"failTasks":        "$failTasks", // 失败任务数
				"avgExecutionTime": 1,
			},
		},
	}
	return successCond
}

// genClusterSubFailTaskMetrics 生成集群维度子任务失败数据的查询条件
func genClusterSubFailTaskMetrics() []bson.M {
	failCond := []bson.M{
		{
			"$match": bson.M{
				"status": "FAILURE",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"taskType":  "$tasktype",
					"clusterId": "$clusterid",
					"message":   "$message",
				},
				"failTasks": bson.M{
					"$sum": 1, // 失败任务数
				},
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"taskType":  "$_id.taskType",
				"clusterId": "$_id.clusterId",
				"message":   "$_id.message",
				"failTasks": 1,
			},
		},
	}
	return failCond
}

// genBusinessSubFailTaskMetrics 生成业务维度子任务失败数据的查询条件
func genBusinessSubFailTaskMetrics() []bson.M {
	failCond := []bson.M{
		{
			"$match": bson.M{
				"status": "FAILURE",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "bcsclustermanagerv2_cluster",
				"localField":   "clusterid",
				"foreignField": "clusterid",
				"as":           "cluster",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$cluster", // 拆开数组
				"preserveNullAndEmptyArrays": true,       // 保留左表中没有匹配的行
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"taskType":   "$tasktype",
					"businessId": "$cluster.businessid",
					"message":    "$message",
				},
				"failTasks": bson.M{
					"$sum": 1, // 失败任务数
				},
			},
		},
		{
			"$project": bson.M{
				"_id":        0,
				"taskType":   "$_id.taskType",
				"businessId": "$_id.businessId",
				"message":    "$_id.message",
				"failTasks":  1,
			},
		},
	}
	return failCond
}
