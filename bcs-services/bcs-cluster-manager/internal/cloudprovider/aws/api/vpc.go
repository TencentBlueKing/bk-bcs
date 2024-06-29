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
	"net"
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
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string, opt *cloudprovider.CommonOption) (
	[]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListVpcs list vpcs
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	client, err := GetEc2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListVpcs GetEc2Client failed, err %s", err.Error())
	}

	filters := []*ec2.Filter{{Name: aws.String("state"), Values: aws.StringSlice([]string{"available"})}}
	if vpcID != "" {
		filters = append(filters, &ec2.Filter{Name: aws.String("vpc-id"), Values: aws.StringSlice([]string{vpcID})})
	}

	vpcs := make([]*ec2.Vpc, 0)
	err = client.DescribeVpcsPages(&ec2.DescribeVpcsInput{Filters: filters},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return !lastPage
		})
	if err != nil {
		return nil, fmt.Errorf("ListVpcs DescribeVpcsPages failed, err %s", err.Error())
	}

	results := make([]*proto.CloudVpc, 0)
	for _, v := range vpcs {
		results = append(results, &proto.CloudVpc{
			VpcId: *v.VpcId,
			Ipv4Cidr: func(v *ec2.Vpc) string {
				if v.CidrBlockAssociationSet != nil {
					return *v.CidrBlockAssociationSet[0].CidrBlock
				}
				return ""
			}(v),
			Ipv6Cidr: func() string {
				if v.Ipv6CidrBlockAssociationSet != nil {
					return *v.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock
				}
				return ""
			}(),
		})
	}

	return results, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID string, zone string, opt *cloudprovider.ListNetworksOption) (
	[]*proto.Subnet, error) {
	client, err := GetEc2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListSubnets GetEc2Client failed, err %s", err.Error())
	}

	cloudSubnets := make([]*ec2.Subnet, 0)
	err = client.DescribeSubnetsPages(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{vpcID}),
			},
			{
				Name:   aws.String("state"),
				Values: aws.StringSlice([]string{"available"}),
			},
		},
	}, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		cloudSubnets = append(cloudSubnets, page.Subnets...)
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("ListSubnets DescribeSubnetsPages failed, err %s", err.Error())
	}

	result := make([]*proto.Subnet, 0)
	for _, v := range cloudSubnets {
		subnet := &proto.Subnet{
			VpcID:      aws.StringValue(v.VpcId),
			SubnetID:   aws.StringValue(v.SubnetId),
			SubnetName: aws.StringValue(v.SubnetId),
			CidrRange:  aws.StringValue(v.CidrBlock),
			Zone:       aws.StringValue(v.AvailabilityZone),
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
	client, err := GetEc2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListSecurityGroups GetEc2Client failed, err %s", err.Error())
	}

	cloudSgs := make([]*ec2.SecurityGroup, 0)
	err = client.DescribeSecurityGroupsPages(&ec2.DescribeSecurityGroupsInput{},
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			cloudSgs = append(cloudSgs, page.SecurityGroups...)
			return !lastPage
		})
	if err != nil {
		return nil, fmt.Errorf("ListSecurityGroups DescribeSecurityGroupsPages failed, err %s", err.Error())
	}

	result := make([]*proto.SecurityGroup, 0)
	for _, s := range cloudSgs {
		result = append(result, &proto.SecurityGroup{
			SecurityGroupName: *s.GroupName,
			SecurityGroupID:   *s.GroupId,
		})
	}

	return result, nil
}

// GetCloudNetworkAccountType 查询用户网络类型
func (vm *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks list bandWidthPacks
func (vm *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
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

// GetVpcIpSurplus get vpc ipSurplus
func (vm *VPCManager) GetVpcIpSurplus(
	vpcId string, ipType string, reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) (uint32, error) {
	return 0, nil
}

// GetOverlayClusterIPSurplus get cluster overlay ipSurplus
func (vm *VPCManager) GetOverlayClusterIPSurplus(clusterId string, opt *cloudprovider.CommonOption) (uint32, error) {
	return 0, nil
}
