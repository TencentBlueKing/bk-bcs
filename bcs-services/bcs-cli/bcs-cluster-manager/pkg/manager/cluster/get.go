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

// Get 获取集群
func (c *ClusterMgr) Get(req types.GetClusterReq) (types.GetClusterResp, error) {
	var (
		resp types.GetClusterResp
		err  error
	)

	servResp, err := c.client.GetCluster(c.ctx, &clustermanager.GetClusterReq{ClusterID: req.ClusterID})
	if err != nil {
		return resp, err
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp = types.GetClusterResp{
		Data: types.Cluster{
			ClusterID:   servResp.Data.ClusterID,
			ProjectID:   servResp.Data.ProjectID,
			BusinessID:  servResp.Data.BusinessID,
			EngineType:  servResp.Data.EngineType,
			IsExclusive: servResp.Data.IsExclusive,
			ClusterType: servResp.Data.ClusterType,
			Creator:     servResp.Data.Creator,
			Updater:     servResp.Data.Updater,
			ManageType:  servResp.Data.ManageType,
			ClusterName: servResp.Data.ClusterName,
			Environment: servResp.Data.Environment,
			Provider:    servResp.Data.Provider,
			Description: servResp.Data.Description,
			ClusterBasicSettings: types.ClusterBasicSettings{
				Version: servResp.Data.ClusterBasicSettings.Version,
			},
			NetworkType: servResp.Data.NetworkType,
			Region:      servResp.Data.Region,
			VpcID:       servResp.Data.VpcID,
			NetworkSettings: types.NetworkSettings{
				CidrStep:      servResp.Data.NetworkSettings.CidrStep,
				MaxNodePodNum: servResp.Data.NetworkSettings.MaxNodePodNum,
				MaxServiceNum: servResp.Data.NetworkSettings.MaxServiceNum,
			},
		},
	}

	return resp, nil
}
