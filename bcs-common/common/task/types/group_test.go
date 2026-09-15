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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// newGroup 构造一个测试用任务组, totals 按顺序描述每个阶段的任务数
func newGroup(totals ...int) *TaskGroup {
	g := NewTaskGroup(GroupInfo{GroupName: "test-group", GroupType: "test"})
	for i, total := range totals {
		g.AddStage(&Stage{Seq: i, Total: total, OnFailure: StageFailureBlock})
	}
	g.TotalCount = 0
	for _, total := range totals {
		g.TotalCount += total
	}
	g.MarkStageDispatched(g.InitialStages())
	g.SetStatus(TaskStatusRunning)
	return g
}

func seqsOf(stages []*Stage) []int {
	result := make([]int, 0, len(stages))
	for _, s := range stages {
		result = append(result, s.Seq)
	}
	return result
}

func TestGroupValidate(t *testing.T) {
	g := NewTaskGroup(GroupInfo{GroupName: "g"})
	assert.ErrorContains(t, g.Validate(), "stages empty")

	g.AddStage(&Stage{Seq: 1, Total: 1})
	g.AddStage(&Stage{Seq: 1, Total: 1})
	assert.ErrorContains(t, g.Validate(), "ascending")

	g2 := NewTaskGroup(GroupInfo{GroupName: "g"})
	g2.AddStage(&Stage{Seq: 0, Total: 1, OnFailure: "whatever"})
	assert.ErrorContains(t, g2.Validate(), "invalid failure policy")

	g3 := NewTaskGroup(GroupInfo{})
	assert.ErrorContains(t, g3.Validate(), "group name is required")

	g4 := newGroup(1, 1)
	assert.NoError(t, g4.Validate())
}

func TestAddStageDefaults(t *testing.T) {
	g := NewTaskGroup(GroupInfo{GroupName: "g"})
	g.AddStage(&Stage{Seq: 0, Total: 1})

	assert.Equal(t, StageStatusNotStarted, g.Stages[0].Status)
	assert.Equal(t, StageFailureBlock, g.Stages[0].OnFailure)
}

func TestInitialStagesOnlyFirstBarrier(t *testing.T) {
	g := newGroup(2, 3, 1)

	assert.Equal(t, []int{0}, seqsOf(g.InitialStages()))
	assert.True(t, g.Stages[0].Dispatched)
	assert.False(t, g.Stages[1].Dispatched)
	assert.Equal(t, StageStatusRunning, g.Stages[0].Status)
}

func TestAdvanceSingleStageAllSuccess(t *testing.T) {
	g := newGroup(2)

	assert.Nil(t, g.Advance(0, TaskStatusSuccess).StageCompleted)

	result := g.Advance(0, TaskStatusSuccess)
	assert.NotNil(t, result.StageCompleted)
	assert.True(t, result.GroupCompleted)
	assert.Empty(t, result.NextStages)
	assert.Equal(t, TaskStatusSuccess, g.GetStatus())
	assert.Equal(t, 2, g.SuccessCount)
	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
}

func TestAdvanceSerialStages(t *testing.T) {
	g := newGroup(2, 2)

	g.Advance(0, TaskStatusSuccess)
	result := g.Advance(0, TaskStatusSuccess)

	assert.Equal(t, []int{1}, seqsOf(result.NextStages))
	assert.False(t, result.GroupCompleted)
	assert.True(t, g.Stages[1].Dispatched)
	assert.Equal(t, 1, g.CurrentBarrier)

	g.Advance(1, TaskStatusSuccess)
	final := g.Advance(1, TaskStatusSuccess)

	assert.True(t, final.GroupCompleted)
	assert.Equal(t, TaskStatusSuccess, g.GetStatus())
	assert.Equal(t, 4, g.SuccessCount)
}

func TestAdvanceBlocksOnFailure(t *testing.T) {
	g := newGroup(2, 3, 4)
	g.Stages[0].BlockMessage = "priority 100 failed"

	g.Advance(0, TaskStatusSuccess)
	result := g.Advance(0, TaskStatusFailure)

	assert.Equal(t, []int{1, 2}, seqsOf(result.BlockedStages))
	assert.Equal(t, []int{1, 2}, result.BlockedSeqs)
	assert.Equal(t, "priority 100 failed", result.BlockedMessage)
	assert.Equal(t, TaskStatusFailure, result.BlockedTaskStatus)
	assert.Empty(t, result.NextStages)
	assert.True(t, result.GroupCompleted)

	assert.Equal(t, TaskStatusFailure, g.GetStatus())
	// 1 个自身失败 + 后续 3 + 4 个被阻断
	assert.Equal(t, 8, g.FailureCount)
	assert.Equal(t, 1, g.SuccessCount)
	assert.Equal(t, StageStatusFailure, g.Stages[1].Status)
	assert.Equal(t, "priority 100 failed", g.Stages[1].Message)
}

