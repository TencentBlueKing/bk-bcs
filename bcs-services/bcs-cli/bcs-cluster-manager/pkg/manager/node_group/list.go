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

// List 获取节点池列表
func (c *NodeGroupMgr) List(req types.ListNodeGroupReq) (types.ListNodeGroupResp, error) {
	var (
		resp types.ListNodeGroupResp
		err  error
	)

	servResp, err := c.client.ListNodeGroup(c.ctx, &clustermanager.ListNodeGroupRequest{})

	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.Data = make([]types.NodeGroup, 0)

	for _, v := range servResp.Data {
		resp.Data = append(resp.Data, types.NodeGroup{
			NodeGroupID:     v.NodeGroupID,
			Name:            v.Name,
			ClusterID:       v.ClusterID,
			Region:          v.Region,
			EnableAutoscale: v.EnableAutoscale,
			AutoScaling: types.AutoScaling{
				MinSize: v.AutoScaling.MinSize,
				MaxSize: v.AutoScaling.MaxSize,
			},
			LaunchTemplate: types.LaunchConfiguration{
				InstanceType: v.LaunchTemplate.InstanceType,
				ImageInfo: types.ImageInfo{
					ImageName: v.LaunchTemplate.ImageInfo.ImageName,
				},
			},
			NodeOS:     v.NodeOS,
			Creator:    v.Creator,
			Updater:    v.Updater,
			CreateTime: v.CreateTime,
			UpdateTime: v.UpdateTime,
			ProjectID:  v.ProjectID,
			Provider:   v.Provider,
			Status:     v.Status,
		})
	}

	return resp, nil
}
