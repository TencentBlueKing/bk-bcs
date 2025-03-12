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
	// PreRegisterSubclusterStepName step name for create cluster
	PreRegisterSubclusterStepName = fedsteps.StepNames{
		Alias: "pre register subcluster into federation manager",
		Name:  "PRE_REGISTER_SUBCLUSTER",
	}
)

// NewPreRegisterSubclusterStep sum step
func NewPreRegisterSubclusterStep() *PreRegisterSubclusterStep {
	return &PreRegisterSubclusterStep{}
}

// PreRegisterSubclusterStep sum step
type PreRegisterSubclusterStep struct{}

// Alias step name
func (s PreRegisterSubclusterStep) Alias() string {
	return PreRegisterSubclusterStepName.Alias
}

// GetName step name
func (s PreRegisterSubclusterStep) GetName() string {
	return PreRegisterSubclusterStepName.Name
}

// DoWork for worker exec task
func (s PreRegisterSubclusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// get common params
	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	subProjectId, ok := t.GetCommonParams(fedsteps.SubProjectIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubProjectIdKey)
	}

	subProjectCode, ok := t.GetCommonParams(fedsteps.SubProjectCodeKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubProjectCodeKey)
	}

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.CreatorKey)
	}

	// get federation cluster
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), fedClusterId)
	if err != nil {
		return err
	}

	subCluster, err := cluster.GetClusterClient().GetCluster(context.Background(), subClusterId)
	if err != nil {
		return err
	}

	// update sub cluster labels
	if err := cluster.GetClusterClient().UpdateSubClusterLabel(context.Background(), subClusterId); err != nil {
		return err
	}

	labels := make(map[string]string, 0)
	labels[cluster.FederationClusterTaskIDLabelKey] = t.GetTaskID()

	// create sub cluster in federation manager
	if err := store.GetStoreModel().CreateSubCluster(context.Background(), &store.SubCluster{
		SubClusterID:               subClusterId,
		SubClusterName:             subCluster.ClusterName,
		FederationClusterID:        fedCluster.FederationClusterID,
		HostClusterID:              fedCluster.HostClusterID,
		ProjectCode:                subProjectCode,
		ProjectID:                  subProjectId,
		IsDeleted:                  false,
		ClusternetClusterID:        "",
		ClusternetClusterName:      "",
		ClusternetClusterNamespace: "",
		Descriptions:               "new sub cluster",
		Labels:                     labels,
		Creator:                    creator,
		Status:                     store.CreatingStatus,
	}); err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s PreRegisterSubclusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
