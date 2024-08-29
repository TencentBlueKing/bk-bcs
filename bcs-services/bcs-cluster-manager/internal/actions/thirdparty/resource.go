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

// Package thirdparty xxx
package thirdparty

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/golang/protobuf/ptypes/wrappers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/daemon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetProviderResourceUsageAction action for get provider resource
type GetProviderResourceUsageAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req  *cmproto.GetProviderResourceUsageRequest
	resp *cmproto.GetProviderResourceUsageResponse

	regionInsTypes map[string][]string
}

// NewGetProviderResourceUsageAction create action
func NewGetProviderResourceUsageAction(model store.ClusterManagerModel) *GetProviderResourceUsageAction {
	return &GetProviderResourceUsageAction{
		model: model,
	}
}

func (ga *GetProviderResourceUsageAction) validate() error {
	if err := ga.req.Validate(); err != nil {
		return err
	}

	if !utils.StringInSlice(ga.req.GetProviderID(), []string{resource.YunTiPool,
		resource.CrPool, resource.SelfPool, resource.BcsResourcePool}) {
		return fmt.Errorf("not support provider[%s]", ga.req.GetProviderID())
	}

	if (ga.req.Region == "" && ga.req.InstanceType == "") && ga.req.GetRatio() == nil && ga.req.GetAvailable() == nil {
		return fmt.Errorf("GetProviderResourceUsageAction paras invalid")
	}

	return nil
}

func (ga *GetProviderResourceUsageAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *GetProviderResourceUsageAction) getProviderRegionInsTypes() error {
	getValue := func(wrap *wrappers.UInt32Value) *int {
		if wrap == nil {
			return nil
		}
		value := int(wrap.GetValue())
		return &value
	}

	pools, err := tresource.GetResourceManagerClient().ListAvailableInsufficientPools(ga.ctx, ga.req.GetProviderID(),
		ga.req.GetRegion(), ga.req.GetInstanceType(), resource.UsageRatio{
			QuotaRatio: getValue(ga.req.Ratio),
			QuotaCount: getValue(ga.req.Available),
		})
	if err != nil {
		blog.Errorf("GetProviderResourceUsageAction ListAvailableInsufficientPools failed: %v", err)
		return err
	}

	// filter region-instanceType
	var (
		regionInsTypes = make(map[string][]string, 0)
	)

	for i := range pools {
		region := pools[i].Region
		insType := pools[i].InstanceType

		v, ok := regionInsTypes[region]
		if !ok {
			if regionInsTypes[region] == nil {
				regionInsTypes[region] = make([]string, 0)
			}
			regionInsTypes[region] = append(regionInsTypes[region], insType)
			continue
		}

		if !utils.StringInSlice(insType, v) {
			regionInsTypes[region] = append(regionInsTypes[region], insType)
		}
	}

	blog.Infof("GetProviderResourceUsageAction getProviderRegionInsTypes %v", regionInsTypes)
	ga.regionInsTypes = regionInsTypes

	return nil
}

func getInstanceTypeFlag(region, instanceType string) string {
	return fmt.Sprintf("%s-%s", region, instanceType)
}

func (ga *GetProviderResourceUsageAction) getBizInfoByPools() error {
	var (
		regionPoolBizData = make(map[string][]*cmproto.BusinessInfo, 0)
		lock              = sync.Mutex{}
	)

	concurency := utils.NewRoutinePool(50)
	defer concurency.Close()

	for region, insTypes := range ga.regionInsTypes {
		concurency.Add(1)

		go func(region string, insTypes []string) {
			defer concurency.Done()

			for i := range insTypes {
				groups, err := daemon.FilterGroupsByRegionInsType(ga.model, region, insTypes[i])
				if err != nil {
					blog.Errorf("GetProviderResourceUsageAction FilterGroupsByRegionInsType failed: %v", err)
					continue
				}

				bizInfo := make([]*cmproto.BusinessInfo, 0)
				for _, group := range groups {
					cls, errLocal := ga.model.GetCluster(ga.ctx, group.ClusterID)
					if errLocal != nil {
						blog.Errorf("getProviderDevicePools GetCluster[%s] failed: %v",
							group.ClusterID, errLocal)
						continue
					}

					proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(cls.ProjectID, true)
					if errLocal != nil {
						blog.Errorf("getProviderDevicePools GetProjectInfo[%s] failed: %v",
							group.ClusterID, errLocal)
						continue
					}

					bizInfo = append(bizInfo, &cmproto.BusinessInfo{
						ProjectId:     proInfo.GetProjectID(),
						ProjectName:   proInfo.GetName(),
						ProjectCode:   proInfo.GetProjectCode(),
						ProjectUsers:  proInfo.GetManagers(),
						BizId:         proInfo.GetBusinessID(),
						BizName:       proInfo.GetBusinessName(),
						ClusterId:     cls.GetClusterID(),
						ClusterName:   cls.ClusterName,
						ClusterRegion: cls.GetRegion(),
						ClusterUsers:  fmt.Sprintf("%s,%s", cls.GetCreator(), cls.GetUpdater()),
						GroupId:       group.GetNodeGroupID(),
						GroupName:     group.GetName(),
						InstanceType:  group.GetLaunchTemplate().GetInstanceType(),
						Zones: func() string {
							if len(group.GetAutoScaling().GetZones()) > 0 {
								return strings.Join(group.GetAutoScaling().GetZones(), ",")
							}

							return ""
						}(),
						ConsumerId: group.GetConsumerID(),
						GroupUsers: fmt.Sprintf("%s,%s", group.GetCreator(), group.GetUpdater()),
						Url: fmt.Sprintf(options.GetGlobalCMOptions().ComponentDeploy.BcsClusterUrl,
							proInfo.GetProjectCode(), group.ClusterID),
					})
				}

				lock.Lock()
				regionPoolBizData[getInstanceTypeFlag(region, insTypes[i])] = bizInfo
				lock.Unlock()
			}

		}(region, insTypes)
	}

	concurency.Wait()

	result, err := utils.MarshalInterfaceToValue(regionPoolBizData)
	if err != nil {
		blog.Errorf("marshal modules err, %s", err.Error())
		return err
	}
	ga.resp.Data = result

	return nil
}

func (ga *GetProviderResourceUsageAction) getProviderBizUsageData() error {
	err := ga.getProviderRegionInsTypes()
	if err != nil {
		return err
	}

	err = ga.getBizInfoByPools()
	if err != nil {
		return err
	}

	return nil
}

// Handle handles resource usage
func (ga *GetProviderResourceUsageAction) Handle(ctx context.Context, req *cmproto.GetProviderResourceUsageRequest,
	resp *cmproto.GetProviderResourceUsageResponse) {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.getProviderBizUsageData(); err != nil {
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("GetProviderResourceUsageAction get provider biz data successfully")
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
