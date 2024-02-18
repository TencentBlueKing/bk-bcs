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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/pkg/errors"
)

/*
	网络接口（network interface controller，NIC）
*/

// GetVmInterfaceAndReturn 查询vm的网卡
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
//
// instanceID - VirtualMachineScaleSetVM.instanceID，而非VirtualMachineScaleSetVM.ID
//
// networkInterfaceName - 网卡名称
func (aks *AksServiceImpl) GetVmInterfaceAndReturn(ctx context.Context, nodeResourceGroup, setName, instanceID,
	networkInterfaceName string) (*armnetwork.Interface, error) {
	resp, err := aks.netClient.GetVirtualMachineScaleSetNetworkInterface(ctx, nodeResourceGroup,
		setName, instanceID, networkInterfaceName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.Interface, nil
}

// GetVirtualNetworks 查询vpc
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// virtualNetworkName - 虚拟网络(AutoScalingGroup.VpcID).
func (aks *AksServiceImpl) GetVirtualNetworks(ctx context.Context, nodeResourceGroup, virtualNetworkName string) (
	*armnetwork.VirtualNetwork, error) {
	resp, err := aks.vnetClient.Get(ctx, nodeResourceGroup, virtualNetworkName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.VirtualNetwork, nil
}

// ListVirtualNetwork 虚拟网络列表
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) ListVirtualNetwork(ctx context.Context, nodeResourceGroup string) (
	[]*armnetwork.VirtualNetwork, error) {
	resp := make([]*armnetwork.VirtualNetwork, 0)
	page := aks.vnetClient.NewListPager(nodeResourceGroup, nil)
	for page.More() {
		nextPage, err := page.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to finish the request")
		}
		resp = append(resp, nextPage.Value...)
	}
	return resp, nil
}

// ListSubnets 虚拟子网列表
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) ListSubnets(ctx context.Context, nodeResourceGroup, vpcName string) (
	[]*armnetwork.Subnet, error) {
	if nodeResourceGroup == "" || vpcName == "" {
		return nil, fmt.Errorf("nodeResourceGroup or vpcName cannot be empty")
	}
	resp := make([]*armnetwork.Subnet, 0)
	page := aks.subnetClient.NewListPager(nodeResourceGroup, vpcName, nil)
	for page.More() {
		nextPage, err := page.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to finish the request")
		}
		resp = append(resp, nextPage.Value...)
	}
	return resp, nil
}

// GetSubnet 虚拟子网
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) GetSubnet(ctx context.Context, nodeResourceGroup, vpcName, subnetName string) (
	*armnetwork.Subnet, error) {
	if nodeResourceGroup == "" || vpcName == "" || subnetName == "" {
		return nil, fmt.Errorf("nodeResourceGroup or vpcName or subnetName cannot be empty")
	}
	resp, err := aks.subnetClient.Get(ctx, nodeResourceGroup, vpcName, subnetName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.Subnet, nil
}

// UpdateSubnet 创建或更新虚拟子网
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) UpdateSubnet(ctx context.Context, nodeResourceGroup, vpcName, subnetName string,
	subnet armnetwork.Subnet) (*armnetwork.Subnet, error) {
	if nodeResourceGroup == "" || vpcName == "" || subnetName == "" {
		return nil, fmt.Errorf("nodeResourceGroup or vpcName or subnetName cannot be empty")
	}
	poller, err := aks.subnetClient.BeginCreateOrUpdate(ctx, nodeResourceGroup, vpcName, subnetName, subnet, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	resp, err := poller.PollUntilDone(ctx, pollFrequency1)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull the result")
	}
	return &resp.Subnet, nil
}

// GetNetworkSecurityGroups 查询安全组
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// networkSecurityGroupName - 安全组名称(LaunchConfiguration.securityGroupIDs).
func (aks *AksServiceImpl) GetNetworkSecurityGroups(ctx context.Context, nodeResourceGroup,
	networkSecurityGroupName string) (*armnetwork.SecurityGroup, error) {
	resp, err := aks.securityGroupsClient.Get(ctx, nodeResourceGroup, networkSecurityGroupName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.SecurityGroup, nil
}

// ListNetworkSecurityGroups 安全组列表
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) ListNetworkSecurityGroups(ctx context.Context, nodeResourceGroup string) (
	[]*armnetwork.SecurityGroup, error) {
	pager := aks.securityGroupsClient.NewListPager(nodeResourceGroup, nil)
	result := make([]*armnetwork.SecurityGroup, 0)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		result = append(result, next.Value...)
	}

	return result, nil
}

// ListNetworkSecurityGroupsAll 全量安全组列表
func (aks *AksServiceImpl) ListNetworkSecurityGroupsAll(ctx context.Context) (
	[]*armnetwork.SecurityGroup, error) {
	pager := aks.securityGroupsClient.NewListAllPager(nil)
	result := make([]*armnetwork.SecurityGroup, 0)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		result = append(result, next.Value...)
	}

	return result, nil
}

// ListSetInterfaceAndReturn 查询set中vm的网卡
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) ListSetInterfaceAndReturn(ctx context.Context, nodeResourceGroup, setName string) (
	[]*armnetwork.Interface, error) {
	resp := make([]*armnetwork.Interface, 0)
	poller := aks.netClient.NewListVirtualMachineScaleSetNetworkInterfacesPager(nodeResourceGroup, setName, nil)
	for poller.More() {
		nextPage, err := poller.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to finish the request")
		}
		resp = append(resp, nextPage.Value...)
	}
	return resp, nil
}
