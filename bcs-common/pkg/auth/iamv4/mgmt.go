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
	"strconv"
)

// ListSpace 查询管理空间 GET /api/v1/open/mgmt/systems/{system_id}/spaces/
func (c *Client) ListSpace(ctx context.Context, systemID string, page PageQuery) (*PageResult[Space], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data PageResult[Space]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/spaces/", url.PathEscape(systemID)),
		page.query(), nil, nil, &data)
	return &data, err
}

// CreateSpace 创建管理空间 POST /api/v1/open/mgmt/systems/{system_id}/spaces/
func (c *Client) CreateSpace(ctx context.Context, systemID string, req CreateSpaceRequest) (*IDData, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("name", req.Name); err != nil {
		return nil, err
	}
	if err := requireNonEmpty("description", req.Description); err != nil {
		return nil, err
	}
	if len(req.Managers) == 0 {
		return nil, fmt.Errorf("managers is required")
	}
	if len(req.PermissionScope) == 0 {
		return nil, fmt.Errorf("permission_scope is required")
	}
	var data IDData
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/spaces/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return &data, err
}

// RetrieveSpace 查询管理空间详情 GET /api/v1/open/mgmt/systems/{system_id}/spaces/{space_id}/
func (c *Client) RetrieveSpace(ctx context.Context, systemID string, spaceID int64) (*Space, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if spaceID <= 0 {
		return nil, fmt.Errorf("space_id is required")
	}
	var data Space
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/spaces/%s/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(spaceID, 10))),
		nil, nil, nil, &data)
	return &data, err
}

// ListGroup 查询用户组 GET /api/v1/open/mgmt/systems/{system_id}/spaces/{space_id}/groups/
func (c *Client) ListGroup(ctx context.Context, systemID string, spaceID int64, page PageQuery) (*PageResult[Group], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if spaceID <= 0 {
		return nil, fmt.Errorf("space_id is required")
	}
	var data PageResult[Group]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/spaces/%s/groups/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(spaceID, 10))),
		page.query(), nil, nil, &data)
	return &data, err
}

// CreateGroup 创建用户组 POST /api/v1/open/mgmt/systems/{system_id}/spaces/{space_id}/groups/
func (c *Client) CreateGroup(ctx context.Context, systemID string, spaceID int64, operator string,
	req CreateGroupRequest) (*IDData, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if spaceID <= 0 {
		return nil, fmt.Errorf("space_id is required")
	}
	header, err := operatorHeader(operator)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("name", req.Name); err != nil {
		return nil, err
	}
	if req.PermissionExpiredAt <= 0 {
		return nil, fmt.Errorf("permission_expired_at is required")
	}
	var data IDData
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/spaces/%s/groups/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(spaceID, 10))),
		nil, req, header, &data)
	return &data, err
}

// ListGroupMember 查询用户组成员 GET /api/v1/open/mgmt/systems/{system_id}/groups/{group_id}/members/
func (c *Client) ListGroupMember(ctx context.Context, systemID string, groupID int64,
	page PageQuery) (*PageResult[GroupMember], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if groupID <= 0 {
		return nil, fmt.Errorf("group_id is required")
	}
	var data PageResult[GroupMember]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/groups/%s/members/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(groupID, 10))),
		page.query(), nil, nil, &data)
	return &data, err
}

// AddGroupMember 添加用户组成员 POST /api/v1/open/mgmt/systems/{system_id}/groups/{group_id}/members/
func (c *Client) AddGroupMember(ctx context.Context, systemID string, groupID int64, operator string,
	req GroupMembersRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if groupID <= 0 {
		return fmt.Errorf("group_id is required")
	}
	header, err := operatorHeader(operator)
	if err != nil {
		return err
	}
	if len(req.Members) == 0 {
		return fmt.Errorf("members is required")
	}
	return c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/groups/%s/members/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(groupID, 10))),
		nil, req, header, nil)
}

