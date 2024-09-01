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

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/mem"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/stretchr/testify/assert"
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
