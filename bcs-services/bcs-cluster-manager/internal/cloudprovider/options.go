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

package cloudprovider

import (
	"errors"
	"fmt"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// task err
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
	// ErrCloudNodeVPCDiffWithClusterResponse for node VPC different cluster VPC
	ErrCloudNodeVPCDiffWithClusterResponse = "node[%s] VPC is different from cluster VPC"
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

const (
	// TaskStatusInit INIT task status
	TaskStatusInit = "INITIALIZING"
	// TaskStatusRunning running task status
	TaskStatusRunning = "RUNNING"
	// TaskStatusSuccess task success
	TaskStatusSuccess = "SUCCESS"
	// TaskStatusSkip task skip
	TaskStatusSkip = "SKIP"
	// TaskStatusPartFailure task part failure
	TaskStatusPartFailure = "PART_FAILURE"
	// TaskStatusFailure task failed
	TaskStatusFailure = "FAILURE"
	// TaskStatusTimeout task run timeout
	TaskStatusTimeout = "TIMEOUT"
	// TaskStatusForceTerminate force task terminate
	TaskStatusForceTerminate = "FORCETERMINATE"
	// TaskStatusNotStarted force task terminate
	TaskStatusNotStarted = "NOTSTARTED"
)

// CommonOption for all option
type CommonOption struct {
	// request ID for tracing
	RequestID string

	// Account fit all cloud AKSK
	Account *proto.Account
	// region information for cloudprovider
	// region is unnecessary in some api
	Region string

	// CommonConf for cloud confInfo
	CommonConf CloudConf
}

// CloudConf for cloud other confInfo
type CloudConf struct {
	// CloudInternalEnable cloud internal conf
	CloudInternalEnable bool
	// CloudDomain for cloud domain
	CloudDomain string
	// MachineDomain for instance domain
	MachineDomain string
	// VpcDomain for vpc domain
	VpcDomain string
}

// InitClusterConfigOption init cluster default cloud config
type InitClusterConfigOption struct {
	// CommonOption for common options
	Common *CommonOption
	// Cloud for cluster
	Cloud *proto.Cloud
	// ClusterVersion for cluster version
	ClusterVersion string
}

// SyncClusterCloudInfoOption sync cluster cloud info
type SyncClusterCloudInfoOption struct {
	// CommonOption for common options
	Common *CommonOption
	// Cloud for cluster
	Cloud *proto.Cloud
	// ImportMode import mode
	ImportMode *proto.ImportCloudMode
	// ClusterVersion for cluster version
	ClusterVersion string
}

// GetNodeOption for GetNodeByIP options
type GetNodeOption struct {
	// CommonOption for common options
	Common *CommonOption
	// ClusterVPCID for cluster vpc
	ClusterVPCID string
	// ClusterID for cluster id
	ClusterID string
}

// ListNodesOption for ListNodesByIP options
type ListNodesOption struct {
	// CommonOption for common options
	Common *CommonOption
	// ClusterVPCID for cluster vpc
	ClusterVPCID string
	// ClusterID for cluster id
	ClusterID string
	// NodeTemplateID for node templateID
	NodeTemplateID string
}

// TaskOptions option for create specified task
type TaskOptions struct {
	Cloud    *proto.Cloud
	Project  *proto.Project
	Cluster  *proto.Cluster
	Operator string
}

// CreateClusterOption create cluster option
type CreateClusterOption struct {
	CommonOption
	Reinstall    bool
	InitPassword string
	Operator     string
	// cloud is used for cloudprovider template
	Cloud        *proto.Cloud
	WorkerNodes  []string
	MasterNodes  []string
	NodeTemplate *proto.NodeTemplate
}

// CreateVirtualClusterOption create virtual cluster option
type CreateVirtualClusterOption struct {
	CommonOption
	Operator string
	// cloud is used for cloudprovider template
	HostCluster *proto.Cluster
	Cloud       *proto.Cloud
	Namespace   *proto.NamespaceInfo
}

// ImportClusterOption import cluster option
type ImportClusterOption struct {
	CommonOption
	// cloud is used for cloudprovider template
	Cloud     *proto.Cloud
	CloudMode *proto.ImportCloudMode
	NodeIPs   []string
	Operator  string
}

// GetZoneListOption get zone list option
type GetZoneListOption struct {
	CommonOption
	VpcId string
	State string
}

// DeleteMode xxx
type DeleteMode string

// String toString
func (dm DeleteMode) String() string {
	return string(dm)
}

const (
	// Terminate terminate mode
	Terminate DeleteMode = "terminate"
	// Retain retain mode
	Retain DeleteMode = "retain"
)

// DeleteClusterOption delete cluster option
type DeleteClusterOption struct {
	CommonOption
	// force delete cluster
	IsForce bool
	// DeleteMode instance deleteMode(terminate/retain)
	DeleteMode DeleteMode
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// Cluster used for cloudprovider
	Cluster *proto.Cluster
	// LatsClusterStatus last cluster status
	LatsClusterStatus string
}

// DeleteVirtualClusterOption delete virtual cluster option
type DeleteVirtualClusterOption struct {
	CommonOption
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// HostCluster used for cloudprovider
	HostCluster *proto.Cluster
	// NamespaceInfo used for virtual in hostCluster
	Namespace *proto.NamespaceInfo
}

// GetNodesOption create cluster option
type GetNodesOption struct {
	CommonOption
}

// ClusterGroupOption cluster/group option
type ClusterGroupOption struct {
	CommonOption
	// Cluster xxx
	Cluster *proto.Cluster
	// Group xxx
	Group *proto.NodeGroup
}

// GetClusterOption get cluster option
type GetClusterOption struct {
	CommonOption
	// Cluster xxx
	Cluster *proto.Cluster
}

// ListClusterOption list cluster option
type ListClusterOption struct {
	CommonOption
}

// CheckClusterCIDROption check cluster CIDR
type CheckClusterCIDROption struct {
	CommonOption
	CurrentNodeCnt  uint64
	IncomingNodeCnt uint64
	ExternalNode    bool
}

// AddNodesOption add nodes to cluster option
type AddNodesOption struct {
	CommonOption
	Reinstall    bool
	InitPassword string
	Login        *proto.NodeLoginInfo
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// NodeTemplate
	NodeTemplate *proto.NodeTemplate
	// setting NodeGroupID means add to specified NodeGroup
	NodeGroupID string
	// node scheduler status
	NodeSchedule bool
}

// DeleteNodesOption create cluster option
type DeleteNodesOption struct {
	CommonOption
	// Operator user
	Operator   string
	IsForce    bool
	DeleteMode string
	// cloud is used for cloudprovider template
	Cloud        *proto.Cloud
	NodeTemplate *proto.NodeTemplate
}

// EnableExternalNodeOption enable cluster external node
type EnableExternalNodeOption struct {
	CommonOption
	// EnablePara 开启关闭第三方节点参数
	EnablePara *EnableExternalNodePara
	// Operator user
	Operator string
}

// EnableExternalNodePara paras
type EnableExternalNodePara struct {
	// NetworkType 集群网络插件类型，支持：Flannel、CiliumBGP、CiliumVXLan
	NetworkType string
	// ClusterCIDR Pod CIDR
	ClusterCIDR string
	// SubnetId 子网ID
	SubnetId string
	// 是否开启第三方节点池支持(true: 开启第三方节点 false: 关闭第三方节点)
	Enabled bool
}

// AddExternalNodesOption add external nodes to cluster option
type AddExternalNodesOption struct {
	CommonOption
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// Cluster clusterInfo
	Cluster *proto.Cluster
}

// DeleteExternalNodesOption delete cluster external nodes
type DeleteExternalNodesOption struct {
	CommonOption
	// Operator user
	Operator string
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// Cluster clusterInfo
	Cluster *proto.Cluster
}

// CreateNodeGroupOption create nodegroup option
type CreateNodeGroupOption struct {
	CommonOption
	// Cluster clusterInfo
	Cluster *proto.Cluster
	// PoolInfo for resourcePool
	PoolInfo ResourcePoolData
	// OnlyData only update data, not build task
	OnlyData bool
}

// ResourcePoolData xxx
type ResourcePoolData struct {
	Provider       string
	ResourcePoolID string
}

// Validate resourcePool data
func (rpd ResourcePoolData) Validate() error {
	if rpd.Provider == "" {
		return fmt.Errorf("ResourcePoolData provider or poolID empty")
	}

	return nil
}

// DeleteNodeGroupOption delete nodegroup option
type DeleteNodeGroupOption struct {
	CommonOption
	IsForce bool
	// reserve all nodes in cluster
	ReserveNodesInCluster bool
	// move all nodes out of cluster, clean all containers
	// but keep nodes running
	ReservedNodeInstance bool
	// move all node out of cluster and delete all nodes
	CleanInstanceInCluster bool
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// cluster is used for clusterInfo
	Cluster *proto.Cluster
	// AsOption for moduleInfo
	AsOption *proto.ClusterAutoScalingOption
	// Operator
	Operator string
	// OnlyData only update data, not build task
	OnlyData bool
}

// UpdateNodeGroupOption create nodegroup option
type UpdateNodeGroupOption struct {
	CommonOption
	// cloud is used for cloudprovider template
	Cloud *proto.Cloud
	// cluster is used for clusterInfo
	Cluster *proto.Cluster
	// OnlyData only update data, not build task
	OnlyData bool
}

// MoveNodesOption move nodes to NodeGroup management
type MoveNodesOption struct {
	CommonOption
	Cluster *proto.Cluster
}

// RemoveNodesOption remove nodes from NodeGroup,
// nodes are still in Cluster
type RemoveNodesOption struct {
	CommonOption
	Cloud   *proto.Cloud
	Cluster *proto.Cluster
}

// CleanNodesOption clean nodes in NodeGroup option
type CleanNodesOption struct {
	CommonOption
	Cluster *proto.Cluster
	Cloud   *proto.Cloud
	// AsOption for moduleInfo
	AsOption *proto.ClusterAutoScalingOption
	Operator string
}

// CleanNodesResponse response for clean nodes in NodeGroup
type CleanNodesResponse struct {
	ClusterID       string
	ResponseOrderID []string
	SuccNodes       []string
	FailedNodes     []string
}

// UpdateDesiredNodeOption update desired node
type UpdateDesiredNodeOption struct {
	CommonOption
	Cluster   *proto.Cluster
	Cloud     *proto.Cloud
	NodeGroup *proto.NodeGroup
	// AsOption for moduleInfo
	AsOption *proto.ClusterAutoScalingOption
	Operator string
	Manual   bool
}

// SwitchNodeGroupAutoScalingOption switch nodegroup auto scaling
type SwitchNodeGroupAutoScalingOption struct {
	CommonOption
	Cluster *proto.Cluster
	Cloud   *proto.Cloud
}

// ScalingResponse response for UpdateDesired nodes
type ScalingResponse struct {
	ResponseID   string
	ScalingUp    uint32
	CapableNodes []string
	Data         string
}

// CreateScalingOption create NodeGroup option
type CreateScalingOption struct {
	CommonOption
}

// DeleteScalingOption create NodeGroup option
type DeleteScalingOption struct {
	CommonOption
}

// UpdateScalingOption create NodeGroup option
type UpdateScalingOption struct {
	CommonOption
}

// CheckEndpointStatusOption check cluster endpoint status option
type CheckEndpointStatusOption struct {
	CommonOption
	ResourceGroupName string
}

// AddSubnetsToClusterOption add subnet to cluster option
type AddSubnetsToClusterOption struct {
	CommonOption
	Cluster *proto.Cluster
}

// GetMasterSuggestedMachinesOption master suggested machine
type GetMasterSuggestedMachinesOption struct {
	CommonOption
	Cpu   int
	Mem   int
	Zones []string
}

// StepInfo step parameter
type StepInfo struct {
	StepMethod string
	StepName   string
}

// InstanceInfo for get instance type
type InstanceInfo struct {
	Region       string
	Zone         string
	NodeFamily   string
	Cpu          uint32
	Memory       uint32
	BizID        string
	Provider     string
	ResourceType string
}

// MachineConfig instance config
type MachineConfig struct {
	Cpu int
	Mem int
	Gpu int
}
