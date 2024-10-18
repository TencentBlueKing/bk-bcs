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

// Package business xxx
package business

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"net"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

// SubnetUsedIpCount 子网已使用IP数目统计
func SubnetUsedIpCount(ctx context.Context, opt *cloudprovider.ListNetworksOption, subnetID string) (uint32, error) {
	var usedIPCount uint32

	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return 0, fmt.Errorf("init AksService failed, %v", err)
	}

	interfaceList, err := client.ListNetworkNicAll(ctx)
	if err != nil {
		return 0, err
	}

	for _, nic := range interfaceList {
		if nic == nil || nic.Properties == nil {
			continue
		}

		for _, ipConfig := range nic.Properties.IPConfigurations {
			if ipConfig.Properties != nil && ipConfig.Properties.Subnet != nil &&
				*ipConfig.Properties.Subnet.ID == subnetID {
				usedIPCount++
			}
		}
	}

	return usedIPCount, nil
}

// GetVpcCIDRBlocks 获取vpc所属的cidr段(包括普通辅助cidr、容器辅助cidr)
func GetVpcCIDRBlocks(opt *cloudprovider.CommonOption, vpcId, resourceGroup string) ([]*net.IPNet, error) {
	vpcCli, err := api.NewAksServiceImplWithCommonOption(opt)
	if err != nil {
		return nil, err
	}

	vpcSet, err := vpcCli.GetVirtualNetworks(context.Background(), resourceGroup, vpcId)
	if err != nil {
		return nil, err
	}
	if vpcSet == nil {
		return nil, fmt.Errorf("GetVpcCIDRBlocks GetVirtualNetworks[%s] empty", vpcId)
	}

	cidrs := make([]string, 0)

	for _, v := range vpcSet.Properties.AddressSpace.AddressPrefixes {
		cidrs = append(cidrs, *v)
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
func GetAllocatedSubnetsByVpc(opt *cloudprovider.CommonOption, vpcId, resourceGroup string) ([]*net.IPNet, error) {
	vpcCli, err := api.NewAksServiceImplWithCommonOption(opt)
	if err != nil {
		return nil, err
	}

	subnets, err := vpcCli.ListSubnets(context.Background(), resourceGroup, vpcId)
	if err != nil {
		return nil, err
	}

	var ret []*net.IPNet
	for _, subnet := range subnets {
		if subnet.Properties != nil && subnet.Properties.AddressPrefix != nil {
			_, c, err := net.ParseCIDR(*subnet.Properties.AddressPrefix)
			if err != nil {
				return ret, err
			}
			ret = append(ret, c)
		}
	}
	return ret, nil
}

// GetFreeIPNets return free subnets
func GetFreeIPNets(opt *cloudprovider.CommonOption, vpcId, resourceGroup string) ([]*net.IPNet, error) {
	// 获取vpc cidr blocks
	allBlocks, err := GetVpcCIDRBlocks(opt, vpcId, resourceGroup)
	if err != nil {
		return nil, err
	}

	// 获取vpc 已使用子网列表
	allSubnets, err := GetAllocatedSubnetsByVpc(opt, vpcId, resourceGroup)
	if err != nil {
		return nil, err
	}

	// 空闲IP列表
	return cidrtree.GetFreeIPNets(allBlocks, nil, allSubnets), nil
}

// AllocateSubnet allocate directrouter subnet
func AllocateSubnet(opt *cloudprovider.CommonOption, vpcId, resourceGroup string,
	mask int, clusterId, subnetName string) (*cidrtree.Subnet, error) {
	frees, err := GetFreeIPNets(opt, vpcId, resourceGroup)
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
	vpcCli, err := api.NewAksServiceImplWithCommonOption(opt)
	if err != nil {
		return nil, err
	}

	subnet := armnetwork.Subnet{
		Name: to.Ptr(subnetName), // nolint
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix: to.Ptr(sub.String()), // nolint
		},
	}
	// 更新和创建subnet为同一个接口
	ret, err := vpcCli.UpdateSubnet(context.Background(), resourceGroup, vpcId, subnetName, subnet)
	if err != nil {
		return nil, err
	}

	return subnetFromVpcSubnet(opt, ret, vpcId, resourceGroup), err
}

// subnetFromVpcSubnet trans vpc subnet to local subnet
func subnetFromVpcSubnet(opt *cloudprovider.CommonOption, info *armnetwork.Subnet, vpcId,
	resourceGroupName string) *cidrtree.Subnet {
	s := &cidrtree.Subnet{}
	if info == nil {
		return s
	}
	s.ID = *info.ID
	s.Name = *info.Name
	if info.Properties != nil && info.Properties.AddressPrefix != nil {
		_, s.IPNet, _ = net.ParseCIDR(*info.Properties.AddressPrefix)
	}
	s.VpcID = vpcId

	netOpt := &cloudprovider.ListNetworksOption{
		CommonOption:      *opt,
		ResourceGroupName: resourceGroupName,
	}
	s.AvailableIps = func() uint64 {
		totalIPs, errLocal := utils.ConvertCIDRToStep(*info.Properties.AddressPrefix)
		if errLocal != nil {
			return 0
		}

		usedIpCnt, errLocal := SubnetUsedIpCount(context.Background(), netOpt, *info.ID)
		if errLocal != nil {
			return 0
		}

		return uint64(totalIPs - usedIpCnt - 5) // 减去5个系统保留的IP地址
	}()

	return s
}

// AllocateClusterVpcCniSubnets 集群分配所需的vpc-cni子网资源
func AllocateClusterVpcCniSubnets(ctx context.Context, clusterId, vpcId string,
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
func CheckConflictFromVpc(opt *cloudprovider.CommonOption, vpcId, cidr, resourceGroupName string) ([]string, error) {
	ipNets, err := GetVpcCIDRBlocks(opt, vpcId, resourceGroupName)
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
