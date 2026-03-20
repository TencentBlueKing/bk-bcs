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
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// CheckFederationInstalledStepName step name for create cluster
	CheckFederationInstalledStepName = fedsteps.StepNames{
		Alias: "check federation modules installation",
		Name:  "CHECK_FEDERATION_INSTALLED",
	}
)

// NewCheckFederationInstalledStep sum step
func NewCheckFederationInstalledStep() *CheckFederationInstalledStep {
	return &CheckFederationInstalledStep{}
}

// CheckFederationInstalledStep sum step
type CheckFederationInstalledStep struct{}

// Alias step name
func (s CheckFederationInstalledStep) Alias() string {
	return CheckFederationInstalledStepName.Alias
}

// GetName step name
func (s CheckFederationInstalledStep) GetName() string {
	return CheckFederationInstalledStepName.Name
}

// DoWork for worker exec task
func (s CheckFederationInstalledStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	projectId, ok := t.GetCommonParams(fedsteps.ProjectIdKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist common param projectId", t.TaskID)
	}

	clusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist common param clusterId", t.TaskID)
	}

	isIntalled, err := helm.GetHelmClient().IsInstalledForFederation(context.Background(), &helm.ReleaseBaseOptions{
		ProjectID: projectId,
		ClusterID: clusterId,
	})
	if err != nil {
		return err
	}
	if isIntalled {
		return fmt.Errorf("federation modules are already installed")
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s CheckFederationInstalledStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
