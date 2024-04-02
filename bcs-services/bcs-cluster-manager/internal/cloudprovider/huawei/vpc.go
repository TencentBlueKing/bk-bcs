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

	model2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	model3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
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
	client, err := api.GetVpc2Client(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	cloudVpcs, err := client.ListVpcsByID(vpcID)
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

	return vpcs, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	client, err := api.GetVpc2Client(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListSubnets(&model2.ListSubnetsRequest{
		VpcId: &vpcID,
	})
	if err != nil {
		return nil, err
	}

	subnetZone := ""
	subnetZoneName := ""
	subnets := make([]*proto.Subnet, 0)

	// 获取可用区
	zones, err := api.GetAvailabilityZones(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	for _, s := range *rsp.Subnets {
		for k, v := range zones {
			if v.ZoneName == s.AvailabilityZone {
				subnetZone = fmt.Sprintf("%d", k+1)
				subnetZoneName = fmt.Sprintf("可用区%d", k+1)
			}
		}

		rps2, err2 := client.ListPrivateips(&model2.ListPrivateipsRequest{SubnetId: s.Id})
		if err2 != nil {
			return nil, err
		}

		total, _ := calculateAvailableIPs(s.Cidr)

		subnets = append(subnets, &proto.Subnet{
			VpcID:                   s.VpcId,
			SubnetID:                s.Id,
			SubnetName:              s.Name,
			CidrRange:               s.Cidr,
			Ipv6CidrRange:           s.CidrV6,
			Zone:                    subnetZone,
			ZoneName:                subnetZoneName,
			AvailableIPAddressCount: uint64(total - len(*rps2.Privateips)),
		})
	}

	return subnets, nil
}

// ListSecurityGroups list security groups
func (vm *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	client, err := api.GetVpc3Client(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListSecurityGroups(&model3.ListSecurityGroupsRequest{})
	if err != nil {
		return nil, err
	}

	sgs := make([]*proto.SecurityGroup, 0)
	for _, v := range *rsp.SecurityGroups {
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

// calculateAvailableIPs takes a CIDR range and returns the number of available IP addresses.
func calculateAvailableIPs(cidr string) (int, error) {
	// Parse the CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CIDR: %w", err)
	}

	// Calculate the number of available IPs
	ones, _ := ipNet.Mask.Size()
	availableIPs := (1 << uint(32-ones)) - 2

	return availableIPs, nil
}
