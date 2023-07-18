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
	"math"
	"strconv"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
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
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)

	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.MachineDomain
	}

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

	// DescribeZones
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

// GetCloudRegions get regionInfo
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetRegionsInfo failed: %v", err)
		return nil, err
	}

	// DescribeRegions
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

// GetNodeInstanceByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeInstanceByIP(ip string, opt *cloudprovider.CommonOption) (*cvm.Instance, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetNodeInstanceByIP failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	var ips []*string
	ips = append(ips, common.StringPtr(ip))
	req.Filters = append(req.Filters, &cvm.Filter{
		Name:   common.StringPtr("private-ip-address"),
		Values: ips,
	})
	// DescribeInstances
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

	return response.InstanceSet[0], nil
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
	// DescribeInstances
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
	if opt.ClusterVPCID != "" && !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
		return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
	}

	return node, nil
}

// GetImageInfoByImageID xxx
func (nm *NodeManager) GetImageInfoByImageID(imageID string, opt *cloudprovider.CommonOption) (*cvm.Image, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetImageInfoByImageID failed, %s", err.Error())
		return nil, err
	}

	req := cvm.NewDescribeImagesRequest()
	req.ImageIds = append(req.ImageIds, common.StringPtr(imageID))

	// DescribeImages
	resp, err := client.DescribeImages(req)
	if err != nil {
		blog.Errorf("cvm client DescribeImages %s failed, %s", imageID, err.Error())
		return nil, err
	}
	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeImages %s but lost response information", imageID)
		return nil, cloudprovider.ErrCloudLostResponse
	}

	if len(response.ImageSet) <= 0 {
		blog.Errorf("cvm client DescribeImages %s failed", imageID)
		return nil, fmt.Errorf("not found image[%s]", imageID)
	}

	return response.ImageSet[0], nil
}

// GetCVMImageIDByImageName xxx
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetCVMImageIDByImageName failed, %s", err.Error())
		return "", err
	}

	var (
		initOffset   uint64
		imageIDList  = make([]string, 0)
		imageListLen = 100
	)

	for {
		if imageListLen != 100 {
			break
		}
		req := cvm.NewDescribeImagesRequest()
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

// ListNodeInstancesByIP get cloud node instance by ips
func (nm *NodeManager) ListNodeInstancesByIP(ips []string, opt *cloudprovider.CommonOption) ([]*cvm.Instance, error) {
	ipChunks := utils.SplitStringsChunks(ips, maxFilterValues)
	blog.Infof("ListNodeInstancesByIP ipChunks %+v", ipChunks)

	var (
		instanceList = make([]*cvm.Instance, 0)
		lock         = sync.Mutex{}
	)
	barrier := utils.NewRoutinePool(20)
	defer barrier.Close()

	for _, chunk := range ipChunks {
		if len(chunk) > 0 {
			barrier.Add(1)
			go func(ips []string) {
				defer barrier.Done()
				nodes, err := nm.transIPsToInstances(ips, opt)
				if err != nil {
					blog.Errorf("ListNodeInstancesByIP failed: %v", err)
					return
				}
				if len(nodes) == 0 {
					return
				}
				lock.Lock()
				instanceList = append(instanceList, nodes...)
				lock.Unlock()
			}(chunk)
		}
	}

	barrier.Wait()
	return instanceList, nil
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	ipChunks := utils.SplitStringsChunks(ips, maxFilterValues)
	blog.Infof("ListNodesByIP ipChunks %+v", ipChunks)

	var (
		nodeList = make([]*proto.Node, 0)
		lock     = sync.Mutex{}
	)
	barrier := utils.NewRoutinePool(20)
	defer barrier.Close()

	for _, chunk := range ipChunks {
		if len(chunk) > 0 {
			barrier.Add(1)
			go func(ips []string) {
				defer barrier.Done()
				nodes, err := nm.transIPsToNodes(ips, opt)
				if err != nil {
					blog.Errorf("ListNodesByIP failed: %v", err)
					return
				}
				if len(nodes) == 0 {
					return
				}
				lock.Lock()
				nodeList = append(nodeList, nodes...)
				lock.Unlock()
			}(chunk)
		}
	}

	barrier.Wait()
	return nodeList, nil
}

// ListNodeInstancesByInstanceID list cloud node instance by instanceID
func (nm *NodeManager) ListNodeInstancesByInstanceID(ids []string, opt *cloudprovider.CommonOption) ([]*cvm.Instance, error) {
	idChunks := utils.SplitStringsChunks(ids, limit)
	instanceList := make([]*cvm.Instance, 0)

	blog.Infof("ListNodeInstancesByInstanceID ipChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := nm.transInstanceIDsToInstances(chunk, opt)
			if err != nil {
				blog.Errorf("ListNodeInstancesByInstanceID failed: %v", err)
				return nil, err
			}
			if len(nodes) == 0 {
				continue
			}
			instanceList = append(instanceList, nodes...)
		}
	}
	return instanceList, nil
}

