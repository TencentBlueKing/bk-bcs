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
	"strings"
	"sync"
	"time"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
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
		groups = append(groups, &proto.SecurityGroup{SecurityGroupName: *v.Name})
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
	opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