func TestAdvanceContinuePolicyKeepsGoing(t *testing.T) {
	g := newGroup(2, 2)
	g.Stages[0].OnFailure = StageFailureContinue

	g.Advance(0, TaskStatusFailure)
	result := g.Advance(0, TaskStatusSuccess)

	assert.Equal(t, []int{1}, seqsOf(result.NextStages))
	assert.Empty(t, result.BlockedStages)
	assert.False(t, result.GroupCompleted)

	g.Advance(1, TaskStatusSuccess)
	final := g.Advance(1, TaskStatusSuccess)

	assert.True(t, final.GroupCompleted)
	// 有失败任务, 任务组终态为失败
	assert.Equal(t, TaskStatusFailure, g.GetStatus())
	assert.Equal(t, 1, g.FailureCount)
}

func TestAdvanceSkipPolicyRevokesRest(t *testing.T) {
	g := newGroup(1, 3)
	g.Stages[0].OnFailure = StageFailureSkip
	g.Stages[0].BlockMessage = "blocked"

	result := g.Advance(0, TaskStatusFailure)

	assert.Equal(t, TaskStatusRevoked, result.BlockedTaskStatus)
	assert.Equal(t, []int{1}, result.BlockedSeqs)
	assert.Equal(t, StageStatusSkipped, g.Stages[1].Status)
	assert.Equal(t, 3, g.SkippedCount)
	assert.Equal(t, 1, g.FailureCount)
	assert.True(t, result.GroupCompleted)
}

func TestStartWithPreviousSharesBarrier(t *testing.T) {
	g := NewTaskGroup(GroupInfo{GroupName: "g"})
	g.AddStage(&Stage{Seq: 0, Total: 1})
	g.AddStage(&Stage{Seq: 1, Total: 1, StartWithPrevious: true})
	g.AddStage(&Stage{Seq: 2, Total: 1})
	g.MarkStageDispatched(g.InitialStages())

	assert.Equal(t, []int{0, 1}, seqsOf(g.InitialStages()))

	// 同栅栏内另一阶段未完成时不推进
	result := g.Advance(0, TaskStatusSuccess)
	assert.NotNil(t, result.StageCompleted)
	assert.Empty(t, result.NextStages)
	assert.False(t, g.Stages[2].Dispatched)

	result = g.Advance(1, TaskStatusSuccess)
	assert.Equal(t, []int{2}, seqsOf(result.NextStages))
}

func TestStartWithPreviousContinue(t *testing.T) {
	g := NewTaskGroup(GroupInfo{GroupName: "g"})
	// 立即下发且不参与阻断的阶段, 与首个业务阶段同栅栏
	g.AddStage(&Stage{Seq: 0, Total: 1, OnFailure: StageFailureContinue})
	g.AddStage(&Stage{Seq: 1, Total: 1, StartWithPrevious: true, OnFailure: StageFailureBlock})
	g.AddStage(&Stage{Seq: 2, Total: 1, OnFailure: StageFailureBlock})
	g.MarkStageDispatched(g.InitialStages())

	g.Advance(0, TaskStatusFailure)
	result := g.Advance(1, TaskStatusSuccess)

	assert.Equal(t, []int{2}, seqsOf(result.NextStages))
	assert.Empty(t, result.BlockedStages)
}

func TestAdvanceOnTerminalStageIsNoop(t *testing.T) {
	g := newGroup(1, 1)

	g.Advance(0, TaskStatusSuccess)
	before := *g.Stages[0]

	result := g.Advance(0, TaskStatusSuccess)

	assert.Nil(t, result.StageCompleted)
	assert.Equal(t, before.Completed, g.Stages[0].Completed)
	assert.Equal(t, 1, g.SuccessCount)
}

func TestAdvanceUnknownStageIsNoop(t *testing.T) {
	g := newGroup(1)

	result := g.Advance(99, TaskStatusSuccess)

	assert.Nil(t, result.StageCompleted)
	assert.Equal(t, 0, g.SuccessCount)
}

func TestReconcileRebuildsCounts(t *testing.T) {
	g := newGroup(2, 2)
	// 模拟崩溃: 阶段 0 的任务实际已全部成功, 但组计数未更新
	result := g.Reconcile(map[int]*StageCounter{
		0: {Total: 2, Succeeded: 2},
		1: {Total: 2, Pending: 2},
	})

	assert.Equal(t, []int{1}, seqsOf(result.NextStages))
	assert.Equal(t, 2, g.SuccessCount)
	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
	assert.True(t, g.Stages[1].Dispatched)
}

func TestReconcileRedispatchesBarrier(t *testing.T) {
	g := newGroup(2, 2)
	// 模拟崩溃: 首个阶段已落库但从未真正下发
	g.Stages[0].Dispatched = false
	g.Stages[0].Status = StageStatusNotStarted

	result := g.Reconcile(map[int]*StageCounter{
		0: {Total: 2, Pending: 2},
		1: {Total: 2, Pending: 2},
	})

	assert.Equal(t, []int{0}, seqsOf(result.NextStages))
	assert.Equal(t, 0, g.CurrentBarrier)
}

