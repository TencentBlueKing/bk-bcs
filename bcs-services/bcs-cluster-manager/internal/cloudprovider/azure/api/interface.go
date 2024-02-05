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

package api

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// Converter aks 与 bcs 内部结构转换器
type Converter interface {
	// AgentPoolToNodeGroup Azure代理节点池 转换 为BCS节点池
	AgentPoolToNodeGroup(pool *armcontainerservice.AgentPool, group *proto.NodeGroup) error
	// NodeGroupToAgentPool 为BCS节点池 转换 Azure代理节点池(仅用于创建)
	NodeGroupToAgentPool(group *proto.NodeGroup, pool *armcontainerservice.AgentPool) error
	// SetToNodeGroup Azure虚拟规模集 转换 为BCS节点池
	SetToNodeGroup(set *armcompute.VirtualMachineScaleSet, group *proto.NodeGroup) error
	// NodeGroupToSet 为BCS节点池 转换 Azure虚拟规模集(仅用于创建)
	NodeGroupToSet(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) error
	// VmToNode Azure节点 转换 为BCS节点
	VmToNode(vm *armcompute.VirtualMachineScaleSetVM, node *proto.Node) error
	// NodeToVm 为BCS节点 转换 Azure节点；(仅用于修改)
	NodeToVm(node *proto.Node, vm *armcompute.VirtualMachineScaleSetVM) error
}

// AksService aks service
type AksService interface {
	Converter
	ClusterService
	AgentPoolService
	SetInstanceService
	ResourceSKUsService
	NetworkInterfaceService
	VirtualMachineScaleSetService
}

// ClusterService 集群(cluster Pool)
type ClusterService interface {
	// GetCluster 查询集群
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*armcontainerservice.ManagedCluster,
		error)

	// GetClusterWithName 查询集群
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterWithName(ctx context.Context, resourceGroupName, resourceName string) (*armcontainerservice.ManagedCluster,
		error)

	// ListClusterByLocation 根据集群位置查询
	//
	// location - 位置
	ListClusterByLocation(ctx context.Context, location string) ([]*armcontainerservice.ManagedCluster, error)

	// ListClusterByResourceGroupName 根据集群资源组名称查询
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	ListClusterByResourceGroupName(ctx context.Context, location, resourceGroupName string) (
		[]*armcontainerservice.ManagedCluster, error)

	// DeleteCluster 删除集群
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	DeleteCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error

	// DeleteClusterWithName 删除集群
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	DeleteClusterWithName(ctx context.Context, resourceGroupName, resourceName string) error

	// GetClusterAdminCredentials 获取集群管理凭证
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterAdminCredentials(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
		[]*armcontainerservice.CredentialResult, error)

	// GetClusterAdminCredentialsWithName 获取集群管理凭证
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterAdminCredentialsWithName(ctx context.Context, resourceGroupName, resourceName string) (
		[]*armcontainerservice.CredentialResult, error)

	// GetClusterUserCredentials 获取集群用户凭证
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterUserCredentials(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
		[]*armcontainerservice.CredentialResult, error)

	// GetClusterUserCredentialsWithName 获取集群用户凭证
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterUserCredentialsWithName(ctx context.Context, resourceGroupName, resourceName string) (
		[]*armcontainerservice.CredentialResult, error)

	// GetClusterMonitoringUserCredentials 获取集群监控凭证
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterMonitoringUserCredentials(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
		[]*armcontainerservice.CredentialResult, error)
	// GetClusterMonitorUserCredWithName 获取集群监控凭证(GetClusterMonitoringUserCredentialsWithName)
	//
	// resourceGroupName - 资源组名称(Account.resourceGroupName)
	//
	// resourceName - K8S名称(Cluster.SystemID).
	GetClusterMonitorUserCredWithName(ctx context.Context, resourceGroupName, resourceName string) (
		[]*armcontainerservice.CredentialResult, error)
}

