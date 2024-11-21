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

// Package tasks xxx
package tasks

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/common/steps"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
)

/* 测试任务 */

// NewExampleTask build example task
func NewExampleTask(a, b string) task.TaskBuilder {
	return &example{
		a: a,
		b: b,
	}
}

// example task
type example struct {
	a string
	b string
}

// Name 任务名称
func (st *example) Name() string {
	return utils.TestExample.GetTaskName()
}

// Type 任务类型
func (st *example) Type() string {
	return utils.TestExample.GetTaskType(ProviderName)
}

// Steps 构建任务step
func (st *example) Steps(defineSteps []task.StepBuilder) []*types.Step {
	stepList := make([]*types.Step, 0)

	// step1: sum step
	step1 := steps.NewSumStep().BuildStep([]task.KeyValue{
		{
			Key:   steps.SumA,
			Value: st.a,
		},
		{
			Key:   steps.SumB,
			Value: st.b,
		},
	}, types.WithMaxExecutionSeconds(10))

	// step2: hello step
	step2 := steps.NewHelloStep().BuildStep(nil)

	stepList = append(stepList, step1, step2)
	return stepList
}

func (st *example) BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error) {
	t := types.NewTask(&info, opts...)
	if len(st.Steps(nil)) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range st.Steps(nil) {
		t.Steps[step.GetAlias()] = step
		t.StepSequence = append(t.StepSequence, step.GetAlias())
	}
	t.CurrentStep = t.StepSequence[0]

	return t, nil
}
