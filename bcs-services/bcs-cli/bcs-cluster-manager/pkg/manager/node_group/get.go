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
 *
 */

package nodegroup

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Get 查询节点池
func (c *NodeGroupMgr) Get(req types.GetNodeGroupReq) (types.GetNodeGroupResp, error) {
	var (
		resp types.GetNodeGroupResp
		err  error
	)

	servResp, err := c.client.GetNodeGroup(c.ctx, &clustermanager.GetNodeGroupRequest{
		NodeGroupID: req.NodeGroupID,
	})

	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.Data = types.NodeGroup{
		NodeGroupID:     servResp.Data.NodeGroupID,
		Name:            servResp.Data.Name,
		ClusterID:       servResp.Data.ClusterID,
		Region:          servResp.Data.Region,
		EnableAutoscale: servResp.Data.EnableAutoscale,
		AutoScaling: types.AutoScaling{
			MinSize: servResp.Data.AutoScaling.MinSize,
			MaxSize: servResp.Data.AutoScaling.MaxSize,
		},
		LaunchTemplate: types.LaunchConfiguration{
			InstanceType: servResp.Data.LaunchTemplate.InstanceType,
			ImageInfo: types.ImageInfo{
				ImageName: servResp.Data.LaunchTemplate.ImageInfo.ImageName,
			},
		},
		NodeOS:     servResp.Data.NodeOS,
		Creator:    servResp.Data.Creator,
		Updater:    servResp.Data.Updater,
		CreateTime: servResp.Data.CreateTime,
		UpdateTime: servResp.Data.UpdateTime,
		ProjectID:  servResp.Data.ProjectID,
		Provider:   servResp.Data.Provider,
		Status:     servResp.Data.Status,
	}

	return resp, nil
}
