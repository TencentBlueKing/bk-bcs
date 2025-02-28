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
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

/*
	虚拟规模集
*/

// CreateSet 创建虚拟机规模集.
func (aks *AksServiceImpl) CreateSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	*armcompute.VirtualMachineScaleSet, error) {
	return nil, errors.New("no implement")
}

// CreateSetWithName 从名称创建虚拟机规模集(不建议手动创建).
// set - 虚拟机规模集.
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) CreateSetWithName(ctx context.Context, set *armcompute.VirtualMachineScaleSet,
	resourceGroupName, setName string) (*armcompute.VirtualMachineScaleSet, error) {
	poller, err := aks.setClient.BeginCreateOrUpdate(ctx, resourceGroupName, setName, *set, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,setName:%s",
			resourceGroupName, setName)
	}
	resp, err := poller.PollUntilDone(ctx, pollFrequency5)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,setName:%s",
			resourceGroupName, setName)
	}
	return &resp.VirtualMachineScaleSet, nil
}

// DeleteSet 删除虚拟机规模集.
func (aks *AksServiceImpl) DeleteSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	asg := info.NodeGroup.AutoScaling
	return aks.DeleteSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
}

// DeleteSetWithName 从名称删除虚拟机规模集.
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) DeleteSetWithName(ctx context.Context, resourceGroupName, setName string) error {
	poller, err := aks.setClient.BeginDelete(ctx, resourceGroupName, setName, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,setName:%s", resourceGroupName,
			setName)
	}
	if _, err = poller.PollUntilDone(ctx, pollFrequency4); err != nil {
		return errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,setName:%s", resourceGroupName,
			setName)
	}
	return nil
}

// BatchDeleteVMs 批量删除节点
// instanceIDs - 实例ID.
func (aks *AksServiceImpl) BatchDeleteVMs(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	instanceIDs []string) error {
	asg := info.NodeGroup.AutoScaling
	return aks.BatchDeleteVMsWithName(ctx, asg.AutoScalingName, asg.AutoScalingID, instanceIDs)
}

// BatchDeleteVMsWithName 批量删除节点
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
// instanceIDs - 节点ID.
func (aks *AksServiceImpl) BatchDeleteVMsWithName(ctx context.Context, resourceGroupName, setName string,
	instanceIDs []string) error {
	vmIDs := armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: make([]*string, len(instanceIDs)),
	}
	for i := range instanceIDs {
		vmIDs.InstanceIDs[i] = to.Ptr(instanceIDs[i])
	}
	// 删除
	var (
		err    error
		poller *runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse]
	)

	retry.Do(func() error {
		poller, err = aks.setClient.BeginDeleteInstances(ctx, resourceGroupName, setName, vmIDs, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to finish the request")
		}

		return err
	}, retry.Attempts(3), retry.Delay(time.Second))

	// 轮询
	if _, err = poller.PollUntilDone(ctx, pollFrequency5); err != nil {
		return errors.Wrapf(err, "failed to pull the result")
	}
	return nil
}

// UpdateSet 修改虚拟机规模集.
func (aks *AksServiceImpl) UpdateSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	*armcompute.VirtualMachineScaleSet, error) {
	asg := info.NodeGroup.AutoScaling
	nodeResourceGroup, ok := info.Cluster.ExtraInfo[common.NodeResourceGroup]
	if !ok || len(nodeResourceGroup) == 0 {
		return nil, errors.New("cluster extraInfo not nodeResourceGroup")
	}
	set, err := aks.GetSetWithName(ctx, nodeResourceGroup, info.NodeGroup.AutoScaling.AutoScalingID)
	if err != nil {
		return nil, errors.Wrapf(err, "call GetSetWithName failed")
	}
	if err = aks.NodeGroupToSet(info.NodeGroup, set); err != nil {
		return nil, errors.Wrapf(err, "call NodeGroupToSet failed")
	}
	return aks.UpdateSetWithName(ctx, set, asg.AutoScalingName, asg.AutoScalingID)
}

// UpdateSetNodeNum 修改虚拟机规模集节点数量扩容
func (aks *AksServiceImpl) UpdateSetNodeNum(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeNum int64) (*armcompute.VirtualMachineScaleSet, error) {

	asg := info.NodeGroup.AutoScaling
	nodeResourceGroup, ok := info.Cluster.ExtraInfo[common.NodeResourceGroup]
	if !ok || len(nodeResourceGroup) == 0 {
		return nil, errors.New("cluster extraInfo not nodeResourceGroup")
	}

	set, err := aks.GetSetWithName(ctx, nodeResourceGroup, info.NodeGroup.AutoScaling.AutoScalingID)
	if err != nil {
		return nil, errors.Wrapf(err, "call GetSetWithName failed")
	}
	set.SKU.Capacity = to.Ptr(*set.SKU.Capacity + nodeNum)

	return aks.UpdateSetWithName(ctx, set, asg.AutoScalingName, asg.AutoScalingID)
}

// UpdateSetWithName 从名称修改虚拟机规模集.
// set - 虚拟机规模集.
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) UpdateSetWithName(ctx context.Context, set *armcompute.VirtualMachineScaleSet,
	resourceGroupName, setName string) (*armcompute.VirtualMachineScaleSet, error) {
	poller, err := aks.setClient.BeginCreateOrUpdate(ctx, resourceGroupName, setName, *set, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,setName:%s",
			resourceGroupName, setName)
	}
	resp, err := poller.PollUntilDone(ctx, pollFrequency3)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,setName:%s",
			resourceGroupName, setName)
	}
	return &resp.VirtualMachineScaleSet, nil
}

// GetSet 获取虚拟机规模集.
func (aks *AksServiceImpl) GetSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	*armcompute.VirtualMachineScaleSet, error) {
	asg := info.NodeGroup.AutoScaling
	return aks.GetSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
}

// GetSetWithName 从名称获取虚拟机规模集.
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) GetSetWithName(ctx context.Context, resourceGroupName, setName string) (
	*armcompute.VirtualMachineScaleSet, error) {
	resp, err := aks.setClient.Get(ctx, resourceGroupName, setName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,setName:%s",
			resourceGroupName, setName)
	}
	return &resp.VirtualMachineScaleSet, nil
}

// ListSet 获取虚拟机规模集列表.
func (aks *AksServiceImpl) ListSet(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	[]*armcompute.VirtualMachineScaleSet, error) {
	return aks.ListSetWithName(ctx, info.NodeGroup.AutoScaling.AutoScalingName)
}

// ListSetWithName 从名称获取虚拟机规模集列表.
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
func (aks *AksServiceImpl) ListSetWithName(ctx context.Context, resourceGroupName string) (
	[]*armcompute.VirtualMachineScaleSet, error) {
	resp := make([]*armcompute.VirtualMachineScaleSet, 0)
	pager := aks.setClient.NewListPager(resourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		resp = append(resp, nextResult.Value...)
	}
	return resp, nil
}

// MatchNodeGroup 匹配节点池
//
// resourceGroupName - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) MatchNodeGroup(ctx context.Context, resourceGroupName, poolName string) (
	*armcompute.VirtualMachineScaleSet, error) {
	pager := aks.setClient.NewListPager(resourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		for i, v := range nextResult.Value {
			name, ok := v.Tags[aksManagedPoolName]
			if ok && *name == poolName {
				return nextResult.Value[i], nil
			}
		}
	}
	return nil, nil
}
