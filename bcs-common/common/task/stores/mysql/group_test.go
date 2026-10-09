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

package mysql

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// newGroupStore 需要真实 MySQL, 未配置 MYSQL_DSN 时跳过
func newGroupStore(t *testing.T) (iface.Store, iface.GroupStore, context.Context) {
	t.Helper()

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("skip test without mysql dsn")
	}

	store, err := New(dsn)
	require.NoError(t, err)

	groupStore, ok := store.(iface.GroupStore)
	require.True(t, ok, "mysql store should implement GroupStore")

	ctx := context.Background()
	require.NoError(t, store.EnsureTable(ctx))
	require.NoError(t, groupStore.EnsureGroupTable(ctx))

	return store, groupStore, ctx
}

// buildGroup 构造一个 stageTotals 描述的任务组, 返回任务组与按阶段分组的任务
func buildGroup(t *testing.T, stageTotals []int, policies ...types.StageFailurePolicy) (
	*types.TaskGroup, map[int][]*types.Task) {
	t.Helper()

	group := types.NewTaskGroup(types.GroupInfo{
		GroupType: "group-store-test",
		GroupName: "group-store-test",
		Creator:   "bcs",
	})

	tasksByStage := make(map[int][]*types.Task, len(stageTotals))
	all := 0
	for seq, total := range stageTotals {
		policy := types.StageFailureBlock
		if seq < len(policies) {
			policy = policies[seq]
		}
		group.AddStage(&types.Stage{
			Seq:          seq,
			Name:         "stage",
			Total:        total,
			OnFailure:    policy,
			BlockMessage: "blocked by previous stage",
		})

		for i := 0; i < total; i++ {
			task := types.NewTask(types.TaskInfo{TaskType: "t", TaskName: "t", Creator: "bcs"})
			task.Steps = []*types.Step{types.NewStep("step", "noop")}
			task.SetGroup(group.GroupID, seq)
			tasksByStage[seq] = append(tasksByStage[seq], task)
			all++
		}
	}

	group.TotalCount = all
	group.SetStatus(types.TaskStatusRunning)
	group.MarkStageDispatched(group.InitialStages())
	return group, tasksByStage
}

func flatten(tasksByStage map[int][]*types.Task) []*types.Task {
	all := make([]*types.Task, 0)
	for seq := 0; seq < len(tasksByStage); seq++ {
		all = append(all, tasksByStage[seq]...)
	}
	return all
}

// finishTask 模拟任务执行完成: 先写任务终态, 再推进任务组
func finishTask(ctx context.Context, t *testing.T, store iface.Store, groupStore iface.GroupStore,
	task *types.Task, status string) *iface.GroupAdvanceResult {
	t.Helper()

	task.SetStatus(status).SetMessage("done")
	require.NoError(t, store.UpdateTask(ctx, task))

	result, err := groupStore.AdvanceGroup(ctx, task.GetGroupID(), task.GetTaskID(), status)
	require.NoError(t, err)
	return result
}

func TestGroupStoreCreateAndGet(t *testing.T) {
	_, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 3})

	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	got, err := groupStore.GetGroup(ctx, group.GroupID)
	require.NoError(t, err)

	assert.Equal(t, group.GroupID, got.GroupID)
	assert.Equal(t, 5, got.TotalCount)
	require.Len(t, got.Stages, 2)
	assert.Equal(t, 2, got.Stages[0].Total)
	assert.True(t, got.Stages[0].Dispatched)
	assert.False(t, got.Stages[1].Dispatched)
}

func TestGroupStoreCreateIsAtomic(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 2})

	// 让最后一个任务违反 task_id 唯一索引, 触发落库失败
	all := flatten(tasksByStage)
	all[len(all)-1].TaskID = all[0].TaskID

	require.Error(t, groupStore.CreateGroup(ctx, group, all))

	// 任务组、阶段与任何任务都不应残留, 否则已落库的任务会永远停在待执行
	_, err := groupStore.GetGroup(ctx, group.GroupID)
	assert.Error(t, err)

	for _, task := range tasksByStage[0] {
		_, err = store.GetTask(ctx, task.TaskID)
		assert.Error(t, err, "任务 %s 不应被创建", task.TaskID)
	}

	ids, err := groupStore.ListStageTaskIDs(ctx, group.GroupID, []int{0, 1}, nil)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestGroupStoreAdvanceDispatches(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 2})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	// 阶段内第一个任务完成, 阶段未结束
	result := finishTask(ctx, t, store, groupStore, tasksByStage[0][0], types.TaskStatusSuccess)
	assert.Empty(t, result.NextTaskIDs)
	assert.Nil(t, result.Advance.StageCompleted)

	// 阶段内最后一个任务完成, 下发下一阶段
	result = finishTask(ctx, t, store, groupStore, tasksByStage[0][1], types.TaskStatusSuccess)
	assert.NotNil(t, result.Advance.StageCompleted)
	assert.ElementsMatch(t,
		[]string{tasksByStage[1][0].TaskID, tasksByStage[1][1].TaskID},
		result.NextTaskIDs)
	assert.Equal(t, 2, result.Group.SuccessCount)
}

func TestGroupStoreAdvanceIsIdempotent(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 1})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	finishTask(ctx, t, store, groupStore, tasksByStage[0][0], types.TaskStatusSuccess)

	// 同一个任务重复推进不应被重复计数
	again, err := groupStore.AdvanceGroup(ctx, group.GroupID, tasksByStage[0][0].TaskID, types.TaskStatusSuccess)
	require.NoError(t, err)
	assert.Nil(t, again.Advance.StageCompleted)
	assert.Empty(t, again.NextTaskIDs)
	assert.Equal(t, 1, again.Group.SuccessCount)
	assert.Equal(t, 1, again.Group.Stages[0].Completed)
}

