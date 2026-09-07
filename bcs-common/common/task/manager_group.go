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

package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/log"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	istore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// ErrGroupNotSupported 当前 Store 未实现 GroupStore, 不支持任务组编排
var ErrGroupNotSupported = errors.New("task group orchestration is not supported by current store")

// GroupStore 返回底层的任务组存储, 未实现时返回 nil
func (m *TaskManager) GroupStore() istore.GroupStore {
	return m.groupStore
}

// EnsureGroupTable 创建任务组相关表
func (m *TaskManager) EnsureGroupTable(ctx context.Context, dst ...any) error {
	if m.groupStore == nil {
		return ErrGroupNotSupported
	}
	return m.groupStore.EnsureGroupTable(ctx, dst...)
}

// Enqueue 把已落库但尚未下发的任务（处于 INITIALIZING 状态）投入执行队列。
func (m *TaskManager) Enqueue(task *types.Task) error {
	if status := task.GetStatus(); status != types.TaskStatusInit {
		return fmt.Errorf("task %s is not pending to dispatch, status: %s", task.GetTaskID(), status)
	}
	return m.dispatchAt(task, "")
}

// DispatchGroup 创建任务组并下发首个栅栏内的任务。
//
// 组内全部任务与编排计划在同一个事务内落库, 之后仅首个栅栏的任务进入执行队列,
// 后续阶段由任务终态回调自动推进。
func (m *TaskManager) DispatchGroup(ctx context.Context, group *types.TaskGroup, tasks []*types.Task) error {
	if m.groupStore == nil {
		return ErrGroupNotSupported
	}
	if group == nil {
		return errors.New("task group is required")
	}
	if err := group.Validate(); err != nil {
		return err
	}
	if len(tasks) == 0 {
		return errors.New("task group has no task")
	}

	if err := bindTasksToGroup(group, tasks); err != nil {
		return err
	}

	now := time.Now()
	group.Start = now
	group.LastUpdate = now
	group.SetStatus(types.TaskStatusRunning).SetMessage("task group running")

	initial := group.InitialStages()
	group.MarkStageDispatched(initial)

	// 创建任务组、阶段与组内全部任务
	if err := m.groupStore.CreateGroup(ctx, group, tasks); err != nil {
		return err
	}

	// 如果没有设置 StartWithPrevious 的 stage，则理论上 stage
	// 如果存在设置了 StartWithPrevious 的阶段, 则需要把它们与首个栅栏内的任务一起投入执行队列
	initialSeqs := make(map[int]struct{}, len(initial))
	for _, stage := range initial {
		initialSeqs[stage.Seq] = struct{}{}
	}

	var firstErr error
	for _, task := range tasks {
		if _, ok := initialSeqs[task.GetStageSeq()]; !ok {
			continue
		}
		// 下发任务到执行队列
		if err := m.Enqueue(task); err != nil {
			log.ERROR.Printf("group[%s] enqueue task[%s] failed: %v", group.GroupID, task.GetTaskID(), err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// bindTasksToGroup 建立任务与阶段的归属关系, 并由框架统计每个阶段的任务总数
func bindTasksToGroup(group *types.TaskGroup, tasks []*types.Task) error {
	stageTotals := make(map[int]int, len(group.Stages))
	for _, task := range tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		// 验证任务所属阶段是否存在
		if _, ok := group.GetStage(task.GetStageSeq()); !ok {
			return fmt.Errorf("task %s refers to unknown stage %d", task.GetTaskID(), task.GetStageSeq())
		}
		// 建立任务与任务组的归属关系
		task.GroupID = group.GroupID
		stageTotals[task.GetStageSeq()]++
	}

	for _, stage := range group.Stages {
		// 统计每个阶段的任务总数
		stage.Total = stageTotals[stage.Seq]
		if stage.Total == 0 {
			return fmt.Errorf("stage %d has no task", stage.Seq)
		}
	}
	// 统计任务组内总任务数
	group.TotalCount = len(tasks)
	return nil
}

// GetGroup 获取任务组
func (m *TaskManager) GetGroup(ctx context.Context, groupID string) (*types.TaskGroup, error) {
	if m.groupStore == nil {
		return nil, ErrGroupNotSupported
	}
	return m.groupStore.GetGroup(ctx, groupID)
}

// ListGroup 查询任务组列表
func (m *TaskManager) ListGroup(ctx context.Context, opt *istore.ListGroupOption) (
	*istore.Pagination[types.TaskGroup], error) {
	if m.groupStore == nil {
		return nil, ErrGroupNotSupported
	}
	return m.groupStore.ListGroup(ctx, opt)
}

// RevokeGroup 撤销任务组: 已下发的任务通过 revoker 中断, 尚未下发的任务直接终结
func (m *TaskManager) RevokeGroup(ctx context.Context, groupID string) error {
	if m.groupStore == nil {
		return ErrGroupNotSupported
	}

	group, err := m.groupStore.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if group.IsTerminal() {
		return nil
	}

	seqs := stageSeqs(group)
	if m.cfg != nil && m.cfg.Revoker != nil {
		running, lErr := m.groupStore.ListStageTaskIDs(ctx, groupID, seqs, []string{types.TaskStatusRunning})
		if lErr != nil {
			return lErr
		}
		for _, taskID := range running {
			if rErr := m.cfg.Revoker.Revoke(ctx, taskID); rErr != nil {
				log.ERROR.Printf("group[%s] revoke task[%s] failed: %v", groupID, taskID, rErr)
			}
		}
	}

	pending, err := m.groupStore.ListStageTaskIDs(ctx, groupID, seqs, types.PendingTaskStatus)
	if err != nil {
		return err
	}
	if _, err = m.groupStore.BatchUpdateTaskStatus(ctx, pending, types.PendingTaskStatus,
		types.TaskStatusRevoked, "task group has been revoked"); err != nil {
		return err
	}

	group.SetStatus(types.TaskStatusRevoked).SetMessage("task group has been revoked")
	group.End = time.Now()
	return m.groupStore.UpdateGroup(ctx, group)
}

// RetryGroup 重试任务组内所有未成功的任务。
//
// 从第一个存在失败任务的阶段开始重置, 已成功的任务不会被重复执行,
// 重置后由对账流程按实际任务状态重新下发。
func (m *TaskManager) RetryGroup(ctx context.Context, groupID string) error {
	if m.groupStore == nil {
		return ErrGroupNotSupported
	}

	group, err := m.groupStore.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}

	fromSeq, ok := firstRetryableStage(group)
	if !ok {
		return nil
	}

	if err = m.groupStore.ResetGroupStages(ctx, groupID, fromSeq); err != nil {
		return err
	}
	return m.ReconcileGroup(ctx, groupID)
}

// firstRetryableStage 返回第一个需要重试的阶段序号
func firstRetryableStage(group *types.TaskGroup) (int, bool) {
	for _, stage := range group.Stages {
		if stage.Failed > 0 || stage.Skipped > 0 || !stage.IsTerminal() {
			return stage.Seq, true
		}
	}
	return 0, false
}

func stageSeqs(group *types.TaskGroup) []int {
	seqs := make([]int, 0, len(group.Stages))
	for _, stage := range group.Stages {
		seqs = append(seqs, stage.Seq)
	}
	return seqs
}

// ReconcileGroup 以任务存储中的实际终态重建任务组进度, 并补发或补齐后续阶段。
func (m *TaskManager) ReconcileGroup(ctx context.Context, groupID string) error {
	if m.groupStore == nil {
		return ErrGroupNotSupported
	}

	result, err := m.groupStore.ReconcileGroup(ctx, groupID)
	if err != nil {
		return err
	}
	return m.applyGroupAdvance(ctx, result)
}

// tryAdvanceGroup 任务到达终态后推进所属任务组。
//
// 推进失败不影响任务自身的执行结果, 由恢复器按实际任务终态对账补偿,
// 否则会导致已经成功的步骤被重复执行。
func (m *TaskManager) tryAdvanceGroup(task *types.Task) {
	groupID := task.GetGroupID()
	if groupID == "" || m.groupStore == nil || !types.IsTaskTerminal(task.GetStatus()) {
		return
	}

	ctx := context.Background()
	result, err := m.groupStore.AdvanceGroup(ctx, groupID, task.GetTaskID(), task.GetStatus())
	if err != nil {
		log.ERROR.Printf("group[%s] advance by task[%s] failed: %v", groupID, task.GetTaskID(), err)
		return
	}
	if err = m.applyGroupAdvance(ctx, result); err != nil {
		log.ERROR.Printf("group[%s] apply advance failed: %v", groupID, err)
	}
}

// applyGroupAdvance 下发推进结果中待执行的任务, 并触发任务组级回调。
// 下发先于业务回调, 业务回调失败不影响编排继续推进。
func (m *TaskManager) applyGroupAdvance(ctx context.Context, result *istore.GroupAdvanceResult) error {
	if result == nil || result.Group == nil || result.Advance == nil {
		return nil
	}

	var firstErr error
	if err := m.enqueueTaskIDs(ctx, result.NextTaskIDs); err != nil {
		firstErr = err
	}

	cbExecutor := m.groupCallbackExecutors[istep.GroupCallbackName(result.Group.GetCallback())]
	if cbExecutor == nil {
		return firstErr
	}

	c := istep.NewGroupContext(ctx, m.store, result.Group)
	if len(result.BlockedTaskIDs) > 0 {
		if err := cbExecutor.OnStageBlocked(c, result.BlockedTaskIDs); err != nil {
			log.ERROR.Printf("group[%s] OnStageBlocked failed: %v", result.Group.GroupID, err)
		}
	}

	if result.Advance.GroupCompleted {
		m.tryGroupCompleteCallback(ctx, cbExecutor, c, result.Group)
	}
	return firstErr
}

func (m *TaskManager) tryGroupCompleteCallback(ctx context.Context, cbExecutor istep.GroupCallbackExecutor,
	c *istep.GroupContext, group *types.TaskGroup) {

	if err := cbExecutor.OnGroupComplete(c); err != nil {
		log.ERROR.Printf("group[%s] OnGroupComplete failed: %v", group.GroupID, err)
		group.CallbackResult = types.CallbackResultFailure
		group.CallbackMessage = err.Error()
	} else {
		group.CallbackResult = types.CallbackResultSuccess
	}

	if err := m.groupStore.UpdateGroup(ctx, group); err != nil {
		log.ERROR.Printf("group[%s] save callback result failed: %v", group.GroupID, err)
	}
}

// enqueueTaskIDs 读回任务后投入执行队列。
// 必须从存储读回而不是直接使用内存对象, 否则新建任务的零值时间会被写出为非法时间。
func (m *TaskManager) enqueueTaskIDs(ctx context.Context, taskIDs []string) error {
	var firstErr error
	for _, taskID := range taskIDs {
		task, err := GetGlobalStorage().GetTask(ctx, taskID)
		if err != nil {
			log.ERROR.Printf("get task[%s] for enqueue failed: %v", taskID, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err = m.Enqueue(task); err != nil {
			log.ERROR.Printf("enqueue task[%s] failed: %v", taskID, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
