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

package api

import (
	"context"
	"fmt"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager("google", &VPCManager{})
	})
}

// VPCManager is the manager for VPC
type VPCManager struct{}

// ListVpcs list vpcs
func (c *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.CommonOption) ([]*proto.CloudVpc, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListSubnets list vpc subnets
func (vm VPCManager) ListSubnets(vpcID string, opt *cloudprovider.CommonOption) ([]*proto.Subnet, error) {
	locationList := strings.Split(opt.Region, "-")
	if len(locationList) == 3 {
		opt.Region = strings.Join(locationList[:2], "-")
	}

	client, err := NewComputeServiceClient(opt)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	subnets, err := client.ListSubnetworks(context.Background(), opt.Region)
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}

	result := make([]*proto.Subnet, 0)
	for _, v := range subnets.Items {
		networkInfo := strings.Split(v.Network, "/")
		if vpcID != networkInfo[len(networkInfo)-1] {
			continue
		}
		regionInfo := strings.Split(v.Region, "/")
		result = append(result, &proto.Subnet{
			VpcID:         networkInfo[len(networkInfo)-1],
			SubnetID:      v.Name,
			SubnetName:    v.Name,
			CidrRange:     v.IpCidrRange,
			Ipv6CidrRange: v.Ipv6CidrRange,
			Zone:          regionInfo[len(regionInfo)-1],
		})
	}
	return result, nil
}

// ListSecurityGroups list security groups
func (vm VPCManager) ListSecurityGroups(opt *cloudprovider.CommonOption) ([]*proto.SecurityGroup, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCloudNetworkAccountType 查询用户网络类型
func (vm VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks list bandWidthPacks
func (vm VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
