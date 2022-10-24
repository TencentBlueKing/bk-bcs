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

package pkg

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

const (
	grantPermissionUrl    = "/v1/permissions"
	getPermissionUrl      = "/v1/permissions"
	revokePermissionUrl   = "/v1/permissions"
	verifyPermissionUrl   = "/v1/permissions/verify"
	verifyPermissionV2Url = "/v2/permissions/verify"
)

// GrantPermissionResponse defines the response of grant permission
type GrantPermissionResponse struct {
	Result  bool             `json:"result"`
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *PermissionsResp `json:"data"`
}

// GrantPermission request grant permission from bcs-user-manager
func (c *UserManagerClient) GrantPermission(reqBody string) (*GrantPermissionResponse, error) {
	reqForm := new(types.BcsPermission)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(grantPermissionUrl, http.MethodPost, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "grant permission with '%s' failed", reqForm)
	}
	resp := new(GrantPermissionResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "grant permission unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

//GetPermissionResponse defines the response of get permission
type GetPermissionResponse struct {
	Result  bool               `json:"result"`
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    []*PermissionsResp `json:"data"`
}

// PermissionsResp query response for
type PermissionsResp struct {
	ResourceType string `json:"resource_type"`
	Resource     string `json:"resource"`
	Role         string `json:"role"`
}

// GetPermissionForm request form
type GetPermissionForm struct {
	UserName     string `json:"user_name" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
}

// GetPermission request get permission from bcs-user-manager
func (c *UserManagerClient) GetPermission(reqBody string) (*GetPermissionResponse, error) {
	reqForm := new(GetPermissionForm)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "get permission form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(getPermissionUrl, http.MethodGet, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "get permission with '%s' failed", reqForm)
	}
	resp := new(GetPermissionResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get permission unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// RevokePermissionResponse defines the response of revoke permission
type RevokePermissionResponse struct {
	Result  bool             `json:"result"`
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *PermissionsResp `json:"data"`
}

// RevokePermission request revoke permission from bcs-user-manager
func (c *UserManagerClient) RevokePermission(reqBody string) (*RevokePermissionResponse, error) {
	reqForm := new(types.BcsPermission)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "revoke permission form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(revokePermissionUrl, http.MethodDelete, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "revoke permission with '%s' failed", reqForm)
	}
	resp := new(RevokePermissionResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "revoke permission unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// VerifyPermissionForm request form for permission
type VerifyPermissionForm struct {
	UserToken    string `json:"user_token" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	Resource     string `json:"resource"`
	Action       string `json:"action" validate:"required"`
}

// VerifyPermissionResponseBody http verify response
type VerifyPermissionResponseBody struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

// VerifyPermissionResponse defines the response of verify permission
type VerifyPermissionResponse struct {
	Result  bool                          `json:"result"`
	Code    int                           `json:"code"`
	Message string                        `json:"message"`
	Data    *VerifyPermissionResponseBody `json:"data"`
}

// VerifyPermission request verify permission from bcs-user-manager
func (c *UserManagerClient) VerifyPermission(reqBody string) (*VerifyPermissionResponse, error) {
	reqForm := new(VerifyPermissionForm)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "release tke cidrs form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(verifyPermissionUrl, http.MethodGet, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "verify permission with '%s' failed", reqForm)
	}
	resp := new(VerifyPermissionResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "verify permission unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// VerifyPermissionReq for permission v2 request
type VerifyPermissionReq struct {
	UserToken    string `json:"user_token" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	// clusterType mesos/k8s when ResourceType="cluster"
	ClusterType string `json:"cluster_type"`
	ProjectID   string `json:"project_id"`
	ClusterID   string `json:"cluster_id"`
	RequestURL  string `json:"request_url"`

	Resource string `json:"resource"`
	Action   string `json:"action" validate:"required"`
}

// VerifyPermissionV2Response defines the response of verify permission v2
type VerifyPermissionV2Response struct {
	Result  bool                          `json:"result"`
	Code    int                           `json:"code"`
	Message string                        `json:"message"`
	Data    *VerifyPermissionResponseBody `json:"data"`
}

// VerifyPermissionV2 request verify permission from bcs-user-manager v2
func (c *UserManagerClient) VerifyPermissionV2(reqBody string) (*VerifyPermissionV2Response, error) {
	reqForm := new(VerifyPermissionReq)
	if err := json.Unmarshal([]byte(reqBody), reqForm); err != nil {
		return nil, errors.Wrapf(err, "release tke cidrs form json unmarshal failed with '%s'", reqBody)
	}
	bs, err := c.do(verifyPermissionV2Url, http.MethodGet, nil, reqForm)
	if err != nil {
		return nil, errors.Wrapf(err, "verify permission v2 with '%s' failed", reqForm)
	}
	resp := new(VerifyPermissionV2Response)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "verify permission v2 unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}
