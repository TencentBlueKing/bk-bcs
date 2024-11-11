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
	"net"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
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
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	client, err := NewComputeServiceClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}

	networks, err := client.ListNetworks(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list networks failed, err %s", err.Error())
	}

	result := make([]*proto.CloudVpc, 0)
	for _, v := range networks.Items {
		if vpcID != "" && vpcID != v.Name {
			continue
		}
		result = append(result, &proto.CloudVpc{
			Name:     v.Name,
			VpcId:    fmt.Sprint(v.Id),
			Ipv4Cidr: v.IPv4Range,
		})
	}

	return result, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	locationList := strings.Split(opt.Region, "-")
	if len(locationList) == 3 {
		opt.Region = strings.Join(locationList[:2], "-")
	}

	client, err := NewComputeServiceClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	subnets, err := client.ListSubnetworks(context.Background(), opt.Region)
	if err != nil {
		return nil, fmt.Errorf("list subnets failed, err %s", err.Error())
	}

	result := make([]*proto.Subnet, 0)
	for _, v := range subnets.Items {
		networkInfo := strings.Split(v.Network, "/")
		if vpcID != "" && vpcID != networkInfo[len(networkInfo)-1] {
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
func (vm *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCloudNetworkAccountType 查询用户网络类型
func (vm *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks list bandWidthPacks
func (vm *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckConflictInVpcCidr check cidr if conflict with vpc cidrs
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CheckConflictInVpcCidrOption) ([]string, error) {
	locationList := strings.Split(opt.Region, "-")
	if len(locationList) == 3 {
		opt.Region = strings.Join(locationList[:2], "-")
	}

	client, err := NewComputeServiceClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	subnets, err := client.ListSubnetworks(context.Background(), opt.Region)
	if err != nil {
		return nil, fmt.Errorf("list subnets failed, err %s", err.Error())
	}

	if len(subnets.Items) == 0 {
		return nil, fmt.Errorf("subnet not found")
	}

	ipNets := make([]*net.IPNet, 0)
	for _, v := range subnets.Items {
		networkInfo := strings.Split(v.Network, "/")
		if vpcID != "" && vpcID != networkInfo[len(networkInfo)-1] {
			continue
		}

		for _, ipRange := range v.SecondaryIpRanges {
			_, c, err := net.ParseCIDR(ipRange.IpCidrRange)
			if err != nil {
				return nil, err
			}

			ipNets = append(ipNets, c)
		}
	}

	_, c, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	conflictCidrs := make([]string, 0)
	for i := range ipNets {
		if cidrtree.CidrContains(ipNets[i], c) || cidrtree.CidrContains(c, ipNets[i]) {
			conflictCidrs = append(conflictCidrs, ipNets[i].String())
		}
	}

	return conflictCidrs, nil
}

// AllocateOverlayCidr allocate overlay cidr
func (vm *VPCManager) AllocateOverlayCidr(vpcId string, cluster *proto.Cluster, cidrLens []uint32,
	reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, nil
}

// AddClusterOverlayCidr add cidr to cluster
func (vm *VPCManager) AddClusterOverlayCidr(clusterId string, cidrs []string, opt *cloudprovider.CommonOption) error {
	return nil
}

// GetVpcIpUsage get vpc ipTotal & ipSurplus
func (vm *VPCManager) GetVpcIpUsage(
	vpcId string, ipType string, reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) (uint32, uint32, error) {
	return 0, 0, nil
}

// GetClusterIpUsage get cluster ip usage
func (vm *VPCManager) GetClusterIpUsage(clusterId string, ipType string, opt *cloudprovider.CommonOption) (
	uint32, uint32, error) {
	return 0, 0, nil
}
