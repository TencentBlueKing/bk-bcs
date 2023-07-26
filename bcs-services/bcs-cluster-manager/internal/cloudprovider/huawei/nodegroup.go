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

package huawei

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

var groupMgr sync.Once

func init() {
	groupMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

// NodeGroup node group management for blueking resource pool solution
type NodeGroup struct {
}

// CreateNodeGroup create nodegroup by cloud provider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (*proto.Task,
	error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildCreateNodeGroupTask(group, opt)
	if err != nil {
		blog.Errorf("build CreateNodeGroup task for cluster %s with cloud provider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// DeleteNodeGroup delete nodegroup by cloud provider api, all nodes belong to NodeGroup
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

// UpdateNodeGroup update specified nodegroup configuration
func (ng *NodeGroup) UpdateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) error {
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return err
	}

	cceCli, err := api.NewCceClient(opt)
	if err != nil {
		blog.Errorf("UpdateNodeGroup[%s]: get cce client failed, %s", err.Error())
		return err
	}

	//获取节点池信息
	rsp, err := cceCli.GetClusterNodePool(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("GetClusterNodePool[%s]: get cluster nodePool failed, %s", err.Error())
		return err
	}

	_, err = cceCli.UpdateNodePool(api.GenerateModifyClusterNodePoolInput(group, cluster.SystemID, rsp))
	if err != nil {
		return err
	}

	return nil
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode,
	error) {
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return nil, err
	}

	cceCli, err := api.NewCceClient(opt)
	if err != nil {
		blog.Errorf("GetNodesInGroup[%s]: get cce client  failed, %s", err.Error())
		return nil, err
	}

	nodes, err := cceCli.ListClusterNodePoolNodes(cluster.SystemID, group.CloudNodeGroupID)
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
			node.Status = "RUNNING"
		case model.GetNodeStatusPhaseEnum().BUILD.Value(),
			model.GetNodeStatusPhaseEnum().BUILD.Value(),
			model.GetNodeStatusPhaseEnum().INSTALLING.Value(),
			model.GetNodeStatusPhaseEnum().INSTALLED.Value(),
			model.GetNodeStatusPhaseEnum().UPGRADING.Value():
			node.Status = "INITIALIZATION"
		case model.GetNodeStatusPhaseEnum().ERROR.Value():
			node.Status = "FAILED"
		}

		node.NodeGroupID = group.NodeGroupID
		node.ClusterID = group.ClusterID
		groupNodes = append(groupNodes, node)
	}

	return groupNodes, nil
}

// MoveNodesToGroup add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// RemoveNodesFromGroup remove nodes from MIG, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return err
	}

	if len(nodes) == 0 || opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return fmt.Errorf("invalid request")
	}

	client, err := api.NewCceClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create as client failed, err: %s", err.Error())
		return err
	}

	ids := make([]string, 0)
	for _, v := range nodes {
		ids = append(ids, v.NodeID)
	}

	return client.RemoveNodePoolNodes(cluster.SystemID, ids, group.LaunchTemplate.InitLoginPassword)
}

// CleanNodesInGroup clean specified nodes in NodeGroup,
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

// UpdateDesiredNodes update nodegroup desired node
func (ng *NodeGroup) UpdateDesiredNodes(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*cloudprovider.ScalingResponse, error) {
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
		return nil, fmt.Errorf("%d %s task(s) is still running", len(taskList), taskType)
	}

	return &cloudprovider.ScalingResponse{
		ScalingUp: desired,
	}, nil
}

// SwitchNodeGroupAutoScaling switch nodegroup autoscaling
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
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
// cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateAutoScalingOption update cluster autoscaling option, cloudprovider will update
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

// SwitchAutoScalingOptionStatus switch cluster autoscaling option status
func (ng *NodeGroup) SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchAutoScalingOptionStatus %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchAutoScalingOptionStatusTask(scalingOption, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchAutoScalingOptionStatus task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}
