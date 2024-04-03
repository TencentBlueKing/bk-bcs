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

// Package huawei xxx
package huawei

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
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

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, nil
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	return nil, nil
}

// GetCVMImageIDByImageName get imageID by imageName
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	return "", nil
}

// GetCloudRegions get cloud regions
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	client, err := api.GetIamClient(opt)
	if err != nil {
		return nil, err
	}

	cloudRegions, err := client.GetCloudRegions()
	if err != nil {
		return nil, err
	}

	regions := make([]*proto.RegionInfo, 0)
	for _, v := range cloudRegions {
		regions = append(regions, &proto.RegionInfo{
			Region:      v.Id,
			RegionName:  v.Locales.ZhCn,
			RegionState: "AVAILABLE",
		})
	}

	return regions, nil
}

// ListExternalNodesByIP list node by IP set
func (nm *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairs describe all ssh keyPairs
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	client, err := api.NewKpsClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetAllUsableKeypairs()
	if err != nil {
		return nil, err
	}

	kps := make([]*proto.KeyPair, 0)
	for _, v := range rsp {
		kps = append(kps, &proto.KeyPair{
			KeyID:   *v.Keypair.Name,
			KeyName: *v.Keypair.Name,
		})
	}

	return kps, nil
}

// GetExternalNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetResourceGroups resource groups list
func (nm *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetZoneList get zoneList
func (nm *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	zones, err := api.GetAvailabilityZones(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	zoneInfos := make([]*proto.ZoneInfo, 0)
	for k, v := range zones {
		zoneInfos = append(zoneInfos, &proto.ZoneInfo{
			ZoneID:    v.ZoneName,
			Zone:      fmt.Sprintf("%d", k+1),
			ZoneName:  fmt.Sprintf("可用区%d", k+1),
			ZoneState: "AVAILABLE",
		})
	}

	return zoneInfos, nil
}

// ListNodeInstanceType list node type by zone and node family
func (nm *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	client, err := api.NewEcsClient(opt)
	if err != nil {
		return nil, err
	}

	flavors, err := client.GetAllFlavors(info.Zone)
	if err != nil {
		return nil, err
	}

	instanceTypes := make([]*proto.InstanceType, 0)
	for _, v := range *flavors {
		var (
			name string
			gpu  uint32
		)

		if v.OsExtraSpecs.Ecsperformancetype != nil {
			name = api.ConvertPerformanceType(*v.OsExtraSpecs.Ecsperformancetype)
			if v.OsExtraSpecs.Ecsgeneration != nil {
				name += *v.OsExtraSpecs.Ecsgeneration
			} else {
				var tmp []string
				if strings.Contains(v.Name, "-") {
					tmp = strings.Split(v.Name, "-")
				} else if strings.Contains(v.Name, ".") {
					tmp = strings.Split(v.Name, ".")
				}
				if len(tmp) > 0 {
					name += tmp[0]
				}
			}
		}

		cpu, _ := strconv.Atoi(v.Vcpus)
		status := ""
		if v.OsExtraSpecs.Condoperationstatus != nil {
			status = *v.OsExtraSpecs.Condoperationstatus
		}
		if v.OsExtraSpecs.Infogpuname != nil {
			res := strings.Split(*v.OsExtraSpecs.Infogpuname, "*")
			if len(res) > 0 {
				i, _ := strconv.Atoi(res[0])
				gpu = uint32(i)
			}
		}
		zones := make([]string, 0)
		if v.OsExtraSpecs.Condoperationaz != nil {
			res := strings.Split(*v.OsExtraSpecs.Condoperationaz, ",")
			for _, y := range res {
				zone := strings.Split(y, "(")
				if len(zone) > 0 {
					zones = append(zones, zone[0])
				}

			}
		}
		instanceTypes = append(instanceTypes, &proto.InstanceType{
			NodeType:   v.Name,
			TypeName:   name,
			NodeFamily: *v.OsExtraSpecs.ResourceType,
			Cpu:        uint32(cpu),
			Memory:     uint32(v.Ram / 1024),
			Gpu:        gpu,
			Status:     status,
			Zones:      zones,
		})
	}

	return instanceTypes, nil
}

// ListOsImage get osimage list
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
