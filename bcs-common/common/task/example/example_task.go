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

// Package main xxx
package main

import (
	"fmt"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

/******************************************************************
************ 构建演示任务 ***********
******************************************************************/

var (
	// ExampleTask task
	ExampleTask istep.TaskName = "测试任务"
	// TestTask task for test
	TestTask istep.TaskType = "TestTask"
)

// NewExampleTask build example task
func NewExampleTask(a, b string) *Example {
	return &Example{
		a: a,
		b: b,
	}
}

// Example task
type Example struct {
	a string
	b string
}

// Name 任务名称
func (st *Example) Name() string {
	return ExampleTask.String()
}

// Type 任务类型
func (st *Example) Type() string {
	return TestTask.String()
}

// Steps 构建任务step
func (st *Example) Steps() []*types.Step {
	steps := make([]*types.Step, 0)

	// step1: sum step
	step1 := SumStep{}.BuildStep([]istep.KeyValue{
		{
			Key:   sumA,
			Value: st.a,
		},
		{
			Key:   sumB,
			Value: st.b,
		},
	}, types.WithMaxExecutionSeconds(10))

	// step2: hello step
	step2 := HelloStep{}.BuildStep(nil)

	steps = append(steps, step1, step2)
	return steps
}

// BuildTask build task
func (st *Example) BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error) {
	t := types.NewTask(info, opts...)
	if len(st.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	t.Steps = st.Steps()
	t.CurrentStep = t.Steps[0].GetName()

	return t, nil
}
