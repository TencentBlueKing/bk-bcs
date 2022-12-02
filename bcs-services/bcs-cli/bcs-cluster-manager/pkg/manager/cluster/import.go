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
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Import 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)
func (c *ClusterMgr) Import(req types.ImportClusterReq) error {
	resp, err := c.client.ImportCluster(c.ctx, &clustermanager.ImportClusterReq{
		ClusterName: req.ClusterName,
		Provider:    req.Provider,
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		Environment: req.Environment,
		EngineType:  req.EngineType,
		IsExclusive: &wrapperspb.BoolValue{
			Value: req.IsExclusive,
		},
		ClusterType: req.ClusterType,
		Creator:     "bcs",
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
