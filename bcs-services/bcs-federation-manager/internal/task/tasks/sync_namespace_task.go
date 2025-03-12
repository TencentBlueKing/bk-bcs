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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	steps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps"
)

var (
	// HandleNamespaceQuotaTaskName step name for create cluster
	HandleNamespaceQuotaTaskName = TaskNames{
		Name: "handle quota and subCluster namespace",
		Type: "HANDLE QUOTA AND SUBCLUSTER NAMESPACE",
	}
)

// NewHandleNamespaceQuotaTask new federation task
func NewHandleNamespaceQuotaTask(opt *HandleNamespaceQuotaOptions) *HandleNamespaceQuota {
	return &HandleNamespaceQuota{
		opt: opt,
	}
}

// HandleNamespaceQuotaOptions handle quota options
type HandleNamespaceQuotaOptions struct {
	HandleType    string // create or update
	FedClusterId  string // 联邦proxy集群id
	HostClusterId string // 联邦host集群id
	Namespace     string // 命名空间
	Parameter     string
}

// HandleNamespaceQuota federation task
type HandleNamespaceQuota struct {
	opt *HandleNamespaceQuotaOptions
}

// Name return name of task
func (i *HandleNamespaceQuota) Name() string {
	return HandleNamespaceQuotaTaskName.Name
}

// Type 任务类型
func (i *HandleNamespaceQuota) Type() string {
	return HandleNamespaceQuotaTaskName.Type
}

// Steps build steps for task
func (i *HandleNamespaceQuota) Steps() []*types.Step {
	allSteps := make([]*types.Step, 0)

	checkNamespaceQuotaParamStep := steps.NewCheckNamespaceQuotaStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.ParameterKey, Value: i.opt.Parameter},
	})

	allSteps = append(allSteps, checkNamespaceQuotaParamStep)

	reqMap := make(map[string]string)
	_ = json.Unmarshal([]byte(i.opt.Parameter), &reqMap)

	for key, reqListStr := range reqMap {
		switch key {
		case fedsteps.SubClusterForTaiji:
			handleTaijiNamespaceStep := steps.NewHandleTaijiNamespaceStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.ParameterKey, Value: reqListStr},
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
				{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
			})
			allSteps = append(allSteps, handleTaijiNamespaceStep)
		case fedsteps.SubClusterForHunbu:
			handleHunbuNamespaceStep := steps.NewHandleHunbuNamespaceStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
				{Key: fedsteps.ParameterKey, Value: reqListStr},
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
			})
			allSteps = append(allSteps, handleHunbuNamespaceStep)
		case fedsteps.SubClusterForSuanli:
			handleSuanliNamespaceStep := steps.NewHandleSuanliNamespaceStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.ParameterKey, Value: reqListStr},
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
				{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
			})
			allSteps = append(allSteps, handleSuanliNamespaceStep)
		case fedsteps.ClusterQuotaKey:
			handleFederationNamespaceQuotaStep := steps.NewHandleFederationNamespaceQuotaStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
				{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
				{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
				{Key: fedsteps.ParameterKey, Value: reqListStr},
			})
			allSteps = append(allSteps, handleFederationNamespaceQuotaStep)
		case fedsteps.SubClusterForNormal:
			handleNormalNamespaceStep := steps.NewHandleNormalNamespaceStep().BuildStep([]task.KeyValue{
				{Key: fedsteps.ParameterKey, Value: reqListStr},
				{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
				{Key: fedsteps.HandleTypeKey, Value: i.opt.HandleType},
			})
			allSteps = append(allSteps, handleNormalNamespaceStep)
		}
	}

	// 更新状态step
	updateFederationNamespaceStatusStep := steps.NewUpdateFederationNamespaceStatusStep().BuildStep([]task.KeyValue{
		{Key: fedsteps.HostClusterIdKey, Value: i.opt.HostClusterId},
		{Key: fedsteps.NamespaceKey, Value: i.opt.Namespace},
	})

	allSteps = append(allSteps, updateFederationNamespaceStatusStep)

	return allSteps
}

// BuildTask build task with steps
func (i *HandleNamespaceQuota) BuildTask(creator string, opts ...types.TaskOption) (*types.Task, error) {
	if i.opt.HostClusterId == "" || i.opt.Parameter == "" {
		return nil, fmt.Errorf("handle namespace and quota task parameter is empty")
	}

	t := types.NewTask(&types.TaskInfo{
		TaskIndex: fmt.Sprintf("%s/%s", i.opt.HostClusterId, i.opt.Namespace),
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

	// record task params
	t.AddCommonParams(fedsteps.FedClusterIdKey, i.opt.FedClusterId).
		AddCommonParams(fedsteps.HostClusterIdKey, i.opt.HostClusterId).
		AddCommonParams(fedsteps.NamespaceKey, i.opt.Namespace)

	return t, nil
}
