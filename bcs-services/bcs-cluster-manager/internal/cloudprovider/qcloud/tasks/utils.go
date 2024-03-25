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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

func transImageNameToImageID(cmOption *cloudprovider.CommonOption, imageName string) (string, error) { // nolint
	imageID, err := business.GetCVMImageIDByImageName(imageName, cmOption)
	if err == nil {
		return imageID, nil
	}

	return imageName, nil
}

func transIPsToInstances(cmOption *cloudprovider.ListNodesOption, ips []string) (map[string]*cmproto.Node, error) {
	nodes, err := business.ListNodesByIP(ips, cmOption)
	if err != nil {
		return nil, err
	}

	instances := make(map[string]*cmproto.Node, 0)
	for i := range nodes {
		instances[nodes[i].InnerIP] = nodes[i]
	}

	return instances, nil
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

// updateNodeStatusByNodeID set node status
func updateFailedNodeStatusByNodeID(ctx context.Context, insInfos map[string]business.InstanceInfo, status string) error { // nolint
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	if len(insInfos) == 0 {
		blog.Infof("updateFailedNodeStatusByNodeID[%s] failed: insInfos empty", taskId)
		return nil
	}

	for id, data := range insInfos {
		node, err := cloudprovider.GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			blog.Errorf("updateFailedNodeStatusByNodeID[%s] GetNode[%s] failed: %v", taskId, id, err)
			continue
		}
		node.Status = status
		if data.FailedReason != "" {
			node.FailedReason = data.FailedReason
		}
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("updateFailedNodeStatusByNodeID[%s] UpdateNode[%s] failed: %v", taskId, id, err)
			continue
		}
	}

	return nil
}

// updateNodeStatusByNodeID set node status
func updateNodeStatusByNodeID(idList []string, status, reason string) error { // nolint
	if len(idList) == 0 {
		return nil
	}

	for _, id := range idList {
		node, err := cloudprovider.GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			continue
		}
		node.Status = status
		if reason != "" {
			node.FailedReason = reason
		}
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}

	return nil
}

// updateNodeIPByNodeID set node innerIP
func updateNodeIPByNodeID(ctx context.Context, clusterId string, n business.InstanceInfo) error { // nolint
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	if n.NodeId == "" || n.NodeIp == "" {
		blog.Errorf("updateNodeIPByNodeID[%s] nodeId[%s] nodeIp[%s] empty", taskId, n.NodeId, n.NodeIp)
		return fmt.Errorf("updateNodeIPByNodeID data[%s:%s] empty", n.NodeId, n.NodeIp)
	}

	blog.Infof("updateNodeIPByNodeID[%s] cluster[%s] nodeId[%s] nodeIp[%s] vpcId[%s]",
		taskId, clusterId, n.NodeId, n.NodeIp, n.VpcId)

	node, err := cloudprovider.GetStorageModel().GetClusterNode(context.Background(), clusterId, n.NodeId)
	if err != nil {
		blog.Errorf("updateNodeIPByNodeID[%s] failed: %v", taskId, err)
		return err
	}
	node.InnerIP = n.NodeIp
	node.VPC = n.VpcId
	err = cloudprovider.GetStorageModel().UpdateClusterNodeByNodeID(context.Background(), node)
	if err != nil {
		blog.Errorf("updateNodeIPByNodeID[%s] failed: %v", taskId, err)
		return err
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

func importClusterNodesToCM(ctx context.Context, ipList []string, opt *cloudprovider.ListNodesOption) error {
	nodes, err := business.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
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

		n.ClusterID = opt.ClusterID
		n.Status = common.StatusRunning
		if node == nil {
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
	if len(cls.GetNetworkSettings().GetClusterIPv4CIDR()) > 0 {
		cidr, err := cloudprovider.GetStorageModel().GetTkeCidr(context.Background(),
			cls.VpcID, cls.NetworkSettings.ClusterIPv4CIDR)
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
			updateCidr.UpdateTime = time.Now().Format(time.RFC3339)
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

// GetExternalNgScriptType get external nodeGroup script type (true: inter; false extra)
func GetExternalNgScriptType(ng *cmproto.NodeGroup) bool {
	if ng.GetExtraInfo() == nil {
		return false
	}

	_, ok := ng.GetExtraInfo()[common.ScriptInterType.String()]
	if ok {
		return ok
	}

	return false
}
