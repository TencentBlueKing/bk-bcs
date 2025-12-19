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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

/*
	实例(Virtual Machine Scale Set Virtual Machine)
*/

// DeleteInstance 删除实例.
func (aks *AksServiceImpl) DeleteInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	node *proto.Node) error {
	id := node.NodeID
	start := strings.IndexByte(id, '/')
	end := strings.LastIndexByte(id, '/')
	asg := info.NodeGroup.AutoScaling
	return aks.DeleteInstanceWithName(ctx, asg.AutoScalingName, asg.AutoScalingID, id[start+1:end])
}

// DeleteInstanceWithName 删除实例.
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
// instanceID - VirtualMachineScaleSetVM.instanceID()，而非VirtualMachineScaleSetVM.ID
func (aks *AksServiceImpl) DeleteInstanceWithName(ctx context.Context, nodeResourceGroup, setName,
	instanceID string) error {
	poller, err := aks.vmClient.BeginDelete(ctx, nodeResourceGroup, setName, instanceID, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,setName:%s", nodeResourceGroup,
			setName)
	}
	if _, err = poller.PollUntilDone(ctx, pollFrequency4); err != nil {
		return errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,setName:%s", nodeResourceGroup,
			setName)
	}
	return nil
}

// UpdateInstance 修改实例.
func (aks *AksServiceImpl) UpdateInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	node *proto.Node) (*proto.Node, error) {
	var (
		asg = info.NodeGroup.AutoScaling
		// attention
		nodeResourceGroup = asg.AutoScalingName
		setName           = asg.AutoScalingID
		id                = node.NodeID
		start             = strings.IndexByte(id, '/')
		end               = strings.LastIndexByte(id, '/')
	)
	vm, err := aks.GetInstanceAndReturn(ctx, nodeResourceGroup, setName, id[start+1:end])
	if err != nil {
		return nil, errors.Wrapf(err, "call GetInstanceAndReturn failed")
	}
	if err = aks.NodeToVm(node, vm); err != nil {
		return nil, errors.Wrapf(err, "call NodeToVm failed")
	}
	vm, err = aks.UpdateInstanceAndReturn(ctx, vm, nodeResourceGroup, setName)
	if err != nil {
		return nil, errors.Wrapf(err, "call UpdateInstanceAndReturn failed")
	}
	if err = aks.VmToNode(vm, node); err != nil {
		return node, errors.Wrapf(err, "call VmToNdeo failed")
	}
	return node, nil
}

// UpdateInstanceWithName 从名称修改实例.
func (aks *AksServiceImpl) UpdateInstanceWithName(ctx context.Context, vm *armcompute.VirtualMachineScaleSetVM,
	nodeResourceGroup, setName string) (*proto.Node, error) {
	vm, err := aks.UpdateInstanceAndReturn(ctx, vm, nodeResourceGroup, setName)
	if err != nil {
		return nil, errors.Wrapf(err, "call UpdateInstanceAndReturn failed")
	}
	node := new(proto.Node)
	if err = aks.VmToNode(vm, node); err != nil {
		return node, errors.Wrapf(err, "call VmToNode failed")
	}
	return node, nil
}

// UpdateInstanceAndReturn 从名称修改实例.
func (aks *AksServiceImpl) UpdateInstanceAndReturn(ctx context.Context, vm *armcompute.VirtualMachineScaleSetVM,
	nodeResourceGroup, setName string) (*armcompute.VirtualMachineScaleSetVM, error) {
	poller, err := aks.vmClient.BeginUpdate(ctx, nodeResourceGroup, setName, *vm.InstanceID, *vm, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	resp, err := poller.PollUntilDone(ctx, pollFrequency1)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull the result")
	}
	return &resp.VirtualMachineScaleSetVM, nil
}

// GetInstance 获取实例.
func (aks *AksServiceImpl) GetInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	node *proto.Node) (*proto.Node, error) {
	id := node.NodeID
	start := strings.IndexByte(id, '/')
	end := strings.LastIndexByte(id, '/')
	asg := info.NodeGroup.AutoScaling
	return aks.GetInstanceWithName(ctx, asg.AutoScalingName, asg.AutoScalingID, id[start+1:end])
}