// AgentPoolService ...	Agent Pool Service 代理节点池
// NOCC:golint/interfacecomment(检查工具规则误报)
type AgentPoolService interface {
	// CreatePool 创建节点池.
	CreatePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup, error)

	// CreatePoolWithName 从名称创建节点池.
	//
	// pool - 代理节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	CreatePoolWithName(ctx context.Context, pool *armcontainerservice.AgentPool, resourceName, poolName string,
		group *proto.NodeGroup) (*proto.NodeGroup, error)

	// CreatePoolAndReturn 从名称创建节点池.
	// pool - 代理节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	CreatePoolAndReturn(ctx context.Context, pool *armcontainerservice.AgentPool, resourceName, poolName string) (
		*armcontainerservice.AgentPool, error)

	// DeletePool 删除节点池.
	DeletePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error

	// DeletePoolWithName 从名称删除节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	DeletePoolWithName(ctx context.Context, resourceName, poolName string) error

	// UpdatePool 修改节点池(全量修改).
	UpdatePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup, error)

	// UpdatePoolAndReturn 从名称修改节点池.
	//
	// pool - 代理节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	UpdatePoolAndReturn(ctx context.Context, pool *armcontainerservice.AgentPool, resourceName, poolName string) (
		*armcontainerservice.AgentPool, error)

	// GetPool 获取节点池.
	GetPool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup, error)

	// GetPoolWithName 从名称获取节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	GetPoolWithName(ctx context.Context, resourceName, poolName string, group *proto.NodeGroup) (*proto.NodeGroup, error)

	// GetPoolAndReturn 从名称获取节点池.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	GetPoolAndReturn(ctx context.Context, resourceName, poolName string) (*armcontainerservice.AgentPool, error)

	// ListPool 获取节点池列表.
	ListPool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) ([]*proto.NodeGroup, error)

	// ListPoolWithName 从名称获取节点池列表.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	ListPoolWithName(ctx context.Context, resourceName string) ([]*proto.NodeGroup, error)

	// ListPoolAndReturn 从名称获取节点池列表.
	//
	// resourceName - K8S名称(Cluster.SystemID).
	ListPoolAndReturn(ctx context.Context, resourceName string) ([]*armcontainerservice.AgentPool, error)
}

// VirtualMachineScaleSetService 虚拟机规模集(Virtual Machine Scale Set)
type VirtualMachineScaleSetService interface {
	// CreateSet 创建虚拟机规模集.
	CreateSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*armcompute.VirtualMachineScaleSet, error)

	// CreateSetWithName 从名称创建虚拟机规模集.
	//
	// set - 虚拟机规模集.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	CreateSetWithName(ctx context.Context, set *armcompute.VirtualMachineScaleSet, nodeResourceGroup, setName string) (
		*armcompute.VirtualMachineScaleSet, error)

	// DeleteSet 删除虚拟机规模集.
	DeleteSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error

	// DeleteSetWithName 从名称删除虚拟机规模集.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	DeleteSetWithName(ctx context.Context, nodeResourceGroup, setName string) error

	// BatchDeleteVMs 批量删除节点
	//
	// instanceIDs - 实例ID.
	BatchDeleteVMs(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, instanceIDs []string) error

	// BatchDeleteVMsWithName 批量删除节点
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	//
	// instanceIDs - 节点ID.
	BatchDeleteVMsWithName(ctx context.Context, nodeResourceGroup, setName string, instanceIDs []string) error

	// UpdateSet 修改虚拟机规模集.
	UpdateSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*armcompute.VirtualMachineScaleSet, error)

	// UpdateSetWithName 从名称修改虚拟机规模集.
	//
	// set - 虚拟机规模集.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	UpdateSetWithName(ctx context.Context, set *armcompute.VirtualMachineScaleSet, nodeResourceGroup, setName string) (
		*armcompute.VirtualMachineScaleSet, error)

	// GetSet 获取虚拟机规模集.
	GetSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*armcompute.VirtualMachineScaleSet, error)

	// GetSetWithName 从名称获取虚拟机规模集.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	GetSetWithName(ctx context.Context, nodeResourceGroup, setName string) (*armcompute.VirtualMachineScaleSet, error)

	// ListSet 获取虚拟机规模集列表.
	ListSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) ([]*armcompute.VirtualMachineScaleSet, error)

	// ListSetWithName 从名称获取虚拟机规模集列表.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	ListSetWithName(ctx context.Context, nodeResourceGroup string) ([]*armcompute.VirtualMachineScaleSet, error)

	// MatchNodeGroup 匹配节点池
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
	MatchNodeGroup(ctx context.Context, nodeResourceGroup, poolName string) (*armcompute.VirtualMachineScaleSet, error)
}

