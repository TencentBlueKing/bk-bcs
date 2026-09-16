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

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// HeaderBkAPIAuthorization 蓝鲸网关应用身份
	HeaderBkAPIAuthorization = "X-Bkapi-Authorization"
	// HeaderTenantID 租户 ID
	HeaderTenantID = "X-Bk-Tenant-Id"
	// HeaderIAMOperator IAM 管理类接口操作人
	HeaderIAMOperator = "X-Bkiam-Operator"
	// DefaultTenantID 默认租户
	DefaultTenantID = "default"

	// SubjectUser 授权对象：用户
	SubjectUser = "user"
	// SubjectDepartment 授权对象：部门
	SubjectDepartment = "department"

	// ExpireTypeAt 指定过期时间点
	ExpireTypeAt = "at"
	// ExpireTypeDuration 指定相对时长（秒）
	ExpireTypeDuration = "duration"

	// IAMPathAttr 资源拓扑路径属性
	IAMPathAttr = "_bk_iam_path_"
)

var (
	// ErrServerNotInit client 未初始化
	ErrServerNotInit = errors.New("iamv4 client not init")
	// ErrEmptyOperator 管理类接口缺少操作人
	ErrEmptyOperator = errors.New("X-Bkiam-Operator is required")
)

// Options IAM V4 客户端配置
type Options struct {
	// SystemID 默认接入系统 ID，可被各接口入参覆盖
	SystemID string
	// AppCode 蓝鲸应用 ID
	AppCode string
	// AppSecret 蓝鲸应用密钥，禁止写入日志
	AppSecret string
	// GateWayHost bkiam 网关地址，例如 https://bkiam.example.com/prod
	GateWayHost string
	// TenantID 租户 ID，空则使用 default
	TenantID string
}

func (opt *Options) validate() error {
	if opt == nil {
		return ErrServerNotInit
	}
	if strings.TrimSpace(opt.AppCode) == "" || strings.TrimSpace(opt.AppSecret) == "" {
		return fmt.Errorf("AppCode/AppSecret required")
	}
	if strings.TrimSpace(opt.GateWayHost) == "" {
		return fmt.Errorf("GateWayHost required")
	}
	if opt.TenantID == "" {
		opt.TenantID = DefaultTenantID
	}
	opt.GateWayHost = strings.TrimRight(opt.GateWayHost, "/")
	return nil
}

// APIError IAM V4 接口错误
type APIError struct {
	StatusCode int
	RequestID  string
	Code       string
	Message    string
}

// Error 实现 error；不回显完整响应体，避免泄露内部细节
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.RequestID == "" {
		return fmt.Sprintf("iamv4 api error: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("iamv4 api error: status=%d code=%s request_id=%s message=%s",
		e.StatusCode, e.Code, e.RequestID, e.Message)
}

// PageQuery 列表分页
type PageQuery struct {
	Page     int
	PageSize int
}

func (p PageQuery) query() url.Values {
	q := url.Values{}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.PageSize > 0 {
		q.Set("page_size", strconv.Itoa(p.PageSize))
	}
	return q
}

// PageResult 分页结果
type PageResult[T any] struct {
	Count   int `json:"count"`
	Results []T `json:"results"`
}

