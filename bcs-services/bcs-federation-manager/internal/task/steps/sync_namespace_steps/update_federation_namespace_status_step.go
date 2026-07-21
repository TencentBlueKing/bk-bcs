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
	"time"

	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// UpdateFederationNamespaceStatusStepName step name for create cluster
	UpdateFederationNamespaceStatusStepName = fedsteps.StepNames{
		Alias: "update federation namespace status",
		Name:  "UPDATE_FEDERATION_NAMESPACE_STATUS",
	}
)

// NewUpdateFederationNamespaceStatusStep new step for UpdateFederationNamespaceStatus
// NOCC:tosa/fn_length(设计如此)
func NewUpdateFederationNamespaceStatusStep() *UpdateFederationNamespaceStatusStep {
	return &UpdateFederationNamespaceStatusStep{}
}

// UpdateFederationNamespaceStatusStep x
type UpdateFederationNamespaceStatusStep struct{}

// Alias step name
func (s UpdateFederationNamespaceStatusStep) Alias() string {
	return UpdateFederationNamespaceStatusStepName.Alias
}

// GetName step name
func (s UpdateFederationNamespaceStatusStep) GetName() string {
	return UpdateFederationNamespaceStatusStepName.Name
}

// DoWork for worker exec task
func (s UpdateFederationNamespaceStatusStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	hostClusterID, ok := step.GetParam(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	namespace, ok := step.GetParam(fedsteps.NamespaceKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.NamespaceKey)
	}

	blog.Infof("UpdateFederationNamespaceStatusStep is running, taskId: %s, hostClusterID: %s, namespace: %s",
		t.GetTaskID(), hostClusterID, namespace)

	if err := retry.Do(func() error {
		federationClusterNamespace, err := cluster.GetClusterClient().GetNamespace(hostClusterID, namespace)
		if err != nil {
			blog.Errorf("UpdateFederationNamespaceStatusStep GetNamespace failed, taskId: %s, "+
				"hostClusterID: %s, namespace: %s, err: %s",
				t.GetTaskID(), hostClusterID, namespace, err.Error())
			return err
		}

		if federationClusterNamespace.Annotations == nil {
			federationClusterNamespace.Annotations = make(map[string]string)
		}
		federationClusterNamespace.Annotations[cluster.CreateNamespaceTaskId] = t.GetTaskID()
		federationClusterNamespace.Annotations[cluster.HostClusterNamespaceStatus] = cluster.NamespaceSuccess
		federationClusterNamespace.Annotations[cluster.NamespaceUpdateTimestamp] = time.Now().Format(time.RFC3339)
		err = cluster.GetClusterClient().UpdateNamespace(hostClusterID, federationClusterNamespace)
		if err != nil {
			blog.Errorf("UpdateFederationNamespaceStatusStep UpdateNamespace failed, taskId: %s, "+
				"hostClusterID: %s, namespace: %s, err: %s",
				t.GetTaskID(), hostClusterID, namespace, err.Error())
			return err
		}

		return nil
	}, retry.Attempts(fedsteps.DefaultAttemptTimes), retry.Delay(fedsteps.DefaultRetryDelay*time.Minute),
		retry.DelayType(retry.BackOffDelay), retry.MaxDelay(fedsteps.DefaultMaxDelay*time.Minute)); err != nil {
		return err
	}

	blog.Infof("UpdateFederationNamespaceStatusStep success, taskId: %s, stepName: %s, "+
		"hostClusterID: %s, namespace: %s",
		t.GetTaskID(), step.GetName(), hostClusterID, namespace)
	return nil
}

// BuildStep build step
func (s UpdateFederationNamespaceStatusStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
