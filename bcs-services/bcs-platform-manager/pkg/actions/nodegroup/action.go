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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
)

// NodeGroupAction nodegroup action interface
type NodeGroupAction interface { // nolint
	ListNodeGroup(ctx context.Context, req *types.ListNodeGroupReq) (*types.ListNodeGroupResp, error)
	GetNodeGroup(ctx context.Context, req *types.GetNodeGroupReq) (*types.NodeGroup, error)
	EnableNodeGroupAutoScale(ctx context.Context, req *types.EnableNodeGroupAutoScaleReq) (bool, error)
	DisableNodeGroupAutoScale(ctx context.Context, req *types.DisableNodeGroupAutoScaleReq) (bool, error)
	UpdateNodeGroup(ctx context.Context, req *types.UpdateNodeGroupReq) (bool, error)
	UpdateGroupMinMaxSize(ctx context.Context, req *types.UpdateGroupMinMaxSizeReq) (bool, error)
}

// Action action for nodegroup
type Action struct{}

// NewNodeGroupAction new nodegroup action
func NewNodeGroupAction() NodeGroupAction {
	return &Action{}
}

// ListNodeGroup list nodegroup
func (a *Action) ListNodeGroup(ctx context.Context, req *types.ListNodeGroupReq) (*types.ListNodeGroupResp, error) {
	/*nodegroupData, err := clustermgr.ListNodeGroup(ctx, &clustermanager.ListNodeGroupV2Request{
		Name:      req.Name,
		ClusterID: req.ClusterID,
		Region:    req.Region,
		ProjectID: req.ProjectID,
		Limit:     req.Limit,
		Page:      req.Page,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := make([]*types.ListNodeGroupData, 0)
	for _, nodeGroup := range nodegroupData.Results {
		as := &types.AutoScalingGroup{}
		if nodeGroup.AutoScaling != nil {
			as = &types.AutoScalingGroup{
				MinSize:     nodeGroup.AutoScaling.MinSize,
				MaxSize:     nodeGroup.AutoScaling.MaxSize,
				DesiredSize: nodeGroup.AutoScaling.DesiredSize,
			}
		}

		lt := &types.LaunchConfiguration{}
		if nodeGroup.LaunchTemplate != nil {
			lt.InstanceType = nodeGroup.LaunchTemplate.InstanceType
			if nodeGroup.LaunchTemplate.ImageInfo != nil {
				lt.ImageInfo = &types.ImageInfo{
					ImageID:   nodeGroup.LaunchTemplate.ImageInfo.ImageID,
					ImageName: nodeGroup.LaunchTemplate.ImageInfo.ImageName,
					ImageType: nodeGroup.LaunchTemplate.ImageInfo.ImageType,
					ImageOs:   nodeGroup.LaunchTemplate.ImageInfo.ImageOs,
				}
			}
		}

		result = append(result, &types.ListNodeGroupData{
			NodeGroupID:    nodeGroup.NodeGroupID,
			Name:           nodeGroup.Name,
			ClusterID:      nodeGroup.ClusterID,
			Status:         nodeGroup.Status,
			AutoScaling:    as,
			LaunchTemplate: lt,
		})
	}

	return &types.ListNodeGroupResp{
		Total:   nodegroupData.Count,
		Results: result,
	}, nil*/
	return nil, nil
}

