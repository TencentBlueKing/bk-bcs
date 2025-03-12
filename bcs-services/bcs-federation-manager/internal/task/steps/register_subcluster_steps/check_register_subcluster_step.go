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
	// CheckRegisterSubclusterStepName step name for create cluster
	CheckRegisterSubclusterStepName = fedsteps.StepNames{
		Alias: "check for register subclsuter",
		Name:  "CHECK_REGISTER_SUBCLUSTER",
	}
)

// NewCheckRegisterSubclusterStep sum step
func NewCheckRegisterSubclusterStep() *CheckRegisterSubclusterStep {
	return &CheckRegisterSubclusterStep{}
}

// CheckRegisterSubclusterStep sum step
type CheckRegisterSubclusterStep struct{}

// Alias step name
func (s CheckRegisterSubclusterStep) Alias() string {
	return CheckRegisterSubclusterStepName.Alias
}

// GetName step name
func (s CheckRegisterSubclusterStep) GetName() string {
	return CheckRegisterSubclusterStepName.Name
}

// DoWork for worker exec task
func (s CheckRegisterSubclusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// get federation cluster and replace proxy clusterId to host clusterId
	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), fedClusterId)
	if err != nil {
		return fmt.Errorf("get federation cluster from federationmanager failed, err: %s", err.Error())
	}

	// get sub cluster
	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	// check if managed cluster is already exist
	obj, iErr := cluster.GetClusterClient().GetManagedCluster(fedCluster.HostClusterID, subClusterId)
	if iErr != nil {
		blog.Errorf("get managed cluster failed, err: %s", iErr.Error())
		return iErr
	}
	if obj != nil {
		return fmt.Errorf("managed cluster is already exist")
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s CheckRegisterSubclusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