// SetInstanceService 规模集实例(Virtual Machine Scale Set Virtual Machine)
type SetInstanceService interface {
	// DeleteInstance 删除实例.
	DeleteInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, node *proto.Node) error

	// DeleteInstanceWithName 删除实例.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	//
	// instanceID - VirtualMachineScaleSetVM.instanceID，而非VirtualMachineScaleSetVM.ID
	DeleteInstanceWithName(ctx context.Context, nodeResourceGroup, setName, instanceID string) error

	// UpdateInstance 修改实例.
	UpdateInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, node *proto.Node) (*proto.Node, error)

	// UpdateInstanceWithName 从名称修改实例.
	UpdateInstanceWithName(ctx context.Context, vm *armcompute.VirtualMachineScaleSetVM, nodeResourceGroup,
		setName string) (*proto.Node, error)

	// UpdateInstanceAndReturn 从名称修改实例.
	UpdateInstanceAndReturn(ctx context.Context, vm *armcompute.VirtualMachineScaleSetVM, nodeResourceGroup,
		setName string) (*armcompute.VirtualMachineScaleSetVM, error)

	// GetInstance 获取实例.
	GetInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, node *proto.Node) (*proto.Node, error)

	// GetInstanceWithName 从名称获取实例.
	GetInstanceWithName(ctx context.Context, nodeResourceGroup, setName, instanceID string) (*proto.Node, error)

	// GetInstanceAndReturn 从名称获取实例.
	GetInstanceAndReturn(ctx context.Context, nodeResourceGroup, setName, instanceID string) (
		*armcompute.VirtualMachineScaleSetVM, error)

	// ListInstance 获取实例列表.
	ListInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) ([]*proto.Node, error)

	// ListInstanceWithName 从名称获取实例列表.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	ListInstanceWithName(ctx context.Context, nodeResourceGroup, setName string) ([]*proto.Node, error)

	// ListInstanceAndReturn 从名称获取实例列表.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	ListInstanceAndReturn(ctx context.Context, nodeResourceGroup, setName string) (
		[]*armcompute.VirtualMachineScaleSetVM, error)

	// ListInstanceByIDAndReturn 从ids获取实例列表.
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	ListInstanceByIDAndReturn(ctx context.Context, nodeResourceGroup, setName string, ids []string) (
		[]*armcompute.VirtualMachineScaleSetVM, error)

	// CheckInstanceType 检查机型是否存在
	CheckInstanceType(ctx context.Context, location, instanceType string) (bool, error)

	// ListVMSize 获取VM机型(ListVMSize)
	//
	// location - 区域名称
	ListVMSize(ctx context.Context, location string) ([]*armcompute.VirtualMachineSize, error)

	// ListOsImage 获取VM操作系统镜像(ListOsImage)
	//
	// location - 区域名称
	//
	// publisher - OS发行商
	//
	// offer - OS提供商.
	ListOsImage(ctx context.Context, location, publisher, offer string,
		options *armcompute.VirtualMachineImagesClientListSKUsOptions) (
		[]*armcompute.VirtualMachineImageResource, error)

	// ListSSHPublicKeys 获取SSH public keys
	ListSSHPublicKeys(ctx context.Context, resourceGroupName string) ([]*armcompute.SSHPublicKeyResource, error)
}

// NetworkInterfaceService 网络接口(Network Interface)
type NetworkInterfaceService interface {
	// GetVmInterfaceAndReturn 查询vm的网卡
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	//
	// instanceID - VirtualMachineScaleSetVM.instanceID，而非VirtualMachineScaleSetVM.ID
	//
	// networkInterfaceName - 网卡名称
	GetVmInterfaceAndReturn(ctx context.Context, nodeResourceGroup, setName, instanceID, networkInterfaceName string) (
		*armnetwork.Interface, error)

	// GetVirtualNetworks 查询虚拟网络
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// virtualNetworkName - 虚拟网络(AutoScalingGroup.VpcID).
	GetVirtualNetworks(ctx context.Context, nodeResourceGroup, virtualNetworkName string) (*armnetwork.VirtualNetwork,
		error)

	// ListVirtualNetwork 虚拟网络列表
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	ListVirtualNetwork(ctx context.Context, nodeResourceGroup string) ([]*armnetwork.VirtualNetwork, error)

	// ListSubnets 虚拟子网列表
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	ListSubnets(ctx context.Context, nodeResourceGroup, vpcName string) ([]*armnetwork.Subnet, error)

	// GetSubnet 虚拟子网
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	GetSubnet(ctx context.Context, nodeResourceGroup, vpcName, subnetName string) (*armnetwork.Subnet, error)

	// UpdateSubnet 创建或更新虚拟子网
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	UpdateSubnet(ctx context.Context, nodeResourceGroup, vpcName, subnetName string,
		subnet armnetwork.Subnet) (*armnetwork.Subnet, error)

	// GetNetworkSecurityGroups 查询安全组
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// networkSecurityGroupName - 安全组名称(LaunchConfiguration.securityGroupIDs).
	GetNetworkSecurityGroups(ctx context.Context, nodeResourceGroup, networkSecurityGroupName string) (
		*armnetwork.SecurityGroup, error)

	ListNetworkSecurityGroups(ctx context.Context, nodeResourceGroup string) ([]*armnetwork.SecurityGroup, error)
	// ListSetInterfaceAndReturn 查询set中vm的网卡
	//
	// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
	//
	// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
	ListSetInterfaceAndReturn(ctx context.Context, nodeResourceGroup, setName string) ([]*armnetwork.Interface, error)
}

// ResourceSKUsService 资源
type ResourceSKUsService interface {
	// ListResourceByLocation 从区域获取资源
	ListResourceByLocation(ctx context.Context, location string) ([]*armcompute.ResourceSKU, error)
}
