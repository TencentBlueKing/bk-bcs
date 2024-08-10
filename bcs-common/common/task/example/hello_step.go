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

const (
	stepName = "你好"
	method   = "hello"
)

// NewHelloStep hello step
func NewHelloStep() *HelloStep {
	return &HelloStep{}
}

// HelloStep hello step
type HelloStep struct{}

// Alias stepAlias
func (s HelloStep) Alias() string {
	return stepName
}

// GetName method name
func (s HelloStep) GetName() string {
	return method
}

// Execute for worker exec task
func (s HelloStep) Execute(c *istep.Context) error {
	fmt.Printf("%s %s %s\n", c.GetTaskID(), c.GetTaskType(), c.GetTaskName())
	return nil
}

// BuildStep build step
func (s HelloStep) BuildStep(kvs []istep.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}

func init() {
	// register step
	istep.Register(method, NewHelloStep())
}
