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

package azure

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// errors
var (
	nodePoolScaleUpErr = errors.New("the status of the aks node pool is scale up, and currently " +
		"no operations can be performed on it")
	nodePoolUpdatingErr = errors.New("the aks node pool status is in the process of being updating " +
		"and no operations can be performed on it right now")
)

var groupMgr sync.Once

func init() {
	groupMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

// NodeGroup nodegroup management in azure
type NodeGroup struct {
}

// CreateNodeGroup 创建节点池 - create nodegroup by cloudprovider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildCreateNodeGroupTask(group, opt)
	if err != nil {
		blog.Errorf("build CreateNodeGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// DeleteNodeGroup 删除节点池 - delete nodegroup by cloudprovider api, all nodes belong to NodeGroup
// will be released. Task is background automatic task
func (ng *NodeGroup) DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildDeleteNodeGroupTask(group, nodes, opt)
	if err != nil {
		blog.Errorf("build DeleteNodeGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// UpdateNodeGroup 更新云上节点池 - update specified nodegroup configuration
func (ng *NodeGroup) UpdateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.UpdateNodeGroupOption) (
	*proto.Task, error) {
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return nil, err
	}

	// create aks client
	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create aks client failed, err: %s", err.Error())
		return nil, err
	}
	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}

	// update agent pool
	if err = ng.updateAgentPoolProperties(client, cluster, group); err != nil {
		return nil, errors.Wrapf(err, "UpdateNodeGroup: call updateAgentPoolProperties failed")
	}
	// update virtual machine scale set
	if err = ng.updateVMSSProperties(client, group); err != nil {
		return nil, errors.Wrapf(err, "UpdateNodeGroup: call updateVMSSProperties failed")
	}

	// update bkCloudName
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		bkBizID, _ := strconv.Atoi(cluster.BusinessID)
		bkModuleID, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
		group.NodeTemplate.Module.ScaleOutModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}

	// note:Azure不支持更换镜像
	//// update imageName
	//if err = ng.updateImageInfo(group); err != nil {
	//	return err
	//}

	return nil, nil
}

// GetNodesInGroup 从云上拉取该节点池的所有节点 - get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.Node,
	error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroupV2 get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroupV2(group *proto.NodeGroup,
	opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode, error) {
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}
	asg := group.AutoScaling
	// new client
	client, err := api.NewAksServiceImplWithCommonOption(opt)
	if err != nil {
		return nil, errors.Wrapf(err, "create aks client failed.")
	}
	// 获取 vm list
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	vmList, err := client.ListInstanceAndReturn(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return nil, errors.Wrapf(err, "ListNodeWithNodeGroup failed.")
	}
	// 获取 interface list
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	interfaceList, err := client.ListSetInterfaceAndReturn(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return nil, errors.Wrapf(err, "ListVMSSsInterfaceWithNodeGroup failed")
	}
	vmIPMap := api.VmMatchInterface(vmList, interfaceList)

	groupNodes := make([]*proto.NodeGroupNode, 0)
	for _, v := range vmList {
		if v.InstanceID == nil {
			continue
		}
		node := transAksNodeToNode(v, vmIPMap)
		node.ClusterID = group.ClusterID
		node.NodeGroupID = group.NodeGroupID
		groupNodes = append(groupNodes, node)
	}
	return groupNodes, nil
}

// MoveNodesToGroup 添加节点到节点池中 - add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *cloudprovider.MoveNodesOption,
) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when MoveNodesToGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildMoveNodesToGroupTask(nodes, group, opt) // 不支持
	if err != nil {
		blog.Errorf("build MoveNodesToGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// RemoveNodesFromGroup 缩容（保留节点） - remove nodes from NodeGroup, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	// 不支持
	return cloudprovider.ErrCloudNotImplemented
}

// CleanNodesInGroup 缩容（不保留节点） - clean specified nodes in NodeGroup
func (ng *NodeGroup) CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	if len(nodes) == 0 || opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("invalid request")
	}

	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CleanNodesInGroup %s failed, %s",
			cloudName, group.Name, err.Error())
		return nil, err
	}

	task, err := mgr.BuildCleanNodesInGroupTask(nodes, group, opt)
	if err != nil {
		blog.Errorf("build CleanNodesInGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error())
		return nil, err
	}
	return task, nil
}

