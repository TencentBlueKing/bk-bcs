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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// RemoveSubclusterStepName step name for create cluster
	RemoveSubclusterStepName = fedsteps.StepNames{
		Alias: "remove subcluster into federation manager",
		Name:  "REMOVE_SUBCLUSTER",
	}
)

// NewRemoveSubclusterStep sum step
func NewRemoveSubclusterStep() *RemoveSubclusterStep {
	return &RemoveSubclusterStep{}
}

// RemoveSubclusterStep sum step
type RemoveSubclusterStep struct{}

// Alias step name
func (s RemoveSubclusterStep) Alias() string {
	return RemoveSubclusterStepName.Alias
}

// GetName step name
func (s RemoveSubclusterStep) GetName() string {
	return RemoveSubclusterStepName.Name
}

// DoWork for worker exec task
func (s RemoveSubclusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// get common params
	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	updater, ok := t.GetCommonParams(fedsteps.UpdaterKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.UpdaterKey)
	}

	// delete mcls for sub
	err := cluster.GetClusterClient().DeleteManagedCluster(hostClusterId, subClusterId)
	if err != nil {
		return fmt.Errorf("delete sub cluster[%s]'s ManagedCluster in cluster[%s] err: %s", subClusterId, hostClusterId, err.Error())
	}

	// delete registration request for sub, if not delete, will cause sub cluster can not be registered again
	err = cluster.GetClusterClient().DeleteClusterRegistrationRequest(hostClusterId, subClusterId)
	if err != nil {
		return fmt.Errorf("delete sub cluster[%s]'s ClusterRegistrationRequest in cluster[%s] err: %s", subClusterId, hostClusterId, err.Error())
	}

	// delete sub cluster in federation manager
	if err := store.GetStoreModel().DeleteSubCluster(context.Background(), &store.SubClusterDeleteOptions{
		FederationClusterID: fedClusterId,
		SubClusterID:        subClusterId,
		Updater:             updater,
	}); err != nil {
		return err
	}

	// delete sub cluster labels
	if err := cluster.GetClusterClient().DeleteSubClusterLabel(context.Background(), subClusterId); err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s RemoveSubclusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
