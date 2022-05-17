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

const (
	// ServiceCode default bcs
	ServiceCode = "bcs"
	// ClusterProd prod
	ClusterProd = "cluster_prod"
	// ClusterTest test
	ClusterTest = "cluster_test"
)

const (
	// PolicyCreate create permission
	PolicyCreate = "create"
	// PolicyEdit edit permission
	PolicyEdit = "edit"
	// PolicyUse use permission
	PolicyUse = "use"
	// PolicyDelete delete permission
	PolicyDelete = "delete"
	// PolicyView view permission
	PolicyView = "view"
)

const (
	// NO_RES no resource
	NO_RES = "**"
	// ANY_RES any resource
	ANY_RES = "*"
)

var (
	// PolicyList policy list for v0
	PolicyList = []string{"create", "delete", "view", "edit", "use", "deploy", "download"}
	// V3PolicyList for v3 cluster policy
	V3PolicyList = []string{"cluster_create", "cluster_view", "cluster_manage", "cluster_delete", "cluster_use"}

	sharedClusterOpenPolicy = []string{"view", "use", "cluster_view", "cluster_manage", "cluster_use"}
)

// CommonResp common resp
type CommonResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// AccessRequest request
type AccessRequest struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
	Env        string `json:"env"`
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

// GetAllSharedClusterPerm get sharedCluster perm policy for v0/v3
func GetAllSharedClusterPerm() map[string]bool {
	var (
		defaultPerm = make(map[string]bool)
		policyList = make([]string, 0)
	)
	policyList = append(policyList, PolicyList...)
	policyList = append(policyList, V3PolicyList...)

	for _, p := range policyList {
		defaultPerm[p] = false
		if utils.StringInSlice(p, sharedClusterOpenPolicy) {
			defaultPerm[p] = true
		}
	}

	return defaultPerm
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

// GetSharedClusterPerm get sharedCluster perm policy
func GetV0SharedClusterPerm() map[string]bool {
	defaultPerm := make(map[string]bool)
	for _, p := range PolicyList {
		defaultPerm[p] = false
		if utils.StringInSlice(p, sharedClusterOpenPolicy) {
			defaultPerm[p] = true
		}
	}

	return defaultPerm
}

// GetInitPerm init perm policy
func GetInitPerm() map[string]bool {
	defaultPerm := make(map[string]bool)
	for _, p := range PolicyList {
		defaultPerm[p] = false
	}

	return defaultPerm
}

// ExistResourceTypeAndPolicyCode get cluster list in perm list
func ExistResourceTypeAndPolicyCode(permList []PermList, policyResource PolicyResourceType) []string {
	for i := range permList {
		if permList[i].ResourceType == policyResource.ResourceType && permList[i].PolicyCode == policyResource.PolicyCode {
			return permList[i].ResourceCodeList
		}
	}

	return nil
}

// ClusterPermMatrix get init perm resourceType and policy
func ClusterPermMatrix() []*PolicyResourceType {
	permMatrix := make([]*PolicyResourceType, 0)
	for _, policy := range PolicyList {
		permMatrix = append(permMatrix, &PolicyResourceType{
			PolicyCode:   policy,
			ResourceType: ClusterTest,
		})
		permMatrix = append(permMatrix, &PolicyResourceType{
			PolicyCode:   policy,
			ResourceType: ClusterProd,
		})
	}

	return permMatrix
}

func GetClusterResType(env string) string {
	switch env {
	case "prod":
		return ClusterProd
	}

	return ClusterTest
}
