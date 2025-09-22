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

// Package cmdb cmdb operate
package cmdb

import (
	"context"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// DeleteAllByBkBizIDAndBkClusterID 清理cmdb容器数据
// @Summary 清理cmdb容器数据
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cmdb/delete_all [put]
func DeleteAllByBkBizIDAndBkClusterID(ctx context.Context,
	req *types.DeleteAllByBkBizIDAndBkClusterIDReq) (*bool, error) {
	result, err := actions.NewCmdbAction().DeleteAllByBkBizIDAndBkClusterID(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetBusiness 获取business列表
// @Summary 获取business列表
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /cmdb/business [post]
func GetBusiness(ctx context.Context,
	req *types.DeleteAllByBkBizIDAndBkClusterIDReq) (*[]cmdb.GetBusinessRespDataInfo, error) {
	result, err := actions.NewCmdbAction().GetBusiness(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}
