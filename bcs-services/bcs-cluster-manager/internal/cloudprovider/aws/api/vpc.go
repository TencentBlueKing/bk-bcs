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

package api

import (
	"fmt"
	"strings"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager("aws", &VPCManager{})
	})
}

// VPCManager is the manager for VPC
type VPCManager struct{}

// CheckConflictInVpcCidr check cidr conflict
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string, opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListVpcs list vpcs
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	client, err := GetEc2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create aws client failed, err %s", err.Error())
	}

	input := &ec2.DescribeVpcsInput{}
	if vpcID != "" {
		input.VpcIds = []*string{&vpcID}
	}

	cloudVpcs, err := client.DescribeVpcs(input)
	if err != nil {
		return nil, err
	}

	vpcs := make([]*proto.CloudVpc, 0)
	for _, v := range cloudVpcs.Vpcs {
		vpcs = append(vpcs, &proto.CloudVpc{
			VpcId:    *v.VpcId,
			Name:     *v.VpcId,
			Ipv4Cidr: *v.CidrBlock,
		})
	}

	return vpcs, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID string, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	client, err := GetEc2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create aws client failed, err %s", err.Error())
	}

	output, err := client.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}

	result := make([]*proto.Subnet, 0)
	for _, v := range output.Subnets {
		subnet := &proto.Subnet{
			VpcID:                   aws.StringValue(v.VpcId),
			SubnetID:                aws.StringValue(v.SubnetId),
			SubnetName:              aws.StringValue(v.SubnetId),
			CidrRange:               aws.StringValue(v.CidrBlock),
			Zone:                    aws.StringValue(v.AvailabilityZone),
			AvailableIPAddressCount: uint64(aws.Int64Value(v.AvailableIpAddressCount)),
		}

		ipv6CidrBlocks := make([]string, 0)
		for _, y := range v.Ipv6CidrBlockAssociationSet {
			ipv6CidrBlocks = append(ipv6CidrBlocks, aws.StringValue(y.Ipv6CidrBlock))
		}

		if len(ipv6CidrBlocks) > 0 {
			subnet.Ipv6CidrRange = strings.Join(ipv6CidrBlocks, ",")
		}

		result = append(result, subnet)
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
