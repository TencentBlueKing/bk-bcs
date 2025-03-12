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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// CheckNamespaceQuotaStepName step name for check parameter
	CheckNamespaceQuotaStepName = fedsteps.StepNames{
		Alias: "check namespace quota param",
		Name:  "CHECK_NAMESPACE_QUOTA_PARAM",
	}
)

// NewCheckNamespaceQuotaStep x
func NewCheckNamespaceQuotaStep() *CheckNamespaceQuotaStep {
	return &CheckNamespaceQuotaStep{}
}

// CheckNamespaceQuotaStep x
type CheckNamespaceQuotaStep struct{}

// Alias step name
func (s CheckNamespaceQuotaStep) Alias() string {
	return CheckNamespaceQuotaStepName.Alias
}

// GetName step name
func (s CheckNamespaceQuotaStep) GetName() string {
	return CheckNamespaceQuotaStepName.Name
}

// DoWork for worker exec task
func (s CheckNamespaceQuotaStep) DoWork(t *types.Task) error {
	blog.Infof("check namespace quota parameter task is run")

	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	parameter, ok := step.GetParam(fedsteps.ParameterKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ParameterKey)
	}

	reqMap := make(map[string]string)
	err := json.Unmarshal([]byte(parameter), &reqMap)
	if err != nil {
		return err
	}

	blog.Infof("CheckNamespaceQuotaStep taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(),
		t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s CheckNamespaceQuotaStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
