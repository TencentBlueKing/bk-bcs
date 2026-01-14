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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// InstallBcsUnifiedApiserverStepName step name for create cluster
	InstallBcsUnifiedApiserverStepName = fedsteps.StepNames{
		Alias: "install bcs-unified-apiserver",
		Name:  "INSTALL_BCS_UNIFIED_APISERVER",
	}
)

// NewInstallBcsUnifiedApiserverStep sum step
func NewInstallBcsUnifiedApiserverStep() *InstallBcsUnifiedApiserverStep {
	return &InstallBcsUnifiedApiserverStep{}
}

// InstallBcsUnifiedApiserverStep sum step
type InstallBcsUnifiedApiserverStep struct{}

// Alias step name
func (s InstallBcsUnifiedApiserverStep) Alias() string {
	return InstallBcsUnifiedApiserverStepName.Alias
}

// GetName step name
func (s InstallBcsUnifiedApiserverStep) GetName() string {
	return InstallBcsUnifiedApiserverStepName.Name
}

// DoWork for worker exec task
func (s InstallBcsUnifiedApiserverStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	projectId, ok := t.GetCommonParams(fedsteps.ProjectIdKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist common param projectId", t.TaskID)
	}

	clusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist common param clusterId", t.TaskID)
	}

	userToken, ok := step.GetParam(fedsteps.UserTokenKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist param userToken", t.TaskID)
	}

	lbId, ok := step.GetParam(fedsteps.LoadBalancerIdKey)
	if !ok {
		return fmt.Errorf("task[%s] not exist param lbId", t.TaskID)
	}

	// create namespace
	var err error
	var ns string
	if helm.GetHelmClient().GetFederationCharts().Apiserver != nil {
		ns = helm.GetHelmClient().GetFederationCharts().Apiserver.ReleaseNamespace
	}
	if ns == "" {
		return fmt.Errorf("bcs-unified-apiserver release namespace is empty")
	}
	err = cluster.GetClusterClient().CreateNamespace(clusterId, ns)
	if err != nil {
		return err
	}

	// create helm release
	err = helm.GetHelmClient().InstallUnifiedApiserver(context.Background(), &helm.BcsUnifiedApiserverOptions{
		ReleaseBaseOptions: helm.ReleaseBaseOptions{
			ProjectID:       projectId,
			ClusterID:       clusterId,
			SkipWhenExisted: true,
		},
		LoadBalancerId: lbId,
		UserToken:      userToken,
	})
	if err != nil {
		return err
	}

	// waiting for bcs unified apiserver ingress installed
	var address string
	if err := retry.Do(func() error {
		addr, iErr := fedsteps.GetBcsUnifiedApiserverAddress(clusterId)
		if iErr != nil {
			blog.Warnf("get bcs unified apiserver address failed, err: %v", iErr)
			return iErr
		}
		address = addr
		return nil
	}, retry.Attempts(5), retry.Delay(1*time.Minute), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(10*time.Minute)); err != nil {
		return err
	}

	// save bcs unified apiserver address to task
	step.SetParamMulti(map[string]string{
		fedsteps.BcsUnifiedApiserverAddressKey: address,
	})

	// check cluster connection health
	if err := retry.Do(func() error {
		err := fedsteps.CheckClusterConnection(address)
		if err != nil {
			blog.Warnf("check cluster[%s] connection failed, err: %v", clusterId, err)
			return err
		}
		return nil
	}, retry.Attempts(5), retry.Delay(1*time.Minute), retry.DelayType(retry.FixedDelay)); err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s InstallBcsUnifiedApiserverStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
