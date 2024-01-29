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

// Package business xxx
package business

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetCVMImageIDByImageName get image info by image name
func GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetCVMImageIDByImageName failed, %s", err.Error())
		return "", err
	}

	cloudImages, err := client.ListImages()
	if err != nil {
		blog.Errorf("GetCVMImageIDByImageName cvm ListImages %s failed, %s", imageName, err.Error())
		return "", err
	}

	var (
		imageIDList = make([]string, 0)
	)
	for _, image := range cloudImages {
		if *image.ImageName == imageName {
			imageIDList = append(imageIDList, *image.ImageId)
		}
	}
	blog.Infof("GetCVMImageIDByImageName successful %v", imageIDList)

	if len(imageIDList) == 0 {
		return "", fmt.Errorf("GetCVMImageIDByImageName[%s] failed: imageIDList empty", imageName)
	}

	return imageIDList[0], nil
}

// GetCloudRegions get cloud regions
func GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetRegionsInfo failed: %v", err)
		return nil, err
	}

	cloudRegions, err := client.GetCloudRegions()
	if err != nil {
		blog.Errorf("GetCloudRegions failed, %s", err.Error())
		return nil, err
	}

	regions := make([]*proto.RegionInfo, 0)
	for i := range cloudRegions {
		regions = append(regions, &proto.RegionInfo{
			Region:      *cloudRegions[i].Region,
			RegionName:  *cloudRegions[i].RegionName,
			RegionState: *cloudRegions[i].RegionState,
		})
	}

	return regions, nil
}

// ListNodesByInstanceID list node by instanceIDs
func ListNodesByInstanceID(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	idChunks := utils.SplitStringsChunks(ids, common.Limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByInstanceID ipChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := TransInstanceIDsToNodes(chunk, opt)
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

// ListNodesByIP list node by IP set
func ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	ipChunks := utils.SplitStringsChunks(ips, common.MaxFilterValues)
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
				nodes, err := TransIPsToNodes(ips, opt)
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

// ListExternalNodesByIP list node by IP set
func ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
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
			node.NodeType = common.IDC.String()

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// InstanceToNode parse Instance information in qcloud to Node in clustermanager
// @param Instance: qcloud instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(inst *cvm.Instance, zoneInfo map[string]*proto.ZoneInfo) *proto.Node {
	var zone *proto.ZoneInfo
	// zone may be nil when api qps limit exceed or zone not exist
	if zoneInfo != nil {
		zone = zoneInfo[*inst.Placement.Zone]
	}

	node := &proto.Node{
		NodeID:       *inst.InstanceId,
		InstanceType: *inst.InstanceType,
		CPU:          uint32(*inst.CPU),
		Mem:          uint32(*inst.Memory),
		GPU:          0,
		VPC:          *inst.VirtualPrivateCloud.VpcId,
		ZoneID:       *inst.Placement.Zone,
		Zone: func() uint32 {
			if zone != nil {
				zoneID, _ := strconv.ParseUint(zone.ZoneID, 10, 32)
				return uint32(zoneID)
			}
			return 0
		}(),
		InnerIPv6: utils.SlicePtrToString(inst.IPv6Addresses),
		ZoneName: func() string {
			if zone != nil {
				return zone.ZoneName
			}
			return ""
		}(),
		ChargeType: *inst.InstanceChargeType,
	}
	return node
}

// GetZoneInfoByRegion region: ap-nanjing/ap-shenzhen
func GetZoneInfoByRegion(opt *cloudprovider.CommonOption) (map[string]*proto.ZoneInfo,
	map[string]*proto.ZoneInfo, error) {
	cvmClient, err := api.GetCVMClient(opt)
	if err != nil {
		return nil, nil, fmt.Errorf("getZoneInfoByRegion GetZoneInfo failed: %v", err)
	}

	zones, err := cvmClient.DescribeZones()
	if err != nil {
		return nil, nil, fmt.Errorf("getZoneInfoByRegion GetZoneInfo failed: %v", err)
	}

	var (
		zoneMap   = make(map[string]*proto.ZoneInfo)
		zoneIdMap = make(map[string]*proto.ZoneInfo)
	)

	for i := range zones {
		if _, ok := zoneMap[*zones[i].Zone]; !ok {
			// zoneID, _ := strconv.ParseUint(zones[i].ZoneID, 10, 32)
			zoneMap[*zones[i].Zone] = &proto.ZoneInfo{
				ZoneID:   *zones[i].ZoneId,
				Zone:     *zones[i].Zone,
				ZoneName: *zones[i].ZoneName,
			}
		}

		if _, ok := zoneIdMap[*zones[i].ZoneId]; !ok {
			zoneIdMap[*zones[i].ZoneId] = &proto.ZoneInfo{
				ZoneID:   *zones[i].ZoneId,
				Zone:     *zones[i].Zone,
				ZoneName: *zones[i].ZoneName,
			}
		}
	}

	return zoneIdMap, zoneMap, nil
}

// TransInstanceIDsToNodes trans IDList to Nodes
func TransInstanceIDsToNodes(
	ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := api.GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when transInstanceIDsToNodes failed, %s", err.Error())
		return nil, err
	}

	cloudInstances, err := client.GetInstancesById(ids)
	if err != nil {
		blog.Errorf("cvm client GetInstancesById len(%d) failed, %s", len(ids), err.Error())
		return nil, err
	}
	_, zoneInfo, err := GetZoneInfoByRegion(opt.Common)
	if err != nil {
		blog.Errorf("cvm client GetZoneInfoByRegion failed: %v", err)
	}

	var (
		nodeMap = make(map[string]*proto.Node)
		nodes   []*proto.Node
	)

	for _, inst := range cloudInstances {
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

// TransIPsToNodes trans IPList to Nodes, filter max 5 values
func TransIPsToNodes(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := api.GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when transIPsToNodes failed, %s", err.Error())
		return nil, err
	}
	cloudInstances, err := client.GetInstancesByIp(ips)
	if err != nil {
		blog.Errorf("cvm client transIPsToNodes GetInstancesByIp len(%d) "+
			"ip address failed, %s", len(ips), err.Error())
		return nil, err
	}
	_, zoneInfo, err := GetZoneInfoByRegion(opt.Common)
	if err != nil {
		blog.Errorf("cvm client transIPsToNodes GetZoneInfoByRegion failed: %v", err)
	}

	var (
		nodeMap = make(map[string]*proto.Node)
		nodes   []*proto.Node
	)

	for _, ip := range ips {
		for _, inst := range cloudInstances {
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
