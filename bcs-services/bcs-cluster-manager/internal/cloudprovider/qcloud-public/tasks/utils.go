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

package tasks

import (
	"context"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
)

// updateClusterSystemID set cluster systemID
// nolint
func updateClusterSystemID(clusterID string, systemID string) error {
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	cluster.SystemID = systemID
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// updateNodeStatusByNodeID set node status
func updateNodeStatusByNodeID(idList []string, status string) error { // nolint
	if len(idList) == 0 {
		return nil
	}

	for _, id := range idList {
		node, err := cloudprovider.GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			continue
		}
		node.Status = status
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}

	return nil
}

func transInstanceIPToNodes(ipList []string, opt *cloudprovider.ListNodesOption) ([]*cmproto.Node, error) {
	nodes, err := business.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
		Common:       opt.Common,
		ClusterVPCID: opt.ClusterVPCID,
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// updateNodeGroupCloudNodeGroupID set nodegroup cloudNodeGroupID
func updateNodeGroupCloudNodeGroupID(nodeGroupID string, newGroup *cmproto.NodeGroup) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = newGroup.AutoScaling.VpcID
	}
	if group.LaunchTemplate != nil {
		group.LaunchTemplate.InstanceChargeType = newGroup.LaunchTemplate.InstanceChargeType
	}
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

// updateNodeGroupDesiredSize set nodegroup desired size
func updateNodeGroupDesiredSize(nodeGroupID string, desiredSize uint32) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	if group.AutoScaling == nil {
		group.AutoScaling = &cmproto.AutoScalingGroup{}
	}
	group.AutoScaling.DesiredSize = desiredSize
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}
