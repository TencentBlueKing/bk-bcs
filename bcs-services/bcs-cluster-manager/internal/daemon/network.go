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

// Package daemon xxx
package daemon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	autoAllocateTcClusterCidrLockKey = "autoAllocateTcClusterCidr"
)

// autoAllocateTcClusterCidr 自动分配 tencentCloud自研云集群vpc-cni模式子网资源
func (d *Daemon) autoAllocateTcClusterCidr(error chan<- error) {
	d.lock.Lock(autoAllocateTcClusterCidrLockKey, []lock.LockOption{lock.LockTTL(time.Second * 10)}...) // nolint
	defer d.lock.Unlock(autoAllocateTcClusterCidrLockKey)                                               // nolint

	defer utils.RecoverPrintStack("autoAllocateTcClusterCidr")

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"status":   common.StatusRunning,
		"provider": tencentCloud,
	})

	clusterList, err := d.model.ListCluster(d.ctx, cond, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("autoAllocateTcClusterCidr ListCluster failed: %v", err)
		error <- err
		return
	}

	// 聚合同地域同vpc集群列表
	regionVpcClusters := make(map[string][]*cmproto.Cluster, 0)

	for i := range clusterList {
		err = checkClusterAutoScaleCidrValidate(clusterList[i])
		if err != nil {
			blog.Error(err)
			continue
		}

		regionVpc := fmt.Sprintf("%s:%s", clusterList[i].Region, clusterList[i].GetVpcID())

		if regionVpcClusters[regionVpc] == nil {
			regionVpcClusters[regionVpc] = make([]*cmproto.Cluster, 0)
		}

		regionVpcClusters[regionVpc] = append(regionVpcClusters[regionVpc], clusterList[i])
	}

	concurency := utils.NewRoutinePool(20)
	defer concurency.Close()

	for _, clusters := range regionVpcClusters {
		concurency.Add(1)
		go func(clusters []*cmproto.Cluster) {
			defer concurency.Done()
			for i := range clusters {
				errLocal := checkClusterAutoScaleCidrValidate(clusters[i])
				if errLocal != nil {
					blog.Errorf("autoAllocateTcClusterCidr checkClusterAutoScaleCidrValidate[%s:%s:%s] failed: %v",
						clusters[i].GetRegion(), clusters[i].GetVpcID(), clusters[i].GetClusterID(), errLocal)
					continue
				}

				errLocal = allocateSubnetsToCluster(d.ctx, d.model, clusters[i])
				if errLocal != nil {
					blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] failed: %v",
						clusters[i].GetRegion(), clusters[i].GetVpcID(), clusters[i].GetClusterID(), errLocal)
					continue
				}

				blog.Infof("autoAllocateTcClusterCidr[%s:%s:%s] successful",
					clusters[i].GetRegion(), clusters[i].GetVpcID(), clusters[i].GetClusterID())
			}

		}(clusters)
	}
	concurency.Wait()
}

func checkClusterAutoScaleCidrValidate(cluster *cmproto.Cluster) error {
	errStr := "autoAllocateTcClusterCidr checkClusterAutoScaleCidrValidate"
	if cluster.GetRegion() == "" || cluster.GetVpcID() == "" {
		return fmt.Errorf("%s cluster[%s] region or vpc empty", errStr, cluster.GetClusterID())
	}

	if !cluster.GetNetworkSettings().GetEnableVPCCni() {
		return fmt.Errorf("%s cluster[%s] platform not enable vpc-cni", errStr, cluster.GetClusterID())
	}

	if cluster.GetNetworkSettings().GetStatus() == common.StatusInitialization {
		return fmt.Errorf("%s cluster[%s] doing enable/disable vpc-cni", errStr, cluster.GetClusterID())
	}

	if cluster.GetNetworkSettings().GetStatus() == common.TaskStatusFailure {
		return fmt.Errorf("%s cluster[%s] enable/disable vpc-cni failure", errStr, cluster.GetClusterID())
	}

	if cluster.GetNetworkSettings().GetSubnetSource() == nil ||
		len(cluster.GetNetworkSettings().GetSubnetSource().GetNew()) == 0 {
		return fmt.Errorf("%s cluster[%s] need to allocate resource empty", errStr, cluster.GetClusterID())
	}

	return nil
}

