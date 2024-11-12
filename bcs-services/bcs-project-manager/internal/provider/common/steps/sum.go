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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

const (
	stepSumName = "求和任务"
	sumMethod   = "sum"
)

var (
	// SumA xxx
	SumA task.ParamKey = "sumA"
	// SumB xxx
	SumB task.ParamKey = "sumB"
	// SumC xxx
	SumC task.ParamKey = "sumC"
)

// NewSumStep sum step
func NewSumStep() task.StepBuilder {
	return &sumStep{}
}

// sumStep sum step
type sumStep struct{}

// GetName step name
func (s sumStep) GetName() string {
	return sumMethod
}

// Alias step method
func (s sumStep) Alias() string {
	return sumMethod
}

// DoWork for worker exec task
func (s sumStep) DoWork(task *types.Task) error {
	step, exist := task.GetStep(s.Alias())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", task.TaskID, s.Alias())
	}

	a := step.Params[SumA.String()]
	b := step.Params[SumB.String()]

	a1, _ := strconv.Atoi(a)
	b1, _ := strconv.Atoi(b)

	c := a1 + b1
	task.AddCommonParams(SumC.String(), fmt.Sprintf("%v", c))

	logging.Info("%s %s %s sumC: %v\n", task.GetTaskID(), task.GetTaskType(), step.GetName(), c)

	return nil
}

// BuildStep build step
func (s sumStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.Alias(), s.GetName(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
