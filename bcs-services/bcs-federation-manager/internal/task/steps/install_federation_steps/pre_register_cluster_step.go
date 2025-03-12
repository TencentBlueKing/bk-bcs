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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// PreRegisterClusterStepName step name for create cluster
	PreRegisterClusterStepName = fedsteps.StepNames{
		Alias: "pre register cluster into cluster manager and federation manager",
		Name:  "PRE_REGISTER_CLUSTER",
	}
)

// NewPreRegisterClusterStep pre register cluster step
func NewPreRegisterClusterStep() *PreRegisterClusterStep {
	return &PreRegisterClusterStep{}
}

// PreRegisterClusterStep pre register cluster step
type PreRegisterClusterStep struct{}

// Alias step name
func (s PreRegisterClusterStep) Alias() string {
	return PreRegisterClusterStepName.Alias
}

// GetName step name
func (s PreRegisterClusterStep) GetName() string {
	return PreRegisterClusterStepName.Name
}

// DoWork for worker exec task
func (s PreRegisterClusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.CreatorKey)
	}

	federationClusterName, ok := step.GetParam(fedsteps.FederationClusterNameKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FederationClusterNameKey)
	}

	businessId, ok := step.GetParam(fedsteps.FederationBusinessIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FederationBusinessIdKey)
	}

	projectId, ok := step.GetParam(fedsteps.FederationProjectIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ProjectIdKey)
	}

	projectCode, ok := step.GetParam(fedsteps.FederationProjectCodeKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ProjectCodeKey)
	}

	env, ok := step.GetParam(fedsteps.FederationClusterEnvKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FederationClusterEnvKey)
	}

	desc, ok := step.GetParam(fedsteps.FederationClusterDescriptionKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FederationClusterDescriptionKey)
	}

	labelsStr, ok := step.GetParam(fedsteps.FederationClusterLabelsStrKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FederationClusterLabelsStrKey)
	}
	labels := make(map[string]string)
	err := json.Unmarshal([]byte(labelsStr), &labels)
	if err != nil {
		return fmt.Errorf("unmarshal federation cluster labels failed, err: %v", err)
	}

	// inject federation identify label
	labels[cluster.FederationClusterTypeLabelKeyFedCluster] = cluster.FederationClusterTypeLabelValueTrue
	labels[cluster.FederationClusterTaskIDLabelKey] = t.GetTaskID()

	// create a new cluster in clustermanager as a federationCluster and get its id
	federationClusterId, err := cluster.GetClusterClient().CreateFederationCluster(context.Background(), &cluster.FederationClusterCreateReq{
		ClusterName: federationClusterName,
		Creator:     creator,
		ProjectId:   projectId,
		BusinessId:  businessId,
		Environment: env,
		Description: desc,
		Labels:      labels,
	})
	if err != nil {
		return err
	}
	// set federation cluster Id
	t.AddCommonParams(fedsteps.FedClusterIdKey, federationClusterId)

	// bind task id to federation cluster
	extras := labels
	extras[cluster.FederationClusterTaskIDLabelKey] = t.GetTaskID()

	// create federation cluster in federation manager
	if err := store.GetStoreModel().CreateFederationCluster(context.Background(), &store.FederationCluster{
		HostClusterID:         hostClusterId,
		FederationClusterID:   federationClusterId,
		FederationClusterName: federationClusterName,
		ProjectCode:           projectCode,
		ProjectID:             projectId,
		IsDeleted:             false,
		Descriptions:          desc,
		Extras:                extras,
		Creator:               creator,
		Status:                store.CreatingStatus,
	}); err != nil {
		return err
	}

	return nil

}

// BuildStep build step
func (s PreRegisterClusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
