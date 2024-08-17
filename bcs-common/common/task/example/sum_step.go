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
	"strconv"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

const (
	stepSumName = "求和任务"
	sumMethod   = "sum"
)

var (
	sumA istep.ParamKey = "sumA"
	sumB istep.ParamKey = "sumB"
	sumC istep.ParamKey = "sumC"
)

// NewSumStep sum step
func NewSumStep() *SumStep {
	return &SumStep{}
}

// SumStep sum step
type SumStep struct{}

// Alias step name
func (s SumStep) Alias() string {
	return stepSumName
}

// GetName step name
func (s SumStep) GetName() string {
	return sumMethod
}

// Execute for worker exec task
func (s SumStep) Execute(c *istep.Context) error {
	a, _ := c.GetParam(sumA.String())
	b, _ := c.GetParam(sumB.String())

	a1, _ := strconv.Atoi(a)
	b1, _ := strconv.Atoi(b)

	c1 := a1 + b1
	_ = c.AddCommonParam(sumC.String(), fmt.Sprintf("%v", c1))

	fmt.Printf("%s %s %s sumC: %v\n", c.GetTaskID(), c.GetTaskType(), c.GetName(), c)

	return nil
}

// BuildStep build step
func (s SumStep) BuildStep(kvs []istep.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}

func init() {
	// register step
	istep.Register(sumMethod, NewSumStep())
}