// GetNodeGroup get nodegroup
func (a *Action) GetNodeGroup(ctx context.Context, req *types.GetNodeGroupReq) (*types.NodeGroup, error) {
	nodegroup, err := clustermgr.GetNodeGroup(ctx, &clustermanager.GetNodeGroupRequest{
		NodeGroupID: req.NodeGroupID,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.NodeGroup{
		NodeGroupID:      nodegroup.NodeGroupID,
		Name:             nodegroup.Name,
		ClusterID:        nodegroup.ClusterID,
		Region:           nodegroup.Region,
		EnableAutoscale:  nodegroup.EnableAutoscale,
		Labels:           nodegroup.Labels,
		Taints:           nodegroup.Taints,
		NodeOS:           nodegroup.NodeOS,
		Creator:          nodegroup.Creator,
		Updater:          nodegroup.Updater,
		CreateTime:       nodegroup.CreateTime,
		UpdateTime:       nodegroup.UpdateTime,
		ProjectID:        nodegroup.ProjectID,
		Provider:         nodegroup.Provider,
		Status:           nodegroup.Status,
		ConsumerID:       nodegroup.ConsumerID,
		CloudNodeGroupID: nodegroup.CloudNodeGroupID,
		Tags:             nodegroup.Tags,
		NodeGroupType:    nodegroup.NodeGroupType,
		ExtraInfo:        nodegroup.ExtraInfo,
	}, nil
}

// EnableNodeGroupAutoScale enable nodegroup auto scale
func (a *Action) EnableNodeGroupAutoScale(ctx context.Context, req *types.EnableNodeGroupAutoScaleReq) (bool, error) {
	result, err := clustermgr.EnableNodeGroupAutoScale(ctx, &clustermanager.EnableNodeGroupAutoScaleRequest{
		NodeGroupID: req.NodeGroupID,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// DisableNodeGroupAutoScale enable nodegroup auto scale
func (a *Action) DisableNodeGroupAutoScale(ctx context.Context, req *types.DisableNodeGroupAutoScaleReq) (bool, error) {
	result, err := clustermgr.DisableNodeGroupAutoScale(ctx, &clustermanager.DisableNodeGroupAutoScaleRequest{
		NodeGroupID: req.NodeGroupID,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateNodeGroup update nodegroup
func (a *Action) UpdateNodeGroup(ctx context.Context, req *types.UpdateNodeGroupReq) (bool, error) { // nolint
	nodegroup, err := clustermgr.GetNodeGroup(ctx, &clustermanager.GetNodeGroupRequest{
		NodeGroupID: req.NodeGroupID,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	updateReq := &clustermanager.UpdateNodeGroupRequest{
		NodeGroupID:    req.NodeGroupID,
		ClusterID:      nodegroup.ClusterID,
		Name:           req.Name,
		Region:         nodegroup.Region,
		AutoScaling:    nodegroup.AutoScaling,
		LaunchTemplate: nodegroup.LaunchTemplate,
		NodeTemplate:   nodegroup.NodeTemplate,
		Labels:         req.Labels,
		Taints:         req.Taints,
		Tags:           req.Tags,
		NodeOS:         nodegroup.NodeOS,
		Updater:        req.Updater,
		Provider:       nodegroup.Provider,
		ConsumerID:     nodegroup.ConsumerID,
		Desc:           req.Desc,
		OnlyUpdateInfo: req.OnlyUpdateInfo,
		ExtraInfo:      nodegroup.ExtraInfo,
	}

	if req.EnableAutoscale != nil {
		updateReq.EnableAutoscale.Value = *req.EnableAutoscale
	}
	if req.BkCloudID != nil {
		updateReq.BkCloudID.Value = *req.BkCloudID
	}
	if req.CloudAreaName != nil {
		updateReq.CloudAreaName.Value = *req.CloudAreaName
	}

	result, err := clustermgr.UpdateNodeGroup(ctx, updateReq)
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateGroupMinMaxSize update nodegroup min max size
func (a *Action) UpdateGroupMinMaxSize(ctx context.Context, req *types.UpdateGroupMinMaxSizeReq) (bool, error) {
	result, err := clustermgr.UpdateGroupMinMaxSize(ctx, &clustermanager.UpdateGroupMinMaxSizeRequest{
		NodeGroupID: req.NodeGroupID,
		MinSize:     req.MinSize,
		MaxSize:     req.MaxSize,
		Operator:    req.Operator,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}
