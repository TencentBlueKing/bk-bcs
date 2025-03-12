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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/kubeconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// RegisterClusterStepName step name for create cluster
	RegisterClusterStepName = fedsteps.StepNames{
		Alias: "register cluster into cluster manager and federation manager",
		Name:  "REGISTER_CLUSTER",
	}
)

// NewRegisterClusterStep sum step
func NewRegisterClusterStep() *RegisterClusterStep {
	return &RegisterClusterStep{}
}

// RegisterClusterStep register cluster step
type RegisterClusterStep struct{}

// Alias step name
func (s RegisterClusterStep) Alias() string {
	return RegisterClusterStepName.Alias
}

// GetName step name
func (s RegisterClusterStep) GetName() string {
	return RegisterClusterStepName.Name
}

// DoWork for worker exec task
func (s RegisterClusterStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.ClusterIdKey)
	}

	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.FedClusterIdKey)
	}

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.CreatorKey)
	}

	// waiting for bcs unified apiserver ingress installed
	var address string
	if err := retry.Do(func() error {
		addr, iErr := fedsteps.GetBcsUnifiedApiserverAddress(hostClusterId)
		if iErr != nil {
			blog.Warnf("get bcs unified apiserver address failed, err: %v", iErr)
			return iErr
		}
		address = addr
		return nil
	}, retry.Attempts(3), retry.Delay(1*time.Minute), retry.DelayType(retry.FixedDelay)); err != nil {
		return err
	}

	// check cluster connection health
	if err := retry.Do(func() error {
		err := fedsteps.CheckClusterConnection(address)
		if err != nil {
			blog.Warnf("check cluster[%s] connection failed, err: %v", hostClusterId, err)
			return err
		}
		return nil
	}, retry.Attempts(3), retry.Delay(1*time.Minute), retry.DelayType(retry.FixedDelay)); err != nil {
		return err
	}

	// get federation cluster from store
	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), fedClusterId)
	if err != nil {
		return err
	}

	// update federation cluster credentials
	if err := cluster.GetClusterClient().UpdateFederationClusterCredentials(context.Background(), fedClusterId,
		kubeconfig.NewConfigForRegister(address).Yaml()); err != nil {
		return err
	}

	// update federation cluster status
	if err := cluster.GetClusterClient().UpdateFederationClusterStatus(context.Background(), fedClusterId, cluster.ClusterStatusRunning); err != nil {
		return err
	}

	// update store federation cluster status
	fedCluster.Status = store.RunningStatus
	if err := store.GetStoreModel().UpdateFederationCluster(context.Background(), fedCluster, creator); err != nil {
		return fmt.Errorf("update federation cluster[%s] in federation manager store failed, err: %v", fedCluster.FederationClusterID, err)
	}

	// update host cluster label
	if err := cluster.GetClusterClient().UpdateHostClusterLabel(context.Background(), hostClusterId); err != nil {
		return fmt.Errorf("update host cluster[%s] labels in cluster manager failed, err: %v", hostClusterId, err)
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s RegisterClusterStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
