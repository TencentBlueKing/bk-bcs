/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package business

import (
	"fmt"
	"net"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// GetVpcCIDRBlocks 获取vpc所属的cidr段
func GetVpcCIDRBlocks(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		return nil, err
	}

	vpcSet, err := vpcCli.DescribeVpcs([]string{vpcId}, nil)
	if err != nil {
		return nil, err
	}
	if len(vpcSet) == 0 {
		return nil, fmt.Errorf("GetVpcCIDRBlocks DescribeVpcs[%s] empty", vpcId)
	}

	cidrs := []string{*vpcSet[0].CidrBlock}
	for _, v := range vpcSet[0].AssistantCidrSet {
		if v.AssistantType != nil && *v.AssistantType == 0 && v.CidrBlock != nil {
			cidrs = append(cidrs, *v.CidrBlock)
		}
	}

	var ret []*net.IPNet
	for _, v := range cidrs {
		_, c, err := net.ParseCIDR(v)
		if err != nil {
			return ret, err
		}
		ret = append(ret, c)
	}
	return ret, nil

}

// GetAllocatedSubnetsByVpc 获取vpc已分配的子网cidr段
func GetAllocatedSubnetsByVpc(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		return nil, err
	}

	filter := make([]*api.Filter, 0)
	filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{vpcId}})
	subnets, err := vpcCli.DescribeSubnets(nil, filter)
	if err != nil {
		return nil, err
	}

	var ret []*net.IPNet
	for _, subnet := range subnets {
		if subnet.CidrBlock != nil {
			_, c, err := net.ParseCIDR(*subnet.CidrBlock)
			if err != nil {
				return ret, err
			}
			ret = append(ret, c)
		}
	}
	return ret, nil
}

// GetFreeIPNets return free globalrouter subnets
func GetFreeIPNets(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	allBlocks, err := GetVpcCIDRBlocks(opt, vpcId)
	if err != nil {
		return nil, err
	}

	allSubnets, err := GetAllocatedSubnetsByVpc(opt, vpcId)
	if err != nil {
		return nil, err
	}

	return cidrtree.GetFreeIPNets(allBlocks, allSubnets), nil
}

// AllocateSubnet allocate directrouter subnet
func AllocateSubnet(opt *cloudprovider.CommonOption, vpcId, zone string,
	mask int, subnetName string) (*cidrtree.Subnet, error) {
	frees, err := GetFreeIPNets(opt, vpcId)
	if err != nil {
		return nil, err
	}
	sub, err := cidrtree.AllocateFromFrees(mask, frees)
	if err != nil {
		return nil, err
	}

	if subnetName == "" {
		subnetName = "bcs-subnet-" + utils.RandomString(8)
	}

	// create vpc subnet
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		return nil, err
	}
	ret, err := vpcCli.CreateSubnet(vpcId, subnetName, zone, sub)
	if err != nil {
		return nil, err
	}

	return subnetFromVpcSubnet(ret), err
}

// subnetFromVpcSubnet trans vpc subnet to local subnet
func subnetFromVpcSubnet(info *vpc.Subnet) (n *cidrtree.Subnet) {
	s := &cidrtree.Subnet{}
	if info == nil {
		return s
	}
	s.ID = *info.SubnetId
	if info.CidrBlock != nil {
		_, s.IPNet, _ = net.ParseCIDR(*info.CidrBlock)
	}
	s.Name = *info.SubnetName
	s.VpcID = *info.VpcId
	s.Zone = *info.Zone
	s.CreatedTime = *info.CreatedTime
	s.AvaliableIP = *info.AvailableIpAddressCount
	return s
}
