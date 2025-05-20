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

// Package storage defines the api client for bcs-storage
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/httpclient"
)

// Resource cluster resource
type Resource struct {
	// UpdateTime   string                 `json:"updateTime"`
	// CreateTime   string                 `json:"createTime"`
	ClusterID    string `json:"clusterId"`
	ResourceType string `json:"resourceType"`
	// ResourceName string                 `json:"resourceName"`
	Data map[string]interface{} `json:"data"`
}

// ClusterResourceResult cluster resource result
type ClusterResourceResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []*Resource `json:"data"`
	Total   int         `json:"total"`
}

// ClusteredNamespaces cluster namespaces
type ClusteredNamespaces struct {
	ClusterID  string   `json:"clusterId"`
	Namespaces []string `json:"namespaces"`
}

// ListMultiClusterResourcesReq list clustered namespaces request
type ListMultiClusterResourcesReq struct {
	Kind                string                `json:"-"`
	Offset              int                   `json:"offset"`
	Limit               int                   `json:"limit"`
	ClusteredNamespaces []ClusteredNamespaces `json:"clusteredNamespaces"`
	Field               string                `json:"field"`
	Conditions          []*operator.Condition `json:"conditions"`
}

var defaultLimit = 1000

// ListAllMultiClusterResources list all multi cluster resources
func ListAllMultiClusterResources(ctx context.Context, req ListMultiClusterResourcesReq) ([]*Resource, error) {
	result := make([]*Resource, 0)
	if len(req.ClusteredNamespaces) == 0 {
		return result, nil
	}
	// reset
	req.Limit = defaultLimit
	req.Offset = 0
	for {
		resources, _, err := ListMultiClusterResources(ctx, req)
		if err != nil {
			return nil, err
		}
		result = append(result, resources...)
		if len(resources) < defaultLimit {
			break
		}
		req.Offset += defaultLimit
	}
	return result, nil
}

// ListMultiClusterResources list multi cluster resources
func ListMultiClusterResources(ctx context.Context, req ListMultiClusterResourcesReq) ([]*Resource, int, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/storage/k8s/dynamic/multicluster_resources/%s", config.G.BCSAPIGW.Host, req.Kind)

	resp, err := httpclient.GetClient().R().
		SetContext(ctx).
		SetHeaders(contextx.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.G.BCSAPIGW.AuthToken).
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, fmt.Errorf("http code %d != 200", resp.StatusCode())
	}

	var result ClusterResourceResult
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, 0, err
	}

	if result.Code != 0 {
		return nil, 0, fmt.Errorf("resp code %d != 0, %s", result.Code, result.Message)
	}

	return result.Data, result.Total, nil
}
