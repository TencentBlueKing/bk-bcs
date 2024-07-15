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

package huawei

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

var groupMgr sync.Once

func init() {
	groupMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

// NodeGroup nodegroup management in huawei
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

	client, err := api.NewCceClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("UpdateNodeGroup[]: get cce client failed, %s", err.Error())
		return nil, err
	}

	// 获取节点池信息
	rsp, err := client.GetClusterNodePool(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("GetClusterNodePool[]: get cluster nodePool failed, %s", err.Error())
		return nil, err
	}

	_, err = client.UpdateNodePoolV2(api.GenerateModifyClusterNodePoolInput(group, cluster.SystemID, rsp))
	if err != nil {
		return nil, err
	}

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
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return nil, err
	}

	client, err := api.NewCceClient(opt)
	if err != nil {
		blog.Errorf("GetNodesInGroup[]: get cce client  failed, %s", err.Error())
		return nil, err
	}

	nodes, err := client.ListClusterNodePoolNodes(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("GetNodeGroupInstances failed, err: %s", err.Error())
		return nil, err
	}

	groupNodes := make([]*proto.NodeGroupNode, 0)
	for _, v := range nodes {
		node := &proto.NodeGroupNode{NodeID: *v.Metadata.Uid}

		if *v.Status.PrivateIPv6IP != "" {
			node.InnerIP = *v.Status.PrivateIPv6IP
		}
		if *v.Status.PrivateIP != "" {
			node.InnerIP = *v.Status.PrivateIP
		}

		switch v.Status.Phase.Value() {
		case model.GetNodeStatusPhaseEnum().ACTIVE.Value():
			node.Status = common.StatusRunning
		case model.GetNodeStatusPhaseEnum().BUILD.Value(),
			model.GetNodeStatusPhaseEnum().INSTALLING.Value(),
			model.GetNodeStatusPhaseEnum().UPGRADING.Value():
			node.Status = common.StatusInitialization
		// case model.GetNodeStatusPhaseEnum().ERROR.Value():
		//	node.Status = "FAILED"
		default:
			node.Status = v.Status.Phase.Value()
		}

		node.NodeGroupID = group.NodeGroupID
		node.ClusterID = group.ClusterID
		groupNodes = append(groupNodes, node)
	}

	return groupNodes, nil
}

// MoveNodesToGroup 添加节点到节点池中 - add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *cloudprovider.MoveNodesOption,
) (*proto.Task, error) {
	// 不支持
	return nil, cloudprovider.ErrCloudNotImplemented
}

// RemoveNodesFromGroup 缩容（保留节点） - remove nodes from NodeGroup, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	// 移除的节点会重装节点的操作系统，并清理节点上的CCE组件
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

	taskType := cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid":   opt.Cluster.ClusterID,
		"tasktype":    taskType,
		"nodegroupid": group.NodeGroupID,
		"status":      cloudprovider.TaskStatusRunning,
	})
	taskList, err := cloudprovider.GetStorageModel().ListTask(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("UpdateDesiredNodes failed: %v", err)
		return nil, err
	}
	if len(taskList) != 0 {
		return nil, fmt.Errorf("gke task(%d) %s is still running", len(taskList), taskType)
	}

	needScaleOutNodes := desired - group.GetAutoScaling().GetDesiredSize()

	blog.Infof("cluster[%s] nodeGroup[%s] current nodes[%d] desired nodes[%d] needNodes[%s]",
		group.ClusterID, group.NodeGroupID, group.GetAutoScaling().GetDesiredSize(), desired, needScaleOutNodes)

	if desired <= group.GetAutoScaling().GetDesiredSize() {
		return nil, fmt.Errorf("NodeGroup %s current nodes %d larger than or equel to desired %d nodes",
			group.Name, group.GetAutoScaling().GetDesiredSize(), desired)
	}

	return &cloudprovider.ScalingResponse{
		ScalingUp: needScaleOutNodes,
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
func (ng *NodeGroup) GetExternalNodeScript(group *proto.NodeGroup, internal bool) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// CheckResourcePoolQuota check resource pool quota when revise group limit
func (ng *NodeGroup) CheckResourcePoolQuota(region, instanceType string, groupId string) error {
	return nil
}
