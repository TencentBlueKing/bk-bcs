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

// Package iface defines the interface for store.
package iface

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// ListOption ...
type ListOption struct {
	TaskID        string
	TaskType      string
	TaskName      string
	TaskIndex     string
	TaskIndexType string
	CurrentStep   string
	Status        string
	StatusList    []string // support multiple statuses
	Creator       string
	CreatedGte    *time.Time     // CreatedGte create time greater or equal to
	CreatedLte    *time.Time     // CreatedLte create time less or equal to
	Sort          map[string]int // Sort map for sort list results
	Offset        int64          // Offset offset for list results
	Limit         int64          // Limit limit for list results
	TaskIDs       []string       // TaskIDs task ids for list results
}

// SortField 排序字段。相比 map, 切片能按元素顺序确定多字段排序的优先级
type SortField struct {
	// Field 排序字段名, 必须命中存储层的可排序字段白名单
	Field string
	// Desc 为 true 时降序, 零值为升序
	Desc bool
}

// Pagination generic pagination for list results
type Pagination[T any] struct {
	Count int64 `json:"count"`
	Items []*T  `json:"items"`
}

// PatchOption 主要实时更新params, payload信息
type PatchOption struct {
	Task        *types.Task
	CurrentStep *types.Step
}

// ListGroupOption 任务组列表查询条件
type ListGroupOption struct {
	GroupID        string
	GroupType      string
	GroupName      string
	GroupIndex     string
	GroupIndexType string
	Status         string
	StatusList     []string // support multiple statuses
	Creator        string
	CreatedGte     *time.Time
	CreatedLte     *time.Time
	Sort           []SortField // Sort 多字段排序, 按切片顺序决定优先级
	Offset         int64
	Limit          int64
}

// GroupAdvanceResult 一次任务组推进在存储层的结果
type GroupAdvanceResult struct {
	// Group 推进后的任务组
	Group *types.TaskGroup
	// Advance 阶段级推进结果
	Advance *types.AdvanceResult
	// NextTaskIDs 需要投递到执行队列的任务
	NextTaskIDs []string
	// BlockedTaskIDs 本次因阻断被直接终结的任务
	BlockedTaskIDs []string
}

// GroupStore 任务组编排的持久化能力, 属于可选接口。
//
// Store 实现同时实现本接口后, TaskManager 才支持任务组编排相关 API；
// 未实现的 Store 不受影响, 原有单任务能力保持不变。
type GroupStore interface {
	// EnsureGroupTable 创建任务组相关表
	EnsureGroupTable(ctx context.Context, dst ...any) error

	// BatchCreateTask 在单个事务内批量创建任务, 要么全部成功要么全部回滚
	BatchCreateTask(ctx context.Context, tasks []*types.Task) error

	// BatchUpdateTaskStatus 批量流转任务状态, 仅当任务当前状态命中 fromStatus 才更新。
	// 返回实际发生流转的任务数, 用于调用方判断是否需要连带回收资源。
	BatchUpdateTaskStatus(ctx context.Context, taskIDs []string, fromStatus []string,
		toStatus string, message string) (int64, error)

	// CreateGroup 在单个事务内创建任务组、阶段与组内全部任务,
	// 保证「任务落库」与「编排计划落库」的原子性
	CreateGroup(ctx context.Context, group *types.TaskGroup, tasks []*types.Task) error

	GetGroup(ctx context.Context, groupID string) (*types.TaskGroup, error)
	ListGroup(ctx context.Context, opt *ListGroupOption) (*Pagination[types.TaskGroup], error)
	UpdateGroup(ctx context.Context, group *types.TaskGroup) error

	// AdvanceGroup 依据组内某个任务的终态推进任务组。
	// 计数推进、阶段流转与被阻断任务的批量终结在同一个事务内完成。
	// 同一个任务重复推进只会被计入一次。
	AdvanceGroup(ctx context.Context, groupID, taskID, taskStatus string) (*GroupAdvanceResult, error)

	// ReconcileGroup 以任务存储中的实际终态重建任务组计数。
	// 框架不会自行调用它, 由调用方在重试等明确的时机触发。
	ReconcileGroup(ctx context.Context, groupID string) (*GroupAdvanceResult, error)

	// ResetGroupStages 把 fromSeq 及其之后阶段内所有未成功的任务重置为待执行,
	// 并把这些阶段恢复成未下发状态, 供任务组重试使用。已成功的任务不受影响。
	ResetGroupStages(ctx context.Context, groupID string, fromSeq int) error

	// ListStageTaskIDs 列出任务组指定阶段内的任务 ID, statusList 为空表示不限状态
	ListStageTaskIDs(ctx context.Context, groupID string, stageSeqs []int, statusList []string) ([]string, error)
}

// Store model for TaskManager
type Store interface {
	EnsureTable(ctx context.Context, dst ...any) error
	CreateTask(ctx context.Context, task *types.Task) error
	ListTask(ctx context.Context, opt *ListOption) (*Pagination[types.Task], error)
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	UpdateTask(ctx context.Context, task *types.Task) error
}
