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

package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

// Update 更新集群
func (c *ClusterMgr) Update(req types.UpdateClusterReq) error {
	resp, err := c.client.UpdateCluster(c.ctx, &clustermanager.UpdateClusterReq{
		ClusterID:   req.ClusterID,
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		EngineType:  req.EngineType,
		IsExclusive: &wrapperspb.BoolValue{Value: req.IsExclusive},
		ClusterType: req.ClusterType,
		Updater:     req.Updater,
		ManageType:  req.ManageType,
		ClusterName: req.ClusterName,
		Environment: req.Environment,
		Provider:    req.Provider,
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
		Master: req.Master,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