// ListNodesByInstanceID list node by instanceIDs
func (nm *NodeManager) ListNodesByInstanceID(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	idChunks := utils.SplitStringsChunks(ids, limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByInstanceID ipChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := nm.transInstanceIDsToNodes(chunk, opt)
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

// transInstanceIDsToInstances trans IDList to cloud node instances
func (nm *NodeManager) transInstanceIDsToInstances(ids []string, opt *cloudprovider.CommonOption) ([]*cvm.Instance, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when transInstanceIDsToInstances failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var idList []*string
	for _, id := range ids {
		idList = append(idList, common.StringPtr(id))
	}
	// instanceIDs max 100
	req.InstanceIds = append(req.InstanceIds, idList...)

	// DescribeInstances
	resp, err := client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ids), err.Error())
		return nil, err
	}
	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ids))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		*response.RequestId, len(ids), *response.TotalCount)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		//* no data response
		return nil, nil
	}
	if len(response.InstanceSet) != len(ids) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d", *response.RequestId,
			len(ids), len(response.InstanceSet))
	}

	return response.InstanceSet, nil
}

// transInstanceIDsToNodes trans IDList to Nodes
func (nm *NodeManager) transInstanceIDsToNodes(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when GetNodeByIP failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var idList []*string
	for _, id := range ids {
		idList = append(idList, common.StringPtr(id))
	}
	// instanceIDs max 100
	req.InstanceIds = append(req.InstanceIds, idList...)

	// DescribeInstances
	resp, err := client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ids), err.Error())
		return nil, err
	}
	//check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ids))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		response.RequestId, len(ids), *response.TotalCount,
	)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		//* no data response
		return nil, nil
	}
	if len(response.InstanceSet) != len(ids) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d")
	}
	zoneInfo, err := GetZoneInfoByRegion(client, opt.Common.Region)
	if err != nil {
		blog.Errorf("cvm client ListNodesByIP failed: %v", err)
	}

	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, inst := range response.InstanceSet {
		node := InstanceToNode(inst, zoneInfo)
		// clean duplicated Node if user input multiple ip that
		// belong to one cvm instance
		if _, ok := nodeMap[node.NodeID]; ok {
			continue
		}

		nodeMap[node.NodeID] = node
		// default get first privateIP
		node.InnerIP = *inst.PrivateIpAddresses[0]
		node.Region = opt.Common.Region

		// check node vpc and cluster vpc
		if opt.ClusterVPCID != "" && !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
			return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// transIPsToInstances trans IPList to cloud Instance, filter max 5 values
func (nm *NodeManager) transIPsToInstances(ips []string, opt *cloudprovider.CommonOption) ([]*cvm.Instance, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when transIPsToInstances failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var ipList []*string
	for _, ip := range ips {
		ipList = append(ipList, common.StringPtr(ip))
	}

	// filters values max 5
	req.Filters = append(req.Filters, &cvm.Filter{
		Name:   common.StringPtr("private-ip-address"),
		Values: ipList,
	})

	// DescribeInstances
	resp, err := client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ips), err.Error())
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ips))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		*response.RequestId, len(ips), *response.TotalCount)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		//* no data response
		return nil, nil
	}
	if len(response.InstanceSet) != len(ips) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d", *response.RequestId,
			len(ips), len(response.InstanceSet))
	}

	return response.InstanceSet, nil
}

