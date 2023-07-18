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

// CommonResp common resp
type CommonResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// AccessRequest request
type AccessRequest struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
	Env        string `json:"env_name"`
}

// AccessTokenResp response
type AccessTokenResp struct {
	CommonResp
	Data *AccessTokenInfo `json:"data"`
}

// AccessTokenInfo data
type AccessTokenInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    uint32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ResourceRequest request
type ResourceRequest struct {
	ProjectID    string `json:"project_id"`
	ServiceCode  string `json:"service_code"`
	ResourceCode string `json:"resource_code"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Creator      string `json:"creator"`
}

// VerifyRequest verify user permission
type VerifyRequest struct {
	ProjectID    string `json:"project_id"`
	ServiceCode  string `json:"service_code"`
	PolicyCode   string `json:"policy_code"`
	ResourceCode string `json:"resource_code"`
	ResourceType string `json:"resource_type"`
	UserID       string `json:"user_id"`
}

// PermRequest perm request
type PermRequest struct {
	ProjectID       string                `json:"project_id"`
	ServiceCode     string                `json:"service_code"`
	PolicyList      []*PolicyResourceType `json:"policy_resource_type_list"`
	UserID          string                `json:"user_id"`
	IsExactResource int                   `json:"is_exact_resource"`
}

// PolicyResourceType policy/resource
type PolicyResourceType struct {
	PolicyCode   string `json:"policy_code"`
	ResourceType string `json:"resource_type"`
}

// PermResp response
type PermResp struct {
	CommonResp
	Data []PermList `json:"data"`
}

// PermList resource/perm
type PermList struct {
	PolicyCode       string   `json:"policy_code"`
	ResourceCodeList []string `json:"resource_code_list"`
	ResourceType     string   `json:"resource_type"`
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
