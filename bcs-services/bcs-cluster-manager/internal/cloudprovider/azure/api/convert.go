package api

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

/*
	转换
*/
// NodeGroupToAgentPool 为BCS节点池 转换 Azure代理节点池(仅用于创建)
func (aks *AksServiceImpl) NodeGroupToAgentPool(group *proto.NodeGroup, pool *armcontainerservice.AgentPool) error {
	if group == nil || pool == nil {
		return errors.New("NodeGroupToAgentPool method group or pool cannot be empty")
	}
	if err := aks.toPoolPreCheck(group, pool); err != nil {
		return errors.Wrapf(err, "check group or pool failed")
	}
	converter := newNodeGroupToAgentPoolConverter(group, pool)
	converter.convert()
	return nil
}

// toPoolPreCheck 前置检查
func (aks *AksServiceImpl) toPoolPreCheck(group *proto.NodeGroup, pool *armcontainerservice.AgentPool) error {
	if pool == nil {
		return cloudprovider.ErrAgentPoolEmpty
	}
	if group == nil {
		return cloudprovider.ErrNodeGroupEmpty
	}
	if group.AutoScaling == nil {
		return cloudprovider.ErrNodeGroupAutoScalingLost
	}
	if group.LaunchTemplate == nil {
		return cloudprovider.ErrNodeGroupLaunchTemplateLost
	}
	if group.NodeTemplate == nil {
		return cloudprovider.ErrNodeGroupNodeTemplateLost
	}
	// 机型检查
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	if ok, err := aks.CheckInstanceType(ctx, group.Region, group.LaunchTemplate.InstanceType); err != nil || !ok {
		return err
	}
	return nil
}

// AgentPoolToNodeGroup Azure代理节点池 转换 为BCS节点池(仅用于创建)
func (aks *AksServiceImpl) AgentPoolToNodeGroup(pool *armcontainerservice.AgentPool, group *proto.NodeGroup) error {
	if group == nil || pool == nil {
		return errors.New("AgentPoolToNodeGroup method group or pool cannot be empty")
	}
	converter := newPoolToNodeGroupConverter(pool, group)
	converter.convert()
	return nil
}

// SetToNodeGroup Azure虚拟规模集 转换 为BCS节点池
func (aks *AksServiceImpl) SetToNodeGroup(set *armcompute.VirtualMachineScaleSet, group *proto.NodeGroup) error {
	if set == nil || group == nil {
		return errors.New("SetToNodeGroup method set or group cannot be empty")
	}
	converter := newSetToNodeGroupConverter(set, group)
	converter.convert()
	return nil
}

// NodeGroupToSet 为BCS节点池 转换 Azure虚拟规模集(仅用于创建)
func (aks *AksServiceImpl) NodeGroupToSet(group *proto.NodeGroup, set *armcompute.VirtualMachineScaleSet) error {
	if set == nil || group == nil {
		return errors.New("NodeGroupToSet method set or group cannot be empty")
	}
	converter := newNodeGroupToSetConverter(group, set)
	converter.convert()
	return nil
}

// VmToNode Azure节点 转换 为BCS节点
func (aks *AksServiceImpl) VmToNode(vm *armcompute.VirtualMachineScaleSetVM, node *proto.Node) error {
	if vm == nil || node == nil {
		return errors.New("VmToNode method vm or node cannot be empty")
	}
	converter := newVmToNodeConverter(vm, node)
	converter.convert()
	return nil
}

// NodeToVm 为BCS节点 转换 Azure节点；(仅用于修改)
func (aks *AksServiceImpl) NodeToVm(node *proto.Node, vm *armcompute.VirtualMachineScaleSetVM) error {
	if vm == nil || node == nil {
		return errors.New("NodeToVm method vm or node cannot be empty")
	}
	converter := newNodeToVmConverter(node, vm)
	converter.convert()
	return nil
}
