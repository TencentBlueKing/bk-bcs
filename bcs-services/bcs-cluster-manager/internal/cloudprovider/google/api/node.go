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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"google.golang.org/api/compute/v1"
)

const (
	limit = 100
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeManager("google", &NodeManager{})
	})
}

// NodeManager define node manager
type NodeManager struct {
}

// GetNodeByIP get specified Node by innerIP address
func (n *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByIP list node by IP set
func (n *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCVMImageIDByImageName get imageID by imageName
func (n *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// GetCloudRegions get cloud regions
func (n *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	client, err := NewComputeServiceClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	regions, err := client.ListRegions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}
	return regions, nil
}

// GetZoneList get zoneList by region
func (n *NodeManager) GetZoneList(opt *cloudprovider.CommonOption) ([]*proto.ZoneInfo, error) {
	client, err := NewComputeServiceClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	zones, err := client.ListZones(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}
	return zones, nil
}

// ListNodeInstanceType list node type by zone and node family
func (n *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListOsImage get osimage list
func (n *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeByIP get specified Node by innerIP address
func (n *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListExternalNodesByIP list node by IP set
func (n *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByInstanceID list node by instance id
func (n *NodeManager) ListNodesByInstanceID(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	idChunks := utils.SplitStringsChunks(ids, limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByInstanceID idChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := n.transInstanceIDsToNodes(chunk, opt)
			if err != nil {
				blog.Errorf("ListNodesByInstanceID failed: %v", err)
				return nil, err
			}
			if len(nodes) == 0 {
				continue
			}
			nodeList = append(nodeList, nodes...)
		}
	}

	return nodeList, nil
}

// ListKeyPairs keyPairs list
func (n *NodeManager) ListKeyPairs(opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// transInstanceIDsToNodes trans IDList to Nodes
func (n *NodeManager) transInstanceIDsToNodes(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node,
	error) {
	client, err := NewComputeServiceClient(opt.Common)
	if err != nil {
		blog.Errorf("create ComputeServiceClient failed when GetNodeByIP, %s", err.Error())
		return nil, err
	}

	insList, err := client.ListZoneInstanceWithFilter(context.Background(), InstanceNameFilter(ids))
	if err != nil {
		blog.Errorf("ListZoneInstanceWithFilter failed, %s", err.Error())
		return nil, err
	}
	instances := insList.Items
	// check response data
	blog.Infof("ListZoneInstanceWithFilter len(%d) id response num %d", len(ids), len(instances))
	if len(instances) == 0 {
		// * no data response
		return nil, nil
	}
	if len(instances) != len(ids) {
		blog.Warnf("ListZoneInstanceWithFilter expect %d, but got %d", len(ids), len(instances))
	}

	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, inst := range instances {
		node := InstanceToNode(client, inst)
		// clean duplicated Node if user input multiple ip that
		// belong to one instance
		if _, ok := nodeMap[node.NodeID]; ok {
			continue
		}
		nodeMap[node.NodeID] = node
		node.InnerIP = inst.NetworkInterfaces[0].NetworkIP
		node.Region = opt.Common.Region
		// check node vpc and cluster vpc
		if !strings.Contains(node.VPC, opt.ClusterVPCID) {
			return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// InstanceToNode parse Instance information in gcloud to Node in clustermanager
// @param Instance: gcloud instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(cli *ComputeServiceClient, ins *compute.Instance) *proto.Node {
	zoneInfo, _ := GetGCEResourceInfo(ins.Zone)
	zone, _ := cli.GetZone(context.Background(), zoneInfo[len(zoneInfo)-1])
	node := &proto.Node{}
	if zoneInfo != nil {
		node.ZoneID = zone.Zone
		zoneID, _ := strconv.Atoi(zone.ZoneID)
		node.Zone = uint32(zoneID)
	}
	machineInfo, _ := GetGCEResourceInfo(ins.MachineType)
	node.NodeID = ins.Name
	node.InstanceType = machineInfo[len(machineInfo)-1]
	node.VPC = ins.NetworkInterfaces[0].Network
	return node
}
