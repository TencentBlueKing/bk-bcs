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

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

const (
	clusterUrl = "/clustermanager/v1/cluster/%s"
)

type (
	// GetClusterRequest 获取集群请求
	GetClusterRequest struct {
		ClusterID string `json:"clusterID,omitempty"`
	}
)

// GetCluster Query the cluster information of the specified Cluster ID
func (p *ProjectManagerClient) GetCluster(in *GetClusterRequest) (*clustermanager.GetClusterResp, error) {
	v, err := query.Values(in)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := p.do(fmt.Sprintf(clusterUrl, in.ClusterID), http.MethodGet, v, nil)
	if err != nil {
		return nil, fmt.Errorf("get cluster failed: %v", err)
	}
	resp := new(clustermanager.GetClusterResp)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get cluster unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("get cluster response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}
