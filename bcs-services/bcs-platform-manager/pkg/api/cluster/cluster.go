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

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ListCluster 获取cluster列表
// @Summary 获取cluster列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.ListClusterRsp
// @Router  /cluster [get]
func ListCluster(ctx context.Context, req *types.ListClusterReq) (*types.ListClusterResp, error) {
	result, err := actions.NewClusterAction().ListCluster(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetClusterOverview 获取cluster总览信息
// @Summary 获取cluster总览信息
// @Tags    Logs
// @Produce json
// @Success 200 {struct} types.GetClusterResp
// @Router  /cluster/{clusterID}/overview [get]
func GetClusterOverview(ctx context.Context, req *types.GetClusterOverviewReq) (*types.GetClusterOverviewResp, error) {
	result, err := actions.NewClusterAction().GetClusterOverview(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetCluster 获取cluster详情
// @Summary 获取cluster详情
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.GetClusterResp
// @Router  /cluster/{clusterID} [get]
func GetCluster(ctx context.Context, req *types.GetClusterReq) (*types.GetClusterResp, error) {
	result, err := actions.NewClusterAction().GetCluster(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetClusterBasicInfo 获取cluster basic info
// @Summary 获取cluster详情
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.GetClusterBasicInfoResp
// @Router  /cluster/{clusterID}/basicinfo [get]
func GetClusterBasicInfo(ctx context.Context, req *types.GetClusterBasicInfoReq) (
	*types.GetClusterBasicInfoResp, error) {
	result, err := actions.NewClusterAction().GetClusterBasicInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetClusterNewworkConfig 获取cluster network config
// @Summary 获取cluster详情
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.GetClusterNetworkConfigResp
// @Router  /cluster/{clusterID}/networkconfig [get]
func GetClusterNewworkConfig(ctx context.Context, req *types.GetClusterNetworkConfigReq) (
	*types.GetClusterNetworkConfigResp, error) {
	result, err := actions.NewClusterAction().GetClusterNetworkConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetClusterControlPlaneConfig 获取cluster control plane config
// @Summary 获取cluster详情
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.GetClusterControlPlaneConfigResp
// @Router  /cluster/{clusterID}/controlplaneconfig [get]
func GetClusterControlPlaneConfig(ctx context.Context, req *types.GetClusterControlPlaneConfigReq) (
	*types.GetClusterControlPlaneConfigResp, error) {
	result, err := actions.NewClusterAction().GetClusterControlPlaneConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateClusterBasicInfo 更新cluster基本信息
// @Summary 更新cluster
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/basicinfo [put]
func UpdateClusterBasicInfo(ctx context.Context, req *types.UpdateClusterBasicInfoReq) (*bool, error) {
	result, err := actions.NewClusterAction().UpdateClusterBasicInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateClusterNetworkConfig 更新cluster网络配置
// @Summary 更新cluster
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/networkconfig [put]
func UpdateClusterNetworkConfig(ctx context.Context, req *types.UpdateClusterNetworkConfigReq) (*bool, error) {
	result, err := actions.NewClusterAction().UpdateClusterNetworkConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateClusterControlPlaneConfig 更新cluster控制面配置
// @Summary 更新cluster
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/controlplaneconfig [put]
func UpdateClusterControlPlaneConfig(ctx context.Context, req *types.UpdateClusterControlPlaneConfigReq) (*bool, error) {
	result, err := actions.NewClusterAction().UpdateClusterControlPlaneConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateClusterOperator 更新cluster 创建人和更新人
// @Summary 更新cluster 创建人和更新人
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/operator [put]
func UpdateClusterOperator(ctx context.Context, req *types.UpdateClusterOperatorReq) (*bool, error) {
	result, err := actions.NewClusterAction().UpdateClusterOperator(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateClusterProjectBusiness 更新cluster 项目或业务ID
// @Summary 更新cluster 项目或业务ID
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/projectidorbusinessid [put]
func UpdateClusterProjectBusiness(ctx context.Context, req *types.UpdateClusterProjectBusinessReq) (*bool, error) {
	result, err := actions.NewClusterAction().UpdateClusterProjectBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddClusterCidr 集群添加cidr
// @Summary 集群添加cidr
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/cidrs [post]
func AddClusterCidr(ctx context.Context, req *types.AddClusterCidrReq) (*bool, error) {
	result, err := actions.NewClusterAction().AddClusterCidr(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddSubnetToCluster 集群添加子网资源
// @Summary 集群添加子网资源
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cluster/{clusterID}/subnets [post]
func AddSubnetToCluster(ctx context.Context, req *types.AddSubnetToClusterReq) (*bool, error) {
	result, err := actions.NewClusterAction().AddSubnetToCluster(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