// transIPsToNodes trans IPList to Nodes, filter max 5 values
func (nm *NodeManager) transIPsToNodes(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when transIPsToNodes failed, %s", err.Error())
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var ipList []*string
	for _, ip := range ips {
		ipList = append(ipList, common.StringPtr(ip))
	}

	// filters values max 5
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
			if opt.ClusterVPCID != "" && !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
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
		InnerIPv6:    utils.SlicePtrToString(inst.IPv6Addresses),
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

// ListNodeInstanceType list node type by zone and node family
func (nm *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("ListNodeInstanceType zone: %s, nodeFamily: %s, cpu: %d, memory: %d",
		info.Zone, info.NodeFamily, info.Cpu, info.Memory)

	if options.GetEditionInfo().IsInnerEdition() {
		return nm.getInnerInstanceTypes(info)
	}

	return nm.getCloudInstanceType(info, opt)
}

// getCloudInstanceType get cloud instance type and filter instanceType by cpu&mem size
func (nm *NodeManager) getCloudInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("getCloudInstanceType %+v", info)
	list, err := nm.DescribeZoneInstanceConfigInfos(info.Zone, info.NodeFamily, "", opt)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.InstanceType, 0)
	for _, item := range list {
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
		result = append(result, item)
	}
	return result, nil
}

// getInnerInstanceTypes get inner instance types info
func (nm *NodeManager) getInnerInstanceTypes(info cloudprovider.InstanceInfo) (
	[]*proto.InstanceType, error) {
	blog.Infof("getInnerInstanceTypes %+v", info)

	targetTypes, err := tresource.GetResourceManagerClient().GetInstanceTypes(context.Background(),
		info.Region, resource.InstanceSpec{
			ProjectID: info.ProjectID,
			BizID:     info.BizID,
			Version:   info.Version,
			Cpu:       info.Cpu,
			Mem:       info.Memory,
			Provider:  info.Provider,
		})
	if err != nil {
		blog.Errorf("resourceManager ListNodeInstanceType failed: %v", err)
		return nil, err
	}
	blog.Infof("getInnerInstanceTypes successful[%+v]", targetTypes)

	var instanceTypes = make([]*proto.InstanceType, 0)
	for _, t := range targetTypes {
		instanceTypes = append(instanceTypes, &proto.InstanceType{
			NodeType:       t.NodeType,
			TypeName:       t.TypeName,
			NodeFamily:     t.NodeFamily,
			Cpu:            t.Cpu,
			Memory:         t.Memory,
			Gpu:            t.Gpu,
			Status:         t.Status, // SOLD_OUT
			UnitPrice:      0,
			Zones:          t.Zones,
			Provider:       t.Provider,
			ResourcePoolID: t.ResourcePoolID,
			SystemDisk: func() *proto.DataDisk {
				if t.SystemDisk == nil {
					return nil
				}

				return &proto.DataDisk{
					DiskType: t.SystemDisk.DiskType,
					DiskSize: t.SystemDisk.DiskSize,
				}
			}(),
			DataDisks: func() []*proto.DataDisk {
				disks := make([]*proto.DataDisk, 0)
				for i := range t.DataDisks {
					disks = append(disks, &proto.DataDisk{
						DiskType: t.DataDisks[i].DiskType,
						DiskSize: t.DataDisks[i].DiskSize,
					})
				}
				return disks
			}(),
		})
	}

	blog.Infof("getInnerInstanceTypes successful[%+v]", instanceTypes)
	return instanceTypes, nil
}

