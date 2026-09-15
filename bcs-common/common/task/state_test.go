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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/mem"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

func TestIsReadyToStep(t *testing.T) {
	globalStorage = mem.New()

	info := types.TaskInfo{
		TaskType: "example-test",
		TaskName: "example",
		Creator:  "bcs",
	}
	task := types.NewTask(info)
	stepName := "step1"
	state := NewState(task, stepName)
	step, err := state.isReadyToStep(stepName)
	if assert.Error(t, err) {
		assert.True(t, strings.Contains(err.Error(), "not exist"))
		assert.Nil(t, step)
	}

	steps := []*types.Step{
		types.NewStep("step1", "hello"),
	}
	task.Steps = steps
	step, err = state.isReadyToStep(stepName)

	assert.NoError(t, err)
	assert.Equal(t, stepName, step.Name)
	assert.Equal(t, types.TaskStatusRunning, step.Status)
	assert.Equal(t, types.TaskStatusRunning, task.Status)
}

func newIgnoreTestTask(stepNames ...string) *types.Task {
	task := types.NewTask(types.TaskInfo{
		TaskType: "example-test",
		TaskName: "example",
		Creator:  "bcs",
	})
	steps := make([]*types.Step, 0, len(stepNames))
	for _, name := range stepNames {
		steps = append(steps, types.NewStep(name, "hello"))
	}
	task.Steps = steps
	return task
}

// settleStep 走一遍「步骤就绪 -> 步骤收尾」的完整流程
func settleStep(t *testing.T, task *types.Task, stepName string, ignoreMsg string) {
	t.Helper()

	state := NewState(task, stepName)
	step, err := state.isReadyToStep(stepName)
	assert.NoError(t, err)
	assert.NotNil(t, step)
	state.step = step

	if ignoreMsg == "" {
		state.updateStepSuccess(time.Now())
		return
	}
	state.updateStepIgnored(time.Now(), ignoreMsg)
}

func TestIgnoredStepSetsTaskIgnored(t *testing.T) {
	globalStorage = mem.New()
	task := newIgnoreTestTask("step1", "step2")

	// 第一步幂等忽略后任务继续执行, 不提前终结
	settleStep(t, task, "step1", "进程已在运行, 无需重复启动")
	assert.Equal(t, types.TaskStatusIgnored, task.Steps[0].Status)
	assert.Equal(t, "进程已在运行, 无需重复启动", task.Steps[0].Message)
	assert.Equal(t, types.TaskStatusRunning, task.GetStatus())

	// 最后一步正常成功, 任务终态收敛为 IGNORED 而非 SUCCESS
	settleStep(t, task, "step2", "")
	assert.Equal(t, types.TaskStatusSuccess, task.Steps[1].Status)
	assert.Equal(t, types.TaskStatusIgnored, task.GetStatus())
	assert.True(t, types.IsTaskTerminal(task.GetStatus()))
}

func TestAllStepsSuccessStaysSuccess(t *testing.T) {
	globalStorage = mem.New()
	task := newIgnoreTestTask("step1", "step2")

	settleStep(t, task, "step1", "")
	settleStep(t, task, "step2", "")

	assert.Equal(t, types.TaskStatusSuccess, task.GetStatus())
}

func TestIgnoredLastStepSetsTaskIgnored(t *testing.T) {
	globalStorage = mem.New()
	task := newIgnoreTestTask("step1")

	settleStep(t, task, "step1", "进程已停止")

	assert.Equal(t, types.TaskStatusIgnored, task.GetStatus())
	assert.Equal(t, "task finished with ignored steps", task.GetMessage())
}

// 任务被重复投递时, 已忽略的步骤必须视同完成, 否则会被当作失败步骤
func TestReadyStepSkipsIgnored(t *testing.T) {
	globalStorage = mem.New()
	task := newIgnoreTestTask("step1")
	task.Steps[0].SetStatus(types.TaskStatusIgnored)
	task.SetStatus(types.TaskStatusRunning)

	state := NewState(task, "step1")
	step, err := state.isReadyToStep("step1")

	assert.NoError(t, err)
	assert.Nil(t, step)
	assert.Equal(t, types.TaskStatusIgnored, task.GetStatus())
}

func TestIgnoredTaskIsTerminated(t *testing.T) {
	task := newIgnoreTestTask("step1")
	task.SetStatus(types.TaskStatusIgnored)

	assert.True(t, NewState(task, "step1").isTaskTerminated())
}
