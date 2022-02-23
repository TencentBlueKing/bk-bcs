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
	"fmt"
	"strconv"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		//init Node
		cloudprovider.InitNodeManager("qcloud", &NodeManager{})
	})
}

// GetCVMClient get cvm client from common option
func GetCVMClient(opt *cloudprovider.CommonOption) (*cvm.Client, error) {
	if opt == nil || len(opt.Key) == 0 || len(opt.Secret) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Key, opt.Secret)

	cpf := profile.NewClientProfile()

	return cvm.NewClient(credential, opt.Region, cpf)
}

// NodeManager CVM relative API management
type NodeManager struct {
}

// GetZoneList get zoneList
func (nm *NodeManager) GetZoneList(opt *cloudprovider.CommonOption) ([]*proto.ZoneInfo, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetZoneList failed: %v", err)
		return nil, err
	}

	req := cvm.NewDescribeZonesRequest()
	resp, err := client.DescribeZones(req)
	if err != nil {
		blog.Errorf("cvm client GetZoneList failed, %s", err.Error())
		return nil, err
	}

	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client GetZoneList but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client GetZoneList response num %d",
		response.RequestId, *response.TotalCount)

	if *response.TotalCount == 0 || len(response.ZoneSet) == 0 {
		//* no data response
		return nil, nil
	}

	zones := make([]*proto.ZoneInfo, 0)
	for i := range response.ZoneSet {
		zones = append(zones, &proto.ZoneInfo{
			ZoneID:    *response.ZoneSet[i].ZoneId,
			Zone:      *response.ZoneSet[i].Zone,
			ZoneName:  *response.ZoneSet[i].ZoneName,
			ZoneState: *response.ZoneSet[i].ZoneState,
		})
	}

	return zones, nil
}

// GetRegionsInfo get regionInfo
func (nm *NodeManager) GetRegionsInfo(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetRegionsInfo failed: %v", err)
		return nil, err
	}

	req := cvm.NewDescribeRegionsRequest()
	resp, err := client.DescribeRegions(req)
	if err != nil {
		blog.Errorf("cvm client DescribeRegions failed, %s", err.Error())
		return nil, err
	}

	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeRegions but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeRegions response num %d",
		response.RequestId, *response.TotalCount)

	if *response.TotalCount == 0 || len(response.RegionSet) == 0 {
		//* no data response
		return nil, nil
	}

	regions := make([]*proto.RegionInfo, 0)
	for i := range response.RegionSet {
		regions = append(regions, &proto.RegionInfo{
			Region:      *response.RegionSet[i].Region,
			RegionName:  *response.RegionSet[i].RegionName,
			RegionState: *response.RegionSet[i].RegionState,
		})
	}

	return regions, nil
}

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	client, err := GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when GetNodeByIP failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	var ips []*string
	ips = append(ips, common.StringPtr(ip))
	req.Filters = append(req.Filters, &cvm.Filter{
		Name:   common.StringPtr("private-ip-address"),
		Values: ips,
	})
	resp, err := client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance %s failed, %s", ip, err.Error())
		return nil, err
	}
	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance %s but lost response information", ip)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance %s response num %d",
		response.RequestId, ip, *response.TotalCount,
	)
	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		//* no data response
		return nil, cloudprovider.ErrCloudNoHost
	}
	zoneInfo, err := GetZoneInfoByRegion(client, opt.Common.Region)
	if err != nil {
		blog.Errorf("cvm client GetNodeByIP failed: %v", err)
	}

	node := InstanceToNode(response.InstanceSet[0], zoneInfo)
	node.InnerIP = ip
	node.Region = opt.Common.Region

	// check node vpc and cluster vpc
	if !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
		return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
	}

	return node, nil
}