func TestGroupStoreBlocksTasksInTx(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{1, 2, 2})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	result := finishTask(ctx, t, store, groupStore, tasksByStage[0][0], types.TaskStatusFailure)

	expectedBlocked := []string{
		tasksByStage[1][0].TaskID, tasksByStage[1][1].TaskID,
		tasksByStage[2][0].TaskID, tasksByStage[2][1].TaskID,
	}
	assert.ElementsMatch(t, expectedBlocked, result.BlockedTaskIDs)
	assert.True(t, result.Advance.GroupCompleted)
	assert.Equal(t, types.TaskStatusFailure, result.Group.GetStatus())
	assert.Equal(t, 5, result.Group.FailureCount)

	// 被阻断的任务已在同一个事务内落库为失败
	for _, taskID := range expectedBlocked {
		task, err := store.GetTask(ctx, taskID)
		require.NoError(t, err)
		assert.Equal(t, types.TaskStatusFailure, task.GetStatus())
		assert.Equal(t, "blocked by previous stage", task.GetMessage())
	}
}

func TestGroupStoreContinuePolicy(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{1, 1}, types.StageFailureContinue)
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	result := finishTask(ctx, t, store, groupStore, tasksByStage[0][0], types.TaskStatusFailure)

	assert.Empty(t, result.BlockedTaskIDs)
	assert.Equal(t, []string{tasksByStage[1][0].TaskID}, result.NextTaskIDs)
}

func TestGroupStoreReconcileAfterCrash(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 2})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	// 模拟崩溃: 任务终态已落库, 但任务组进度从未推进
	for _, task := range tasksByStage[0] {
		task.SetStatus(types.TaskStatusSuccess).SetMessage("done")
		require.NoError(t, store.UpdateTask(ctx, task))
	}

	result, err := groupStore.ReconcileGroup(ctx, group.GroupID)
	require.NoError(t, err)

	assert.Equal(t, 2, result.Group.SuccessCount)
	assert.Equal(t, types.StageStatusSuccess, result.Group.Stages[0].Status)
	assert.ElementsMatch(t,
		[]string{tasksByStage[1][0].TaskID, tasksByStage[1][1].TaskID},
		result.NextTaskIDs)

	// 对账已收敛这些任务, 迟到的回调不会再重复计数
	late, err := groupStore.AdvanceGroup(ctx, group.GroupID, tasksByStage[0][0].TaskID, types.TaskStatusSuccess)
	require.NoError(t, err)
	assert.Equal(t, 2, late.Group.SuccessCount)
}

func TestGroupStoreReconcileCascade(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{1, 2})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	tasksByStage[0][0].SetStatus(types.TaskStatusFailure)
	require.NoError(t, store.UpdateTask(ctx, tasksByStage[0][0]))

	result, err := groupStore.ReconcileGroup(ctx, group.GroupID)
	require.NoError(t, err)

	assert.Len(t, result.BlockedTaskIDs, 2)
	assert.True(t, result.Advance.GroupCompleted)
	assert.Equal(t, types.TaskStatusFailure, result.Group.GetStatus())
}

func TestGroupStoreResetStagesForRetry(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)
	group, tasksByStage := buildGroup(t, []int{2, 1})
	require.NoError(t, groupStore.CreateGroup(ctx, group, flatten(tasksByStage)))

	finishTask(ctx, t, store, groupStore, tasksByStage[0][0], types.TaskStatusSuccess)
	finishTask(ctx, t, store, groupStore, tasksByStage[0][1], types.TaskStatusFailure)

	require.NoError(t, groupStore.ResetGroupStages(ctx, group.GroupID, 0))

	// 成功的任务保持不变, 失败的任务回到待执行
	succeeded, err := store.GetTask(ctx, tasksByStage[0][0].TaskID)
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusSuccess, succeeded.GetStatus())

	failed, err := store.GetTask(ctx, tasksByStage[0][1].TaskID)
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusInit, failed.GetStatus())

	// 重置后对账应重新下发这个失败任务
	result, err := groupStore.ReconcileGroup(ctx, group.GroupID)
	require.NoError(t, err)
	assert.Equal(t, []string{tasksByStage[0][1].TaskID}, result.NextTaskIDs)
}

func TestBatchUpdateFromStatus(t *testing.T) {
	store, groupStore, ctx := newGroupStore(t)

	pending := types.NewTask(types.TaskInfo{TaskType: "t", TaskName: "t", Creator: "bcs"})
	pending.Steps = []*types.Step{types.NewStep("step", "noop")}

	running := types.NewTask(types.TaskInfo{TaskType: "t", TaskName: "t", Creator: "bcs"})
	running.Steps = []*types.Step{types.NewStep("step", "noop")}
	running.SetStatus(types.TaskStatusRunning)

	require.NoError(t, groupStore.BatchCreateTask(ctx, []*types.Task{pending, running}))

	affected, err := groupStore.BatchUpdateTaskStatus(ctx,
		[]string{pending.TaskID, running.TaskID}, types.PendingTaskStatus,
		types.TaskStatusFailure, "canceled")
	require.NoError(t, err)

	// 只有待执行的任务被流转
	assert.Equal(t, int64(1), affected)

	got, err := store.GetTask(ctx, running.TaskID)
	require.NoError(t, err)
	assert.Equal(t, types.TaskStatusRunning, got.GetStatus())
}
