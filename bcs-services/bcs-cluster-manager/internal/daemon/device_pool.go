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

// Package daemon for daemon
package daemon

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	poolTotal             = "pool_total"
	poolAvailable         = "pool_available"
	poolOversoldTotal     = "pool_oversold_total"
	poolOversoldAvailable = "pool_oversold_available"
	poolUsed              = "pool_used"

	groupUsed  = "group_used"
	groupQuota = "group_quota"
)

func (d *Daemon) reportRegionInsTypeUsage(error chan<- error) {
	regionInstypes, err := tresource.GetResourceManagerClient().GetRegionInstanceTypesFromPools(d.ctx,
		resource.YunTiPool)
	if err != nil {
		blog.Errorf("reportRegionInsTypeUsage GetRegionInstanceTypesFromPools failed: %v", err)
		error <- err
		return
	}

	concurency := utils.NewRoutinePool(30)
	defer concurency.Close()

	for region, insTypes := range regionInstypes {
		concurency.Add(1)
		go func(region string, insTypes []string) {
			defer concurency.Done()

			for i := range insTypes {
				pools, errLocal := GetRegionDevicePoolDetail(d.model, region, insTypes[i], nil)
				if errLocal != nil {
					blog.Errorf("reportRegionInsTypeUsage[%s:%s] GetRegionDevicePoolDetail: %v",
						region, insTypes[i], errLocal)
					continue
				}

				// report data
				for _, pool := range pools {

					err = SetResourceDevicePoolData(pool.PoolId, pool)
					if err != nil {
						blog.Errorf("reportRegionInsTypeUsage[%s:%s] SetResourceDevicePoolData[%s]: %v %+v",
							region, insTypes[i], pool.PoolId, pool, err)
					}

					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						poolTotal, float64(pool.Total))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						poolAvailable, float64(pool.Available))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						poolOversoldTotal, float64(pool.OversoldTotal))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						poolOversoldAvailable, float64(pool.OversoldAvailable))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						poolUsed, float64(pool.Used))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						groupQuota, float64(pool.GroupQuota))
					metrics.ReportRegionInsTypeNum(region, pool.Zone, insTypes[i], pool.PoolId,
						groupUsed, float64(pool.GroupUsed))

				}

			}
		}(region, insTypes)
	}

	concurency.Wait()
}

// GetRegionDevicePoolDetail get region device pool detail
func GetRegionDevicePoolDetail(model store.ClusterManagerModel, region string, instanceType string,
	filterGroupIds []string) ([]*resource.DevicePoolInfo, error) {
	filterGroups, err := FilterGroupsByRegionInsType(model, region, instanceType)
	if err != nil {
		blog.Errorf("GetRegionDevicePoolDetail[%s:%s] FilterGroupsByRegionInsType failed: %v",
			region, instanceType, err)
		return nil, err
	}

	// 地域-机型 维度的 资源池 和 可用区列表
	zonePools, resourceZones, err := tresource.GetResourceManagerClient().ListRegionZonePools(context.Background(),
		resource.YunTiPool, region, instanceType)
	if err != nil {
		blog.Errorf("GetRegionDevicePoolDetail[%s:%s] ListRegionZonePools failed: %v", region, instanceType, err)
		return nil, err
	}

	if len(resourceZones) == 0 || len(zonePools) == 0 {
		blog.Errorf("region[%s] instanceType[%s] 无可用区机型", region, instanceType)
		return nil, fmt.Errorf("region[%s] instanceType[%s] 无可用区机型", region, instanceType)
	}

	// 当前需要如何分配 (只要机器足够即可，机器不够的情况，简单按照平均即可)
	for _, group := range filterGroups {

		if utils.StringInSlice(group.NodeGroupID, filterGroupIds) {
			continue
		}

		nodesDistribution, curDistribution, _, errLocal := GetGroupCurAndPredictNodes(model,
			group.NodeGroupID, resourceZones)
		if errLocal != nil {
			blog.Errorf("nodeGroup[%s] GetRegionDevicePoolDetail[%s:%s] GetGroupCurAndPredictNodes failed: %v",
				group.GetNodeGroupID(), region, instanceType, errLocal)
			continue
		}

		for zone := range nodesDistribution {
			_, ok := zonePools[zone]
			if ok {
				zonePools[zone].GroupQuota += nodesDistribution[zone]
			}
		}

		for zone := range curDistribution {
			_, ok := zonePools[zone]
			if ok {
				zonePools[zone].GroupUsed += curDistribution[zone]
			}
		}
	}

	pools := make([]*resource.DevicePoolInfo, 0)

	for i := range zonePools {
		blog.Infof("GetRegionDevicePoolDetail region[%s] zone[%s] instanceType[%s] pool[%s] "+
			"poolTotal[%v] poolAvailable[%v] poolOversoldTotal[%v] poolOversoldAvailable[%v] poolUsed[%v] "+
			"groupQuota[%v] groupUsed[%v]", region, zonePools[i].Zone, instanceType, zonePools[i].PoolId,
			zonePools[i].Total, zonePools[i].Available, zonePools[i].OversoldTotal, zonePools[i].OversoldAvailable,
			zonePools[i].Used, zonePools[i].GroupQuota, zonePools[i].GroupUsed)
		pools = append(pools, zonePools[i])
	}

	return pools, nil
}
