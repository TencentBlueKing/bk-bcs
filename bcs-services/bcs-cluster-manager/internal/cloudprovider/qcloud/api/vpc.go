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
	"net"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NewVPCClient init VPC client
func NewVPCClient(opt *cloudprovider.CommonOption) (*VpcClient, error) {
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

	return &VpcClient{client: cli}, nil
}

// VpcClient xxx
type VpcClient struct {
	client *vpc.Client
}

// DescribeSecurityGroups describe security groups (https://cloud.tencent.com/document/api/215/15808)
func (v *VpcClient) DescribeSecurityGroups(securityGroupIds []string, filters []*Filter) (
	[]*SecurityGroup, error) {
	blog.Infof("DescribeSecurityGroups input: %s, %s", utils.ToJSONString(securityGroupIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeSecurityGroupsRequest()
	req.Limit = common.StringPtr(strconv.Itoa(limit))

	if len(securityGroupIds) > 0 {
		req.SecurityGroupIds = common.StringPtrs(securityGroupIds)
	}

	if len(filters) > 0 {
		req.Filters = make([]*vpc.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &vpc.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
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

// DescribeVpcs describe vpcs (https://cloud.tencent.com/document/api/215/15778)
// 参数不支持同时指定VpcIds和Filters，仅需要指定其中1个参数即可
func (v *VpcClient) DescribeVpcs(vpcIds []string, filters []*Filter) (
	[]*vpc.Vpc, error) {
	blog.Infof("DescribeVpcs input: %s, %s", utils.ToJSONString(vpcIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeVpcsRequest()
	req.Limit = common.StringPtr(strconv.Itoa(limit))

	if len(vpcIds) > 0 {
		req.VpcIds = common.StringPtrs(vpcIds)
	}

	if len(filters) > 0 {
		req.Filters = make([]*vpc.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &vpc.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
	}

	var (
		got, total = 0, 0
		first      = true
		vpcs       = make([]*vpc.Vpc, 0)
	)
	for got < total || first {
		first = false
		req.Offset = common.StringPtr(strconv.Itoa(got))
		resp, err := v.client.DescribeVpcs(req)
		if err != nil {
			blog.Errorf("DescribeVpcs failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeVpcs resp is nil")
			return nil, fmt.Errorf("DescribeVpcs resp is nil")
		}
		blog.Infof("DescribeVpcs success, requestID: %s", resp.Response.RequestId)

		vpcs = append(vpcs, resp.Response.VpcSet...)

		got += len(resp.Response.VpcSet)
		total = int(*resp.Response.TotalCount)
	}
	return vpcs, nil
}

// DescribeSubnets describe subnets (https://cloud.tencent.com/document/api/215/15784)
func (v *VpcClient) DescribeSubnets(subnetIds []string, filters []*Filter) (
	[]*vpc.Subnet, error) {
	blog.Infof("DescribeSubnets input: %s, %s", utils.ToJSONString(subnetIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeSubnetsRequest()
	req.Limit = common.StringPtr(strconv.Itoa(limit))

	if len(subnetIds) > 0 {
		req.SubnetIds = common.StringPtrs(subnetIds)
	}

	if len(filters) > 0 {
		req.Filters = make([]*vpc.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &vpc.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
	}

	got, total := 0, 0
	first := true
	subnets := make([]*vpc.Subnet, 0)
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
		// convertSubnet(resp.Response.SubnetSet)...
		subnets = append(subnets, resp.Response.SubnetSet...)
		got += len(resp.Response.SubnetSet)
		total = int(*resp.Response.TotalCount)
	}
	return subnets, nil
}

// DescribeBandwidthPackages describe 带宽包资源 (https://cloud.tencent.com/document/product/215/19209)
func (v *VpcClient) DescribeBandwidthPackages(bwpIds []string, filters []*Filter) (
	[]*vpc.BandwidthPackage, error) {
	blog.Infof("DescribeBandwidthPackages input: %s, %s", utils.ToJSONString(bwpIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeBandwidthPackagesRequest()
	req.Limit = common.Uint64Ptr(uint64(limit))

	if len(bwpIds) > 0 {
		req.BandwidthPackageIds = common.StringPtrs(bwpIds)
	}

	if len(filters) > 0 {
		req.Filters = make([]*vpc.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &vpc.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
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

// DescribeAssistantCidr describe assistant cidr (https://cloud.tencent.com/document/product/215/43274)
// 参数不支持同时指定VpcIds和Filters，仅需要指定其中1个参数即可。返回符合条件的Cidr数组
func (v *VpcClient) DescribeAssistantCidr(vpcIds []string, filters []*Filter) (
	[]*vpc.AssistantCidr, error) {
	blog.Infof("DescribeAssistantCidr input: %s, %s", utils.ToJSONString(vpcIds),
		utils.ToJSONString(filters))

	req := vpc.NewDescribeAssistantCidrRequest()
	req.Limit = common.Uint64Ptr(limit)

	if len(vpcIds) > 0 {
		req.VpcIds = common.StringPtrs(vpcIds)
	}

	if len(filters) > 0 {
		req.Filters = make([]*vpc.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &vpc.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
	}

	var (
		got, total uint64 = 0, 0
		first             = true
		cidrs             = make([]*vpc.AssistantCidr, 0)
	)
	for got < total || first {
		first = false
		req.Offset = common.Uint64Ptr(got)
		resp, err := v.client.DescribeAssistantCidr(req)
		if err != nil {
			blog.Errorf("DescribeAssistantCidr failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeAssistantCidr resp is nil")
			return nil, fmt.Errorf("DescribeAssistantCidr resp is nil")
		}
		blog.Infof("DescribeAssistantCidr success, requestID: %s", resp.Response.RequestId)

		cidrs = append(cidrs, resp.Response.AssistantCidrSet...)

		got += uint64(len(resp.Response.AssistantCidrSet))
		total = *resp.Response.TotalCount
	}

	return cidrs, nil
}

// CheckAssistantCidr 检测cidr冲突 (https://cloud.tencent.com/document/product/215/43277)
func (v *VpcClient) CheckAssistantCidr(vpcId string, news []string, olds []string) (
	[]*vpc.ConflictSource, error) {
	blog.Infof("CheckAssistantCidr input: %s, %s, %s", vpcId, utils.ToJSONString(news),
		utils.ToJSONString(olds))

	req := vpc.NewCheckAssistantCidrRequest()
	req.VpcId = common.StringPtr(vpcId)

	if len(news) > 0 {
		req.NewCidrBlocks = common.StringPtrs(news)
	}
	// req.OldCidrBlocks = common.StringPtrs(olds)

	resp, err := v.client.CheckAssistantCidr(req)
	if err != nil {
		blog.Errorf("CheckAssistantCidr failed, err: %s", err.Error())
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("CheckAssistantCidr resp is nil")
		return nil, fmt.Errorf("CheckAssistantCidr resp is nil")
	}
	blog.Infof("CheckAssistantCidr success, requestID: %s", *resp.Response.RequestId)

	fmt.Printf("CheckAssistantCidr success, requestID: %s\n", *resp.Response.RequestId)
	fmt.Printf("%+v\n", resp.Response.ConflictSourceSet)
	return resp.Response.ConflictSourceSet, nil
}

// CreateSubnet create subnet in vpc
func (v *VpcClient) CreateSubnet(vpcId, subnetName, zone string, subnet *net.IPNet, enableIPv6 bool) (*vpc.Subnet, error) {
	request := vpc.NewCreateSubnetRequest()
	request.VpcId = common.StringPtr(vpcId)
	request.SubnetName = common.StringPtr(subnetName)
	request.Zone = common.StringPtr(zone)
	request.CidrBlock = common.StringPtr(subnet.String())

	resp, err := v.client.CreateSubnet(request)
	if err != nil {
		blog.Errorf("CreateSubnet failed, err: %s", err.Error())
		return nil, err
	}

	if resp == nil || resp.Response == nil {
		return nil, fmt.Errorf("CreateSubnet resp is nil")
	}

	if enableIPv6 {
		if resp.Response.Subnet == nil || resp.Response.Subnet.SubnetId == nil {
			return nil, fmt.Errorf("CreateSubnet response subnet or subnet ID is nil")
		}

		ipv6Cidr, err := v.allocateIpv6SubnetCidr(vpcId)
		if err != nil {
			blog.Errorf("CreateSubnet allocateIpv6SubnetCidr failed, err: %s, rollback and delete subnet %s", err.Error(), *resp.Response.Subnet.SubnetId)
			_ = v.DeleteSubnet(*resp.Response.Subnet.SubnetId)
			return nil, err
		}

		blog.Infof("CreateSubnet assign IPv6 cidr[%s] for subnet[%s]", ipv6Cidr, *resp.Response.Subnet.SubnetId)
		assignReq := vpc.NewAssignIpv6SubnetCidrBlockRequest()
		assignReq.VpcId = common.StringPtr(vpcId)
		assignReq.Ipv6SubnetCidrBlocks = []*vpc.Ipv6SubnetCidrBlock{
			{
				SubnetId:      resp.Response.Subnet.SubnetId,
				Ipv6CidrBlock: common.StringPtr(ipv6Cidr),
			},
		}

		_, err = v.client.AssignIpv6SubnetCidrBlock(assignReq)
		if err != nil {
			blog.Errorf("AssignIpv6SubnetCidrBlock failed, err: %s, rollback and delete subnet %s", err.Error(), *resp.Response.Subnet.SubnetId)
			_ = v.DeleteSubnet(*resp.Response.Subnet.SubnetId)
			return nil, err
		}
	}

	return resp.Response.Subnet, nil
}

// DeleteSubnet delete subnet in vpc
func (v *VpcClient) DeleteSubnet(subnetId string) error {
	request := vpc.NewDeleteSubnetRequest()
	request.SubnetId = common.StringPtr(subnetId)

	_, err := v.client.DeleteSubnet(request)
	if err != nil {
		blog.Errorf("DeleteSubnet failed, err: %s", err.Error())
		return err
	}

	return nil
}

// DescribeNetworkAccountTypeRequest 查询用户网络类型
func (v *VpcClient) DescribeNetworkAccountTypeRequest() (string, error) {
	req := vpc.NewDescribeNetworkAccountTypeRequest()

	resp, err := v.client.DescribeNetworkAccountType(req)
	if err != nil {
		blog.Errorf("DescribeNetworkAccountType failed: %v", err)
		return "", err
	}

	return *resp.Response.NetworkAccountType, nil
}

// allocateIpv6SubnetCidr 从VPC的IPv6 CIDR中分配一个空闲的/64子网段
func (v *VpcClient) allocateIpv6SubnetCidr(vpcId string) (string, error) {
	// 查询VPC的IPv6 CIDR
	vpcs, err := v.DescribeVpcs([]string{vpcId}, nil)
	if err != nil {
		return "", fmt.Errorf("DescribeVpcs[%s] failed: %v", vpcId, err)
	}
	if len(vpcs) == 0 || vpcs[0].Ipv6CidrBlock == nil || *vpcs[0].Ipv6CidrBlock == "" {
		return "", fmt.Errorf("vpc[%s] has no IPv6 cidr block, please enable IPv6 for the VPC first", vpcId)
	}

	vpcIpv6Cidr := *vpcs[0].Ipv6CidrBlock
	_, vpcNet, err := net.ParseCIDR(vpcIpv6Cidr)
	if err != nil {
		return "", fmt.Errorf("parse vpc IPv6 cidr[%s] failed: %v", vpcIpv6Cidr, err)
	}

	// 查询VPC下已有子网已分配的IPv6段
	usedIpv6Cidrs := make([]*net.IPNet, 0)
	subnets, err := v.DescribeSubnets(nil, []*Filter{
		{Name: "vpc-id", Values: []string{vpcId}},
	})
	if err != nil {
		return "", fmt.Errorf("DescribeSubnets for vpc[%s] failed: %v", vpcId, err)
	}
	for _, sub := range subnets {
		if sub.Ipv6CidrBlock != nil && *sub.Ipv6CidrBlock != "" {
			_, ipNet, parseErr := net.ParseCIDR(*sub.Ipv6CidrBlock)
			if parseErr == nil {
				usedIpv6Cidrs = append(usedIpv6Cidrs, ipNet)
			}
		}
	}

	// 从VPC IPv6 CIDR中找一个空闲的/64子网段
	ipv6Cidr, err := findFreeIpv6Subnet(vpcNet, usedIpv6Cidrs)
	if err != nil {
		return "", fmt.Errorf("find free IPv6 subnet in vpc[%s] failed: %v", vpcId, err)
	}

	return ipv6Cidr, nil
}

// findFreeIpv6Subnet 在VPC IPv6 CIDR范围内寻找一个空闲的/64子网段
func findFreeIpv6Subnet(vpcNet *net.IPNet, used []*net.IPNet) (string, error) {
	if vpcNet.IP.To4() != nil {
		return "", fmt.Errorf("vpc network %s is not a valid IPv6 network", vpcNet.String())
	}

	// IPv6子网段固定为/64
	const subnetBits = 64
	vpcOnes, _ := vpcNet.Mask.Size()

	if vpcOnes > subnetBits {
		return "", fmt.Errorf("vpc IPv6 cidr mask /%d is smaller than subnet /%d", vpcOnes, subnetBits)
	}

	// 可用的子网数量 = 2^(64 - vpcMaskLen)
	subnetCount := uint64(1) << (subnetBits - vpcOnes)

	// Limit candidates check to prevent CPU exhaustion on very large VPC CIDR blocks
	const maxCandidates = 65536
	if subnetCount > maxCandidates {
		subnetCount = maxCandidates
	}

	// 将VPC起始IP转为uint64用于递增（取前8字节即IPv6地址的高64位）
	vpcIP := vpcNet.IP.To16()
	base := ipv6HighBits(vpcIP)

	for i := uint64(0); i < subnetCount; i++ {
		candidateIP := make(net.IP, 16)
		copy(candidateIP, vpcIP)
		setIpv6HighBits(candidateIP, base+i)

		candidateNet := &net.IPNet{
			IP:   candidateIP,
			Mask: net.CIDRMask(subnetBits, 128),
		}
		cidr := candidateNet.String()

		// 检查是否已被占用
		conflict := false
		for _, u := range used {
			if netsOverlap(candidateNet, u) {
				conflict = true
				break
			}
		}
		if !conflict {
			return cidr, nil
		}
	}

	return "", fmt.Errorf("no free IPv6 /%d subnet available in vpc cidr %s (checked %d candidates)", subnetBits, vpcNet.String(), subnetCount)
}

// ipv6HighBits 取IPv6地址高64位作为uint64
func ipv6HighBits(ip net.IP) uint64 {
	ip = ip.To16()
	return uint64(ip[0])<<56 | uint64(ip[1])<<48 | uint64(ip[2])<<40 | uint64(ip[3])<<32 |
		uint64(ip[4])<<24 | uint64(ip[5])<<16 | uint64(ip[6])<<8 | uint64(ip[7])
}

// setIpv6HighBits 设置IPv6地址高64位
func setIpv6HighBits(ip net.IP, val uint64) {
	ip[0] = byte(val >> 56)
	ip[1] = byte(val >> 48)
	ip[2] = byte(val >> 40)
	ip[3] = byte(val >> 32)
	ip[4] = byte(val >> 24)
	ip[5] = byte(val >> 16)
	ip[6] = byte(val >> 8)
	ip[7] = byte(val)
}

// netsOverlap 检查两个网络段是否有重叠
func netsOverlap(a, b *net.IPNet) bool {
	return a.Contains(b.IP) || b.Contains(a.IP)
}