// DescribeInstanceTypeConfigs describe instance type configs
// https://cloud.tencent.com/document/api/213/17378
func (nm *NodeManager) DescribeInstanceTypeConfigs(filters []*Filter, opt *cloudprovider.CommonOption) (
	[]*InstanceTypeConfig, error) {
	blog.Infof("DescribeInstanceTypeConfigs input: %s", utils.ToJSONString(filters))
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeInstanceTypeConfigs failed: %v", err)
		return nil, err
	}
	req := cvm.NewDescribeInstanceTypeConfigsRequest()
	for _, v := range filters {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}
	resp, err := client.DescribeInstanceTypeConfigs(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstanceTypeConfigs failed: %v", err)
		return nil, err
	}

	if resp == nil || resp.Response == nil {
		blog.Errorf("cvm client DescribeInstanceTypeConfigs lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	result := make([]*InstanceTypeConfig, 0)
	for _, v := range resp.Response.InstanceTypeConfigSet {
		result = append(result, &InstanceTypeConfig{
			Zone:           v.Zone,
			InstanceType:   v.InstanceType,
			InstanceFamily: v.InstanceFamily,
			CPU:            v.CPU,
			GPU:            v.GPU,
			Memory:         v.Memory,
			FPGA:           v.FPGA,
		})
	}
	blog.Infof("DescribeInstanceTypeConfigs success, result: %s", utils.ToJSONString(result))
	return result, nil
}

// DescribeZoneInstanceConfigInfos describe zone instance config infos
// https://cloud.tencent.com/document/api/213/17378
func (nm *NodeManager) DescribeZoneInstanceConfigInfos(zone, instanceFamily, instanceType string, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("DescribeZoneInstanceConfigInfos input: zone/%s, instanceFamily/%s, instanceType/%s", zone, instanceFamily, instanceType)
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeZoneInstanceConfigInfos failed: %v", err)
		return nil, err
	}
	req := cvm.NewDescribeZoneInstanceConfigInfosRequest()
	req.Filters = make([]*cvm.Filter, 0)
	// 按量计费
	req.Filters = append(req.Filters, &cvm.Filter{
		Name: common.StringPtr("instance-charge-type"), Values: common.StringPtrs([]string{"POSTPAID_BY_HOUR"})})
	if len(zone) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("zone"), Values: common.StringPtrs([]string{zone})})
	}
	if len(instanceFamily) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("instance-family"), Values: common.StringPtrs([]string{instanceFamily})})
	}
	if len(instanceType) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("instance-type"), Values: common.StringPtrs([]string{instanceType})})
	}
	resp, err := client.DescribeZoneInstanceConfigInfos(req)
	if err != nil {
		blog.Errorf("cvm client DescribeZoneInstanceConfigInfos failed: %v", err)
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("cvm client DescribeZoneInstanceConfigInfos lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	result := make([]*proto.InstanceType, 0)
	instanceMap := make(map[string][]string) // instanceType: []zone
	for _, v := range resp.Response.InstanceTypeQuotaSet {
		if _, ok := instanceMap[*v.InstanceType]; ok {
			instanceMap[*v.InstanceType] = append(instanceMap[*v.InstanceType], *v.Zone)
			continue
		}
		instanceMap[*v.InstanceType] = append(instanceMap[*v.InstanceType], *v.Zone)
		t := &proto.InstanceType{}
		if v.InstanceType != nil {
			t.NodeType = *v.InstanceType
		}
		if v.TypeName != nil {
			t.TypeName = *v.TypeName
		}
		if v.InstanceFamily != nil {
			t.NodeFamily = *v.InstanceFamily
		}
		if v.Cpu != nil {
			t.Cpu = uint32(*v.Cpu)
		}
		if v.Memory != nil {
			t.Memory = uint32(*v.Memory)
		}
		if v.Gpu != nil {
			t.Gpu = uint32(*v.Gpu)
		}
		if v.Price != nil && v.Price.UnitPrice != nil {
			t.UnitPrice = float32(*v.Price.UnitPrice)
		}
		if v.Status != nil {
			t.Status = *v.Status
		}
		result = append(result, t)
	}
	for i := range result {
		result[i].Zones = instanceMap[result[i].NodeType]
	}
	blog.Infof("DescribeZoneInstanceConfigInfos success, result: %s", utils.ToJSONString(result))
	return result, nil
}

// DescribeInstances describe instances
// https://cloud.tencent.com/document/api/213/15728
func (nm *NodeManager) DescribeInstances(ins []string, filters []*Filter, opt *cloudprovider.CommonOption) (
	[]*proto.Node, error) {
	blog.Infof("DescribeInstances input: %s, %s", utils.ToJSONString(ins),
		utils.ToJSONString(filters))
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeInstances failed: %v", err)
		return nil, err
	}
	req := cvm.NewDescribeInstancesRequest()
	req.InstanceIds = common.StringPtrs(ins)
	req.Limit = common.Int64Ptr(limit)
	req.Filters = make([]*cvm.Filter, 0)
	for _, v := range filters {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}
	got, total := 0, 0
	first := true
	nodes := make([]*proto.Node, 0)
	zoneInfo, err := GetZoneInfoByRegion(client, opt.Region)
	if err != nil {
		blog.Errorf("cvm client GetZoneInfoByRegion failed: %v", err)
		return nil, err
	}
	for got < total || first {
		first = false
		req.Offset = common.Int64Ptr(int64(got))
		resp, err := client.DescribeInstances(req)
		if err != nil {
			blog.Errorf("DescribeInstances failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeInstances resp is nil")
			return nil, fmt.Errorf("DescribeInstances resp is nil")
		}
		blog.Infof("DescribeInstances success, requestID: %s", *resp.Response.RequestId)
		for _, v := range resp.Response.InstanceSet {
			node := &proto.Node{NodeID: *v.InstanceId}
			if v.InstanceType != nil {
				node.InstanceType = *v.InstanceType
			}
			if v.CPU != nil {
				node.CPU = uint32(*v.CPU)
			}
			if v.Memory != nil {
				node.Mem = uint32(*v.Memory)
			}
			if v.InstanceState != nil {
				node.Status = *v.InstanceState
			}
			if len(v.PrivateIpAddresses) > 0 {
				node.InnerIP = *v.PrivateIpAddresses[0]
			}
			if v.GPUInfo != nil && v.GPUInfo.GPUCount != nil {
				node.GPU = uint32(*v.GPUInfo.GPUCount)
			}
			if v.Placement != nil && v.Placement.Zone != nil {
				node.ZoneID = *v.Placement.Zone
				node.Zone = zoneInfo[*v.Placement.Zone]
			}
			if v.VirtualPrivateCloud != nil && v.VirtualPrivateCloud.VpcId != nil {
				node.VPC = *v.VirtualPrivateCloud.VpcId
			}
			if v.LoginSettings != nil && v.LoginSettings.Password != nil {
				node.Passwd = *v.LoginSettings.Password
			}
			nodes = append(nodes, node)
		}
		got += len(resp.Response.InstanceSet)
		total = int(*resp.Response.TotalCount)
	}
	return nodes, nil
}

