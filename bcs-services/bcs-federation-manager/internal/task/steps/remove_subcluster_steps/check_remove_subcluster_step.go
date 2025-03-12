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

// Package steps include all steps for federation manager
package steps

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// CheckRemoveSubclusterStepName step name for create cluster
	CheckRemoveSubclusterStepName = fedsteps.StepNames{
		Alias: "check for remove subclsuter",
		Name:  "CHECK_REMOVE_SUBCLUSTER",
	}
)

// NewCheckRemoveSubclusterStep sum step
func NewCheckRemoveSubclusterStep() *CheckRemoveSubclusterStep {
	return &CheckRemoveSubclusterStep{}
}

// CheckRemoveSubclusterStep sum step
type CheckRemoveSubclusterStep struct{}

// Alias step name
func (s CheckRemoveSubclusterStep) Alias() string {
	return CheckRemoveSubclusterStepName.Alias
}

// GetName step name
func (s CheckRemoveSubclusterStep) GetName() string {
	return CheckRemoveSubclusterStepName.Name
}

// DoWork for worker exec task
func (s CheckRemoveSubclusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// nothing todo
	// all params is set when build task

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s CheckRemoveSubclusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
