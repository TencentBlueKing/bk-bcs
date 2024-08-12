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

// Package cloudprovider xxx
package cloudprovider

import (
	"fmt"
)

// TaskType taskType
type TaskType string

// String toString
func (tt TaskType) String() string {
	return string(tt)
}

var (
	// CreateCluster task
	CreateCluster TaskType = "CreateCluster"
	// CreateVirtualCluster task
	CreateVirtualCluster TaskType = "CreateVirtualCluster"
	// ImportCluster task
	ImportCluster TaskType = "ImportCluster"
	// DeleteCluster task
	DeleteCluster TaskType = "DeleteCluster"
	// DeleteVirtualCluster task
	DeleteVirtualCluster TaskType = "DeleteVirtualCluster"
	// AddNodesToCluster task
	AddNodesToCluster TaskType = "AddNodesToCluster"
	// RemoveNodesFromCluster task
	RemoveNodesFromCluster TaskType = "RemoveNodesFromCluster"
	// AddExternalNodesToCluster task
	AddExternalNodesToCluster TaskType = "AddExternalNodesToCluster"
	// RemoveExternalNodesFromCluster task
	RemoveExternalNodesFromCluster TaskType = "RemoveExternalNodesFromCluster"
	// SwitchClusterUnderlayNetwork task
	SwitchClusterUnderlayNetwork TaskType = "SwitchClusterUnderlayNetwork"

	// CreateNodeGroup task
	CreateNodeGroup TaskType = "CreateNodeGroup"
	// UpdateNodeGroup task
	UpdateNodeGroup TaskType = "UpdateNodeGroup"
	// DeleteNodeGroup task
	DeleteNodeGroup TaskType = "DeleteNodeGroup"
	// MoveNodesToNodeGroup task
	MoveNodesToNodeGroup TaskType = "MoveNodesToNodeGroup"

	// SwitchNodeGroupAutoScaling task
	SwitchNodeGroupAutoScaling TaskType = "SwitchNodeGroupAutoScaling"

	// UpdateAutoScalingOption task
	UpdateAutoScalingOption TaskType = "UpdateAutoScalingOption"
	// SwitchAutoScalingOptionStatus task
	SwitchAutoScalingOptionStatus TaskType = "SwitchAutoScalingOptionStatus"

	// UpdateNodeGroupDesiredNode task
	UpdateNodeGroupDesiredNode TaskType = "UpdateNodeGroupDesiredNode"

	// ApplyInstanceMachinesTask apply instance subTask
	ApplyInstanceMachinesTask TaskType = "ApplyInstanceMachinesTask"
	// ApplyExternalNodeMachinesTask apply external node subTask
	ApplyExternalNodeMachinesTask TaskType = "ApplyExternalNodeMachinesTask"
	// CheckExternalNodesEmptyTask check external node empty
	CheckExternalNodesEmptyTask TaskType = "CheckExternalNodesEmptyTask"

	// CleanNodeGroupNodes task
	CleanNodeGroupNodes TaskType = "CleanNodeGroupNodes"

	// DebugBkSopsTask task
	DebugBkSopsTask TaskType = "DebugBkSopsTask"
)

// GetTaskType by cloud
func GetTaskType(cloud string, taskName TaskType) string {
	return fmt.Sprintf("%s-%s", cloud, taskName.String())
}

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
	// CreateClusterTask xx
	// cluster manager task
	CreateClusterTask TaskName = "创建集群"
	// CreateVirtualClusterTask xx
	CreateVirtualClusterTask TaskName = "创建虚拟集群"
	// ImportClusterTask xx
	ImportClusterTask TaskName = "纳管集群"
	// DeleteClusterTask xx
	DeleteClusterTask TaskName = "删除集群"
	// DeleteVirtualClusterTask xx
	DeleteVirtualClusterTask TaskName = "删除虚拟集群"
	// AddNodesToClusterTask xx
	AddNodesToClusterTask TaskName = "集群上架节点"
	// RemoveNodesFromClusterTask xx
	RemoveNodesFromClusterTask TaskName = "集群下架节点"
	// AddExternalNodesToClusterTask xx
	AddExternalNodesToClusterTask TaskName = "集群上架第三方节点"
	// RemoveExternalNodesFromClusterTask xx
	RemoveExternalNodesFromClusterTask TaskName = "集群下架第三方节点"
	// SwitchClusterUnderlayNetworkTask xx
	SwitchClusterUnderlayNetworkTask TaskName = "切换集群underlay网络模式"

	// CreateNodeGroupTask task
	// nodeGroup manager task
	CreateNodeGroupTask TaskName = "创建节点池"
	// UpdateDesiredNodesTask task
	UpdateDesiredNodesTask TaskName = "扩容节点池"
	// CleanNodesInGroupTask task
	CleanNodesInGroupTask TaskName = "缩容节点池"
	// UpdateNodeGroupTask task
	UpdateNodeGroupTask TaskName = "更新节点池"
	// DeleteNodeGroupTask task
	DeleteNodeGroupTask TaskName = "删除节点池"
	// SwitchNodeGroupAutoScalingTask task
	SwitchNodeGroupAutoScalingTask TaskName = "开启/关闭节点池"
	// UpdateAutoScalingOptionTask task
	UpdateAutoScalingOptionTask TaskName = "更新集群自动扩缩容配置"
	// SwitchAutoScalingOptionStatusTask task
	SwitchAutoScalingOptionStatusTask TaskName = "开启/关闭集群自动扩缩容"

	// DebugBkSopsTaskName task
	DebugBkSopsTaskName TaskName = "调试标准运维任务"
)