// UpdateDesiredNodes 扩容 - update nodegroup desired node
func (ng *NodeGroup) UpdateDesiredNodes(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (res *cloudprovider.ScalingResponse, err error) {
	if group == nil || opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("invalid request")
	}
	if err = ng.scaleUpPreCheck(opt.Cluster.ClusterID, opt.Cloud.CloudID, group.NodeGroupID); err != nil { //扩容前置检查
		return nil, errors.Wrapf(err, "nodeGroupID: %s unable to sacle up", group.NodeGroupID)
	}

	// scaling nodes with desired, first get all node for status filtering
	// check if nodes are already in cluster
	goodNodes, err := cloudprovider.ListNodesInClusterNodePool(opt.Cluster.ClusterID, group.NodeGroupID)
	if err != nil {
		return nil, errors.Wrapf(err, "cloudprovider aks get NodeGroup %s all Nodes failed", group.NodeGroupID)
	}

	// check incoming nodes
	inComingNodes, err := cloudprovider.GetNodesNumWhenApplyInstanceTask(
		opt.Cluster.ClusterID,
		group.NodeGroupID,
		cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode),
		cloudprovider.TaskStatusRunning,
		[]string{cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.ApplyInstanceMachinesTask)},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "UpdateDesiredNodes GetNodesNumWhenApplyInstanceTask failed.")
	}
	if inComingNodes > 0 {
		return nil, errors.Wrapf(nodePoolScaleUpErr, "nodeGroupID: %s is scale up, incoming nodes %d",
			group.NodeGroupID, inComingNodes)
	}

	// cluster current node
	current := len(goodNodes) + inComingNodes
	nodeNames := make([]string, 0)
	for _, node := range goodNodes {
		nodeNames = append(nodeNames, node.InnerIP)
	}
	blog.Infof("NodeGroup %s has total nodes %d, current capable nodes %d, current incoming nodes %d, details %v",
		group.NodeGroupID, len(goodNodes), current, inComingNodes, nodeNames)

	if current >= int(desired) {
		blog.Infof("NodeGroup %s current capable nodes %d larger than desired %d nodes, nothing to do",
			group.NodeGroupID, current, desired)
		res = &cloudprovider.ScalingResponse{
			ScalingUp:    0,
			CapableNodes: nodeNames,
		}
		err = fmt.Errorf("NodeGroup %s UpdateDesiredNodes nodes %d larger than desired %d nodes",
			group.NodeGroupID, current, desired)
		return res, err
	}

	// current scale nodeNum
	scalingUp := int(desired) - current
	return &cloudprovider.ScalingResponse{
		ScalingUp:    uint32(scalingUp),
		CapableNodes: nodeNames,
	}, nil
}

