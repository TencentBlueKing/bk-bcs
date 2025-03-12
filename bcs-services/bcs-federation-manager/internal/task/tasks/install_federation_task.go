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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	steps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/install_federation_steps"
)

var (
	// InstallFederationTaskName step name for create cluster
	InstallFederationTaskName = TaskNames{
		Name: "install federation modules",
		Type: "INSTALL_FEDERATION",
	}
)

// NewInstallFederationTask new install federation task
func NewInstallFederationTask(opt *InstallFederationOptions) *InstallFederation {
	return &InstallFederation{
		opt: opt,
	}
}

// InstallFederationOptions options for install federation task
type InstallFederationOptions struct {
	ProjectId                    string
	ClusterId                    string
	UserToken                    string
	LbId                         string
	FederationBusinessId         string
	FederationProjectId          string
	FederationProjectCode        string
	FederationClusterName        string
	FederationClusterEnv         string
	FederationClusterDescription string
	FederationClusterLabels      map[string]string
}

// InstallFederation install federation task
type InstallFederation struct {
	opt *InstallFederationOptions
}

// Name 任务名字
func (i *InstallFederation) Name() string {
	return InstallFederationTaskName.Name
}

// Type 任务类型
func (i *InstallFederation) Type() string {
	return InstallFederationTaskName.Type
}

// Steps build steps for task
func (i *InstallFederation) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	// step0: check federation installed
	step0 := steps.NewCheckFederationInstalledStep().BuildStep([]task.KeyValue{})

	// step1: pre register cluster
	labels := i.opt.FederationClusterLabels
	if labels == nil {
		labels = make(map[string]string)
	}
	labelsStr, _ := json.Marshal(labels)
	step1 := steps.NewPreRegisterClusterStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.FederationProjectIdKey, Value: i.opt.FederationProjectId},
		{Key: fedsteps.FederationProjectCodeKey, Value: i.opt.FederationProjectCode},
		{Key: fedsteps.FederationClusterNameKey, Value: i.opt.FederationClusterName},
		{Key: fedsteps.FederationBusinessIdKey, Value: i.opt.FederationBusinessId},
		{Key: fedsteps.FederationClusterEnvKey, Value: i.opt.FederationClusterEnv},
		{Key: fedsteps.FederationClusterDescriptionKey, Value: i.opt.FederationClusterDescription},
		{Key: fedsteps.FederationClusterLabelsStrKey, Value: string(labelsStr)},
	})

	// step2: install clusternet-hub
	step2 := steps.NewInstallClusternetHubStep().BuildStep([]task.KeyValue{})

	// step3: install clusternet-scheduler
	step3 := steps.NewInstallClusternetSchedulerStep().BuildStep([]task.KeyValue{})

	// step4: install clusternet-controller
	step4 := steps.NewInstallClusternetControllerStep().BuildStep([]task.KeyValue{})

	// step5: install bcs-unified-apiserver
	step5 := steps.NewInstallBcsUnifiedApiserverStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.UserTokenKey, Value: i.opt.UserToken},
		{Key: fedsteps.LoadBalancerIdKey, Value: i.opt.LbId},
	})

	// step6: register federation cluster to cluster manager and store to federation manager
	step6 := steps.NewRegisterClusterStep().BuildStep([]task.KeyValue{})

	// step7: create register token
	step7 := steps.NewCreateRegisterToken().BuildStep([]task.KeyValue{})

	// all
	allSteps = append(allSteps, step0, step1, step2, step3, step4, step5, step6, step7)
	return allSteps
}

// BuildTask build task with steps
func (i *InstallFederation) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if i.opt == nil {
		return nil, fmt.Errorf("install federation task options empty")
	}
	if i.opt.ProjectId == "" || i.opt.ClusterId == "" || i.opt.UserToken == "" {
		return nil, fmt.Errorf("install federation task options empty, projectId: %s, clusterId: %s",
			i.opt.ProjectId, i.opt.ClusterId)
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: i.opt.ClusterId,
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

	t.AddCommonParams(fedsteps.ProjectIdKey, i.opt.ProjectId).
		AddCommonParams(fedsteps.ClusterIdKey, i.opt.ClusterId).
		AddCommonParams(fedsteps.CreatorKey, creator).
		AddCommonParams(fedsteps.FederationProjectCodeKey, i.opt.FederationProjectCode).
		AddCommonParams(fedsteps.FederationProjectIdKey, i.opt.FederationProjectId)

	// set callback func
	t.SetCallback(steps.NewInstallFederationCallBack().GetName())
	return t, nil
}

// BeforeRetryInstallFederation retry task
func BeforeRetryInstallFederation(ctx context.Context, t *types.Task) error {

	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		// fed cluster is not created, skip, PRE_REGISTER_CLUSTER is not executed
		return nil
	}

	fedCluster, err := store.GetStoreModel().GetFederationCluster(ctx, fedClusterId)
	if err != nil {
		// if cluster is not existed, skip, PRE_REGISTER_CLUSTER is not executed
		return nil
	}

	// update status to initialization
	if err := cluster.GetClusterClient().UpdateFederationClusterStatus(ctx,
		fedClusterId, cluster.ClusterStatusInitialization); err != nil {
		return fmt.Errorf("update cluster %s status failed, err %s", fedClusterId, err.Error())
	}

	// update federation cluster status to creating
	fedCluster.Status = store.CreatingStatus
	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fmt.Errorf("get %s failed when retry install federation task", fedsteps.CreatorKey)
	}
	if err := store.GetStoreModel().UpdateFederationCluster(ctx, fedCluster, creator); err != nil {
		return fmt.Errorf("update cluster %s status failed, err %s", fedClusterId, err.Error())
	}

	return nil
}
