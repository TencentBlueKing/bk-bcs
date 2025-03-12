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
	"time"

	"github.com/avast/retry-go"
	"github.com/clusternet/clusternet/pkg/apis/clusters/v1beta1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// RegisterSubclusterStepName step name for create cluster
	RegisterSubclusterStepName = fedsteps.StepNames{
		Alias: "register subcluster into federation manager",
		Name:  "REGISTER_SUBCLUSTER",
	}
)

// NewRegisterSubclusterStep sum step
func NewRegisterSubclusterStep() *RegisterSubclusterStep {
	return &RegisterSubclusterStep{}
}

// RegisterSubclusterStep sum step
type RegisterSubclusterStep struct{}

// Alias step name
func (s RegisterSubclusterStep) Alias() string {
	return RegisterSubclusterStepName.Alias
}

// GetName step name
func (s RegisterSubclusterStep) GetName() string {
	return RegisterSubclusterStepName.Name
}

// DoWork for worker exec task
func (s RegisterSubclusterStep) DoWork(t *types.Task) error {
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

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.CreatorKey)
	}

	// get federation cluster
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), fedClusterId)
	if err != nil {
		return err
	}

	// mcls get success means sub cluster is registered in clusternet
	var mcls *v1beta1.ManagedCluster
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := retry.Do(func() error {
		obj, iErr := cluster.GetClusterClient().GetManagedCluster(fedCluster.HostClusterID, subClusterId)
		if iErr != nil {
			blog.Errorf("get managed cluster failed, err: %s", iErr.Error())
			return iErr
		}
		if obj == nil {
			return fmt.Errorf("managed cluster is not found for subcluster[%s] in host cluster[%s]", subClusterId, fedCluster.HostClusterID)
		}
		mcls = obj
		return nil
	}, retry.Attempts(10), retry.Delay(1*time.Minute), retry.DelayType(retry.FixedDelay)); err != nil {
		return err
	}
	if mcls == nil {
		return fmt.Errorf("managed cluster is not found for subcluster[%s] in host cluster[%s]", subClusterId, fedCluster.HostClusterID)
	}

	// get sub cluster which is registered in PRE_REGISTER_SUBCLUSTER
	subCluster, err := store.GetStoreModel().GetSubCluster(context.Background(), fedClusterId, subClusterId)
	if err != nil {
		return err
	}
	if subCluster == nil {
		return fmt.Errorf("sub cluster is not found for subcluster[%s] in federation cluster[%s]", subClusterId, fedClusterId)
	}

	// set mcls infos and subcluster status
	subCluster.ClusternetClusterID = string(mcls.Spec.ClusterID)
	subCluster.ClusternetClusterName = mcls.Name
	subCluster.ClusternetClusterNamespace = mcls.Namespace
	subCluster.Status = store.RunningStatus
	if err := store.GetStoreModel().UpdateSubCluster(context.Background(), subCluster, creator); err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s RegisterSubclusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