func TestReconcileWaitsWhenRunning(t *testing.T) {
	g := newGroup(3, 2)

	result := g.Reconcile(map[int]*StageCounter{
		0: {Total: 3, Succeeded: 1, Pending: 2},
		1: {Total: 2, Pending: 2},
	})

	assert.Empty(t, result.NextStages)
	assert.Empty(t, result.BlockedStages)
	assert.False(t, result.GroupCompleted)
	assert.Equal(t, StageStatusRunning, g.Stages[0].Status)
}

func TestReconcileBlockedCascade(t *testing.T) {
	g := newGroup(2, 3)
	g.Stages[0].BlockMessage = "blocked"

	// 模拟崩溃: 阶段 0 已失败但级联未落库
	result := g.Reconcile(map[int]*StageCounter{
		0: {Total: 2, Succeeded: 1, Failed: 1},
		1: {Total: 3, Pending: 3},
	})

	assert.Equal(t, []int{1}, result.BlockedSeqs)
	assert.Equal(t, "blocked", result.BlockedMessage)
	assert.True(t, result.GroupCompleted)
	assert.Equal(t, TaskStatusFailure, g.GetStatus())
	assert.Equal(t, 4, g.FailureCount)
}

func TestReconcileIsIdempotent(t *testing.T) {
	g := newGroup(2, 2)
	// 两个阶段都已下发且任务全部成功, 重复对账不应改变结果
	g.MarkStageDispatched(g.Stages)
	counters := map[int]*StageCounter{
		0: {Total: 2, Succeeded: 2},
		1: {Total: 2, Succeeded: 2},
	}

	first := g.Reconcile(counters)
	second := g.Reconcile(counters)

	assert.True(t, first.GroupCompleted)
	assert.True(t, second.GroupCompleted)
	assert.Equal(t, 4, g.SuccessCount)
	assert.Equal(t, TaskStatusSuccess, g.GetStatus())
}

func TestGroupCommonPayload(t *testing.T) {
	type payload struct {
		BatchID uint32 `json:"batchID"`
	}

	g := NewTaskGroup(GroupInfo{GroupName: "g"})
	assert.NoError(t, g.SetCommonPayload(&payload{BatchID: 42}))

	got := &payload{}
	assert.NoError(t, g.GetCommonPayload(got))
	assert.Equal(t, uint32(42), got.BatchID)
}

func TestAdvanceIgnoredNotBlocking(t *testing.T) {
	g := newGroup(2, 1)

	assert.Nil(t, g.Advance(0, TaskStatusIgnored).StageCompleted)
	result := g.Advance(0, TaskStatusSuccess)

	// 忽略不阻断后续阶段, 阶段按成功收尾
	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
	assert.Equal(t, []int{1}, seqsOf(result.NextStages))
	assert.Empty(t, result.BlockedStages)

	// 忽略数独立计数, 既不混入成功数也不计入失败数
	assert.Equal(t, 1, g.Stages[0].Ignored)
	assert.Equal(t, 1, g.Stages[0].Succeeded)
	assert.Equal(t, 0, g.Stages[0].Failed)
	assert.Equal(t, 2, g.Stages[0].Completed)
	assert.Equal(t, 1, g.IgnoredCount)
	assert.Equal(t, 1, g.SuccessCount)
	assert.Equal(t, 0, g.FailureCount)
}

func TestAdvanceAllIgnoredGroupSucceeds(t *testing.T) {
	g := newGroup(2)

	g.Advance(0, TaskStatusIgnored)
	result := g.Advance(0, TaskStatusIgnored)

	assert.True(t, result.GroupCompleted)
	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
	// 任务组终态只有 SUCCESS/FAILURE, 全部忽略时收敛为成功
	assert.Equal(t, TaskStatusSuccess, g.GetStatus())
	assert.Equal(t, 2, g.IgnoredCount)
	assert.Equal(t, 0, g.SuccessCount)
}

func TestReconcileIgnoredCompleted(t *testing.T) {
	g := newGroup(2, 1)

	result := g.Reconcile(map[int]*StageCounter{
		0: {Total: 2, Succeeded: 1, Ignored: 1},
	})

	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
	assert.Equal(t, 2, g.Stages[0].Completed)
	assert.Equal(t, 1, g.Stages[0].Ignored)
	assert.Equal(t, 1, g.IgnoredCount)
	assert.Equal(t, 1, g.SuccessCount)
	// 阶段已完成且无失败, 继续下发下一栅栏
	assert.Equal(t, []int{1}, seqsOf(result.NextStages))
}

func TestReconcileIgnoredNotSkipped(t *testing.T) {
	g := newGroup(2)

	// 一个任务被撤销、一个被忽略: 忽略的任务实际执行过, 阶段不应判为 SKIPPED
	g.Reconcile(map[int]*StageCounter{
		0: {Total: 2, Skipped: 1, Ignored: 1},
	})

	assert.Equal(t, StageStatusSuccess, g.Stages[0].Status)
	assert.Equal(t, 1, g.IgnoredCount)
	assert.Equal(t, 1, g.SkippedCount)
}
