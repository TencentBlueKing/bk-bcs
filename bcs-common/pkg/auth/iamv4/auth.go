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
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DirectAuth 直接鉴权 POST /api/v1/open/rbac/authorization/systems/{system_id}/auth/
func (c *Client) DirectAuth(ctx context.Context, systemID string, req DirectAuthRequest) (*DirectAuthResult, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("subject.id", req.Subject.ID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("subject.type", req.Subject.Type); err != nil {
		return nil, err
	}
	if err := requireModelID("action_id", req.ActionID); err != nil {
		return nil, err
	}
	var data DirectAuthResult
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/authorization/systems/%s/auth/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return &data, err
}

// DirectAuthByActions 批量操作鉴权
// POST /api/v1/open/rbac/authorization/systems/{system_id}/auth-by-actions/
func (c *Client) DirectAuthByActions(ctx context.Context, systemID string,
	req AuthByActionsRequest) ([]ActionAuthResult, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("subject.id", req.Subject.ID); err != nil {
		return nil, err
	}
	if len(req.ActionIDs) == 0 {
		return nil, fmt.Errorf("action_ids is required")
	}
	if len(req.ActionIDs) > 20 {
		return nil, fmt.Errorf("action_ids exceeds max 20")
	}
	var data ActionAuthResults
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/authorization/systems/%s/auth-by-actions/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return []ActionAuthResult(data), err
}

// DirectAuthByResources 批量资源鉴权
// POST /api/v1/open/rbac/authorization/systems/{system_id}/auth-by-resources/
func (c *Client) DirectAuthByResources(ctx context.Context, systemID string,
	req AuthByResourcesRequest) ([]ResourceAuthResult, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("subject.id", req.Subject.ID); err != nil {
		return nil, err
	}
	if err := requireModelID("action_id", req.ActionID); err != nil {
		return nil, err
	}
	if len(req.Resources) == 0 {
		return nil, fmt.Errorf("resources is required")
	}
	if len(req.Resources) > 20 {
		return nil, fmt.Errorf("resources exceeds max 20")
	}
	var data ResourceAuthResults
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/authorization/systems/%s/auth-by-resources/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return []ResourceAuthResult(data), err
}

// GeneratePermApplyURL 生成无权限申请 URL
// POST /api/v1/open/application/permission-apply-urls/
func (c *Client) GeneratePermApplyURL(ctx context.Context, req GenerateApplyURLRequest) (*ApplyURLData, error) {
	if req.SystemID == "" {
		req.SystemID = c.opt.SystemID
	}
	if err := requireModelID("system_id", req.SystemID); err != nil {
		return nil, err
	}
	if len(req.Permissions) == 0 {
		return nil, fmt.Errorf("permissions is required")
	}
	var data ApplyURLData
	err := c.do(ctx, http.MethodPost, "/api/v1/open/application/permission-apply-urls/", nil, req, nil, &data)
	return &data, err
}

// AddAuthorization 角色授权 POST /api/v1/open/rbac/mgmt/systems/{system_id}/authorizations/
func (c *Client) AddAuthorization(ctx context.Context, systemID, operator string, items []AuthorizationItem) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	header, err := operatorHeader(operator)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("authorizations is required")
	}
	if len(items) > 20 {
		return fmt.Errorf("authorizations exceeds max 20")
	}
	return c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/mgmt/systems/%s/authorizations/", url.PathEscape(systemID)),
		nil, items, header, nil)
}

// RevokeAuthorization 撤销授权 DELETE /api/v1/open/rbac/mgmt/systems/{system_id}/authorizations/
func (c *Client) RevokeAuthorization(ctx context.Context, systemID, operator string, items []AuthorizationItem) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	header, err := operatorHeader(operator)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("authorizations is required")
	}
	if len(items) > 20 {
		return fmt.Errorf("authorizations exceeds max 20")
	}
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/rbac/mgmt/systems/%s/authorizations/", url.PathEscape(systemID)),
		nil, items, header, nil)
}

// ListAuthorizationSubject 查询角色授权对象
// POST /api/v1/open/rbac/mgmt/systems/{system_id}/authorizations/query-subject/
func (c *Client) ListAuthorizationSubject(ctx context.Context, systemID string,
	req ListAuthorizationSubjectRequest) (*PageResult[AuthorizationSubject], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireModelID("role_id", req.RoleID); err != nil {
		return nil, err
	}
	var data PageResult[AuthorizationSubject]
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/mgmt/systems/%s/authorizations/query-subject/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return &data, err
}
