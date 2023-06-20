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

package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Create 创建集群
func (c *ClusterMgr) Create(req types.CreateClusterReq) (types.CreateClusterResp, error) {
	var (
		resp types.CreateClusterResp
		err  error
	)

	servResp, err := c.client.CreateCluster(c.ctx, &clustermanager.CreateClusterReq{
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		EngineType:  req.EngineType,
		IsExclusive: req.IsExclusive,
		ClusterType: req.ClusterType,
		Creator:     "bcs",
		ManageType:  req.ManageType,
		ClusterName: req.ClusterName,
		Environment: req.Environment,
		Provider:    req.Provider,
		Description: req.Description,
		ClusterBasicSettings: &clustermanager.ClusterBasicSetting{
			Version: req.ClusterBasicSettings.Version,
		},
		NetworkType: req.NetworkType,
		Region:      req.Region,
		VpcID:       req.VpcID,
		NetworkSettings: &clustermanager.NetworkSetting{
			MaxNodePodNum: req.NetworkSettings.MaxNodePodNum,
			MaxServiceNum: req.NetworkSettings.MaxServiceNum,
			CidrStep:      req.NetworkSettings.CidrStep,
		},
		Master:         req.Master,
		OnlyCreateInfo: true,
	})
	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.ClusterID = servResp.Data.GetClusterID()
	resp.TaskID = servResp.Task.GetTaskID()

	return resp, nil
}
