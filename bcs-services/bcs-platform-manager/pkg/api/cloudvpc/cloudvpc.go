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

// Package cloudvpc cloudvpc operate
package cloudvpc

import (
	"context"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/cloudvpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// CreateCloudVPC 创建VPC
// @Summary 创建VPC
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cloudvpc [post]
func CreateCloudVPC(ctx context.Context, req *types.CreateCloudVPCReq) (*bool, error) {
	result, err := actions.NewCloudVPCAction().CreateCloudVPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateCloudVPC 更新VPC
// @Summary 更新VPC
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cloudvpc [put]
func UpdateCloudVPC(ctx context.Context, req *types.UpdateCloudVPCReq) (*bool, error) {
	result, err := actions.NewCloudVPCAction().UpdateCloudVPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCloudVPCRecommendCIDR 获取云VPC推荐CIDR列表
// @Summary 获取云VPC推荐CIDR列表
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cloudvpc [get]
func GetCloudVPCRecommendCIDR(ctx context.Context, req *types.GetCloudVPCRecommendCIDRReq) (*types.GetCloudVPCRecommendCIDRResp, error) {
	result, err := actions.NewCloudVPCAction().GetCloudVPCRecommendCIDR(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
