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
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager("qcloud", &VPCClient{})
	})
}

// NewVPCClient init VPC client
func NewVPCClient(opt *cloudprovider.CommonOption) (*VPCClient, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)
	cpf := profile.NewClientProfile()

	cli, err := vpc.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &VPCClient{vpc: cli}, nil
}

// VPCClient is the client for VPC
type VPCClient struct {
	vpc *vpc.Client
}

// DescribeSecurityGroups describe security groups
// https://cloud.tencent.com/document/api/215/15808
func (c *VPCClient) DescribeSecurityGroups(securityGroupIds []string, filters []*Filter) (
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
		resp, err := c.vpc.DescribeSecurityGroups(req)
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

// DescribeSubnets describe subnets
// https://cloud.tencent.com/document/api/215/15784
func (c *VPCClient) DescribeSubnets(subnetIds []string, filters []*Filter) (
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
		resp, err := c.vpc.DescribeSubnets(req)
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

// ListVPCs list vpc
func (c *VPCClient) ListVPCs(vpcID string, opt *cloudprovider.CommonOption) ([]*proto.CloudVPC, error) {
	blog.Infof("ListSubnets input: vpcID/%s", vpcID)
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListSubnets list vpc subnets
func (c *VPCClient) ListSubnets(vpcID string, opt *cloudprovider.CommonOption) ([]*proto.Subnet, error) {
	blog.Infof("ListSubnets input: vpcID/%s", vpcID)
	vpcCli, err := NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*Filter, 0)
	filter = append(filter, &Filter{Name: "vpc-id", Values: []string{vpcID}})
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
func (c *VPCClient) ListSecurityGroups(opt *cloudprovider.CommonOption) ([]*proto.SecurityGroup, error) {
	vpcCli, err := NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	sgs, err := vpcCli.DescribeSecurityGroups(nil, nil)
	if err != nil {
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
