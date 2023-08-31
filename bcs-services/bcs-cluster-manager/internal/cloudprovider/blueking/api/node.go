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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		//init Node
		cloudprovider.InitNodeManager("blueking", &NodeManager{})
	})
}

// NodeManager CVM relative API management
type NodeManager struct {
}

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	node := &proto.Node{}
	node.InnerIP = ip

	if util.IsIPv6(ip) {
		node.InnerIPv6 = ip
	}
	if util.IsIPv4(ip) {
		node.InnerIP = ip
	}

	node.Region = opt.Common.Region
	return node, nil
}

// GetCloudRegions get regionInfo
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	// blueking cloud not need to implement interface
	return nil, nil
}

// GetZoneList get zoneList
func (nm *NodeManager) GetZoneList(opt *cloudprovider.CommonOption) ([]*proto.ZoneInfo, error) {
	// blueking cloud not need to implement interface
	return nil, nil
}

// GetCVMImageIDByImageName get imageID by imageName
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	// blueking cloud not need to implement interface
	return "", nil
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	var nodes []*proto.Node
	for _, ip := range ips {
		node := &proto.Node{}
		node.InnerIP = ip
		node.Region = opt.Common.Region
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// ListNodeInstanceType list node type by zone and node family
func (nm *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo, opt *cloudprovider.CommonOption) (
	[]*proto.InstanceType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListOsImage list image os
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) (
	[]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	node := &proto.Node{}
	node.InnerIP = ip
	node.Region = opt.Common.Region
	node.NodeType = common.IDC.String()
	return node, nil
}

// ListExternalNodesByIP list node by IP set
func (nm *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	var nodes []*proto.Node
	for _, ip := range ips {
		node := &proto.Node{}
		node.InnerIP = ip
		node.Region = opt.Common.Region
		node.NodeType = common.IDC.String()
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// ListKeyPairs keyPairs list
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.CommonOption) ([]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
