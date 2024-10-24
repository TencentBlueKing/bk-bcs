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

package azure

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
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
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListVpcs create AksService failed, %v", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	vn, err := client.ListVirtualNetwork(ctx, opt.ResourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("ListVpcs ListVirtualNetwork failed, err %s", err.Error())
	}

	result := make([]*proto.CloudVpc, 0)
	for _, v := range vn {
		if vpcID != "" && *v.Name != vpcID {
			continue
		}

		vpc := &proto.CloudVpc{
			Name:  *v.Name,
			VpcId: *v.Name,
		}
		if v.Properties != nil && v.Properties.AddressSpace != nil &&
			len(v.Properties.AddressSpace.AddressPrefixes) > 0 {
			if !strings.Contains(*v.Properties.AddressSpace.AddressPrefixes[0], ":") {
				vpc.Ipv4Cidr = *v.Properties.AddressSpace.AddressPrefixes[0]

				_, ipNet, err := net.ParseCIDR(vpc.Ipv4Cidr)
				ipNum, getIPErr := cidrtree.GetIPNum(ipNet)
				if getIPErr != nil {
					blog.Errorf("vpc GetIPNum failed: %v", err)
					continue
				}
				vpc.AllocateIpNum = ipNum
			} else {
				vpc.Ipv6Cidr = *v.Properties.AddressSpace.AddressPrefixes[0]
			}
		}
		result = append(result, vpc)
	}

	return result, nil
}

// ListSubnets list vpc subnets
func (vm *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListSubnets create AksService failed, %v", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	subnets, err := client.ListSubnets(ctx, opt.ResourceGroupName, vpcID)
	if err != nil {
		return nil, fmt.Errorf("ListSubnets failed, err %s", err.Error())
	}

	result := make([]*proto.Subnet, 0)
	for _, v := range subnets {
		var cidr string
		if v.Properties != nil && v.Properties.AddressPrefix != nil {
			cidr = *v.Properties.AddressPrefix
		}

		result = append(result, &proto.Subnet{
			VpcID:      vpcID,
			SubnetID:   *v.Name,
			SubnetName: *v.Name,
			CidrRange:  cidr,
			AvailableIPAddressCount: func() uint64 {
				totalIPs, errLocal := utils.ConvertCIDRToStep(cidr)
				if errLocal != nil {
					return 0
				}

				usedIpCnt, errLocal := business.SubnetUsedIpCount(context.Background(), opt, *v.ID)
				if errLocal != nil {
					return 0
				}

				return uint64(totalIPs - usedIpCnt - 5) // 减去5个系统保留的IP地址
			}(),
		})
	}
	return result, nil
}

// ListSecurityGroups list security groups
func (vm *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	cli, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("ListKeyPairs create aks client failed, %v", err)
	}

	result, err := cli.ListNetworkSecurityGroups(context.Background(), opt.ResourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("ListSSHPublicKeys failed, %v", err)
	}

	groups := make([]*proto.SecurityGroup, 0)
	for _, v := range result {
		if opt.Region != "" && opt.Region != *v.Location {
			continue
		}
		groups = append(groups, &proto.SecurityGroup{
			SecurityGroupName: *v.Name,
			SecurityGroupID:   *v.Name,
		})
	}

	return groups, nil
}

// GetCloudNetworkAccountType 查询用户网络类型
func (vm *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks list bandWidthPacks
func (vm *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckConflictInVpcCidr check cidr if conflict with vpc cidrs
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CheckConflictInVpcCidrOption) ([]string, error) {
	return business.CheckConflictFromVpc(&opt.CommonOption, vpcID, cidr, opt.ResourceGroupName)
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

// GetVpcIpUsage get vpc ipTotal/ipSurplus
func (vm *VPCManager) GetVpcIpUsage(
	vpcId string, ipType string, reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) (uint32, uint32, error) {
	return 0, 0, nil
}

// GetClusterIpUsage get cluster ip usage
func (vm *VPCManager) GetClusterIpUsage(clusterId string, ipType string, opt *cloudprovider.CommonOption) (
	uint32, uint32, error) {
	return 0, 0, nil
}
