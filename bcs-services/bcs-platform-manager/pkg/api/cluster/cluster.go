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
func ListCluster(ctx context.Context, req *types.ListClusterReq) (*[]*types.ListClusterRsp, error) {
	result, err := actions.NewClusterAction().ListCluster(ctx, req)
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
