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

// Package cluster cluster operate
package cluster

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	projectrmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ClusterAction cluster action interface
type ClusterAction interface { // nolint
	ListCluster(ctx context.Context, req *types.ListClusterReq) ([]*types.ListClusterRsp, error)
	UpdateClusterOperator(ctx context.Context, req *types.UpdateClusterOperatorReq) (bool, error)
	UpdateClusterProjectBusiness(ctx context.Context, req *types.UpdateClusterProjectBusinessReq) (bool, error)
}

// Action action for cloud vpc
type Action struct{}

// NewClusterAction new cluster action
func NewClusterAction() ClusterAction {
	return &Action{}
}

// SortCluster 排序
type SortCluster []*types.ListClusterRsp

// Len 实现sort.Interface接口的Len方法
func (a SortCluster) Len() int { return len(a) }

// Less 实现sort.Interface接口的Less方法，这里我们先按Name排序，如果Name相同则按Age排序
func (a SortCluster) Less(i, j int) bool {
	sortKey := a[0].SortKey
	sortWay := a[0].SortWay

	switch sortKey {
	case "clusterID":
		if sortWay == "desc" {
			return a[i].ClusterID > a[j].ClusterID
		}
		return a[i].ClusterID < a[j].ClusterID
	case "clusterName":
		if sortWay == "desc" {
			return a[i].ClusterName > a[j].ClusterName
		}
		return a[i].ClusterName < a[j].ClusterName
	default:
		return a[i].CreateTime > a[j].CreateTime
	}
}

// Swap 实现sort.Interface接口的Swap方法
func (a SortCluster) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ListCluster list cluster
func (a *Action) ListCluster(ctx context.Context, req *types.ListClusterReq) ([]*types.ListClusterRsp, error) {
	projects, err := projectrmgr.ListAllProject(ctx)
	if err != nil {
		return nil, err
	}

	clusters, err := clustermgr.ListCluster(ctx, &clustermanager.ListClusterV2Req{
		ProjectID:  req.ProjectID,
		BusinessID: req.BusinessID,
		Provider:   req.Provider,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*types.ListClusterRsp, 0)
	for _, cluster := range clusters {
		result = append(result, &types.ListClusterRsp{
			ClusterID:       cluster.ClusterID,
			ClusterName:     cluster.ClusterName,
			Provider:        cluster.Provider,
			Region:          cluster.Region,
			VpcID:           cluster.VpcID,
			ProjectID:       cluster.ProjectID,
			BusinessID:      cluster.BusinessID,
			Environment:     cluster.Environment,
			EngineType:      cluster.EngineType,
			ClusterType:     cluster.ClusterType,
			Creator:         cluster.Creator,
			CreateTime:      cluster.CreateTime,
			UpdateTime:      cluster.UpdateTime,
			ManageType:      cluster.ManageType,
			Status:          cluster.Status,
			Updater:         cluster.Updater,
			Description:     cluster.Description,
			ClusterCategory: cluster.ClusterCategory,
			Link: func() string {
				projectCode := ""
				for _, project := range projects {
					if project.ProjectID == cluster.ProjectID {
						projectCode = project.ProjectCode
						break
					}
				}
				return fmt.Sprintf("%s/bcs/projects/%s/clusters?clusterId=%s",
					config.G.BCS.Host, projectCode, cluster.ClusterID)
			}(),
			SortKey:         req.SortKey,
			SortWay:         req.SortWay,
			Label:           cluster.Labels,
			SystemID:        cluster.SystemID,
			NetworkType:     cluster.NetworkType,
			ModuleID:        cluster.ModuleID,
			IsCommonCluster: cluster.IsCommonCluster,
			IsShared:        cluster.IsShared,
		})
	}

	sort.Sort(SortCluster(result))

	return result, nil
}

// UpdateClusterOperator update cluster operator
func (a *Action) UpdateClusterOperator(ctx context.Context, req *types.UpdateClusterOperatorReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID: req.ClusterID,
		Creator:   req.Creator,
		Updater:   req.Updater,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// UpdateClusterProjectBusiness update cluster project business
func (a *Action) UpdateClusterProjectBusiness(ctx context.Context,
	req *types.UpdateClusterProjectBusinessReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID:  req.ClusterID,
		ProjectID:  req.ProjectID,
		BusinessID: req.BusinessID,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}
