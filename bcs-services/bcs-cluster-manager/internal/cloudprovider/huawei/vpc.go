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
	"net"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/business"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager(cloudName, &VPCManager{})
	})
}

// VPCManager is the manager for VPC
type VPCManager struct{}

// ListVpcs list vpcs
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	client, err := api.NewVpcClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	cloudVpcs, err := client.ListVpcs(func() []string {
		if len(vpcID) == 0 {
			return nil
		}

		return []string{vpcID}
	}())
	if err != nil {
		return nil, err
	}

	vpcs := make([]*proto.CloudVpc, 0)
	for _, v := range cloudVpcs {
		vpcs = append(vpcs, &proto.CloudVpc{
			VpcId:    v.Id,
			Name:     v.Name,
			Ipv4Cidr: v.Cidr,
		})
	}
	// vpc 剩余的可用IP数量

	return vpcs, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	cloudSubnets, err := business.GetCloudSubnetsByVpc(vpcID, opt.CommonOption)
	if err != nil {
		return nil, err
	}
	zones, err := business.GetCloudZones(opt.CommonOption)
	if err != nil {
		return nil, err
	}
	subnets := make([]*proto.Subnet, 0)

	for _, s := range cloudSubnets {
		subnetZone := ""
		subnetZoneName := ""

		switch *s.Scope {
		case api.SubnetScopeAz:
			for _, v := range zones {
				if v.ZoneName == s.AvailabilityZone {
					subnetZone = v.ZoneName
					subnetZoneName = fmt.Sprintf("可用区%d", func() int {
						return business.GetZoneNameByZoneId(opt.Region, v.ZoneName)
					}())
				}
			}
		default:
		}

		cnt, errLocal := business.GetSubnetAvailableIpNum(s.Id, opt.CommonOption)
		if errLocal != nil {
			return nil, errLocal
		}

		subnets = append(subnets, &proto.Subnet{
			VpcID:                   s.VpcId,
			SubnetID:                s.Id,
			SubnetName:              s.Name,
			CidrRange:               s.Cidr,
			Ipv6CidrRange:           s.CidrV6,
			Zone:                    subnetZone,
			ZoneName:                subnetZoneName,
			AvailableIPAddressCount: uint64(cnt),
			HwNeutronSubnetID:       s.NeutronSubnetId,
		})
	}

	return subnets, nil
}

// ListSecurityGroups list security groups
func (vm *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	client, err := api.NewVpcClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	secs, err := client.ListSecurityGroups(nil)
	if err != nil {
		return nil, err
	}

	sgs := make([]*proto.SecurityGroup, 0)
	for _, v := range secs {
		sgs = append(sgs, &proto.SecurityGroup{
			SecurityGroupID:   v.Id,
			SecurityGroupName: v.Name,
			Description:       v.Description,
		})
	}

	return sgs, nil
}

// GetCloudNetworkAccountType 查询用户网络类型
func (vm *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks list bandWidthPacks
func (vm *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	client, err := api.NewEipClient(opt)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetAllBandwidths()
	if err != nil {
		return nil, err
	}

	bandwidths := make([]*proto.BandwidthPackageInfo, 0)
	for _, v := range rsp {
		bandwidths = append(bandwidths, &proto.BandwidthPackageInfo{
			Id:          *v.Id,
			Name:        *v.Name,
			NetworkType: *v.BandwidthType,
			Status:      *v.AdminState,
			Bandwidth:   *v.Size,
		})
	}

	return bandwidths, nil
}

// CheckConflictInVpcCidr check cidr if conflict with vpc cidrs
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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

// GetVpcIpUsage get vpc ipTotal/ipSurplus
func (vm *VPCManager) GetVpcIpUsage(
	vpcId string, ipType string, reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) (uint32, uint32, error) {
	return 0, 0, nil
}

// GetClusterIpUsage get cluster ip usage
func (vm *VPCManager) GetClusterIpUsage(clusterId string, ipType string, opt *cloudprovider.CommonOption) (
	uint32, uint32, error) {
	return 0, 0, nil
}