// GetCVMImageIDByImageName xxx
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetCVMImageIDByImageName failed, %s", err.Error())
		return "", err
	}

	var (
		initOffset   uint64 = 0
		imageIDList         = make([]string, 0)
		imageListLen        = 100
	)

	for {
		if imageListLen != 100 {
			break
		}
		req := cvm.NewDescribeImagesRequest()
		req.Filters = []*cvm.Filter{
			&cvm.Filter{
				Name:   common.StringPtr("image-type"),
				Values: common.StringPtrs([]string{"PRIVATE_IMAGE"}),
			},
		}
		req.Offset = common.Uint64Ptr(initOffset)
		req.Limit = common.Uint64Ptr(uint64(100))

		resp, err := client.DescribeImages(req)
		if err != nil {
			blog.Errorf("cvm client DescribeImages %s failed, %s", imageName, err.Error())
			return "", err
		}
		//check response
		response := resp.Response
		if response == nil {
			blog.Errorf("cvm client DescribeImages %s but lost response information", imageName)
			return "", cloudprovider.ErrCloudLostResponse
		}

		for _, image := range response.ImageSet {
			if *image.ImageName == imageName {
				imageIDList = append(imageIDList, *image.ImageId)
			}
		}

		imageListLen = len(response.ImageSet)
		initOffset = initOffset + 100
	}

	blog.Infof("GetCVMImageIDByImageName successful %v", imageIDList)
	if len(imageIDList) == 0 {
		return "", fmt.Errorf("GetCVMImageIDByImageName[%s] failed: imageIDList empty", imageName)
	}

	return imageIDList[0], nil
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	ipChunks := utils.SplitStringsChunks(ips, limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByIP ipChunks %+v", ipChunks)
	for _, chunk := range ipChunks {
		if len(chunk) > 0 {
			nodes, err := nm.transIPsToNodes(chunk, opt)
			if err != nil {
				blog.Errorf("ListNodesByIP failed: %v", err)
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

// transIPsToNodes trans IPList to Nodes
func (nm *NodeManager) transIPsToNodes(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when GetNodeByIP failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var ipList []*string
	for _, ip := range ips {
		ipList = append(ipList, common.StringPtr(ip))
	}

	req.Filters = append(req.Filters, &cvm.Filter{
		Name:   common.StringPtr("private-ip-address"),
		Values: ipList,
	})
	resp, err := client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ips), err.Error())
		return nil, err
	}
	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ips))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		response.RequestId, len(ips), *response.TotalCount,
	)
	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		//* no data response
		return nil, nil
	}
	if len(response.InstanceSet) != len(ips) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d")
	}
	zoneInfo, err := GetZoneInfoByRegion(client, opt.Common.Region)
	if err != nil {
		blog.Errorf("cvm client ListNodesByIP failed: %v", err)
	}

	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, ip := range ips {
		for _, inst := range response.InstanceSet {
			// ip in instance.PrivateIp list
			found := false
			for _, instIP := range inst.PrivateIpAddresses {
				if ip == *instIP {
					found = true
				}
			}
			if !found {
				continue
			}
			node := InstanceToNode(inst, zoneInfo)
			// clean duplicated Node if user input multiple ip that
			// belong to one cvm instance
			if _, ok := nodeMap[node.NodeID]; ok {
				continue
			}
			nodeMap[node.NodeID] = node
			node.InnerIP = ip
			node.Region = opt.Common.Region

			// check node vpc and cluster vpc
			if !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
				return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
			}

			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// InstanceToNode parse Instance information in qcloud to Node in clustermanager
// @param Instance: qcloud instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(inst *cvm.Instance, zoneInfo map[string]uint32) *proto.Node {
	var zoneID uint32
	if zoneInfo != nil {
		zoneID = zoneInfo[*inst.Placement.Zone]
	}
	node := &proto.Node{
		NodeID:       *inst.InstanceId,
		InstanceType: *inst.InstanceType,
		CPU:          uint32(*inst.CPU),
		Mem:          uint32(*inst.Memory),
		GPU:          0,
		VPC:          *inst.VirtualPrivateCloud.VpcId,
		ZoneID:       *inst.Placement.Zone,
		Zone:         zoneID,
	}
	return node
}

// GetZoneInfoByRegion region: ap-nanjing/ap-shenzhen
func GetZoneInfoByRegion(client *cvm.Client, region string) (map[string]uint32, error) {
	if client == nil {
		return nil, fmt.Errorf("getZoneInfoByRegion client is nil")
	}

	req := cvm.NewDescribeZonesRequest()
	resp, err := client.DescribeZones(req)
	if err != nil {
		return nil, fmt.Errorf("getZoneInfoByRegion GetZoneInfo failed: %v", err)
	}

	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeZones lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeZones response num %d",
		response.RequestId, *response.TotalCount)

	if *response.TotalCount == 0 || len(response.ZoneSet) == 0 {
		//* no data response
		return nil, nil
	}

	zoneIDMap := make(map[string]uint32)
	for i := range response.ZoneSet {
		if _, ok := zoneIDMap[*response.ZoneSet[i].Zone]; !ok {
			zoneID, _ := strconv.ParseUint(*response.ZoneSet[i].ZoneId, 10, 32)
			zoneIDMap[*response.ZoneSet[i].Zone] = uint32(zoneID)
		}
	}

	return zoneIDMap, nil
}
