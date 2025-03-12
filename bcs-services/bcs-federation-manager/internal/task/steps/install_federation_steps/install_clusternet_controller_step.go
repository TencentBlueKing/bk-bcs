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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// InstallClusternetControllerStepName step name for create cluster
	InstallClusternetControllerStepName = fedsteps.StepNames{
		Alias: "install bcs-clusternet-controller",
		Name:  "INSTALL_CLUSTERNET_CONTROLLER",
	}
)

// NewInstallClusternetControllerStep sum step
func NewInstallClusternetControllerStep() *InstallClusternetControllerStep {
	return &InstallClusternetControllerStep{}
}

// InstallClusternetControllerStep sum step
type InstallClusternetControllerStep struct{}

// Alias step name
func (s InstallClusternetControllerStep) Alias() string {
	return InstallClusternetControllerStepName.Alias
}

// GetName step name
func (s InstallClusternetControllerStep) GetName() string {
	return InstallClusternetControllerStepName.Name
}

// DoWork for worker exec task
func (s InstallClusternetControllerStep) DoWork(t *types.Task) error {
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

	// create namespace
	var err error
	var ns string
	if helm.GetHelmClient().GetFederationCharts().Controller != nil {
		ns = helm.GetHelmClient().GetFederationCharts().Controller.ReleaseNamespace
	}
	if ns == "" {
		return fmt.Errorf("clusternet-controller release namespace is empty")
	}
	err = cluster.GetClusterClient().CreateNamespace(clusterId, ns)
	if err != nil {
		return err
	}

	// create helm release
	err = helm.GetHelmClient().InstallClusternetController(&helm.ReleaseBaseOptions{
		ProjectID:       projectId,
		ClusterID:       clusterId,
		SkipWhenExisted: true,
	})
	if err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s InstallClusternetControllerStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
