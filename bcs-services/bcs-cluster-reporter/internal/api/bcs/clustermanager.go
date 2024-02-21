/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcs

import (
	"crypto/tls"
	"net/http"
)

// ClusterManager client
type ClusterManager struct {
	token           string
	url             string
	apiGatewayUrl   string
	apiGatewayToken string
}

// NewClusterManager init clustermanager client
func NewClusterManager(token, apiserver, apiGatewayUrl, apiGatewayToken string) (*ClusterManager, error) {
	return &ClusterManager{
		token:           token,
		url:             apiserver,
		apiGatewayUrl:   apiGatewayUrl,
		apiGatewayToken: apiGatewayToken,
	}, nil
}

// BcsTransport client
type BcsTransport struct {
	token string
}

// RoundTrip xxx
func (t *BcsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Set("accept", "application/json")
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+t.token)
	req.Header = header

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 设置为 true 来禁用证书验证 nolint
		},
	}

	return tr.RoundTrip(req)
}

var (
	_urlMap = map[string]string{
		"GetClusters":         "/bcsapi/v4/clustermanager/v1/cluster",
		"GetNodesByClusterId": "/bcsapi/v4/clustermanager/v1/cluster/%s/node",
		"GetNode":             "/bcsapi/v4/clustermanager/v1/node/%s",
	}
)
