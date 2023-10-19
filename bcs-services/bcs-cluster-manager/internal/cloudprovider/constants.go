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

// TaskName xxx
type TaskName string

// String xxx
func (tn TaskName) String() string {
	return string(tn)
}

// StepName xx
type StepName string

// String xxx
func (sn StepName) String() string {
	return string(sn)
}

var (
	// CreateNodeGroupTask task
	CreateNodeGroupTask TaskName = "创建节点池"

	// UpdateDesiredNodesTask task
	UpdateDesiredNodesTask TaskName = "扩容节点池"
	// ApplyInstanceMachinesStep step
	ApplyInstanceMachinesStep StepName = "申请节点"
	// CheckClusterNodesStatusStep step
	CheckClusterNodesStatusStep StepName = "检测节点状态"
	// InstallGSEAgentStep step
	InstallGSEAgentStep StepName = "安装 GSE Agent"
	// TransferHostModuleStep step
	TransferHostModuleStep StepName = "转移主机模块"

	// CleanNodesInGroupTask task
	CleanNodesInGroupTask TaskName = "缩容节点池"
	// CleanNodeGroupNodesStep step
	CleanNodeGroupNodesStep StepName = "下架节点池节点"
	// RemoveHostFromCMDBStep step
	RemoveHostFromCMDBStep StepName = "移除主机"
)

// ParamKey xxx
type ParamKey string

// String xxx
func (pk ParamKey) String() string {
	return string(pk)
}

var (
	// TaskNameKey xxx
	TaskNameKey ParamKey = "taskName"

	// ProjectIDKey xxx
	ProjectIDKey ParamKey = "projectID"
	// ClusterIDKey xxx
	ClusterIDKey ParamKey = "clusterID"
	// NodeGroupIDKey xxx
	NodeGroupIDKey ParamKey = "nodeGroupID"
	// CloudIDKey xxx
	CloudIDKey ParamKey = "cloudID"

	// PasswordKey xxx
	PasswordKey ParamKey = "password"
	// ScalingKey xxx
	ScalingKey ParamKey = "scaling"
	// OperatorKey xxx
	OperatorKey ParamKey = "operator"
	// BKBizIDKey bk biz id key
	BKBizIDKey ParamKey = "bkBizID"
	// BKCloudIDKey bk cloud id key
	BKCloudIDKey ParamKey = "bkCloudID"
	// BKModuleIDKey bk module id key
	BKModuleIDKey ParamKey = "bkModuleID"

	// NodeIPsKey TODO
	// Task Common Instance
	// NodeIPsKey xxx
	NodeIPsKey ParamKey = "nodeIPs"
	// NodeIDsKey xxx
	NodeIDsKey ParamKey = "nodeIDs"

	// DynamicNodeIPListKey xxx
	DynamicNodeIPListKey ParamKey = "NodeIPList"

	// SuccessNodeIDsKey TODO
	// CVM Instance
	// SuccessNodeIDsKey xxx
	SuccessNodeIDsKey ParamKey = "successNodeIDs"
	// FailedNodeIDsKey xxx
	FailedNodeIDsKey ParamKey = "failedNodeIDs"

	// SuccessClusterNodeIDsKey TODO
	// cloud cluster success & failed Instance
	// SuccessClusterNodeIDsKey xxx
	SuccessClusterNodeIDsKey ParamKey = "successClusterNodeIDs"
	// FailedClusterNodeIDsKey xxx
	FailedClusterNodeIDsKey ParamKey = "failedClusterNodeIDs"
	// DeleteModeKey xxx
	DeleteModeKey ParamKey = "deleteMode"

	// SuccessNodeGroupIDsKey xxx
	SuccessNodeGroupIDsKey ParamKey = "successNodeGroupIDs"
	// FailedNodeGroupIDsKey xxx
	FailedNodeGroupIDsKey ParamKey = "failedNodeGroupIDs"

	// JobTypeKey xxx
	JobTypeKey ParamKey = "jobType"
	// BkSopsTaskUrlKey inject bksops task url
	BkSopsTaskUrlKey ParamKey = "taskUrl"
)
