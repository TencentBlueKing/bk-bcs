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
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	// "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/daemon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
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

	cloudZones, err := client.DescribeZones()
	if err != nil {
		blog.Errorf("GetZoneList failed, %s", err.Error())
		return nil, err
	}

	zones := make([]*proto.ZoneInfo, 0)
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
		})
	}

	return zones, nil
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
func (nm *NodeManager) ListNodeInstanceType(ctx context.Context, info cloudprovider.InstanceInfo,
	opt *cloudprovider.CommonOption) ([]*proto.InstanceType, error) {
	blog.Infof("ListNodeInstanceType zone: %s, nodeFamily: %s, cpu: %d, memory: %d, instanceType: %s",
		info.Zone, info.NodeFamily, info.CPU, info.Memory, info.InstanceType)

	if options.GetEditionInfo().IsInnerEdition() {
		return nm.getInnerInstanceTypes(ctx, info)
	}

	return nm.getCloudInstanceType(info, opt)
}

// getInnerInstanceTypes get inner instance types info
func (nm *NodeManager) getInnerInstanceTypes(ctx context.Context, info cloudprovider.InstanceInfo) ( // nolint
	[]*proto.InstanceType, error) {
	blog.Infof("getInnerInstanceTypes %+v", info)

	quoteGrayMode, err := project.GetProjectManagerClient().CheckProjectQuotaGrayLabel(ctx, info.ProjectID)
	if err != nil {
		blog.Errorf("GetProjectManagerClient GetProjectQuotaGrayLabel[%s] failed: %v",
			info.ProjectID, err)
		return nil, err
	}

	var targetTypes []resource.InstanceType

	if utils.StringInSlice(quoteGrayMode, []string{project.QuotaGrayOverMode, project.QuotaGrayNormalMode}) &&
		info.Provider != resource.SelfPool {
		targetTypes, err = nm.GetInstanceTypeByProjectQuotaList(info.ProjectID, info.Region, info.Provider, info.InstanceType)
		if err != nil {
			blog.Errorf("GetProjectManagerClient GetNodeGroupAndZoneResourceQuotas[%s:%s] failed: %v",
				info.ProjectID, info.Region, err)
			return nil, err
		}

		blog.Infof("GetNodeGroupAndZoneResourceQuotas successful[%+v]", targetTypes)
	} else {
		targetTypes, err = tresource.GetResourceManagerClient().GetInstanceTypes(context.Background(),
			info.Region, resource.InstanceSpec{
				BizID:        info.BizID,
				Cpu:          info.CPU,
				Mem:          info.Memory,
				Provider:     info.Provider,
				ResourceType: info.ResourceType,
				InstanceType: info.InstanceType,
			})
		if err != nil {
			blog.Errorf("resourceManager ListNodeInstanceType failed: %v", err)
			return nil, err
		}
		blog.Infof("getInnerInstanceTypes successful[%+v]", targetTypes)
	}

	var instanceTypes = make([]*proto.InstanceType, 0)
	for _, t := range targetTypes {
		instanceTypes = append(instanceTypes, &proto.InstanceType{
			NodeType:       t.NodeType,
			TypeName:       t.TypeName,
			NodeFamily:     t.NodeFamily,
			Region:         t.Region,
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
			AvailableQuota: uint32(t.OversoldAvailable),
		})
	}

	blog.Infof("getInnerInstanceTypes successful[%+v]", instanceTypes)

	if info.Provider == resource.SelfPool || info.Provider == resource.CrPool ||
		utils.StringInSlice(quoteGrayMode, []string{project.QuotaGrayOverMode, project.QuotaGrayNormalMode}) {
		return instanceTypes, nil
	}

	// 获取当前资源池的使用情况 & 超卖情况
	var (
		barrier = utils.NewRoutinePool(50)
		lock    = sync.Mutex{}
	)
	defer barrier.Close()

	for i := range instanceTypes {
		barrier.Add(1)

		go func(i int) {
			defer func() {
				barrier.Done()
			}()

			poolQuota, exist := daemon.GetResourceDevicePoolData(instanceTypes[i].ResourcePoolID)
			if exist {
				lock.Lock()
				instanceTypes[i].AvailableQuota = func() uint32 {
					if uint32(poolQuota.OversoldTotal) <= uint32(poolQuota.GroupQuota) {
						return 0
					}
					return uint32(poolQuota.OversoldTotal) - uint32(poolQuota.GroupQuota)
				}()
				lock.Unlock()

				return
			}

			blog.Infof("getInnerInstanceTypes region[%s] insType[%s] devicePoolId[%s] not exist",
				info.Region, instanceTypes[i].NodeType, instanceTypes[i].ResourcePoolID)
		}(i)
	}
	barrier.Wait()

	return instanceTypes, nil
}

