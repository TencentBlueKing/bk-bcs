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

package business

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetVpcCIDRBlocks 获取vpc所属的cidr段
func GetVpcCIDRBlocks(opt *cloudprovider.CommonOption, vpcId string, assistantType int) ([]*net.IPNet, error) {
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
		if assistantType < 0 && v.AssistantType != nil && v.CidrBlock != nil {
			cidrs = append(cidrs, *v.CidrBlock)
			continue
		}

		if v.AssistantType != nil && int(*v.AssistantType) == assistantType && v.CidrBlock != nil {
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

// GetFreeIPNets return free subnets
func GetFreeIPNets(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	allBlocks, err := GetVpcCIDRBlocks(opt, vpcId, 0)
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
	mask int, clusterId, subnetName string) (*cidrtree.Subnet, error) {
	frees, err := GetFreeIPNets(opt, vpcId)
	if err != nil {
		return nil, err
	}
	sub, err := cidrtree.AllocateFromFrees(mask, frees)
	if err != nil {
		return nil, err
	}

	if subnetName == "" {
		subnetName = fmt.Sprintf("bcs-subnet-%s-%s", clusterId, utils.RandomString(8))
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

// AllocateClusterVpcCniSubnets 集群分配所需的vpc-cni子网资源
func AllocateClusterVpcCniSubnets(ctx context.Context, clusterId, vpcId string,
	subnets []*proto.NewSubnet, opt *cloudprovider.CommonOption) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	subnetIDs := make([]string, 0)

	for i := range subnets {
		mask := 0
		if subnets[i].Mask > 0 {
			mask = int(subnets[i].Mask)
		} else if subnets[i].IpCnt > 0 {
			lenMask, err := utils.GetMaskLenByNum(utils.IPV4, float64(subnets[i].IpCnt))
			if err != nil {
				blog.Errorf("AllocateClusterVpcCniSubnets[%s] failed: %v", taskID, err)
				continue
			}

			mask = lenMask
		} else {
			mask = utils.DefaultMask
		}

		sub, err := AllocateSubnet(opt, vpcId, subnets[i].Zone, mask, clusterId, "")
		if err != nil {
			blog.Errorf("AllocateClusterVpcCniSubnets[%s] failed: %v", taskID, err)
			continue
		}

		blog.Infof("AllocateClusterVpcCniSubnets[%s] vpc[%s] zone[%s] subnet[%s]",
			taskID, vpcId, subnets[i].Zone, sub.ID)
		subnetIDs = append(subnetIDs, sub.ID)
		time.Sleep(time.Millisecond * 500)
	}

	blog.Infof("AllocateClusterVpcCniSubnets[%s] subnets[%v]", taskID, subnetIDs)
	return subnetIDs, nil
}

// CheckConflictFromVpc check cidr conflict in vpc cidrs
func CheckConflictFromVpc(opt *cloudprovider.CommonOption, vpcId, cidr string) ([]string, error) {
	ipNets, err := GetVpcCIDRBlocks(opt, vpcId, -1)
	if err != nil {
		return nil, err
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

// GetZoneAvailableSubnetsByVpc 获取vpc下某个地域下每个可用区的可用子网
func GetZoneAvailableSubnetsByVpc(opt *cloudprovider.CommonOption, vpcId string) (map[string]uint32, error) {
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

	var (
		availableSubnets = make(map[string]uint32, 0)
	)
	for i := range subnets {
		// subnet is available when default subnet available ipNum eq 10
		if *subnets[i].AvailableIPAddressCount >= 10 {
			availableSubnets[*subnets[i].Zone]++
		}
	}

	return availableSubnets, nil
}
