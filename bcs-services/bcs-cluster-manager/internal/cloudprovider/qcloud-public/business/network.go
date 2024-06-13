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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// GrMaxClusterCidrNum globalRouter 模式下支持的最多cidr段
	GrMaxClusterCidrNum = 10
)

// GetVpcCIDRBlocks 获取vpc所属的cidr段(包括普通辅助cidr、容器辅助cidr)
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

	cidrs := make([]string, 0)
	if assistantType < 0 || assistantType == 0 {
		cidrs = append(cidrs, *vpcSet[0].CidrBlock)
	}

	for _, v := range vpcSet[0].AssistantCidrSet {
		// 获取所有辅助cidr, 不区分是 普通辅助cidr/容器辅助cidr
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
	// 获取vpc cidr blocks
	allBlocks, err := GetVpcCIDRBlocks(opt, vpcId, 0)
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
		if *subnets[i].AvailableIpAddressCount >= 10 {
			availableSubnets[*subnets[i].Zone]++
		}
	}

	return availableSubnets, nil
}

// DeleteDrSubnet delete directRouter subnet by subnetId
func DeleteDrSubnet(opt *cloudprovider.CommonOption, subnetId string) error {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		return err
	}
	return vpcCli.DeleteSubnet(subnetId)
}

// GetDrSubnetInfo get subnet info
func GetDrSubnetInfo(opt *cloudprovider.CommonOption, subnetId string) (*cidrtree.Subnet, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		return nil, err
	}

	subnets, err := vpcCli.DescribeSubnets([]string{subnetId}, nil)
	if err != nil {
		return nil, err
	}
	subnetInfo := subnetFromVpcSubnet(subnets[0])

	cvmCli, err := api.GetCVMClient(opt)
	if err != nil {
		return nil, err
	}
	zoneInfos, err := cvmCli.DescribeZones()
	if err != nil {
		return nil, err
	}
	for _, zoneInfo := range zoneInfos {
		if *zoneInfo.Zone == *subnets[0].Zone {
			subnetInfo.ZoneName = *zoneInfo.ZoneName
		}
	}
	return subnetInfo, nil
}

// ListSubnets list vpc subnets
func ListSubnets(opt *cloudprovider.CommonOption, vpcId string) ([]*cidrtree.Subnet, error) {
	var subnetInfos []*cidrtree.Subnet
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

	if err != nil {
		return subnetInfos, err
	}

	cvmCli, err := api.GetCVMClient(opt)
	if err != nil {
		return nil, err
	}
	zoneInfos, err := cvmCli.DescribeZones()
	if err != nil {
		return nil, err
	}

	zoneNameMap := make(map[string]string)
	for _, zoneInfo := range zoneInfos {
		zoneNameMap[*zoneInfo.Zone] = *zoneInfo.ZoneName
	}

	for _, subnet := range subnets {
		si := subnetFromVpcSubnet(subnet)
		si.ZoneName = zoneNameMap[si.Zone]
		subnetInfos = append(subnetInfos, si)
	}

	return subnetInfos, nil
}

// Global Router IP business handle logic

// AllocateGlobalRouterCidr allocates cidrs for the cluster
func AllocateGlobalRouterCidr(opt *cloudprovider.CommonOption, vpcId string, cluster *proto.Cluster,
	cidrLens []uint32, reservedBlocks []*net.IPNet) ([]string, error) {

	cidrNum := len(cidrLens)
	if cidrNum == 0 {
		return nil, fmt.Errorf("AllocateGlobalRouterCidr[%s:%s] cidrLens is empty",
			cluster.GetClusterID(), vpcId)
	}

	// get overlay cidrs
	cidrBlocks, err := cloudprovider.GetOverlayCidrBlocks(cluster.Provider, vpcId)
	if err != nil {
		return nil, err
	}

	// 获取已经分配的容器辅助cidr
	allocatedCidrBlocks, err := GetVpcCIDRBlocks(opt, vpcId, 1)
	if err != nil {
		return nil, err
	}

	if len(reservedBlocks) > 0 {
		allocatedCidrBlocks = append(allocatedCidrBlocks, reservedBlocks...)
	}

	// if clusterCidr is not empty, allocate cidr in the same cidr first
	if cluster != nil {
		clusterCidrIP, _, errLocal := net.ParseCIDR(cluster.NetworkSettings.ClusterIPv4CIDR)
		if errLocal != nil {
			return nil, errLocal
		}

		// 如果clusterVPC 被包含在了其中一个overlayIP中的话，那么最后将这个overlayIP网段放在最前面。
		for k, cidrBlock := range cidrBlocks {
			if cidrBlock.Contains(clusterCidrIP) {
				if k == 0 {
					break
				} else {
					cidrBlocks[0], cidrBlocks[k] = cidrBlocks[k], cidrBlocks[0]
				}
			}
		}
	}

	allocatableCidrBlock := make([]string, cidrNum)

	for i := 0; i < cidrNum; i++ {
		for k, cidrBlock := range cidrBlocks {
			man := cidrtree.NewCidrManager(cidrBlock, allocatedCidrBlocks)
			ipnet, errLocal := man.Allocate(int(cidrLens[i]))
			if errLocal == nil {
				allocatableCidrBlock[i] = ipnet.String()
				allocatedCidrBlocks = append(allocatedCidrBlocks, ipnet)
				break
			} else if errLocal == cidrtree.ErrNoEnoughFreeSubnet {
				if k == len(cidrBlocks) {
					return nil, cidrtree.ErrNoEnoughFreeSubnet
				}
				continue
			} else {
				return nil, errLocal
			}
		}
	}

	return allocatableCidrBlock, nil
}