// GetInstanceTypeByProjectQuotaList get instanceType from zoneResource by project quota list info
func (nm *NodeManager) GetInstanceTypeByProjectQuotaList(
	projectId, region string, provider string, instanceType string) ([]resource.InstanceType, error) {
	listProjectQuotasData, err := project.GetProjectManagerClient().ListProjectQuotas(projectId,
		project.ProjectQuotaHostType, project.ProjectQuotaProvider)
	if err != nil {
		blog.Errorf("GetProjectManagerClient GetListProjectQuotas[%s:%s:%s] failed: %v", projectId,
			project.ProjectQuotaHostType, project.ProjectQuotaProvider, err)
		return nil, err
	}

	var (
		instanceTypes     = make([]resource.InstanceType, 0)
		projectQuotaLists = listProjectQuotasData.GetResults()
	)

	for _, projectQuota := range projectQuotaLists {
		zoneResources := projectQuota.GetQuota().GetZoneResources()
		if region != "" && region != zoneResources.GetRegion() {
			continue
		}

		if projectQuota.GetStatus() != icommon.StatusRunning {
			continue
		}

		if instanceType != "" {
			gpuNum := zoneResources.GetGpu()
			switch instanceType {
			case icommon.CvmInstanceType:
				if gpuNum != 0 {
					continue
				}
			case icommon.GpuInstanceType:
				if gpuNum == 0 {
					continue
				}
			default:
			}
		}

		availableQuota := zoneResources.GetQuotaNum() - zoneResources.GetQuotaUsed()
		// instanceType sell status
		status := icommon.InstanceSell
		if availableQuota <= 0 {
			status = icommon.InstanceSoldOut
		}

		instanceTypes = append(instanceTypes, resource.InstanceType{
			NodeType:       zoneResources.GetInstanceType(),
			Cpu:            zoneResources.GetCpu(),
			Memory:         zoneResources.GetMem(),
			Gpu:            zoneResources.GetGpu(),
			Status:         status,
			Zones:          []string{zoneResources.GetZoneName()},
			Provider:       provider,
			ResourcePoolID: projectQuota.GetQuotaId(),
			SystemDisk: func() *resource.DataDisk {
				systemDisks := nm.ConvertDataDisk([]*project.DataDisk{zoneResources.GetSystemDisk()})
				if len(systemDisks) > 0 {
					return systemDisks[0]
				}
				return nil
			}(),
			DataDisks:         nm.ConvertDataDisk(zoneResources.GetDataDisks()),
			OversoldAvailable: int32(availableQuota),
			Region:            zoneResources.GetRegion(),
		})
	}

	return instanceTypes, nil
}

// getCloudInstanceType get cloud instance type and filter instanceType by cpu&mem size
func (nm *NodeManager) getCloudInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	blog.Infof("getCloudInstanceType %+v", info)

	client, err := api.GetCVMClient(opt)
	if err != nil {
		blog.Errorf("create CVM client when transInstanceIDsToNodes failed, %s", err.Error())
		return nil, err
	}

	cloudInstanceTypes, err := client.DescribeZoneInstanceConfigInfos(info.Zone, info.NodeFamily, "")
	if err != nil {
		return nil, err
	}

	list := make([]*proto.InstanceType, 0)
	instanceMap := make(map[string][]string) // instanceType: []zone
	for _, v := range cloudInstanceTypes {
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
		if info.CPU > 0 {
			if item.Cpu != info.CPU {
				continue
			}
		}
		if info.Memory > 0 {
			if item.Memory != info.Memory {
				continue
			}
		}
		if info.InstanceType != "" {
			switch info.InstanceType {
			case icommon.CvmInstanceType:
				if item.Gpu != 0 {
					continue
				}
			case icommon.GpuInstanceType:
				if item.Gpu == 0 {
					continue
				}
			default:
			}
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

// GetServiceRoles service roles list
func (nm *NodeManager) GetServiceRoles(opt *cloudprovider.CommonOption, roleType string) (
	[]*proto.ServiceRoleInfo, error) {
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
func (nm *NodeManager) GetExternalNodeByIP(ip string,
	opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	node := &proto.Node{}

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(),
		tenant.ResourceMetaData{ClusterId: opt.ClusterID})
	if err != nil {
		return nil, err
	}

	ips := []string{ip}
	hostData, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(ctx, cmdb.FieldHostIP, ips, cmdb.Page{
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
func (nm *NodeManager) ListExternalNodesByIP(ips []string,
	opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	var nodes []*proto.Node

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(),
		tenant.ResourceMetaData{ClusterId: opt.ClusterID})
	if err != nil {
		return nil, err
	}

	hostDataList, err := cmdb.GetCmdbClient().QueryHostInfoWithoutBiz(ctx, cmdb.FieldHostIP, ips, cmdb.Page{
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

// ListDiskTypes get disk type list
func (nm *NodeManager) ListDiskTypes(instanceTypes []string, zones []string, diskChargeType string, cpu,
	memory uint64, opt *cloudprovider.CommonOption) (
	[]*proto.DiskConfigSet, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodePublicPrefixs get public prefixs list
func (nm *NodeManager) ListNodePublicPrefixs(opt *cloudprovider.ListNodePublicPrefixesOption) (
	[]*proto.NodePublicPrefix, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ConvertDataDisk convert project pb data disk to resource pb data disk
func (nm *NodeManager) ConvertDataDisk(srcDataDisks []*project.DataDisk) []*resource.DataDisk {
	if srcDataDisks == nil {
		return nil
	}
	protoDataDisk := make([]*resource.DataDisk, 0)
	for _, dataDisk := range srcDataDisks {
		protoDataDisk = append(protoDataDisk, &resource.DataDisk{
			DiskType: dataDisk.DiskType,
			DiskSize: dataDisk.DiskSize,
		})
	}
	return protoDataDisk
}