// GetInstanceWithName 从名称获取实例.
func (aks *AksServiceImpl) GetInstanceWithName(ctx context.Context, nodeResourceGroup, setName, instanceID string) (
	*proto.Node, error) {
	vm, err := aks.GetInstanceAndReturn(ctx, nodeResourceGroup, setName, instanceID)
	if err != nil {
		return nil, errors.Wrapf(err, "call GetInstanceAndReturn falied")
	}
	node := new(proto.Node)
	if err = aks.VmToNode(vm, node); err != nil {
		return node, errors.Wrapf(err, "call VmToNode failed")
	}
	return node, nil
}

// GetInstanceAndReturn 从名称获取实例.
func (aks *AksServiceImpl) GetInstanceAndReturn(ctx context.Context, nodeResourceGroup, setName,
	instanceID string) (*armcompute.VirtualMachineScaleSetVM, error) {
	resp, err := aks.vmClient.Get(ctx, nodeResourceGroup, setName, instanceID, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.VirtualMachineScaleSetVM, nil
}

// ListInstance 获取实例列表.
func (aks *AksServiceImpl) ListInstance(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) ([]*proto.Node,
	error) {
	asg := info.NodeGroup.AutoScaling
	return aks.ListInstanceWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
}

// ListInstanceWithName 从名称获取实例列表.
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) ListInstanceWithName(ctx context.Context, nodeResourceGroup, setName string) ([]*proto.Node,
	error) {
	vms, err := aks.ListInstanceAndReturn(ctx, nodeResourceGroup, setName)
	if err != nil {
		return nil, errors.Wrapf(err, "call ListInstanceAndReturn falied")
	}
	resp := make([]*proto.Node, len(vms))
	for i, vm := range vms {
		resp[i] = new(proto.Node)
		if err = aks.VmToNode(vm, resp[i]); err != nil {
			return resp, errors.Wrapf(err, "call VmToNode falied")
		}
	}
	return resp, nil
}

// ListInstanceAndReturn 从名称获取实例列表.
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) ListInstanceAndReturn(ctx context.Context, nodeResourceGroup, setName string) (
	[]*armcompute.VirtualMachineScaleSetVM, error) {
	resp := make([]*armcompute.VirtualMachineScaleSetVM, 0)
	pager := aks.vmClient.NewListPager(nodeResourceGroup, setName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		resp = append(resp, nextResult.Value...)
	}
	return resp, nil
}

// ListInstanceByIDAndReturn 从ids获取实例列表.
//
// nodeResourceGroup - 基础结构资源组(AutoScalingGroup.autoScalingName/Cluster.ExtraInfo["nodeResourceGroup"]).
//
// setName - 虚拟机规模集名称(AutoScalingGroup.autoScalingID).
func (aks *AksServiceImpl) ListInstanceByIDAndReturn(ctx context.Context, nodeResourceGroup, setName string,
	ids []string) ([]*armcompute.VirtualMachineScaleSetVM, error) {
	vmList, err := aks.ListInstanceAndReturn(ctx, nodeResourceGroup, setName)
	if err != nil {
		return nil, errors.Wrapf(err, "call ListInstanceAndReturn falied")
	}
	idSet := make(map[string]bool)
	for i := range ids {
		idSet[ids[i]] = true
	}
	resp := make([]*armcompute.VirtualMachineScaleSetVM, 0)
	for i, vm := range vmList {
		if idSet[*vm.InstanceID] {
			resp = append(resp, vmList[i])
		}
	}
	return resp, nil
}

// CheckInstanceType 检查机型是否存在
func (aks *AksServiceImpl) CheckInstanceType(ctx context.Context, location, instanceType string) (bool, error) {
	skus, err := aks.ListResourceByLocation(ctx, location)
	if err != nil {
		return false, errors.Wrapf(err, "call ListResourceByLocation failed")
	}
	for _, sku := range skus {
		if *sku.Name == instanceType {
			return true, nil
		}
	}
	// 机型不存在
	return false, cloudprovider.ErrVMInstanceType
}

// ListKeyPairs keyPairs list
func (aks *AksServiceImpl) ListKeyPairs(opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
