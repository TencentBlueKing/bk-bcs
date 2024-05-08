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

package qcloud

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeManager(cloudName, &NodeManager{})
	})
}

// NodeManager CVM relative API management
type NodeManager struct {
}

// GetZoneList get zoneList
func (nm *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	client, err := api.GetCVMClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create CVM client when GetZoneList failed: %v", err)
		return nil, err
	}

	var (
		zoneSubnetNum map[string]uint32
	)
	if len(opt.VpcId) > 0 {
		zoneSubnetNum, err = business.GetZoneAvailableSubnetsByVpc(&opt.CommonOption, opt.VpcId)
		if err != nil {
			blog.Errorf("GetZoneList GetZoneAvailableSubnetsByVpc[%s] failed: %v", opt.VpcId, err)
			return nil, err
		}
	}

	cloudZones, err := client.DescribeZones()
	if err != nil {
		blog.Errorf("GetZoneList failed, %s", err.Error())
		return nil, err
	}

	var (
		zones       = make([]*proto.ZoneInfo, 0)
		filterZones = make([]*proto.ZoneInfo, 0)
	)

	for i := range cloudZones {
		zones = append(zones, &proto.ZoneInfo{
			// 可用区ID 30003
			ZoneID: *cloudZones[i].ZoneId,
			// ap-nanjing-3
			Zone: *cloudZones[i].Zone,
			// 可用区描述，例如，南京三区
			ZoneName: *cloudZones[i].ZoneName,
			// 可用区状态，包含AVAILABLE和UNAVAILABLE。AVAILABLE代表可用，UNAVAILABLE代表不可用。
			ZoneState: *cloudZones[i].ZoneState,
			SubnetNum: func() uint32 {
				return zoneSubnetNum[*cloudZones[i].Zone]
			}(),
		})
	}

	if opt.State == "" {
		return zones, nil
	}

	for i := range zones {
		if zones[i].ZoneState == opt.State {
			filterZones = append(filterZones, zones[i])
		}
	}
	return filterZones, nil
}

// GetCloudRegions get regionInfo
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	if opt.Region == "" {
		opt.Region = defaultRegion
	}

	return business.GetCloudRegions(opt)
}

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	client, err := api.GetCVMClient(opt.Common)
	if err != nil {
		blog.Errorf("create CVM client when GetNodeByIP failed, %s", err.Error())
		return nil, err
	}

	instance, err := client.GetNodeInstanceByIP(ip)
	if err != nil {
		blog.Errorf("cvm client GetNodeInstanceByIP %s failed, %s", ip, err.Error())
		return nil, err
	}

	_, zoneInfo, err := business.GetZoneInfoByRegion(opt.Common)
	if err != nil {
		blog.Errorf("cvm client GetNodeByIP failed: %v", err)
	}

	node := business.InstanceToNode(instance, zoneInfo)
	node.InnerIP = ip
	node.Region = opt.Common.Region

	// check node vpc and cluster vpc
	if opt.ClusterVPCID != "" && !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
		return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
	}

	return node, nil
}

// Image info
type Image struct {
	// 镜像ID
	ImageId string
	// 镜像操作系统
	OsName string
	// 镜像类型
	ImageType string
	// 镜像名称
	ImageName string
}

// GetImageInfoByImageID get image by image
func (nm *NodeManager) GetImageInfoByImageID(imageID string, opt *cloudprovider.CommonOption) (*Image, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when GetImageByImageID failed, %s", err.Error())
		return nil, err
	}

	image, err := client.GetImageByImageID(imageID)
	if err != nil {
		blog.Errorf("GetImageByImageID %s failed, %s", imageID, err.Error())
		return nil, err
	}

	return &Image{
		ImageId:   *image.ImageId,
		OsName:    *image.OsName,
		ImageType: *image.ImageType,
		ImageName: *image.ImageName,
	}, nil
}

