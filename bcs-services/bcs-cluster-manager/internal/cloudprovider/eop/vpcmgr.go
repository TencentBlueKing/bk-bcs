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

package eop

import (
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/eop/api"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager(cloudName, &VPCManager{})
	})
}

// VPCManager is the client for VPC
type VPCManager struct {
}

// ListVpcs list vpcs
func (vm *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*cmproto.CloudVpc, error) {
	cli, err := api.NewCTClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("ListVpcs create eck client when failed: %v", err)
		return nil, err
	}

	result, err := cli.ListVpcs(opt.Region)
	if err != nil {
		blog.Errorf("ListVpcs failed: %v", err)
		return nil, err
	}

	vpcs := make([]*cmproto.CloudVpc, 0)

	if vpcID != "" {
		for _, v := range result {
			if v.Name == vpcID {
				vpcs = append(vpcs, &cmproto.CloudVpc{
					VpcId:    strconv.Itoa(int(v.VpcId)),
					Name:     v.Name,
					Ipv4Cidr: v.CidrBlock,
				})
				break
			}
		}
		return vpcs, nil
	}

	for _, v := range result {
		vpcs = append(vpcs, &cmproto.CloudVpc{
			VpcId:    strconv.Itoa(int(v.VpcId)),
			Name:     v.Name,
			Ipv4Cidr: v.CidrBlock,
		})
	}

	return vpcs, nil
}

// ListSubnets get vpc subnets
func (vm *VPCManager) ListSubnets(vpcID string, zone string,
	opt *cloudprovider.ListNetworksOption) ([]*cmproto.Subnet, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCloudNetworkAccountType get accoount type
func (vm *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*cmproto.CloudAccountType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListBandwidthPacks get bandwidth packs
func (vm *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*cmproto.BandwidthPackageInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckConflictInVpcCidr check cidr conflict in vpc
func (vm *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListSecurityGroups list security groups
func (vm *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*cmproto.SecurityGroup, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
