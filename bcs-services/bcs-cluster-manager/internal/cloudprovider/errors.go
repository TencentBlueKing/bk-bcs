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
 *
 */

package cloudprovider

import "errors"

// template
var (
	// ErrCloudNodeVPCDiffWithClusterResponse for node VPC different cluster VPC
	ErrCloudNodeVPCDiffWithClusterResponse = "node[%s] VPC is different from cluster VPC"
)

// tke error
var (
	// ErrCloudCredentialLost credential lost in option
	ErrCloudCredentialLost = errors.New("credential info lost")
	// ErrCloudRegionLost region information lost in option
	ErrCloudRegionLost = errors.New("region info lost")
	// ErrCloudLostResponse lost response information in cloud response
	ErrCloudLostResponse = errors.New("lost response information")
	// ErrCloudNoHost no specified instance
	ErrCloudNoHost = errors.New("no such host in region")
	// ErrCloudNoProvider no specified cloud provider
	ErrCloudNoProvider = errors.New("no such cloudprovider")
	// ErrCloudNotImplemented no implementation
	ErrCloudNotImplemented = errors.New("not implemented")
	// ErrCloudInitFailed init failed
	ErrCloudInitFailed = errors.New("failed to init cloud client")
	// ErrServerIsNil server nil
	ErrServerIsNil = errors.New("server is nil")
)

// aks error
var (
	// ErrClusterEmpty cluster 不能为空
	ErrClusterEmpty = errors.New("cluster cannot be empty")
	// ErrAgentPoolEmpty AgentPool 不能为空
	ErrAgentPoolEmpty = errors.New("agentPool cannot be empty")
	// ErrVirtualMachineScaleSetEmpty VirtualMachineScaleSet 不能为空
	ErrVirtualMachineScaleSetEmpty = errors.New("virtualMachineScaleSet cannot be empty")
	// ErrNodeGroupEmpty nodeGroup 不能为空
	ErrNodeGroupEmpty = errors.New("nodeGroup cannot be empty")
	// ErrNodeGroupAutoScalingLost  nodeGroup 的 autoScaling 不能为空
	ErrNodeGroupAutoScalingLost = errors.New("autoscaling attribute in nodegroup cannot be empty")
	// ErrNodeGroupNodeTemplateLost nodeGroup 的 nodeTemplate 不能为空
	ErrNodeGroupNodeTemplateLost = errors.New("nodeTemplate attribute in nodegroup cannot be empty")
	// ErrNodeGroupLaunchTemplateLost nodeGroup 的 launchTemplate 不能为空
	ErrNodeGroupLaunchTemplateLost = errors.New("launchTemplate attribute in nodegroup cannot be empty")
	// ErrVirtualMachineEmpty VirtualMachine 不能为空
	ErrVirtualMachineEmpty = errors.New("virtual machine cannot be empty")
	// ErrVmInstanceType 机型不存在
	ErrVmInstanceType = errors.New("instance type does not exist")
	// ErrAgentPoolNotMatchesVMSSs 找不到与AgentPool匹配的VMSSs
	ErrAgentPoolNotMatchesVMSSs = errors.New("could not find a matching VMSSs for AgentPool")
)