// GetCVMImageIDByImageName get image by image name
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	return business.GetCVMImageIDByImageName(imageName, opt)
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	ipChunks := utils.SplitStringsChunks(ips, icommon.MaxFilterValues)
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
				nodes, err := business.TransIPsToNodes(ips, opt)
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
	idChunks := utils.SplitStringsChunks(ids, icommon.Limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByInstanceID ipChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := business.TransInstanceIDsToNodes(chunk, opt)
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

// ListNodeInstanceType list node type by zone and node family
func (nm *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("ListNodeInstanceType zone: %s, nodeFamily: %s, cpu: %d, memory: %d",
		info.Zone, info.NodeFamily, info.Cpu, info.Memory)

	return nm.getCloudInstanceType(info, opt)
}

// getCloudInstanceType get cloud instance type and filter instanceType by cpu&mem size
func (nm *NodeManager) getCloudInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("getCloudInstanceType %+v", info)

	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when getCloudInstanceType failed, %s", err.Error())
		return nil, err
	}

	cloudInstanceTypes, err := client.DescribeZoneInstanceConfigInfos(info.Zone, info.NodeFamily, "")
	if err != nil {
		return nil, err
	}

	list := make([]*proto.InstanceType, 0)
	instanceMap := make(map[string][]string) // instanceType: []zone
	for _, v := range cloudInstanceTypes {
		// qcloud filter small instance
		if (v.Cpu != nil && *v.Cpu < 4) || (v.Memory != nil && *v.Memory < 8) {
			continue
		}

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
		list = append(list, t)
	}
	for i := range list {
		list[i].Zones = instanceMap[list[i].NodeType]
	}
	blog.Infof("DescribeZoneInstanceConfigInfos success, result: %s", utils.ToJSONString(list))

	// filter result instanceTypes
	result := make([]*proto.InstanceType, 0)
	for _, item := range list {
		if info.Cpu > 0 && item.Cpu != info.Cpu {
			continue
		}
		if info.Memory > 0 && item.Memory != info.Memory {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// DescribeKeyPairsByID describe ssh keyPairs https://cloud.tencent.com/document/product/213/15699
func (nm *NodeManager) DescribeKeyPairsByID(keyIDs []string,
	opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when DescribeKeyPairs failed: %v", err)
		return nil, err
	}

	idChunks := utils.SplitStringsChunks(keyIDs, icommon.Limit)
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

				LocalKeyPairs, err := client.DescribeKeyPairsByID(ids)
				if err != nil {
					blog.Errorf("DescribeKeyPairs[%v] failed: %v", ids, err)
					return
				}
				if len(LocalKeyPairs) == 0 {
					return
				}

				for i := range LocalKeyPairs {
					lock.Lock()
					keyPairs = append(keyPairs, &proto.KeyPair{
						KeyID:       *LocalKeyPairs[i].KeyId,
						KeyName:     *LocalKeyPairs[i].KeyName,
						Description: *LocalKeyPairs[i].Description,
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
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	client, err := api.GetCVMClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create CVM client when ListKeyPairs failed: %v", err)
		return nil, err
	}

	var (
		keyPairs = make([]*proto.KeyPair, 0)
	)

	cloudKeyPairs, err := client.ListKeyPairs()
	if err != nil {
		blog.Errorf("cvm client DescribeKeyPairs failed, %s", err.Error())
		return nil, err
	}

	for i := range cloudKeyPairs {
		keyPairs = append(keyPairs, &proto.KeyPair{
			KeyID:       *cloudKeyPairs[i].KeyId,
			KeyName:     *cloudKeyPairs[i].KeyName,
			Description: *cloudKeyPairs[i].Description,
		})
	}
	blog.Infof("ListKeyPairs successful")

	return keyPairs, nil
}

// GetResourceGroups resource groups list
func (nm *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodeRoles node roles list
func (nm *NodeManager) GetNodeRoles(opt *cloudprovider.CommonOption) ([]*proto.NodeRoleInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	hostData, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(cmdb.FieldHostIP, ips, cmdb.Page{
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

	hostDataList, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(cmdb.FieldHostIP, ips, cmdb.Page{
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

// ListRuntimeInfo get runtime info list
func (nm *NodeManager) ListRuntimeInfo(opt *cloudprovider.ListRuntimeInfoOption) (map[string][]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