// DescribeImages describe images: PRIVATE_IMAGE: 私有镜像; PUBLIC_IMAGE: 公共镜像 (腾讯云官方镜像)
// https://cloud.tencent.com/document/api/213/15715
func (nm *NodeManager) DescribeImages(imageType string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	blog.Infof("DescribeImages input: %s", imageType)
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeImages failed: %v", err)
		return nil, err
	}
	req := cvm.NewDescribeImagesRequest()
	if imageType != "" {
		req.Filters = []*cvm.Filter{
			{
				Name:   common.StringPtr("image-type"),
				Values: common.StringPtrs([]string{imageType}),
			},
		}
	}
	images := make([]*proto.OsImage, 0)
	got, total := 0, 0
	first := true
	for got < total || first {
		first = false
		req.Offset = common.Uint64Ptr(uint64(got))
		resp, err := client.DescribeImages(req)
		if err != nil {
			blog.Errorf("DescribeImages failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeImages resp is nil")
			return nil, fmt.Errorf("DescribeImages resp is nil")
		}
		blog.Infof("DescribeImages success, requestID: %s", *resp.Response.RequestId)
		for _, v := range resp.Response.ImageSet {
			if v.ImageId == nil {
				continue
			}
			image := &proto.OsImage{
				ImageID: *v.ImageId,
			}
			if v.ImageName != nil {
				image.Alias = *v.ImageName
			}
			if v.Architecture != nil {
				image.Arch = *v.Architecture
			}
			if v.OsName != nil {
				image.OsName = *v.OsName
			}
			if v.ImageType != nil {
				image.Provider = *v.ImageType
			}
			if v.ImageState != nil {
				image.Status = *v.ImageState
			}
			images = append(images, image)
		}
		got += len(resp.Response.ImageSet)
		total = int(*resp.Response.TotalCount)
	}
	return images, nil
}

// DescribeKeyPairsByID describe ssh keyPairs https://cloud.tencent.com/document/product/213/15699
func (nm *NodeManager) DescribeKeyPairsByID(keyIDs []string,
	opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeKeyPairs failed: %v", err)
		return nil, err
	}

	idChunks := utils.SplitStringsChunks(keyIDs, limit)
	blog.Infof("DescribeKeyPairsByID Chunks %+v", idChunks)

	var (
		keyPairs = make([]*proto.KeyPair, 0)
		lock     = sync.Mutex{}
	)

	barrier := utils.NewRoutinePool(20)
	defer barrier.Close()

	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			barrier.Add(1)
			go func(ids []string) {
				defer barrier.Done()

				req := cvm.NewDescribeKeyPairsRequest()
				req.KeyIds = common.StringPtrs(ids)
				req.Limit = common.Int64Ptr(limit)

				resp, err := client.DescribeKeyPairs(req)
				if err != nil {
					blog.Errorf("DescribeKeyPairs[%v] failed: %v", ids, err)
					return
				}
				if len(resp.Response.KeyPairSet) == 0 {
					return
				}

				for i := range resp.Response.KeyPairSet {
					lock.Lock()
					keyPairs = append(keyPairs, &proto.KeyPair{
						KeyID:       *resp.Response.KeyPairSet[i].KeyId,
						KeyName:     *resp.Response.KeyPairSet[i].KeyName,
						Description: *resp.Response.KeyPairSet[i].Description,
					})
					lock.Unlock()
				}
			}(chunk)
		}
	}
	barrier.Wait()

	return keyPairs, nil
}

