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
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	computev1 "google.golang.org/api/compute/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
func (n *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	client, err := NewComputeServiceClient(&opt.CommonOption)
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
	blog.Infof("ListNodeInstanceType zone: %s, nodeFamily: %s, cpu: %d, memory: %d",
		info.Zone, info.NodeFamily, info.Cpu, info.Memory)

	client, err := NewComputeServiceClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}

	var filter string
	if info.NodeFamily != "" {
		filter = fmt.Sprintf("name eq %s-*", info.NodeFamily)
	}
	list, err := client.ListMachineTypes(context.Background(), opt.Region, filter)
	if err != nil {
		return nil, err
	}

	result := convertToInstanceType(list.Items, info.NodeFamily)

	var instanceTypes = make([]*proto.InstanceType, 0)
	for _, item := range result {
		if info.Cpu > 0 {
			if item.Cpu != info.Cpu {
				continue
			}
		}
		if info.Memory > 0 {
			if item.Memory != info.Memory {
				continue
			}
		}
		instanceTypes = append(instanceTypes, item)
	}

	return instanceTypes, nil
}

func convertToInstanceType(mt []*computev1.MachineType, family string) []*proto.InstanceType {
	result := make([]*proto.InstanceType, 0)

	for _, m := range mt {
		nameList := strings.Split(m.Name, "-")
		memGb := math.Ceil(float64(m.MemoryMb / 1024)) // nolint

		insType := &proto.InstanceType{}
		insType.Status = common.InstanceSell
		if m.Deprecated != nil {
			insType.Status = common.InstanceSoldOut
		}
		insType.NodeType = m.Name
		insType.TypeName = nameList[1]
		insType.NodeFamily = family

		insType.Cpu = uint32(m.GuestCpus)
		insType.Memory = uint32(memGb)
		insType.Zones = []string{m.Zone}

		result = append(result, insType)
	}

	return result
}

// ListOsImage get osimage list
func (n *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	// 先返回固定的UBUNTU_CONTAINERD
	images := make([]*proto.OsImage, 0)
	images = append(images, &proto.OsImage{
		ImageID:  "UBUNTU_CONTAINERD",
		Status:   "NORMAL",
		Provider: common.PublicImageProvider,
	})
	return images, nil
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
// checkIP - It may take time when an instance get an ip after scaling up a nodegroup in GKE
func (n *NodeManager) ListNodesByInstanceID(ids []string, opt *cloudprovider.ListNodesOption) (
	[]*proto.Node, error) {
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
func (n *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetResourceGroups resource groups list
func (n *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// transInstanceIDsToNodes trans IDList to Nodes
func (n *NodeManager) transInstanceIDsToNodes(ids []string, opt *cloudprovider.ListNodesOption) (
	[]*proto.Node, error) {
	client, err := NewComputeServiceClient(opt.Common)
	if err != nil {
		blog.Errorf("transInstanceIDsToNodes create ComputeServiceClient failed, %s", err.Error())
		return nil, err
	}
	insList, err := client.ListZoneInstanceWithFilter(context.Background(), InstanceNameFilter(ids))
	if err != nil {
		blog.Errorf("transInstanceIDsToNodes ListZoneInstanceWithFilter failed, %s", err.Error())
		return nil, err
	}

	blog.Infof("transInstanceIDsToNodes desired %d, response %d", len(ids), len(insList.Items))

	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, in := range insList.Items {
		blog.Infof("transInstanceIDsToNodes instance[%s], ip[%s]", in.Name, in.NetworkInterfaces[0].NetworkIP)
		node := InstanceToNode(client, in)
		// clean duplicated Node if user input multiple ip that
		// belong to one instance
		if _, ok := nodeMap[node.NodeID]; ok {
			continue
		}
		nodeMap[node.NodeID] = node
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// InstanceToNode parse Instance information in gcloud to Node in clustermanager
// @param Instance: gcloud instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(cli *ComputeServiceClient, ins *computev1.Instance) *proto.Node {
	zoneInfo, _ := GetGCEResourceInfo(ins.Zone)
	zone, _ := cli.GetZone(context.Background(), zoneInfo[len(zoneInfo)-1])

	node := &proto.Node{}
	node.NodeID = strconv.Itoa(int(ins.Id))
	node.NodeName = ins.Name
	node.InnerIP = ins.NetworkInterfaces[0].NetworkIP

	if zoneInfo != nil {
		node.ZoneID = zone.Zone
		zoneID, _ := strconv.Atoi(zone.ZoneID)
		node.Zone = uint32(zoneID)
		node.ZoneName = zone.ZoneName
	}

	machineInfo, _ := GetGCEResourceInfo(ins.MachineType)
	node.InstanceType = machineInfo[len(machineInfo)-1]

	networkInfo := strings.Split(ins.NetworkInterfaces[0].Subnetwork, "/")
	node.VPC = networkInfo[len(networkInfo)-1]
	node.Status = ins.Status

	return node
}

// ListRuntimeInfo get runtime info list
func (n *NodeManager) ListRuntimeInfo(opt *cloudprovider.ListRuntimeInfoOption) (map[string][]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetServiceRoles service roles list
func (n *NodeManager) GetServiceRoles(opt *cloudprovider.CommonOption, roleType string) (
	[]*proto.ServiceRoleInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