// Subject 授权 / 鉴权对象
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ResourceRef 资源实例引用
type ResourceRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// AuthResource 鉴权资源实例
type AuthResource struct {
	ID         string                 `json:"id"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// System 接入系统
type System struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Managers    []string `json:"managers,omitempty"`
	Clients     []string `json:"clients,omitempty"`
	CallbackURL string   `json:"callback_url,omitempty"`
}

// CreateSystemRequest 注册系统
type CreateSystemRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Managers    []string `json:"managers,omitempty"`
	Clients     []string `json:"clients"`
	CallbackURL string   `json:"callback_url,omitempty"`
}

// UpdateSystemRequest 更新系统
type UpdateSystemRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Managers    []string `json:"managers,omitempty"`
	Clients     []string `json:"clients,omitempty"`
	CallbackURL string   `json:"callback_url,omitempty"`
}

// SystemIDData 系统 ID 响应
type SystemIDData struct {
	ID string `json:"id"`
}

// AuthTokenData 系统回调鉴权 Token
type AuthTokenData struct {
	AuthToken string `json:"auth_token"`
}

// ResourceType 资源类型
type ResourceType struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Ancestors []string `json:"ancestors,omitempty"`
}

// UpdateResourceTypeRequest 更新资源类型
type UpdateResourceTypeRequest struct {
	Name      string   `json:"name,omitempty"`
	Ancestors []string `json:"ancestors,omitempty"`
}

// Action 操作
type Action struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ResourceTypeID string `json:"resource_type_id,omitempty"`
}

// UpdateActionRequest 更新操作（仅允许改名称）
type UpdateActionRequest struct {
	Name string `json:"name"`
}

// RoleAction 角色关联操作
type RoleAction struct {
	ID             string `json:"id"`
	ResourceTypeID string `json:"resource_type_id,omitempty"`
}

// Role 角色
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Actions     []RoleAction `json:"actions"`
}

// UpdateRoleRequest 更新角色基本信息
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// DirectAuthRequest 直接鉴权
type DirectAuthRequest struct {
	Subject  Subject       `json:"subject"`
	ActionID string        `json:"action_id"`
	Resource *AuthResource `json:"resource,omitempty"`
}

// DirectAuthResult 直接鉴权结果
type DirectAuthResult struct {
	Allowed bool `json:"allowed"`
}

// AuthByActionsRequest 批量操作鉴权
type AuthByActionsRequest struct {
	Subject   Subject       `json:"subject"`
	ActionIDs []string      `json:"action_ids"`
	Resource  *AuthResource `json:"resource,omitempty"`
}

// ActionAuthResult 按操作的鉴权结果
type ActionAuthResult struct {
	ActionID string `json:"action_id"`
	Allowed  bool   `json:"allowed"`
}

// ActionAuthResults 兼容 data 为数组或 {action_id: allowed} 对象
type ActionAuthResults []ActionAuthResult

// UnmarshalJSON 解析批量操作鉴权结果
func (r *ActionAuthResults) UnmarshalJSON(b []byte) error {
	items, err := decodeIDAllowed(b, "action_id")
	if err != nil {
		return err
	}
	out := make([]ActionAuthResult, 0, len(items))
	for _, it := range items {
		out = append(out, ActionAuthResult{ActionID: it.id, Allowed: it.allowed})
	}
	*r = out
	return nil
}

// AuthByResourcesRequest 批量资源鉴权
type AuthByResourcesRequest struct {
	Subject   Subject        `json:"subject"`
	ActionID  string         `json:"action_id"`
	Resources []AuthResource `json:"resources"`
}

// ResourceAuthResult 按资源的鉴权结果
type ResourceAuthResult struct {
	ResourceID string `json:"resource_id"`
	Allowed    bool   `json:"allowed"`
}

// ResourceAuthResults 兼容 data 为数组或 {resource_id: allowed} 对象
type ResourceAuthResults []ResourceAuthResult

// UnmarshalJSON 解析批量资源鉴权结果
func (r *ResourceAuthResults) UnmarshalJSON(b []byte) error {
	items, err := decodeIDAllowed(b, "resource_id")
	if err != nil {
		return err
	}
	out := make([]ResourceAuthResult, 0, len(items))
	for _, it := range items {
		out = append(out, ResourceAuthResult{ResourceID: it.id, Allowed: it.allowed})
	}
	*r = out
	return nil
}

type idAllowed struct {
	id      string
	allowed bool
}

func decodeIDAllowed(b []byte, idField string) ([]idAllowed, error) {
	var arr []json.RawMessage
	if err := json.Unmarshal(b, &arr); err == nil {
		out := make([]idAllowed, 0, len(arr))
		for _, raw := range arr {
			var row struct {
				ResourceID string `json:"resource_id"`
				ActionID   string `json:"action_id"`
				ID         string `json:"id"`
				Allowed    bool   `json:"allowed"`
			}
			if err := json.Unmarshal(raw, &row); err != nil {
				return nil, err
			}
			id := row.ResourceID
			if idField == "action_id" && row.ActionID != "" {
				id = row.ActionID
			}
			if id == "" {
				id = row.ActionID
			}
			if id == "" {
				id = row.ID
			}
			out = append(out, idAllowed{id: id, allowed: row.Allowed})
		}
		return out, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, err
	}
	if raw, ok := obj["results"]; ok {
		return decodeIDAllowed(raw, idField)
	}
	out := make([]idAllowed, 0, len(obj))
	for id, raw := range obj {
		var allowed bool
		if err := json.Unmarshal(raw, &allowed); err != nil {
			var wrap struct {
				Allowed bool `json:"allowed"`
			}
			if err := json.Unmarshal(raw, &wrap); err != nil {
				return nil, err
			}
			allowed = wrap.Allowed
		}
		out = append(out, idAllowed{id: id, allowed: allowed})
	}
	return out, nil
}

// ApplyAncestor 申请权限的祖先资源
type ApplyAncestor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ApplyResource 申请权限的资源拓扑
type ApplyResource struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Ancestors []ApplyAncestor `json:"ancestors,omitempty"`
}

// ApplyPermission 申请的操作与资源
type ApplyPermission struct {
	ActionID  string          `json:"action_id"`
	Resources []ApplyResource `json:"resources,omitempty"`
}

// GenerateApplyURLRequest 生成无权限申请 URL
type GenerateApplyURLRequest struct {
	SystemID    string            `json:"system_id"`
	Permissions []ApplyPermission `json:"permissions"`
}

// ApplyURLData 申请 URL
type ApplyURLData struct {
	URL string `json:"url"`
}

// AuthorizationItem 角色授权 / 撤销授权项
type AuthorizationItem struct {
	Subject               Subject       `json:"subject"`
	RoleID                string        `json:"role_id"`
	RelatedResourceTypeID string        `json:"related_resource_type_id,omitempty"`
	Resources             []ResourceRef `json:"resources,omitempty"`
	ExpiredAt             int64         `json:"expired_at,omitempty"`
}

// ListAuthorizationSubjectRequest 查询角色授权对象
type ListAuthorizationSubjectRequest struct {
	RoleID                string       `json:"role_id"`
	RelatedResourceTypeID string       `json:"related_resource_type_id,omitempty"`
	Resource              *ResourceRef `json:"resource,omitempty"`
	Page                  int          `json:"page,omitempty"`
	PageSize              int          `json:"page_size,omitempty"`
}

// AuthorizationSubject 已授权对象
type AuthorizationSubject struct {
	Subject   Subject `json:"subject"`
	ExpiredAt int64   `json:"expired_at"`
}

// PermissionInstance 权限范围实例
type PermissionInstance struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// PermissionResource 权限范围资源
type PermissionResource struct {
	RelatedResourceTypeID string               `json:"related_resource_type_id"`
	IsAny                 bool                 `json:"is_any,omitempty"`
	Instances             []PermissionInstance `json:"instances"`
}

// PermissionScopeItem 空间 / 用户组权限范围
type PermissionScopeItem struct {
	ID        string               `json:"id"`
	Resources []PermissionResource `json:"resources"`
}

// SubjectScope 人员管控范围
type SubjectScope struct {
	Users       []string `json:"users"`
	Departments []string `json:"departments,omitempty"`
}

// CreateSpaceRequest 创建管理空间
type CreateSpaceRequest struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Managers        []string              `json:"managers"`
	PermissionScope []PermissionScopeItem `json:"permission_scope"`
	SubjectScope    SubjectScope          `json:"subject_scope"`
}

// Space 管理空间
type Space struct {
	ID              int64                 `json:"id"`
	Name            string                `json:"name,omitempty"`
	Description     string                `json:"description,omitempty"`
	Managers        []string              `json:"managers,omitempty"`
	PermissionScope []PermissionScopeItem `json:"permission_scope,omitempty"`
	SubjectScope    *SubjectScope         `json:"subject_scope,omitempty"`
}

// IDData 通用 ID 响应
type IDData struct {
	ID int64 `json:"id"`
}

// GroupMember 用户组成员
type GroupMember struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// CreateGroupRequest 创建用户组
type CreateGroupRequest struct {
	Name                string                `json:"name"`
	Description         string                `json:"description"`
	Members             []GroupMember         `json:"members,omitempty"`
	Permissions         []PermissionScopeItem `json:"permissions,omitempty"`
	PermissionExpiredAt int64                 `json:"permission_expired_at"`
	ITSMWorkflowID      string                `json:"itsm_workflow_id,omitempty"`
}

// Group 用户组
type Group struct {
	ID                  int64                 `json:"id"`
	Name                string                `json:"name,omitempty"`
	Description         string                `json:"description,omitempty"`
	Members             []GroupMember         `json:"members,omitempty"`
	Permissions         []PermissionScopeItem `json:"permissions,omitempty"`
	PermissionExpiredAt int64                 `json:"permission_expired_at,omitempty"`
}

// GroupMembersRequest 批量增删组成员
type GroupMembersRequest struct {
	Members []GroupMember `json:"members"`
}

// GroupPermission 用户组权限
type GroupPermission struct {
	ID        string               `json:"id,omitempty"`
	Resources []PermissionResource `json:"resources,omitempty"`
}

// DisplayResourceType 角色默认展示资源层级
type DisplayResourceType struct {
	RelatedResourceTypeID string `json:"related_resource_type_id"`
	DisplayResourceTypeID string `json:"display_resource_type_id"`
}

// UpdateDisplayResourceTypesRequest 全量配置角色展示层级
type UpdateDisplayResourceTypesRequest struct {
	DisplayResourceTypes []DisplayResourceType `json:"display_resource_types"`
}

// CreateRoleTicketRequest 申请角色权限单据
type CreateRoleTicketRequest struct {
	Applicant   string                `json:"applicant"`
	Roles       []PermissionScopeItem `json:"roles"`
	ExpireType  string                `json:"expire_type"`
	ExpireValue int64                 `json:"expire_value"`
	Reason      string                `json:"reason"`
}

// ITSMTicket ITSM 单据
type ITSMTicket struct {
	ID string `json:"id"`
	SN string `json:"sn"`
}

// RoleTicket 角色申请单据
type RoleTicket struct {
	ID         int64      `json:"id"`
	ITSMTicket ITSMTicket `json:"itsm_ticket"`
}

// RoleTicketFailed 申请失败明细
type RoleTicketFailed struct {
	Roles   []PermissionScopeItem `json:"roles"`
	Message string                `json:"message"`
}

// CreateRoleTicketResult 申请单据结果
type CreateRoleTicketResult struct {
	FailedDetails []RoleTicketFailed `json:"failed_details"`
	Tickets       []RoleTicket       `json:"tickets"`
}

// RoleTicketFilter 查询角色申请单据
type RoleTicketFilter struct {
	Page     int
	PageSize int
	// Extra 预留查询条件，按网关实际支持的 query 透传
	Extra url.Values
}

func (f RoleTicketFilter) query() url.Values {
	q := PageQuery{Page: f.Page, PageSize: f.PageSize}.query()
	for k, vs := range f.Extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return q
}