// ListKeyPairs describe all ssh keyPairs https://cloud.tencent.com/document/product/213/15699
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	client, err := GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when ListKeyPairs failed: %v", err)
		return nil, err
	}

	var (
		keyPairs = make([]*proto.KeyPair, 0)

		initOffset int64
		keyListLen = limit
	)

	for {
		if keyListLen != limit {
			break
		}
		req := cvm.NewDescribeKeyPairsRequest()
		req.Offset = common.Int64Ptr(initOffset)
		req.Limit = common.Int64Ptr(limit)

		/*
			for i := range filters {
				req.Filters = append(req.Filters, &cvm.Filter{
					Name:   common.StringPtr(filters[i].Name),
					Values: common.StringPtrs(filters[i].Values),
				})
			}
		*/

		resp, err := client.DescribeKeyPairs(req)
		if err != nil {
			blog.Errorf("cvm client DescribeKeyPairs failed, %s", err.Error())
			continue
		}

		// check response
		response := resp.Response
		if response == nil {
			blog.Errorf("cvm client DescribeKeyPairs but lost response information")
			continue
		}

		for i := range response.KeyPairSet {
			keyPairs = append(keyPairs, &proto.KeyPair{
				KeyID:       *response.KeyPairSet[i].KeyId,
				KeyName:     *response.KeyPairSet[i].KeyName,
				Description: *response.KeyPairSet[i].Description,
			})
		}

		keyListLen = len(response.KeyPairSet)
		initOffset = initOffset + limit
	}

	blog.Infof("ListKeyPairs successful")

	return keyPairs, nil
}

// ListOsImage list image os
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	os := make([]*proto.OsImage, 0)
	for _, v := range utils.ImageOsList {
		if provider == v.Provider {
			os = append(os, v)
		}
	}

	return os, nil
}

// GetExternalNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	node := &proto.Node{}

	ips := []string{ip}
	hostData, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(ips, cmdb.Page{
		Start: 0,
		Limit: len(ips),
	})
	if err != nil {
		blog.Errorf("GetExternalNodeByIP failed: %v", err)
		return nil, err
	}

	node.InnerIP = hostData[0].BKHostInnerIP
	node.CPU = uint32(hostData[0].HostCpu)
	node.Mem = uint32(math.Floor(float64(hostData[0].HostMem) / float64(1024)))
	node.InstanceType = hostData[0].NormalDeviceType
	node.Region = cmdb.GetCityZoneByCityName(hostData[0].IDCCityName)

	node.NodeType = icommon.IDC.String()
	return node, nil
}

// ListExternalNodesByIP list node by IP set
func (nm *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	var nodes []*proto.Node

	hostDataList, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(ips, cmdb.Page{
		Start: 0,
		Limit: len(ips),
	})
	if err != nil {
		blog.Errorf("ListExternalNodesByIP failed: %v", err)
		return nil, err
	}
	hostMap := make(map[string]cmdb.HostDetailData)
	for i := range hostDataList {
		hostMap[hostDataList[i].BKHostInnerIP] = hostDataList[i]
	}

	for _, ip := range ips {
		if host, ok := hostMap[ip]; ok {
			node := &proto.Node{}
			node.InnerIP = host.BKHostInnerIP
			node.CPU = uint32(host.HostCpu)
			node.Mem = uint32(math.Floor(float64(host.HostMem) / float64(1024)))
			node.InstanceType = host.NormalDeviceType
			node.Region = cmdb.GetCityZoneByCityName(host.IDCCityName)
			node.NodeType = icommon.IDC.String()

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}
