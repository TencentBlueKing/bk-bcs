/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import "fmt"

const (
	// Project 项目
	Project = "project"
	// Cluster 集群
	Cluster = "cluster"
	// Namespace 命名空间
	Namespace = "namespace"
	// TemplateSet 模板集
	TemplateSet = "templateset"
	// CloudAccount 云账号
	CloudAccount = "cloud_account"

	methodListInstance      = "list_instance"
	methodFetchInstanceInfo = "fetch_instance_info"

	attrID          = "id"
	attrDisplayName = "display_name"
	attrIAMPath     = "_bk_iam_path_"
	attrApprovers   = "_bk_iam_approvers_"

	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 1000
	maxIDs          = 1000
	maxIDLen        = 256
	maxKeywordLen   = 128
	maxRequires     = 20
	maxBodyBytes    = 1 << 20
)

// CallbackRequest IAM V4 资源回调请求
type CallbackRequest struct {
	Type     string   `json:"type"`
	Method   string   `json:"method"`
	Filter   Filter   `json:"filter"`
	Page     Page     `json:"page"`
	Requires []string `json:"requires"`
}

// Filter 列表 / 拉取过滤条件
type Filter struct {
	Keyword   string         `json:"keyword"`
	IDs       []string       `json:"ids"`
	Parent    ResourceParent `json:"parent"`
	Ancestors []Ancestor     `json:"ancestors"`
}

// ResourceParent 直接上级
type ResourceParent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Ancestor 祖先节点
type Ancestor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Page V4 分页，page 从 1 开始
type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// ListResult 列表结果
type ListResult struct {
	Count   int        `json:"count"`
	Results []Instance `json:"results"`
}

// Instance 资源实例
type Instance struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name,omitempty"`
	IAMPath     string   `json:"_bk_iam_path_,omitempty"`
	Approvers   []string `json:"_bk_iam_approvers_,omitempty"`
}

// SuccessResponse 成功信封，仅含 data
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ErrorResponse 失败信封
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody 失败详情
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type requestError struct {
	status  int
	code    string
	message string
}

func (e *requestError) Error() string {
	return e.message
}

func invalidRequest(msg string) *requestError {
	return &requestError{status: 400, code: "INVALID_REQUEST", message: msg}
}

// InvalidRequest 参数错误，回调返回 HTTP 400
func InvalidRequest(msg string) error {
	return invalidRequest(msg)
}

// CombineNameID 展示名：name(id)
func CombineNameID(name, id string) string {
	return fmt.Sprintf("%s(%s)", name, id)
}

// ProjectPath 集群/云账号的 IAM path
func ProjectPath(projectID string) string {
	if projectID == "" {
		return ""
	}
	return fmt.Sprintf("/project,%s/", projectID)
}

// NamespacePath 命名空间的 IAM path
func NamespacePath(projectID, clusterID string) string {
	if projectID == "" || clusterID == "" {
		return ""
	}
	return fmt.Sprintf("/project,%s/cluster,%s/", projectID, clusterID)
}
