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

// Package bcs xxx
package bcs

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/contextx"
)

// Cluster 集群信息
type Cluster struct {
	ProjectID   string `json:"projectID"`
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	BusinessID  string `json:"businessID"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
	ClusterType string `json:"clusterType"`
}

// GetClusterResponse 集群信息响应
type GetClusterResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *Cluster `json:"data"`
}

// GetCluster 获取集群详情
func GetCluster(ctx context.Context, clusterID string) (*Cluster, error) {
	bcsConf := config.G.BCS
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", bcsConf.Host, clusterID)
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeaders(contextx.GetLaneIDByCtx(ctx)).
		SetAuthToken(bcsConf.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	cluster := new(Cluster)
	if err := component.UnmarshalBKResult(resp, cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}
