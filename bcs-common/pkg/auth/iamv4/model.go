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
	"strings"
)

// CreateSystem 注册系统 POST /api/v1/open/rbac/model/systems/
func (c *Client) CreateSystem(ctx context.Context, req CreateSystemRequest) (*SystemIDData, error) {
	if err := requireModelID("id", req.ID); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("name", req.Name); err != nil {
		return nil, err
	}
	if len(req.Clients) == 0 {
		return nil, fmt.Errorf("clients is required")
	}
	var data SystemIDData
	err := c.do(ctx, http.MethodPost, "/api/v1/open/rbac/model/systems/", nil, req, nil, &data)
	return &data, err
}

// RetrieveSystem 查询系统 GET /api/v1/open/rbac/model/systems/{system_id}/
func (c *Client) RetrieveSystem(ctx context.Context, systemID string) (*System, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data System
	err = c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/", url.PathEscape(systemID)),
		nil, nil, nil, &data)
	return &data, err
}

// UpdateSystem 更新系统 PUT /api/v1/open/rbac/model/systems/{system_id}/
func (c *Client) UpdateSystem(ctx context.Context, systemID string, req UpdateSystemRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	return c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/", url.PathEscape(systemID)),
		nil, req, nil, nil)
}

// RetrieveSystemAuthToken 获取系统回调 Auth Token
// GET /api/v1/open/rbac/model/systems/{system_id}/auth-token/
func (c *Client) RetrieveSystemAuthToken(ctx context.Context, systemID string) (*AuthTokenData, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data AuthTokenData
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/auth-token/", url.PathEscape(systemID)),
		nil, nil, nil, &data)
	return &data, err
}

// ListResourceType 查询资源类型列表
// GET /api/v1/open/rbac/model/systems/{system_id}/resource-types/
func (c *Client) ListResourceType(ctx context.Context, systemID string, page PageQuery) (*PageResult[ResourceType], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data PageResult[ResourceType]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/resource-types/", url.PathEscape(systemID)),
		page.query(), nil, nil, &data)
	return &data, err
}

// BatchCreateResourceType 批量创建资源类型
// POST /api/v1/open/rbac/model/systems/{system_id}/resource-types/
func (c *Client) BatchCreateResourceType(ctx context.Context, systemID string, items []ResourceType) ([]string, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("resource types is required")
	}
	for i, item := range items {
		if err := requireModelID("resource_type.id", item.ID); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if err := requireNonEmpty("resource_type.name", item.Name); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	var data []string
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/resource-types/", url.PathEscape(systemID)),
		nil, items, nil, &data)
	return data, err
}

// UpdateResourceType 更新资源类型
// PUT /api/v1/open/rbac/model/systems/{system_id}/resource-types/{resource_type_id}/
func (c *Client) UpdateResourceType(ctx context.Context, systemID, resourceTypeID string,
	req UpdateResourceTypeRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("resource_type_id", resourceTypeID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/resource-types/%s/",
			url.PathEscape(systemID), url.PathEscape(resourceTypeID)),
		nil, req, nil, nil)
}

// DeleteResourceType 删除资源类型
// DELETE /api/v1/open/rbac/model/systems/{system_id}/resource-types/{resource_type_id}/
func (c *Client) DeleteResourceType(ctx context.Context, systemID, resourceTypeID string) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("resource_type_id", resourceTypeID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/resource-types/%s/",
			url.PathEscape(systemID), url.PathEscape(resourceTypeID)),
		nil, nil, nil, nil)
}

// ListAction 查询操作列表 GET /api/v1/open/rbac/model/systems/{system_id}/actions/
func (c *Client) ListAction(ctx context.Context, systemID string, page PageQuery) (*PageResult[Action], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data PageResult[Action]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/actions/", url.PathEscape(systemID)),
		page.query(), nil, nil, &data)
	return &data, err
}

// BatchCreateAction 批量创建操作 POST /api/v1/open/rbac/model/systems/{system_id}/actions/
func (c *Client) BatchCreateAction(ctx context.Context, systemID string, items []Action) ([]string, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("actions is required")
	}
	for i, item := range items {
		if err := requireModelID("action.id", item.ID); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if err := requireNonEmpty("action.name", item.Name); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if item.ResourceTypeID != "" {
			if err := requireModelID("action.resource_type_id", item.ResourceTypeID); err != nil {
				return nil, fmt.Errorf("items[%d]: %w", i, err)
			}
		}
	}
	var data []string
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/actions/", url.PathEscape(systemID)),
		nil, items, nil, &data)
	return data, err
}

