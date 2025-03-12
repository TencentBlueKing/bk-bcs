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

// Package tasks include all tasks for bcs-federation-manager
package tasks

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	steps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/remove_subcluster_steps"
)

var (
	// RemoveSubclusterTaskName is Remove subcluster task name
	RemoveSubclusterTaskName = TaskNames{
		Name: "remove subcluster from federation cluster",
		Type: "REMOVE_SUBCLUSTER",
	}
)

// NewRemoveSubclusterTask new Remove subcluster task
func NewRemoveSubclusterTask(opt *RemoveSubclusterOptions) *RemoveSubcluster {
	return &RemoveSubcluster{
		opt: opt,
	}
}

// RemoveSubclusterOptions Remove subcluster task options
type RemoveSubclusterOptions struct {
	ProjectId    string
	ClusterId    string
	SubClusterId string
}

// RemoveSubcluster Remove subcluster task
type RemoveSubcluster struct {
	opt *RemoveSubclusterOptions
}

// Name 任务名字
func (i *RemoveSubcluster) Name() string {
	return RemoveSubclusterTaskName.Name
}

// Type 任务类型
func (i *RemoveSubcluster) Type() string {
	return RemoveSubclusterTaskName.Type
}

// Steps build steps for task
func (i *RemoveSubcluster) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	step0 := steps.NewCheckRemoveSubclusterStep().BuildStep([]task.KeyValue{})

	step1 := steps.NewUninstallClusternetAgentStep().BuildStep([]task.KeyValue{})

	step2 := steps.NewUninstallEstimatorAgentStep().BuildStep([]task.KeyValue{})

	step3 := steps.NewRemoveSubclusterStep().BuildStep([]task.KeyValue{})

	// all
	allSteps = append(allSteps, step0, step1, step2, step3)
	return allSteps
}

// BuildTask build task with steps
func (i *RemoveSubcluster) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if i.opt == nil {
		return nil, fmt.Errorf("remove subcluster task options empty")
	}
	if i.opt.ProjectId == "" || i.opt.ClusterId == "" || i.opt.SubClusterId == "" {
		return nil, fmt.Errorf("remove subcluster task options empty, projectId: %s, clusterId: %s,  subclusterId: %s",
			i.opt.ProjectId, i.opt.ClusterId, i.opt.SubClusterId)
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", i.opt.ClusterId, i.opt.SubClusterId),
		TaskType:  i.Type(),
		TaskName:  i.Name(),
		Creator:   creator,
	}, opts...)
	if len(i.Steps()) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range i.Steps() {
		t.Steps[step.GetName()] = step
		t.StepSequence = append(t.StepSequence, step.GetName())
	}
	t.CurrentStep = t.StepSequence[0]

	// federation cluster
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), i.opt.ClusterId)
	if err != nil {
		return nil, fmt.Errorf("get federation cluster from federationmanager failed, err: %s", err.Error())
	}

	// host cluster
	hostCluster, err := cluster.GetClusterClient().GetCluster(context.Background(), fedCluster.HostClusterID)
	if err != nil {
		return nil, fmt.Errorf("get federation cluster from clustermanager failed, err: %s", err.Error())
	}

	// sub cluster ...
	// if sub cluster is not exist, remove task should going on
	// so that we do not need check sub cluster status

	t.AddCommonParams(fedsteps.ProjectIdKey, i.opt.ProjectId).
		AddCommonParams(fedsteps.FedClusterIdKey, fedCluster.FederationClusterID).
		AddCommonParams(fedsteps.HostProjectIdKey, hostCluster.GetProjectID()).
		AddCommonParams(fedsteps.HostClusterIdKey, hostCluster.GetClusterID()).
		AddCommonParams(fedsteps.SubClusterIdKey, i.opt.SubClusterId).
		AddCommonParams(fedsteps.UpdaterKey, creator)

	return t, nil
}
