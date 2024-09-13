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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetVpcCIDRBlocks 获取vpc所属的cidr段(包括普通辅助cidr、容器辅助cidr)
func GetVpcCIDRBlocks(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	client, err := api.GetEc2Client(opt)
	if err != nil {
		return nil, fmt.Errorf("GetVpcCIDRBlocks GetEc2Client failed, %s", err.Error())
	}

	output, err := client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{vpcId}),
	})
	if err != nil {
		return nil, fmt.Errorf("GetVpcCIDRBlocks DescribeVpcs failed, %s", err.Error())
	}

	if len(output.Vpcs) == 0 {
		return nil, fmt.Errorf("GetVpcCIDRBlocks DescribeVpcs failed, vpc does not exist")
	}

	cidr := *output.Vpcs[0].CidrBlock
	_, netIP, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("GetVpcCIDRBlocks ParseCIDR failed, %s", err.Error())
	}

	return []*net.IPNet{netIP}, nil
}

// GetAllocatedSubnetsByVpc 获取vpc已分配的子网cidr段
func GetAllocatedSubnetsByVpc(opt *cloudprovider.CommonOption, vpcId string) ([]*net.IPNet, error) {
	client, err := api.GetEc2Client(opt)
	if err != nil {
		return nil, fmt.Errorf("GetAllocatedSubnetsByVpc GetEc2Client failed, %s", err.Error())
	}

	cloudSubnets := make([]*ec2.Subnet, 0)
	err = client.DescribeSubnetsPages(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{vpcId}),
			},
		},
	}, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		cloudSubnets = append(cloudSubnets, page.Subnets...)
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("GetAllocatedSubnetsByVpc DescribeSubnetsPages failed, %s", err.Error())
	}

	var ret []*net.IPNet
	for _, subnet := range cloudSubnets {
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
	// 获取vpc cidr blocks
	allBlocks, err := GetVpcCIDRBlocks(opt, vpcId)
	if err != nil {
		return nil, err
	}

	// 获取vpc 已使用子网列表
	allSubnets, err := GetAllocatedSubnetsByVpc(opt, vpcId)
	if err != nil {
		return nil, err
	}

	// 空闲IP列表
	return cidrtree.GetFreeIPNets(allBlocks, nil, allSubnets), nil
}

// AllocateSubnet allocate directrouter subnet
func AllocateSubnet(opt *cloudprovider.CommonOption, vpcId, zone string, mask int) (*cidrtree.Subnet, error) {
	frees, err := GetFreeIPNets(opt, vpcId)
	if err != nil {
		return nil, err
	}
	sub, err := cidrtree.AllocateFromFrees(mask, frees)
	if err != nil {
		return nil, err
	}

	// create vpc subnet
	vpcCli, err := api.GetEc2Client(opt)
	if err != nil {
		return nil, err
	}
	ret, err := vpcCli.CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: aws.String(zone),
		CidrBlock:        aws.String(sub.String()),
		VpcId:            aws.String(vpcId),
	})
	if err != nil {
		return nil, err
	}

	return subnetFromVpcSubnet(ret.Subnet), err
}

// subnetFromVpcSubnet trans vpc subnet to local subnet
func subnetFromVpcSubnet(info *ec2.Subnet) *cidrtree.Subnet {
	s := &cidrtree.Subnet{}
	if info == nil {
		return s
	}
	s.ID = *info.SubnetId
	if info.CidrBlock != nil {
		_, s.IPNet, _ = net.ParseCIDR(*info.CidrBlock)
	}
	s.VpcID = *info.VpcId
	s.Zone = *info.AvailabilityZone
	s.AvailableIps = uint64(*info.AvailableIpAddressCount)

	return s
}

// AllocateClusterVpcCniSubnets 集群分配所需的vpc-cni子网资源
func AllocateClusterVpcCniSubnets(ctx context.Context, vpcId string,
	subnets []*proto.NewSubnet, opt *cloudprovider.CommonOption) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	subnetIDs := make([]string, 0)

	for i := range subnets {
		mask := 0 // nolint
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

		sub, err := AllocateSubnet(opt, vpcId, subnets[i].Zone, mask)
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
	ipNets, err := GetVpcCIDRBlocks(opt, vpcId)
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