// SwitchNodeGroupAutoScaling 开/关CA - switch nodegroup auto scaling
func (ng *NodeGroup) SwitchNodeGroupAutoScaling(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchNodeGroupAutoScaling %s failed, %s",
			cloudName, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchNodeGroupAutoScalingTask(group, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchNodeGroupAutoScaling task for nodeGroup %s with cloudprovider %s failed, %s",
			group.NodeGroupID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// CreateAutoScalingOption create cluster autoscaling option, cloudprovider will
// deploy cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.CreateScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateAutoScalingOption %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchAsOptionStatusTask(scalingOption, true, &opt.CommonOption)
	if err != nil {
		blog.Errorf("build CreateAutoScalingOption task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
// cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteAutoScalingOption %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchAsOptionStatusTask(scalingOption, false, &opt.CommonOption)
	if err != nil {
		blog.Errorf("build DeleteAutoScalingOption task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// UpdateAutoScalingOption 更新CA参数 - update cluster autoscaling option, cloudprovider will update
// cluster-autoscaler configuration in backgroup according cloudprovider implementation.
// Implementation is optional.
func (ng *NodeGroup) UpdateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.UpdateScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when UpdateAutoScalingOption %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildUpdateAutoScalingOptionTask(scalingOption, opt)
	if err != nil {
		blog.Errorf("build UpdateAutoScalingOption task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// SwitchAutoScalingOptionStatus 更新CA状态 - switch cluster autoscaling option status
func (ng *NodeGroup) SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchAutoScalingOptionStatus %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchAsOptionStatusTask(scalingOption, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchAutoScalingOptionStatus task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// AddExternalNodeToCluster add external to cluster
func (ng *NodeGroup) AddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteExternalNodeFromCluster remove external node from cluster
func (ng *NodeGroup) DeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeScript get nodegroup external node script
func (ng *NodeGroup) GetExternalNodeScript(group *proto.NodeGroup) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// transAksNodeToNode 节点转换
func transAksNodeToNode(node *armcompute.VirtualMachineScaleSetVM, vmIPMap map[string][]string) *proto.NodeGroupNode {
	n := &proto.NodeGroupNode{NodeID: *node.InstanceID}
	// azure 默认为节点，无法获取master
	properties := node.Properties
	if properties != nil && properties.ProvisioningState != nil {
		switch *properties.ProvisioningState {
		case api.NormalState:
			n.Status = common.StatusRunning
		case api.CreatingState:
			n.Status = common.StatusInitialization
		//case "failed":
		//	n.Status = "FAILED"
		default:
			n.Status = *properties.ProvisioningState
		}
	}
	if list, ok := vmIPMap[*node.Name]; ok && len(list) != 0 {
		n.InnerIP = list[0]
	}
	return n
}

// updateAgentPoolProperties 更新 AKS 代理节点池 - update agent pool
func (ng *NodeGroup) updateAgentPoolProperties(client api.AksService, cluster *proto.Cluster,
	group *proto.NodeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := client.GetPoolAndReturn(ctx, cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetAgentPool api failed")
	}
	if err = checkPoolState(pool); err != nil { // 更新前检查节点池的状态
		return errors.Wrapf(err, "nodeGroupID: %s unable to update agent pool", group.NodeGroupID)
	}

	// 更新 pool
	api.SetAgentPoolFromNodeGroup(group, pool)

	// update agent pool
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdatePoolAndReturn(ctx, pool, cluster.SystemID, *pool.Name); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateAgentPool api failed")
	}

	return nil
}

// updateVMSSProperties 更新虚拟机规模集 - update virtual machine scale set
func (ng *NodeGroup) updateVMSSProperties(client api.AksService, group *proto.NodeGroup) error {
	asg := group.AutoScaling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	set, err := client.GetSetWithName(ctx, asg.AutoScalingName, asg.AutoScalingID)
	if err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call GetSetWithName api failed")
	}

	if group.LaunchTemplate != nil && len(group.LaunchTemplate.UserData) != 0 {
		set.Properties.VirtualMachineProfile.UserData = to.Ptr(group.LaunchTemplate.UserData)
	}
	// 镜像引用-暂时置空处理，若不置空会导致无法更新set
	api.SetImageReferenceNull(set)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err = client.UpdateSetWithName(ctx, set, asg.AutoScalingName, asg.AutoScalingID); err != nil {
		return errors.Wrapf(err, "UpdateNodeGroup: call UpdateSetWithName api failed")
	}

	return nil
}

// scaleUpPreCheck 扩容前置检查
func (ng *NodeGroup) scaleUpPreCheck(clusterID, cloudID, nodeGroupID string) error {
	info, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		return errors.Wrapf(err, "call GetClusterDependBasicInfo failed")
	}

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "new azure client failed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	pool, err := client.GetPoolAndReturn(ctx, info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
	if err != nil {
		return errors.Wrapf(err, "call GetPoolAndReturn failed")
	}

	return checkPoolState(pool)
}

// checkPoolState 更新前，检查节点池的状态
// 如果节点池正在 "更新中" 或 "扩容中"，将无法对其进行操作
func checkPoolState(pool *armcontainerservice.AgentPool) error {
	state := *pool.Properties.ProvisioningState
	if state == api.UpdatingState {
		return errors.Wrapf(nodePoolUpdatingErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	if state == api.ScalingState {
		return errors.Wrapf(nodePoolScaleUpErr, "cloudNodeGroupID: %s", *pool.Name)
	}
	return nil
}