func checkClusterNeedToScaleSubnet(cls *cmproto.Cluster) (map[string]uint64, []string, bool, error) {
	needAllocateSubnets, curSubnetIds, err := getClusterNeedAllocateSubnets(cls)
	if err != nil {
		return nil, nil, false, err
	}

	needScale := false
	for zone := range needAllocateSubnets {
		if needAllocateSubnets[zone] > 0 {
			needScale = true
			break
		}
	}
	if !needScale {
		return needAllocateSubnets, curSubnetIds, false, nil
	}

	return needAllocateSubnets, curSubnetIds, true, nil
}

func getClusterNeedAllocateSubnets(cls *cmproto.Cluster) (map[string]uint64, []string, error) {
	if len(cls.GetNetworkSettings().GetSubnetSource().GetNew()) == 0 {
		return nil, nil, fmt.Errorf("autoAllocateTcClusterCidr getClusterNeedAllocateSubnets[%s:%s:%s] "+
			"subnetSource empty", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID())
	}
	clsSubnets := cls.GetNetworkSettings().GetSubnetSource().GetNew()

	// 获取集群每个可用区子网数容量
	needSubnets := make(map[string]uint64, 0)
	for i := range clsSubnets {
		needSubnets[clsSubnets[i].GetZone()] += uint64(clsSubnets[i].GetIpCnt())
	}

	// 获取集群当前子网数目
	zoneCurSubnetInfo, _, subnetIds, err := business.GetClusterCurrentVpcCniSubnets(cls, true)
	if err != nil {
		return nil, nil, err
	}

	blog.Infof("autoAllocateTcClusterCidr getClusterNeedAllocateSubnets[%s:%s:%s] needSubnets %+v",
		cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), needSubnets)

	for zone := range zoneCurSubnetInfo {
		blog.Infof("autoAllocateTcClusterCidr getClusterNeedAllocateSubnets[%s:%s:%s] curSubnets %v %s %v %v",
			cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), needSubnets, zone,
			zoneCurSubnetInfo[zone].AvailableIps, zoneCurSubnetInfo[zone].TotalIps)
	}

	// 计算每个可用区还需要多少IP数目
	needAllocateSubnetNum := make(map[string]uint64, 0)
	for zone, num := range needSubnets {
		data, ok := zoneCurSubnetInfo[zone]
		if ok {
			if num > data.TotalIps {
				needAllocateSubnetNum[zone] = num - data.TotalIps
			}
		} else {
			needAllocateSubnetNum[zone] = num
		}
	}

	return needAllocateSubnetNum, subnetIds, nil
}

// getClusterAllocatedEmptySubnets 获取集群已分配未使用的子网列表(通过step分类)
func getClusterAllocatedEmptySubnets(cls *cmproto.Cluster, subnetIds []string) (
	map[string]uint64, map[string][]string, error) {
	cmOption, err := cloudprovider.GetCloudCmOptionByCluster(cls)
	if err != nil {
		return nil, nil, err
	}

	subnets, err := business.ListSubnets(cmOption, cls.GetVpcID())
	if err != nil {
		return nil, nil, err
	}

	// 过滤已分配未使用的子网
	var (
		allocatedSubnetsIds    = make(map[string][]string, 0)
		allocatedZoneSubnetNum = make(map[string]uint64, 0)
	)

	for i := range subnets {
		if strings.Contains(subnets[i].Name, cls.GetClusterID()) &&
			!utils.StringContainInSlice(subnets[i].ID, subnetIds) {

			if allocatedSubnetsIds[subnets[i].Zone] == nil {
				allocatedSubnetsIds[subnets[i].Zone] = make([]string, 0)
			}

			allocatedSubnetsIds[subnets[i].Zone] = append(allocatedSubnetsIds[subnets[i].Zone], subnets[i].ID)
			allocatedZoneSubnetNum[subnets[i].Zone] += subnets[i].TotalIps + 3
		}
	}

	blog.Infof("autoAllocateTcClusterCidr getClusterAllocatedEmptySubnets %+v", allocatedZoneSubnetNum)

	return allocatedZoneSubnetNum, allocatedSubnetsIds, nil
}