// GetGrVPCIPSurplus returns free ip num
func GetGrVPCIPSurplus(opt *cloudprovider.CommonOption, cloudId, vpcId string,
	reservedBlocks []*net.IPNet) (uint32, error) {
	allBlocks, err := cloudprovider.GetOverlayCidrBlocks(cloudId, vpcId)
	if err != nil {
		return 0, err
	}
	allSubnets, err := GetVpcCIDRBlocks(opt, vpcId, 1)
	if err != nil {
		return 0, err
	}

	if len(reservedBlocks) > 0 {
		allSubnets = append(allSubnets, reservedBlocks...)
	}

	frees := cidrtree.GetFreeIPNets(allBlocks, nil, allSubnets)

	ipSurplus := uint32(0)
	for _, free := range frees {
		// 计算出当前ip下，可以有多少个子网
		freeIPNum, err := cidrtree.GetIPNum(free)
		if err != nil {
			return 0, err
		}
		ipSurplus += freeIPNum
	}

	return ipSurplus, nil
}

// GetClusterGrCidrs return all cidr info of the cluster by clusterId
func GetClusterGrCidrs(opt *cloudprovider.CommonOption, clusterId string) ([]*cidrtree.Cidr, error) {
	cluster, err := GetTkeCluster(opt, clusterId)
	if err != nil {
		return nil, err
	}

	return GetCidrsFromCluster(cluster)
}

// GetCidrsFromCluster return all cidr info of the cluster by cluster ptr
func GetCidrsFromCluster(cluster *tke.Cluster) ([]*cidrtree.Cidr, error) {
	cidrs := make([]*cidrtree.Cidr, 0)

	clusterCidr := *cluster.ClusterNetworkSettings.ClusterCIDR

	clsCidr, err := cidrtree.StringToCidr(clusterCidr)
	if err != nil {
		return nil, err
	}
	clsCidr.Type = utils.ClusterCIDR
	cidrs = append(cidrs, clsCidr)

	serviceCidr := *cluster.ClusterNetworkSettings.ServiceCIDR
	serCidr, err := cidrtree.StringToCidr(serviceCidr)
	if err != nil {
		return nil, err
	}
	serCidr.Type = utils.ServiceCIDR
	cidrs = append(cidrs, serCidr)

	clusterPropertyMap := make(map[string]interface{})
	// 将Cluster的property信息转换为map格式
	err = json.Unmarshal([]byte(*cluster.Property), &clusterPropertyMap)
	if err != nil {
		return nil, err
	}

	// 检查是否启用了MultiClusterCIDR模式，如果是，则也加入到cidrs中去。
	if _, found := clusterPropertyMap[utils.EnableMultiClusterCIDR]; found {
		multiClusterCIDR := clusterPropertyMap[utils.MultiClusterCIDR]
		multiClusterCIDRList := strings.Split(multiClusterCIDR.(string), ",")
		for _, c := range multiClusterCIDRList {
			cidr, err := cidrtree.StringToCidr(c)
			if err != nil {
				return nil, err
			}
			cidr.Type = utils.MultiClusterCIDR
			cidrs = append(cidrs, cidr)
		}
	}

	return cidrs, nil
}

// GetClusterGrIPSurplus return current available ip number for pod
func GetClusterGrIPSurplus(opt *cloudprovider.CommonOption, clusterId string) (uint32, error) {
	ipSurplus := uint32(0)
	cluster, err := GetTkeCluster(opt, clusterId)
	if err != nil {
		return 0, err
	}

	nodeNum := (uint32)(*cluster.ClusterNodeNum + *cluster.ClusterMaterNodeNum)
	maxNodePodNum := (uint32)(*cluster.ClusterNetworkSettings.MaxNodePodNum)
	maxClusterServiceNum := (uint32)(*cluster.ClusterNetworkSettings.MaxClusterServiceNum)

	cidrs, err := GetCidrsFromCluster(cluster)
	if err != nil {
		return 0, err
	}

	for _, cidr := range cidrs {
		// 如果cidr的type值是ClsuterCIDR或者MultiClusterCIDR其中之一
		if utils.StringInSlice(cidr.Type, []string{utils.ClusterCIDR, utils.MultiClusterCIDR}) {
			ipnum, err := cidr.GetIPNum()
			if err != nil {
				return 0, err
			}
			ipSurplus += ipnum
		}
	}

	if ipSurplus < (maxNodePodNum*nodeNum + maxClusterServiceNum) {
		ipSurplus = 0
	} else {
		ipSurplus = ipSurplus - maxNodePodNum*nodeNum - maxClusterServiceNum
	}

	return ipSurplus, nil
}

