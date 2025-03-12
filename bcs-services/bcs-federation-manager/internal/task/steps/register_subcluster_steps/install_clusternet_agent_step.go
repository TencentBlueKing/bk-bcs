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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// InstallClusternetAgentStepName step name for create cluster
	InstallClusternetAgentStepName = fedsteps.StepNames{
		Alias: "install bcs-clusternet-agent",
		Name:  "INSTALL_CLUSTERNET_AGENT",
	}
)

// NewInstallClusternetAgentStep sum step
func NewInstallClusternetAgentStep() *InstallClusternetAgentStep {
	return &InstallClusternetAgentStep{}
}

// InstallClusternetAgentStep sum step
type InstallClusternetAgentStep struct{}

// Alias step name
func (s InstallClusternetAgentStep) Alias() string {
	return InstallClusternetAgentStepName.Alias
}

// GetName step name
func (s InstallClusternetAgentStep) GetName() string {
	return InstallClusternetAgentStepName.Name
}

// DoWork for worker exec task
func (s InstallClusternetAgentStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// host cluster may not in same project with federation cluster
	hostProjectId, ok := t.GetCommonParams(fedsteps.HostProjectIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostProjectIdKey)
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	userToken, ok := t.GetCommonParams(fedsteps.UserTokenKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.UserTokenKey)
	}

	// 1.获取host集群bootstrap secret

	secret, err := cluster.GetClusterClient().GetBootstrapSecret(&cluster.ResourceGetOptions{
		ClusterId: hostClusterId,
		Namespace: fedsteps.BootstrapTokenNamespace})
	if err != nil {
		return err
	}

	if secret.Data == nil {
		return fmt.Errorf("bootstrap secret not found Data")
	}

	tokenId, ok := secret.Data[fedsteps.BootstrapTokenIdKey]
	if !ok {
		return fmt.Errorf("bootstrap secret[%s/%s] not found token-id", secret.Namespace, secret.Name)
	}

	tokenSecret, ok := secret.Data[fedsteps.BootstrapTokenSecretKey]
	if !ok {
		return fmt.Errorf("bootstrap secret[%s/%s] not found token-secret", secret.Namespace, secret.Name)
	}

	gatewayAddress, ok := t.GetCommonParams(fedsteps.BcsGatewayAddressKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.BcsGatewayAddressKey)
	}

	err = helm.GetHelmClient().InstallClusternetAgent(&helm.BcsClusternetAgentOptions{
		ReleaseBaseOptions: helm.ReleaseBaseOptions{
			ProjectID:       hostProjectId,
			ClusterID:       hostClusterId,
			SkipWhenExisted: true,
		},
		SubClusterId:      subClusterId,
		RegistrationToken: fmt.Sprintf("%s.%s", string(tokenId), string(tokenSecret)),
		BcsGateWayAddress: gatewayAddress,
		UserToken:         userToken,
	})
	if err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s InstallClusternetAgentStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
