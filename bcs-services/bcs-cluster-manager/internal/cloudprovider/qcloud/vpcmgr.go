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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
func (c *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*api.Filter, 0)
	if vpcID != "" {
		filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{vpcID}})
	}

	vpcs, err := vpcCli.DescribeVpcs(nil, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.CloudVpc, 0)
	for _, v := range vpcs {
		result = append(result, &proto.CloudVpc{
			Name:     utils.StringPtrToString(v.VpcName),
			VpcId:    utils.StringPtrToString(v.VpcId),
			Ipv4Cidr: utils.StringPtrToString(v.CidrBlock),
			Ipv6Cidr: utils.StringPtrToString(v.Ipv6CidrBlock),
			Cidrs: func() []*proto.AssistantCidr {
				cidrs := make([]*proto.AssistantCidr, 0)

				for _, c := range v.AssistantCidrSet {
					cidrs = append(cidrs, &proto.AssistantCidr{
						Cidr:     utils.StringPtrToString(c.CidrBlock),
						CidrType: int32(utils.Int64PtrToInt64(c.AssistantType)),
					})
				}

				return cidrs
			}(),
		})
	}
	return result, nil
}

// ListSubnets list vpc subnets
func (c *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	blog.Infof("ListSubnets input: vpcID/%s", vpcID)
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*api.Filter, 0)
	filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{vpcID}})
	if len(zone) > 0 {
		filter = append(filter, &api.Filter{Name: "zone", Values: []string{zone}})
	}

	subnets, err := vpcCli.DescribeSubnets(nil, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.Subnet, 0)
	for _, v := range subnets {
		result = append(result, &proto.Subnet{
			VpcID:                   *v.VpcID,
			SubnetID:                *v.SubnetID,
			SubnetName:              *v.SubnetName,
			CidrRange:               *v.CidrBlock,
			Ipv6CidrRange:           *v.Ipv6CidrBlock,
			Zone:                    *v.Zone,
			AvailableIPAddressCount: *v.AvailableIPAddressCount,
		})
	}
	return result, nil
}

// ListSecurityGroups list security groups
func (c *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	sgs, err := vpcCli.DescribeSecurityGroups(nil, nil)
	if err != nil {
		blog.Errorf("ListSecurityGroups DescribeSecurityGroups failed: %v", err)
		return nil, err
	}

	result := make([]*proto.SecurityGroup, 0)
	for _, v := range sgs {
		result = append(result, &proto.SecurityGroup{
			SecurityGroupID:   v.ID,
			SecurityGroupName: v.Name,
			Description:       v.Desc,
		})
	}

	return result, nil
}

// GetCloudNetworkAccountType 查询用户网络类型
func (c *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	accountType, err := vpcCli.DescribeNetworkAccountTypeRequest()
	if err != nil {
		blog.Errorf("DescribeNetworkAccountType failed: %v", err)
		return nil, err
	}

	return &proto.CloudAccountType{Type: accountType}, nil
}

// ListBandwidthPacks packs
func (c *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	bwps, err := vpcCli.DescribeBandwidthPackages(nil, nil)
	if err != nil {
		blog.Errorf("ListBandwidthPacks describeBandwidthPackages failed: %v", err)
		return nil, err
	}

	result := make([]*proto.BandwidthPackageInfo, 0)
	for _, v := range bwps {
		result = append(result, &proto.BandwidthPackageInfo{
			Id:          *v.BandwidthPackageId,
			Name:        *v.BandwidthPackageName,
			NetworkType: *v.NetworkType,
			Status:      *v.Status,
			Bandwidth: func() int32 {
				if v != nil && v.Bandwidth != nil {
					return int32(*v.Bandwidth)
				}
				return 0
			}(),
		})
	}

	return result, nil
}

// CheckConflictInVpcCidr check cidr if conflict with vpc cidrs
func (c *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