// DeleteGroupMember 删除用户组成员 DELETE /api/v1/open/mgmt/systems/{system_id}/groups/{group_id}/members/
func (c *Client) DeleteGroupMember(ctx context.Context, systemID string, groupID int64, operator string,
	req GroupMembersRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if groupID <= 0 {
		return fmt.Errorf("group_id is required")
	}
	header, err := operatorHeader(operator)
	if err != nil {
		return err
	}
	if len(req.Members) == 0 {
		return fmt.Errorf("members is required")
	}
	return c.do(ctx, http.MethodDelete,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/groups/%s/members/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(groupID, 10))),
		nil, req, header, nil)
}

// ListGroupPermission 查询用户组权限 GET /api/v1/open/mgmt/systems/{system_id}/groups/{group_id}/permissions/
func (c *Client) ListGroupPermission(ctx context.Context, systemID string, groupID int64,
	page PageQuery) (*PageResult[GroupPermission], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if groupID <= 0 {
		return nil, fmt.Errorf("group_id is required")
	}
	var data PageResult[GroupPermission]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/groups/%s/permissions/",
			url.PathEscape(systemID), url.PathEscape(strconv.FormatInt(groupID, 10))),
		page.query(), nil, nil, &data)
	return &data, err
}

// ListRoleDisplayResourceTypes 查询角色展示资源层级
// GET /api/v1/open/mgmt/systems/{system_id}/roles/{role_id}/display-resource-types/
func (c *Client) ListRoleDisplayResourceTypes(ctx context.Context, systemID, roleID string) ([]DisplayResourceType, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return nil, err
	}
	var data []DisplayResourceType
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/roles/%s/display-resource-types/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		nil, nil, nil, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UpdateRoleDisplayResourceTypes 全量配置角色展示资源层级
// PUT /api/v1/open/mgmt/systems/{system_id}/roles/{role_id}/display-resource-types/
func (c *Client) UpdateRoleDisplayResourceTypes(ctx context.Context, systemID, roleID string,
	req UpdateDisplayResourceTypesRequest) error {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return err
	}
	if err := requireModelID("role_id", roleID); err != nil {
		return err
	}
	if len(req.DisplayResourceTypes) == 0 {
		return fmt.Errorf("display_resource_types is required")
	}
	return c.do(ctx, http.MethodPut,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/roles/%s/display-resource-types/",
			url.PathEscape(systemID), url.PathEscape(roleID)),
		nil, req, nil, nil)
}

// ListRoleTicket 查询角色申请单据 GET /api/v1/open/mgmt/systems/{system_id}/role-tickets/
func (c *Client) ListRoleTicket(ctx context.Context, systemID string, filter RoleTicketFilter) (*PageResult[RoleTicket], error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	var data PageResult[RoleTicket]
	err = c.do(ctx, http.MethodGet,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/role-tickets/", url.PathEscape(systemID)),
		filter.query(), nil, nil, &data)
	return &data, err
}

// CreateRoleTicket 创建角色申请单据 POST /api/v1/open/mgmt/systems/{system_id}/role-tickets/
func (c *Client) CreateRoleTicket(ctx context.Context, systemID string,
	req CreateRoleTicketRequest) (*CreateRoleTicketResult, error) {
	systemID, err := c.systemID(systemID)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty("applicant", req.Applicant); err != nil {
		return nil, err
	}
	if len(req.Roles) == 0 {
		return nil, fmt.Errorf("roles is required")
	}
	if req.ExpireType != ExpireTypeAt && req.ExpireType != ExpireTypeDuration {
		return nil, fmt.Errorf("expire_type must be %s or %s", ExpireTypeAt, ExpireTypeDuration)
	}
	if req.ExpireValue <= 0 {
		return nil, fmt.Errorf("expire_value is required")
	}
	if err := requireNonEmpty("reason", req.Reason); err != nil {
		return nil, err
	}
	var data CreateRoleTicketResult
	err = c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/v1/open/mgmt/systems/%s/role-tickets/", url.PathEscape(systemID)),
		nil, req, nil, &data)
	return &data, err
}
