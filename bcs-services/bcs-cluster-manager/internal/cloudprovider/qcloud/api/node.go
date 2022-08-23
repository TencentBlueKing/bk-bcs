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
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
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
			{
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
				nodes, err := nm.transIPsToNodes(chunk, opt)
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
		if !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
			return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
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

// ListNodeInstanceType list node type by zone and node family
func (nm *NodeManager) ListNodeInstanceType(zone, nodeFamily string, cpu, memory uint32, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("ListNodeInstanceType zone: %s, nodeFamily: %s, cpu: %d, memory: %d", zone, nodeFamily, cpu, memory)
	list, err := nm.DescribeZoneInstanceConfigInfos(zone, nodeFamily, "", opt)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.InstanceType, 0)
	for _, item := range list {
		if cpu > 0 {
			if item.Cpu != cpu {
				continue
			}
		}
		if memory > 0 {
			if item.Memory != memory {
				continue
			}
		}
		result = append(result, item)
	}
	return result, nil
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
			t.NodeFamily = *v.TypeName
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

// DescribeImages describe images
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

// ListOsImage list image os
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if provider == cmcommon.MarketImageProvider {
		return imageOsList, nil
	}
	return nm.DescribeImages(provider, opt)
}
