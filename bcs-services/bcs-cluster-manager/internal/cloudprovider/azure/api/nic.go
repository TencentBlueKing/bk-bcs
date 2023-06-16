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

package api

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
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
