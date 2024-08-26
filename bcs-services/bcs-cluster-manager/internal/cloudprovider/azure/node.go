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

package azure

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeManager(cloudName, &NodeManager{})
	})
}

// NodeManager define node manager
type NodeManager struct {
}

// GetNodeByIP 通过IP查询节点 - get specified Node by innerIP address
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
	client, err := api.NewAMClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, err %s", err.Error())
	}
	regions, err := client.ListLocations(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}
	return regions, nil
}

// GetZoneList get zoneList by region
func (n *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	client, err := api.NewAMClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, err %s", err.Error())
	}
	zones, err := client.ListAvailabilityZones(context.Background(), opt.Region)
	if err != nil {
		return nil, fmt.Errorf("list zones failed, err %s", err.Error())
	}
	return zones, nil
}

// ListNodeInstanceType list node type by zone and node family
func (n *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	cli, err := api.NewAksServiceImplWithCommonOption(opt)
	if err != nil {
		return nil, fmt.Errorf("ListNodeInstanceType create aks client failed, %v", err)
	}

	resources, err := cli.ListResourceByLocation(context.Background(), opt.Region)
	if err != nil {
		return nil, fmt.Errorf("ListNodeInstanceType ListResourceByLocation failed, %v", err)
	}

	instanceTypes := make([]*proto.InstanceType, 0)
	for _, v := range resources {
		var cpu, mem, gpu int
		if *v.ResourceType == "virtualMachines" {
			for _, c := range v.Capabilities {
				if *c.Name == "vCPUs" {
					cpu, _ = strconv.Atoi(*c.Value)
				}
				if *c.Name == "MemoryGB" {
					mem, _ = strconv.Atoi(*c.Value)
				}
				if *c.Name == "GPUs" {
					gpu, _ = strconv.Atoi(*c.Value)
				}
			}

			// filter cpu && mem
			if cpu == 0 || mem == 0 || cpu < 4 || mem < 4 {
				continue
			}

			zones := make([]string, 0)
			if len(v.LocationInfo) != 0 {
				for _, z := range v.LocationInfo[0].Zones {
					zones = append(zones, *z)
				}
			}

			instanceTypes = append(instanceTypes, &proto.InstanceType{
				NodeType:   *v.Name,
				TypeName:   *v.Name,
				NodeFamily: *v.Family,
				Cpu:        uint32(cpu),
				Memory:     uint32(mem),
				Gpu:        uint32(gpu),
				Zones:      zones,
				Status: func() string {
					if len(zones) == 0 {
						return common.InstanceSoldOut
					}

					return common.InstanceSell
				}(),
			})
		}
	}

	return instanceTypes, nil
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

// ListKeyPairs keyPairs list
func (n *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	cli, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListKeyPairs create aks client failed, %v", err)
	}

	result, err := cli.ListSSHPublicKeys(context.Background(), opt.ResourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("ListSSHPublicKeys failed, %v", err)
	}

	keys := make([]*proto.KeyPair, 0)
	for _, v := range result {
		keys = append(keys, &proto.KeyPair{KeyName: *v.Name})
	}

	return keys, nil
}

// GetResourceGroups resource groups list
func (n *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	client, err := api.NewResourceGroupsClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, err %s", err.Error())
	}

	groups, err := client.ListResourceGroups(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}

	return groups, nil
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
