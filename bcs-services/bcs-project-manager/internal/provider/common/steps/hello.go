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

// Package steps xxx
package steps

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

const (
	helloStepName = "你好"
	helloMethod   = "hello"
)

// NewHelloStep hello step
func NewHelloStep() task.StepBuilder {
	return &helloStep{}
}

// helloStep hello step
type helloStep struct{}

// GetName stepName
func (s helloStep) GetName() string {
	return helloMethod
}

// Alias method name
func (s helloStep) Alias() string {
	return helloMethod
}

// DoWork for worker exec task
func (s helloStep) DoWork(task *types.Task) error {
	_, ok := task.GetStep(s.GetName())
	if !ok {
		return fmt.Errorf("task %s step %s not exist", task.GetTaskID(), s.GetName())
	}

	// get step params && handle business logic

	fmt.Printf("%s %s %s\n", task.GetTaskID(), task.GetTaskType(), task.GetTaskName())
	return nil
}

// BuildStep build step
func (s helloStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