// UpdateAction 更新操作 PUT /api/v1/open/rbac/model/systems/{system_id}/actions/{action_id}/
func (c *Client) UpdateAction(ctx context.Context, systemID, actionID string, req UpdateActionRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("action_id", actionID); err != nil {
		return err
	}
	if err := requireNonEmpty("name", req.Name); err != nil {
		return err
	}
	return c.do(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/actions/%s/",
			url.PathEscape(systemID), url.PathEscape(actionID)),
		nil, req, nil, nil)
}

// DeleteAction 删除操作 DELETE /api/v1/open/rbac/model/systems/{system_id}/actions/{action_id}/
func (c *Client) DeleteAction(ctx context.Context, systemID, actionID string) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("action_id", actionID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/actions/%s/",
			url.PathEscape(systemID), url.PathEscape(actionID)),
		nil, nil, nil, nil)
}

// ListRole 查询角色列表 GET /api/v1/open/rbac/model/systems/{system_id}/roles/
func (c *Client) ListRole(ctx context.Context, systemID string, page PageQuery) (*PageResult[Role], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data PageResult[Role]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/", url.PathEscape(systemID)),
		page.query(), nil, nil, &data)
	return &data, err
}

// BatchCreateRole 批量创建角色 POST /api/v1/open/rbac/model/systems/{system_id}/roles/
func (c *Client) BatchCreateRole(ctx context.Context, systemID string, items []Role) ([]string, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("roles is required")
	}
	for i, item := range items {
		if err := requireModelID("role.id", item.ID); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if err := requireNonEmpty("role.name", item.Name); err != nil {
			return nil, fmt.Errorf("items[%d]: %w", i, err)
		}
		if len(item.Actions) == 0 {
			return nil, fmt.Errorf("items[%d]: actions is required", i)
		}
	}
	var data []string
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/", url.PathEscape(systemID)),
		nil, items, nil, &data)
	return data, err
}

// UpdateRole 更新角色基本信息 PUT /api/v1/open/rbac/model/systems/{system_id}/roles/{role_id}/
func (c *Client) UpdateRole(ctx context.Context, systemID, roleID string, req UpdateRoleRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/%s/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		nil, req, nil, nil)
}

// DeleteRole 删除角色 DELETE /api/v1/open/rbac/model/systems/{system_id}/roles/{role_id}/
func (c *Client) DeleteRole(ctx context.Context, systemID, roleID string) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return err
	}
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/%s/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		nil, nil, nil, nil)
}

// BatchCreateRoleAction 批量添加角色操作
// POST /api/v1/open/rbac/model/systems/{system_id}/roles/{role_id}/actions/
func (c *Client) BatchCreateRoleAction(ctx context.Context, systemID, roleID string, items []RoleAction) ([]string, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("role actions is required")
	}
	var data []string
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/%s/actions/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		nil, items, nil, &data)
	return data, err
}

// BatchDeleteRoleAction 批量删除角色操作
// DELETE /api/v1/open/rbac/model/systems/{system_id}/roles/{role_id}/actions/?ids=
func (c *Client) BatchDeleteRoleAction(ctx context.Context, systemID, roleID string, actionIDs []string) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return err
	}
	if len(actionIDs) == 0 {
		return fmt.Errorf("action ids is required")
	}
	q := url.Values{}
	q.Set("ids", strings.Join(actionIDs, ","))
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/rbac/model/systems/%s/roles/%s/actions/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		q, nil, nil, nil)
}

// ShareListSystem 查询可共享系统列表 GET /api/v1/open/rbac/share/model/systems/
func (c *Client) ShareListSystem(ctx context.Context, page PageQuery) (*PageResult[System], error) {
	var data PageResult[System]
	err := c.do(ctx, http.MethodGet, "/api/v1/open/rbac/share/model/systems/", page.query(), nil, nil, &data)
	return &data, err
}

// ShareRetrieveSystem 查询可共享系统详情
// 网关实际 path 为 /api/v1/open/rabc/share/model/systems/{system_id}/（历史拼写）
func (c *Client) ShareRetrieveSystem(ctx context.Context, systemID string) (*System, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data System
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/rabc/share/model/systems/%s/", url.PathEscape(systemID)),
		nil, nil, nil, &data)
	return &data, err
}
