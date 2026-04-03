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

// Package nodegroup nodegroup operate
package nodegroup

import (
	"context"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/nodegroup"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ListNodeGroup 获取nodegroup列表
// @Summary 获取nodegroup列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.NodeGroup
// @Router  /nodegroup [get]
func ListNodeGroup(ctx context.Context, req *types.ListNodeGroupReq) (*types.ListNodeGroupResp, error) {
	result, err := actions.NewNodeGroupAction().ListNodeGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetNodeGroup 获取nodegroup
// @Summary 获取nodegroup
// @Tags    Logs
// @Produce json
// @Success 200 {struct} types.NodeGroup
// @Router  /nodegroup/{NodeGroupID}  [get]
func GetNodeGroup(ctx context.Context, req *types.GetNodeGroupReq) (*types.NodeGroup, error) {
	result, err := actions.NewNodeGroupAction().GetNodeGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// EnableNodeGroupAutoScale 开启节点池自动扩缩容
// @Summary 开启节点池自动扩缩容
// @Tags    Logs
// @Produce json
// @Success 200 {bool} types.NodeGroup
// @Router  /nodegroup/{NodeGroupID}/autoscale/enable  [post]
func EnableNodeGroupAutoScale(ctx context.Context, req *types.EnableNodeGroupAutoScaleReq) (*bool, error) {
	result, err := actions.NewNodeGroupAction().EnableNodeGroupAutoScale(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DisableNodeGroupAutoScale 关闭节点池自动扩缩容
// @Summary 关闭节点池自动扩缩容
// @Tags    Logs
// @Produce json
// @Success 200 {bool} types.NodeGroup
// @Router  /nodegroup/{NodeGroupID}/autoscale/disable  [post]
func DisableNodeGroupAutoScale(ctx context.Context, req *types.DisableNodeGroupAutoScaleReq) (*bool, error) {
	result, err := actions.NewNodeGroupAction().DisableNodeGroupAutoScale(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateNodeGroup 更新节点池
// @Summary 更新节点池
// @Tags    Logs
// @Produce json
// @Success 200 {bool} types.NodeGroup
// @Router  /nodegroup/{NodeGroupID}  [put]
func UpdateNodeGroup(ctx context.Context, req *types.UpdateNodeGroupReq) (*bool, error) {
	result, err := actions.NewNodeGroupAction().UpdateNodeGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateGroupMinMaxSize 更新minSize/maxSize信息
// @Summary 更新minSize/maxSize信息
// @Tags    Logs
// @Produce json
// @Success 200 {bool} types.NodeGroup
// @Router  /nodegroup/{NodeGroupID}/boundsize  [post]
func UpdateGroupMinMaxSize(ctx context.Context, req *types.UpdateGroupMinMaxSizeReq) (*bool, error) {
	result, err := actions.NewNodeGroupAction().UpdateGroupMinMaxSize(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
