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

package eop

import (
	"fmt"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/eop/api"
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

// GetExternalNodeByIP get specified Node by innerIP address
func (n *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListExternalNodesByIP list node by IP set
func (n *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodeInstanceType list node type by zone and node family
func (n *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo,
	opt *cloudprovider.CommonOption) ([]*proto.InstanceType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairs keyPairs list
func (n *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	client, err := api.NewCTClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, err %s", err.Error())
	}
	regions, err := client.ListRegions()
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}

	regionInfo := make([]*proto.RegionInfo, 0)
	for _, r := range regions {
		regionInfo = append(regionInfo, &proto.RegionInfo{
			Region:     r.Name,
			RegionName: r.Name,
		})
	}

	return regionInfo, nil
}

// GetZoneList get zoneList by region
func (n *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	client, err := api.NewCTClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, err %s", err.Error())
	}
	regions, err := client.ListRegions()
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}

	zoneInfo := make([]*proto.ZoneInfo, 0)
	for _, r := range regions {
		if r.Name == opt.Region {
			for _, z := range r.Zones {
				zoneInfo = append(zoneInfo, &proto.ZoneInfo{
					Zone:     z.NodeCode,
					ZoneName: z.Name,
				})
			}
		}
	}

	return zoneInfo, nil
}

// ListOsImage get osimage list
func (n *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetResourceGroups resource groups list
func (n *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