// ParamKey xxx
type ParamKey string

// String xxx
func (pk ParamKey) String() string {
	return string(pk)
}

var (
	// TaskNameKey bk-sops 任务名称
	TaskNameKey ParamKey = "taskName"

	// ProjectIDKey xxx
	ProjectIDKey ParamKey = "projectID"
	// ClusterIDKey xxx
	ClusterIDKey ParamKey = "clusterID"
	// HostClusterIDKey xxx
	HostClusterIDKey ParamKey = "hostClusterID"
	// NodeGroupIDKey xxx
	NodeGroupIDKey ParamKey = "nodeGroupID"
	// DeviceRecordIDKey deviceRecord key
	DeviceRecordIDKey ParamKey = "deviceRecordKey"
	// ManualKey xxx
	ManualKey ParamKey = "manual"

	// PoolProvider xxx
	PoolProvider ParamKey = "poolProvider"
	// PoolID xxx
	PoolID ParamKey = "poolID"

	// CloudIDKey xxx
	CloudIDKey ParamKey = "cloudID"
	// NodeTemplateIDKey xxx
	NodeTemplateIDKey ParamKey = "nodeTemplateID"

	// PasswordKey xxx
	PasswordKey ParamKey = "password"
	// UsernameKey xxx
	UsernameKey ParamKey = "username"
	// SecretKey xxx
	SecretKey ParamKey = "secret"
	// PortKey xxx
	PortKey ParamKey = "port"
	// AllowReviseAgent allow revise agent cloudId
	AllowReviseAgent ParamKey = "allow"

	// ScalingNodesNumKey xxx
	ScalingNodesNumKey ParamKey = "scalingNodesNum"
	// OperatorKey xxx
	OperatorKey ParamKey = "operator"
	// UserKey xxx
	UserKey ParamKey = "user"
	// OrderIDKey xxx
	OrderIDKey ParamKey = "orderID"
	// ScriptContentKey xxx
	ScriptContentKey ParamKey = "scriptContent"

	// BKBizIDKey bk biz id key
	BKBizIDKey ParamKey = "bkBizID"
	// BKCloudIDKey bk cloud id key
	BKCloudIDKey ParamKey = "bkCloudID"
	// BKModuleIDKey bk module id key
	BKModuleIDKey ParamKey = "bkModuleID"
	// BKMasterModuleIDKey bk master module id key
	BKMasterModuleIDKey ParamKey = "bkMasterModuleID"

	// DeleteModeKey xxx
	DeleteModeKey ParamKey = "deleteMode"
	// KeepInstanceKey xxx
	KeepInstanceKey ParamKey = "keepInstance"
	// LastClusterStatus xxx
	LastClusterStatus ParamKey = "lastClusterStatus"

	// NodeIPsKey xxx
	// Task Common Instance
	NodeIPsKey ParamKey = "nodeIPs"
	// OriginNodeIPsKey xxx
	OriginNodeIPsKey ParamKey = "originNodeIPs"
	// NodeIDsKey xxx
	NodeIDsKey ParamKey = "nodeIDs"
	// NodeLoginKey xxx
	NodeLoginKey ParamKey = "nodeLogin"
	// FailedTransVpcNodeIDsKey xxx
	FailedTransVpcNodeIDsKey ParamKey = "failedNodeIDs"
	// NodeNamesKey xxx
	NodeNamesKey ParamKey = "nodeNames"
	// DeviceIDsKey xxx
	DeviceIDsKey ParamKey = "deviceIDs"
	// NodeSchedule node schedule status
	NodeSchedule ParamKey = "nodeSchedule"
	// MasterNodeIPsKey nodes
	MasterNodeIPsKey ParamKey = "masterIPs"
	// WorkerNodeIPsKey nodes
	WorkerNodeIPsKey ParamKey = "workerIPs"
	// VpcKey vpcId
	VpcKey ParamKey = "vpc"

	// TerminateChargeNodes terminate nodes
	TerminateChargeNodes ParamKey = "terminateChargeNodes"
	// RetainChargeNodes retain nodes
	RetainChargeNodes ParamKey = "retainChargeNodes"

	// NamespaceKey xxx
	NamespaceKey ParamKey = "namespace"
	// QuotaNameKey xxx
	QuotaNameKey ParamKey = "name"
	// LabelsKey xxx
	LabelsKey ParamKey = "labels"
	// TaintsKey xxx
	TaintsKey ParamKey = "taints"
	// RemoveTaintsKey xxx
	RemoveTaintsKey ParamKey = "RemoveTaints"
	// AnnotationsKey xxx
	AnnotationsKey ParamKey = "annotations"
	// ResourceQuotaKey xxx
	ResourceQuotaKey ParamKey = "resourceQuota"

	// DynamicNodeIPListKey xxx
	DynamicNodeIPListKey ParamKey = "NodeIPList"
	// DynamicNodeScriptKey xxx
	DynamicNodeScriptKey ParamKey = "ExternalNodeScript"
	// DynamicClusterKubeConfigKey xxx
	DynamicClusterKubeConfigKey ParamKey = "KubeConfig"

	// SuccessNodeIDsKey xxx
	// CVM Instance
	SuccessNodeIDsKey ParamKey = "successNodeIDs"
	// FailedNodeIDsKey xxx
	FailedNodeIDsKey ParamKey = "failedNodeIDs"
	// FailureNodeIDsKey xxx
	FailureNodeIDsKey ParamKey = "failureNodeIDs"
	// TimeoutNodeIDsKey xxx
	TimeoutNodeIDsKey ParamKey = "timeoutNodeIDs"

	// FailureReason xxx
	FailureReason ParamKey = "failureReason"

	// SuccessAddClusterNodeIDsKey xxx
	// cloud cluster success & failed Instance
	SuccessAddClusterNodeIDsKey ParamKey = "successAddClusterNodeIDs"
	// SuccessClusterNodeIDsKey xxx
	SuccessClusterNodeIDsKey ParamKey = "successClusterNodeIDs"
	// FailedClusterNodeIDsKey xxx
	FailedClusterNodeIDsKey ParamKey = "failedClusterNodeIDs"
	// FailedClusterNodeReasonKey xxx
	FailedClusterNodeReasonKey ParamKey = "failedClusterNodeReason"

	// SuccessNodeGroupIDsKey xxx
	SuccessNodeGroupIDsKey ParamKey = "successNodeGroupIDs"
	// FailedNodeGroupIDsKey xxx
	FailedNodeGroupIDsKey ParamKey = "failedNodeGroupIDs"

	// JobTypeKey xxx
	JobTypeKey ParamKey = "jobType"

	// MasterIPs xxx
	MasterIPs ParamKey = "masterIPs"
	// MasterIDs xxx
	MasterIDs ParamKey = "masterIDs"
	// CloudSystemID xxx
	CloudSystemID ParamKey = "systemID"

	// BkSopsUrlKey bk-sops depend Key
	// BkSopsUrlKey url
	BkSopsUrlKey ParamKey = "url"
	// BkSopsBizIDKey bizID
	BkSopsBizIDKey ParamKey = "template_biz_id"
	// BkSopsTemplateIDKey template ID
	BkSopsTemplateIDKey ParamKey = "template_id"
	// BkSopsTemplateUserKey template user
	BkSopsTemplateUserKey ParamKey = "template_user"
	// BkSopsTemplateSourceKey templateSource
	BkSopsTemplateSourceKey ParamKey = "template_source"
	// BkSopsConstantsKey constants
	BkSopsConstantsKey ParamKey = "constants"

	// BkSopsTaskUrlKey inject bksops task url
	BkSopsTaskUrlKey ParamKey = "taskUrl"
	// ShowSopsUrlKey show bkSops url
	ShowSopsUrlKey ParamKey = "showUrl"

	// ConnectClusterKey connect cluster result
	ConnectClusterKey ParamKey = "connectCluster"
	// InstallGseAgentKey install gse agent
	InstallGseAgentKey ParamKey = "installGseAgent"

	// cluster kubeConfig parameters

	// IsExtranetKey key
	IsExtranetKey ParamKey = "isExtranet"
	// SubnetIDKey key
	SubnetIDKey ParamKey = "subnetID"

	// VclusterSrcClusterIDKey vcluster depend key
	// SrcClusterIDKey xxx
	VclusterSrcClusterIDKey ParamKey = "vclusterSrcClusterID"
	// VclusterEtcdServersKey xxx
	VclusterEtcdServersKey ParamKey = "vclusterEtcdServers"
	// VclusterEtcdCAKey xxx
	VclusterEtcdCAKey ParamKey = "vclusterEtcdCA"
	// VclusterEtcdClientCertKey xxx
	VclusterEtcdClientCertKey ParamKey = "vclusterEtcdClientCert"
	// VclusterEtcdClientKeyKey xxx
	VclusterEtcdClientKeyKey ParamKey = "vclusterEtcdClientCert"
	// VclusterServiceCIDRKey xxx
	VclusterServiceCIDRKey ParamKey = "vclusterServiceCIDR"

	// SubnetInfoKey xxx
	SubnetInfoKey ParamKey = "subnetInfo"
	// IsStaticIpModeKey xxx
	IsStaticIpModeKey ParamKey = "isStaticIpMode"
	// ClaimExpiredSecondsKey xxx
	ClaimExpiredSecondsKey ParamKey = "claimExpiredSeconds"
)
