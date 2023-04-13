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

package tasks

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

func transImageNameToImageID(cmOption *cloudprovider.CommonOption, imageName string) (string, error) {
	nodeManager := &api.NodeManager{}

	imageID, err := nodeManager.GetCVMImageIDByImageName(imageName, cmOption)
	if err == nil {
		return imageID, nil
	}

	return imageName, nil
}

func transIPsToInstanceID(cmOption *cloudprovider.ListNodesOption, ips []string) ([]string, error) {
	nodeManager := &api.NodeManager{}
	nodes, err := nodeManager.ListNodesByIP(ips, cmOption)
	if err != nil {
		return nil, err
	}

	instanceIDs := make([]string, 0)
	for i := range nodes {
		instanceIDs = append(instanceIDs, nodes[i].NodeID)
	}

	return instanceIDs, nil
}

// updateClusterSystemID set cluster systemID
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

// updateClusterStatus set cluster status
func updateClusterStatus(clusterID string, status string) error {
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		return err
	}

	cluster.Status = status
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// updateNodeStatusByIP set node status
func updateNodeStatusByIP(ipList []string, status string) error {
	if len(ipList) == 0 {
		return nil
	}

	for _, ip := range ipList {
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(context.Background(), ip)
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

// updateNodeStatusByNodeID set node status
func updateNodeStatusByNodeID(idList []string, status string) error {
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
	nodeMgr := api.NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
		Common:       opt.Common,
		ClusterVPCID: opt.ClusterVPCID,
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func importClusterNodesToCM(ctx context.Context, ipList []string, opt *cloudprovider.ListNodesOption) error {
	nodeMgr := api.NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
		Common:       opt.Common,
		ClusterVPCID: opt.ClusterVPCID,
	})
	if err != nil {
		return err
	}

	for _, n := range nodes {
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(ctx, n.InnerIP)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("importClusterNodes GetNodeByIP[%s] failed: %v", n.InnerIP, err)
			// no import node when found err
			continue
		}

		if node == nil {
			n.ClusterID = opt.ClusterID
			n.Status = common.StatusRunning
			err = cloudprovider.GetStorageModel().CreateNode(ctx, n)
			if err != nil {
				blog.Errorf("importClusterNodes CreateNode[%s] failed: %v", n.InnerIP, err)
			}
			continue
		}
		err = cloudprovider.GetStorageModel().UpdateNode(ctx, n)
		if err != nil {
			blog.Errorf("importClusterNodes UpdateNode[%s] failed: %v", n.InnerIP, err)
		}
	}

	return nil
}

// releaseClusterCIDR release cluster CIDR
func releaseClusterCIDR(cls *cmproto.Cluster) error {
	if len(cls.NetworkSettings.ClusterIPv4CIDR) > 0 {
		cidr, err := cloudprovider.GetStorageModel().GetTkeCidr(context.Background(), cls.VpcID,
			cls.NetworkSettings.ClusterIPv4CIDR)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			return err
		}

		if cidr == nil {
			return nil
		}

		if cidr.Cluster == cls.ClusterID && cidr.Status == common.TkeCidrStatusUsed {
			// update cidr and save to DB
			updateCidr := cidr
			updateCidr.Status = common.TkeCidrStatusAvailable
			updateCidr.Cluster = ""
			updateCidr.UpdateTime = time.Now().String()
			err = cloudprovider.GetStorageModel().UpdateTkeCidr(context.Background(), updateCidr)
			if err != nil {
				return err
			}
		}
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
