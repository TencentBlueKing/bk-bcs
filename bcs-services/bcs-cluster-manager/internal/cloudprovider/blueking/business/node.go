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
	"math"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
)

// ListNodesByInstanceIP list node by IP set
func ListNodesByInstanceIP(ips []string) ([]*proto.Node, error) {
	var nodes []*proto.Node

	hostDataList, err := cmdb.GetCmdbClient().QueryAllHostInfoWithoutBiz(ips)
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

// ListNodesByIP list node by IP set
func ListNodesByIP(region string, ips []string) ([]*proto.Node, error) {
	var nodes []*proto.Node
	for _, ip := range ips {
		node := &proto.Node{}
		node.InnerIP = ip
		node.Region = region
		nodes = append(nodes, node)
	}
	return nodes, nil
}
