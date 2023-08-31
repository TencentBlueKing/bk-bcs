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

package auth

import "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

var (
	// PolicyList policy list for v0
	PolicyList = []string{"create", "delete", "view", "edit", "use", "deploy", "download"}
	// V3PolicyList for v3 cluster policy
	V3PolicyList = []string{"cluster_create", "cluster_view", "cluster_manage", "cluster_delete", "cluster_use"}

	sharedClusterOpenPolicy = []string{"view", "use", "cluster_view", "cluster_manage", "cluster_use"}
)

// CommonGateWayResp common resp
type CommonGateWayResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// AccessGateWayRequest request
type AccessGateWayRequest struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
	Env        string `json:"env_name"`
}

// AccessTokenGateWayResp response
type AccessTokenGateWayResp struct {
	CommonGateWayResp
	Data *AccessTokenInfo `json:"data"`
}

// AccessSsmRequest request
type AccessSsmRequest struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
	Env        string `json:"env"`
}

// CommonSsmResp common resp
type CommonSsmResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// AccessTokenSsmResp response
type AccessTokenSsmResp struct {
	CommonSsmResp
	Data *AccessTokenInfo `json:"data"`
}

// AccessTokenInfo data
type AccessTokenInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    uint32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// GetV3SharedClusterPerm get sharedCluster perm policy
func GetV3SharedClusterPerm() map[string]interface{} {
	defaultPerm := make(map[string]interface{})
	for _, p := range V3PolicyList {
		defaultPerm[p] = false
		if utils.StringInSlice(p, sharedClusterOpenPolicy) {
			defaultPerm[p] = true
		}
	}

	return defaultPerm
}