func allocateSubnetsToCluster(ctx context.Context, model store.ClusterManagerModel, cls *cmproto.Cluster) error { // nolint
	cloud, err := model.GetCloud(ctx, cls.GetProvider())
	if err != nil {
		blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] GetCloud failed: %v",
			cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}

	needAllocateSubnets, curSubnetIds, needScale, err := checkClusterNeedToScaleSubnet(cls)
	if err != nil {
		blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster"+
			"[%s:%s:%s] checkClusterNeedToScaleSubnet failed: %v",
			cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}
	if !needScale {
		blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] not need allocate new subnets",
			cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID())
		return nil
	}

	blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] needScale detailInfo %+v",
		cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), needAllocateSubnets)

	cidrSteps := cloud.GetNetworkInfo().GetUnderlayAutoSteps()
	if len(cidrSteps) == 0 {
		cidrSteps = business.DefaultSubnetSteps
	}
	transCidrSteps, _ := utils.Uint32ToInt(cidrSteps)

	// 优先选择已有的子网, 选择已存在子网后再分配新子网
	allocatedEmptyZoneSubnetNum, allocatedEmptyZoneSubnets, err := getClusterAllocatedEmptySubnets(cls, curSubnetIds)
	if err != nil {
		blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] getClusterAllocatedEmptySubnets "+
			"failed: %v", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}
	existedSubnetIds := make([]string, 0)
	for zone := range needAllocateSubnets {
		num, ok := allocatedEmptyZoneSubnetNum[zone]
		if ok {
			if needAllocateSubnets[zone] >= num {
				needAllocateSubnets[zone] -= num
			} else {
				needAllocateSubnets[zone] = 0
			}
			existedSubnetIds = append(existedSubnetIds, allocatedEmptyZoneSubnets[zone]...)
		}
	}
	blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] existedSubnets[%v], "+
		"needAllocateSubnets[%+v]", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(),
		existedSubnetIds, needAllocateSubnets)

	subnetSource := make([]*cmproto.NewSubnet, 0)
	for zone, subNum := range needAllocateSubnets {

		if subNum <= 0 {
			continue
		}

		subs, errLocal := utils.Decompose(int(subNum), transCidrSteps)
		if errLocal != nil {
			blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] decompose[%s:%s] "+
				"failed: %v", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), zone, subNum, errLocal)
			continue
		}

		for _, sub := range subs {
			subnetSource = append(subnetSource, &cmproto.NewSubnet{
				Zone:  zone,
				IpCnt: uint32(sub),
			})
		}
	}

	for i := range subnetSource {
		blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] newSubnet[%s:%s]",
			cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), subnetSource[i].GetZone(), subnetSource[i].GetIpCnt())
	}

	// once check cluster if need to allocate subnets
	_, _, onceCheck, err := checkClusterNeedToScaleSubnet(cls)
	if err != nil {
		blog.Errorf("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] checkClusterNeedToScaleSubnet "+
			"onceCheck failed: %v", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}
	if !onceCheck {
		blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] onceCheck not need "+
			"allocate new subnets", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID())
		return nil
	}

	cmOption, err := cloudprovider.GetCloudCmOptionByCluster(cls)
	if err != nil {
		blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] "+
			"GetCloudCmOptionByCluster failed: %v", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}
	// allocate new subnets
	newSubnetIds, err := business.AllocateClusterVpcCniSubnets(ctx, cls.GetClusterID(),
		cls.GetVpcID(), subnetSource, cmOption)
	if err != nil {
		blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] "+
			"AllocateClusterVpcCniSubnets failed: %v", cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), err)
		return err
	}

	if len(existedSubnetIds) > 0 {
		newSubnetIds = append(newSubnetIds, existedSubnetIds...)
	}

	blog.Infof("autoAllocateTcClusterCidr allocateSubnetsToCluster[%s:%s:%s] newSubnetIds %v",
		cls.GetRegion(), cls.GetVpcID(), cls.GetClusterID(), newSubnetIds)

	return business.AddSubnetsToCluster(cls, newSubnetIds, cmOption)
}
