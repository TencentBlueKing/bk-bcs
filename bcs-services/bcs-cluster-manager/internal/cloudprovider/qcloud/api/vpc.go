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
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager("qcloud", &VPCManager{})
	})
}

// newVPCClient init VPC client
func newVPCClient(opt *cloudprovider.CommonOption) (*vpcClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)
	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.VpcDomain
	}

	cli, err := vpc.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &vpcClient{client: cli}, nil
}

type vpcClient struct {
	client *vpc.Client
}

// describeSecurityGroups describe security groups (https://cloud.tencent.com/document/api/215/15808)
func (v *vpcClient) describeSecurityGroups(securityGroupIds []string, filters []*Filter) (
	[]*SecurityGroup, error) {
	blog.Infof("DescribeSecurityGroups input: %s, %s", utils.ToJSONString(securityGroupIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeSecurityGroupsRequest()
	if securityGroupIds != nil {
		req.SecurityGroupIds = common.StringPtrs(securityGroupIds)
	}
	req.Limit = common.StringPtr(strconv.Itoa(limit))
	req.Filters = make([]*vpc.Filter, 0)
	for _, v := range filters {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}
	got, total := 0, 0
	first := true
	sg := make([]*SecurityGroup, 0)
	for got < total || first {
		first = false
		req.Offset = common.StringPtr(strconv.Itoa(got))
		resp, err := v.client.DescribeSecurityGroups(req)
		if err != nil {
			blog.Errorf("DescribeSecurityGroups failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeSecurityGroups resp is nil")
			return nil, fmt.Errorf("DescribeSecurityGroups resp is nil")
		}
		blog.Infof("DescribeSecurityGroups success, requestID: %s", resp.Response.RequestId)
		for _, v := range resp.Response.SecurityGroupSet {
			sg = append(sg, &SecurityGroup{ID: *v.SecurityGroupId, Name: *v.SecurityGroupName, Desc: *v.SecurityGroupDesc})
		}
		got += len(resp.Response.SecurityGroupSet)
		total = int(*resp.Response.TotalCount)
	}
	return sg, nil
}

// describeSubnets describe subnets (https://cloud.tencent.com/document/api/215/15784)
func (v *vpcClient) describeSubnets(subnetIds []string, filters []*Filter) (
	[]*Subnet, error) {
	blog.Infof("DescribeSubnets input: %s, %s", utils.ToJSONString(subnetIds),
		utils.ToJSONString(filters))
	req := vpc.NewDescribeSubnetsRequest()
	req.SubnetIds = common.StringPtrs(subnetIds)
	req.Limit = common.StringPtr(strconv.Itoa(limit))
	req.Filters = make([]*vpc.Filter, 0)
	for _, v := range filters {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}
	got, total := 0, 0
	first := true
	subnets := make([]*Subnet, 0)
	for got < total || first {
		first = false
		req.Offset = common.StringPtr(strconv.Itoa(got))
		resp, err := v.client.DescribeSubnets(req)
		if err != nil {
			blog.Errorf("DescribeSubnets failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeSubnets resp is nil")
			return nil, fmt.Errorf("DescribeSubnets resp is nil")
		}
		blog.Infof("DescribeSubnets success, requestID: %s", resp.Response.RequestId)
		subnets = append(subnets, convertSubnet(resp.Response.SubnetSet)...)
		got += len(resp.Response.SubnetSet)
		total = int(*resp.Response.TotalCount)
	}
	return subnets, nil
}

// describeBandwidthPackages describe 带宽包资源 (https://cloud.tencent.com/document/product/215/19209)
func (v *vpcClient) describeBandwidthPackages(bwpIds []string, filters []*Filter) (
	[]*vpc.BandwidthPackage, error) {
	blog.Infof("DescribeBandwidthPackages input: %s, %s", utils.ToJSONString(bwpIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeBandwidthPackagesRequest()
	req.BandwidthPackageIds = common.StringPtrs(bwpIds)
	req.Limit = common.Uint64Ptr(uint64(limit))

	req.Filters = make([]*vpc.Filter, 0)
	for _, v := range filters {
		req.Filters = append(req.Filters, &vpc.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}

	var (
		got, total = 0, 0
		first      = true
		bwps       = make([]*vpc.BandwidthPackage, 0)
	)

	for got < total || first {
		first = false
		req.Offset = common.Uint64Ptr(uint64(got))

		resp, err := v.client.DescribeBandwidthPackages(req)
		if err != nil {
			blog.Errorf("DescribeBandwidthPackages failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeBandwidthPackages resp is nil")
			return nil, fmt.Errorf("DescribeBandwidthPackages resp is nil")
		}
		blog.Infof("DescribeBandwidthPackages success, requestID: %s", resp.Response.RequestId)

		bwps = append(bwps, resp.Response.BandwidthPackageSet...)
		got += len(resp.Response.BandwidthPackageSet)

		total = int(*resp.Response.TotalCount)
	}

	return bwps, nil
}

// describeNetworkAccountTypeRequest 查询用户网络类型
func (v *vpcClient) describeNetworkAccountTypeRequest() (string, error) {
	req := vpc.NewDescribeNetworkAccountTypeRequest()

	resp, err := v.client.DescribeNetworkAccountType(req)
	if err != nil {
		blog.Errorf("DescribeNetworkAccountType failed: %v", err)
		return "", err
	}

	return *resp.Response.NetworkAccountType, nil
}

// VPCManager is the manager for VPC
type VPCManager struct{}

// ListSubnets list vpc subnets
func (c *VPCManager) ListSubnets(vpcID string, opt *cloudprovider.CommonOption) ([]*proto.Subnet, error) {
	blog.Infof("ListSubnets input: vpcID/%s", vpcID)
	vpcCli, err := newVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*Filter, 0)
	filter = append(filter, &Filter{Name: "vpc-id", Values: []string{vpcID}})
	subnets, err := vpcCli.describeSubnets(nil, filter)
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
func (c *VPCManager) ListSecurityGroups(opt *cloudprovider.CommonOption) ([]*proto.SecurityGroup, error) {
	vpcCli, err := newVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	sgs, err := vpcCli.describeSecurityGroups(nil, nil)
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
	vpcCli, err := newVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	accountType, err := vpcCli.describeNetworkAccountTypeRequest()
	if err != nil {
		blog.Errorf("DescribeNetworkAccountType failed: %v", err)
		return nil, err
	}

	return &proto.CloudAccountType{Type: accountType}, nil
}

// ListBandwidthPacks packs
func (c *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	vpcCli, err := newVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	bwps, err := vpcCli.describeBandwidthPackages(nil, nil)
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