// AddClusterGrCidr add gr cidrs to the cluster
func AddClusterGrCidr(opt *cloudprovider.CommonOption, clusterId string, cidrs []string) error {
	cidrNum := len(cidrs)
	if cidrNum == 0 {
		return errors.New("cidrNum must be greater than 0")
	}

	clusterCidrs, err := GetClusterGrCidrs(opt, clusterId)
	if err != nil {
		return err
	}

	totalCidrNum := cidrNum
	for _, cidr := range clusterCidrs {
		if utils.StringInSlice(cidr.Type, []string{utils.ClusterCIDR, utils.MultiClusterCIDR}) {
			totalCidrNum++
		}
	}

	if totalCidrNum > GrMaxClusterCidrNum {
		return fmt.Errorf("total cidr number must be less than or equal to %d", GrMaxClusterCidrNum)
	}

	// 调用tke接口添加集群的cidr
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		return err
	}
	err = tkeCli.AddClusterCIDR(clusterId, cidrs, true)
	if err != nil {
		return err
	}
	return nil
}

// GetVpcGrFreeIPNets get vpc cidr free cidrs
func GetVpcGrFreeIPNets(opt *cloudprovider.CommonOption, cloudId, vpcId string,
	reservedBlocks []*net.IPNet) ([]*net.IPNet, error) {
	allBlocks, err := cloudprovider.GetOverlayCidrBlocks(cloudId, vpcId)
	if err != nil {
		return nil, err
	}
	// vpcID中所有的容器辅助cidr网段
	allSubnets, err := GetVpcCIDRBlocks(opt, vpcId, 1)
	if err != nil {
		return nil, err
	}

	// 缓存gr cidr
	cacheSubnets, err := getGrInCacheByVpc(vpcId)
	if err != nil {
		return nil, err
	}

	// 已使用cidr 和 缓存cidr 去重
	subnetMap := make(map[string]*net.IPNet)
	for _, subnet := range allSubnets {
		subnetMap[subnet.String()] = subnet
	}

	// 将不重复的缓存子网添加到allSubnets
	for _, subnet := range cacheSubnets {
		if _, exists := subnetMap[subnet.String()]; !exists {
			allSubnets = append(allSubnets, subnet)

			subnetMap[subnet.String()] = subnet
		}
	}

	return cidrtree.GetFreeIPNets(allBlocks, reservedBlocks, allSubnets), nil
}

// getGrInCacheByVpc get cache gr cidr
func getGrInCacheByVpc(vpcId string) ([]*net.IPNet, error) {
	var subs []string
	err := cloudprovider.GetEtcdModel().List(context.TODO(), vpcId, &subs)
	if err != nil {
		return nil, err
	}

	var subnets []*net.IPNet
	for _, v := range subs {
		_, sub, err := net.ParseCIDR(v)
		if err != nil {
			continue
		}
		subnets = append(subnets, sub)
	}
	return subnets, nil
}

// addGrCidrToCache add cidr to etcd cache
func addGrCidrToCache(sub *cidrtree.Subnet) error {
	key := path.Join(sub.VpcID, sub.ID)
	return cloudprovider.GetEtcdModel().Create(context.TODO(), key, sub.ID)
}

// AllocateGrCidrSubnet allocate gr cidr subnet
func AllocateGrCidrSubnet(opt *cloudprovider.CommonOption, cloudId, vpcId string,
	mask int, reservedBlocks []*net.IPNet) (*cidrtree.Subnet, error) {

	frees, err := GetVpcGrFreeIPNets(opt, cloudId, vpcId, reservedBlocks)
	if err != nil {
		return nil, err
	}
	blog.Infof("AllocateGrSubnet free %v", frees)

	sub, err := cidrtree.AllocateFromFrees(mask, frees)
	if err != nil {
		return nil, err
	}

	ret := &cidrtree.Subnet{
		ID:    sub.String(),
		IPNet: sub,
		VpcID: vpcId,
	}
	err = addGrCidrToCache(ret)
	if err != nil {
		blog.Errorf("AllocateGrSubnet addGrCidrToCache failed: %v", err)
	}

	return ret, err
}
