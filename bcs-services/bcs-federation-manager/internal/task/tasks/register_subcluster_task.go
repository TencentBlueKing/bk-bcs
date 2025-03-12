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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	steps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/register_subcluster_steps"
)

var (
	// RegisterSubclusterTaskName is register subcluster task name
	RegisterSubclusterTaskName = TaskNames{
		Name: "register subcluster to federation cluster",
		Type: "REGISTER_SUBCLUSTER",
	}
)

// NewRegisterSubclusterTask new register subcluster task
func NewRegisterSubclusterTask(opt *RegisterSubclusterOptions) *RegisterSubcluster {
	return &RegisterSubcluster{
		opt: opt,
	}
}

// RegisterSubclusterOptions register subcluster task options
type RegisterSubclusterOptions struct {
	ProjectId      string
	ClusterId      string
	SubClusterId   string
	UserToken      string
	GatewayAddress string
}

// RegisterSubcluster register subcluster task
type RegisterSubcluster struct {
	opt *RegisterSubclusterOptions
}

// Name 任务名字
func (i *RegisterSubcluster) Name() string {
	return RegisterSubclusterTaskName.Name
}

// Type 任务类型
func (i *RegisterSubcluster) Type() string {
	return RegisterSubclusterTaskName.Type
}

// Steps build steps for task
func (i *RegisterSubcluster) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	step0 := steps.NewCheckRegisterSubclusterStep().BuildStep([]task.KeyValue{})

	step1 := steps.NewPreRegisterSubclusterStep().BuildStep([]task.KeyValue{})

	step2 := steps.NewInstallClusternetAgentStep().BuildStep([]task.KeyValue{})

	step3 := steps.NewInstallEstimatorAgentStep().BuildStep([]task.KeyValue{})

	step4 := steps.NewRegisterSubclusterStep().BuildStep([]task.KeyValue{})

	// all
	allSteps = append(allSteps, step0, step1, step2, step3, step4)
	return allSteps
}

// BuildTask build task with steps
func (i *RegisterSubcluster) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if i.opt == nil {
		return nil, fmt.Errorf("register subcluster task options empty")
	}
	if i.opt.ProjectId == "" || i.opt.ClusterId == "" || i.opt.UserToken == "" || i.opt.SubClusterId == "" {
		return nil, fmt.Errorf("register subcluster task options empty, projectId: %s, clusterId: %s, userToken: %s, subclusterId: %s",
			i.opt.ProjectId, i.opt.ClusterId, i.opt.UserToken, i.opt.SubClusterId)
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
		return nil, fmt.Errorf("get host cluster from clustermanager failed, err: %s", err.Error())
	}

	// sub cluster
	subCluster, err := cluster.GetClusterClient().GetCluster(context.Background(), i.opt.SubClusterId)
	if err != nil {
		return nil, fmt.Errorf("get sub cluster from clustermanager failed, err: %s", err.Error())
	}

	// sub cluster's project
	subProject, err := project.GetProjectClient().GetProject(context.Background(), subCluster.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("get sub project from projectmanager failed, err: %s", err.Error())
	}

	t.AddCommonParams(fedsteps.ProjectIdKey, i.opt.ProjectId).
		AddCommonParams(fedsteps.FedClusterIdKey, fedCluster.FederationClusterID).
		AddCommonParams(fedsteps.HostProjectIdKey, hostCluster.GetProjectID()).
		AddCommonParams(fedsteps.HostClusterIdKey, hostCluster.GetClusterID()).
		AddCommonParams(fedsteps.SubProjectIdKey, subProject.GetProjectID()).
		AddCommonParams(fedsteps.SubProjectCodeKey, subProject.GetProjectCode()).
		AddCommonParams(fedsteps.SubClusterIdKey, subCluster.GetClusterID()).
		AddCommonParams(fedsteps.BcsGatewayAddressKey, i.opt.GatewayAddress).
		AddCommonParams(fedsteps.UserTokenKey, i.opt.UserToken).
		AddCommonParams(fedsteps.CreatorKey, creator)

	// set callback func
	t.SetCallback(steps.NewRegisterSubclusterCallBack().GetName())
	return t, nil
}

// BeforeRetryRegisterSubCluster retry task
func BeforeRetryRegisterSubCluster(ctx context.Context, t *types.Task) error {

	taskId := t.GetTaskID()

	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		return fmt.Errorf("can not get fedClusterId from task: %s, step: %s", taskId, t.GetCurrentStep())
	}

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fmt.Errorf("can not get subClusterId from task: %s, step: %s", taskId, t.GetCurrentStep())
	}

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fmt.Errorf("can not get creator from task: %s, step: %s", taskId, t.GetCurrentStep())
	}

	subCluster, err := store.GetStoreModel().GetSubCluster(ctx, fedClusterId, subClusterId)
	if err != nil {
		// if subcluster is not exist, do nothing
		return nil
	}
	// if subcluster is exist, update status to creating
	subCluster.Status = store.CreatingStatus
	if err := store.GetStoreModel().UpdateSubCluster(ctx, subCluster, creator); err != nil {
		return fmt.Errorf("update subCluster %s status failed, err %s", subClusterId, err.Error())
	}
	return nil
}
