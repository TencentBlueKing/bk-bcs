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
 */

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	getAdminUserUrl    = "/v1/users/admin/%s"
	createAdminUserUrl = "/v1/users/admin/%s"

	getSaasUserUrl      = "/v1/users/saas/%s"
	createSaasUserUrl   = "/v1/users/saas/%s"
	refreshSaasTokenUrl = "/v1/users/saas/%s/refresh" // nolint
	getUserInfo         = "/v1/users/info"

	getPlainUserUrl      = "/v1/users/plain/%s"
	createPlainUserUrl   = "/v1/users/plain/%s"
	refreshPlainTokenUrl = "/v1/users/plain/%s/refresh/%s" // nolint

	timeFormatter = "2006-01-02 15:04:05"
)

// GetAdminUserResponse defines the response of get admin user
type GetAdminUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// GetAdminUser request admin user from bcs-user-manager
func (c *UserManagerClient) GetAdminUser(userName string) (*GetAdminUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(getAdminUserUrl, userName), http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get admin user with '%s' failed", userName)
	}
	resp := new(GetAdminUserResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get admin user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CreateAdminUserResponse defines the response of get admin user
type CreateAdminUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// BcsUser user table
type BcsUser struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	Name      string     `json:"name" gorm:"not null"`
	UserType  uint       `json:"user_type"`
	UserToken string     `json:"user_token" gorm:"unique;size:64"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamp null;default:null"` // 用户创建时间
	UpdatedAt time.Time  `json:"updated_at" gorm:"type:timestamp null;default:null"` // user-token刷新时间
	ExpiresAt time.Time  `json:"expires_at" gorm:"type:timestamp null;default:null"` // user-token过期时间
	DeletedAt *time.Time `json:"deleted_at" gorm:"type:timestamp null;default:null"` // user-token删除时间
}

// GetDeletedAtStr return str  "" or str converted from user delete time
func (u *BcsUser) GetDeletedAtStr() string {
	if u != nil && u.DeletedAt != nil {
		return u.DeletedAt.Format(timeFormatter)
	}
	return ""
}

// CreateAdminUser request admin user from bcs-user-manager
func (c *UserManagerClient) CreateAdminUser(userName string) (*CreateAdminUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(createAdminUserUrl, userName), http.MethodPost, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create admin user with '%s' failed", userName)
	}
	resp := new(CreateAdminUserResponse)
	// resp.Data.Name = userName
	// resp.Data.UserType = userManagerModels.AdminUser
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create admin user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// GetSaasUserResponse defines the response of get saas user
type GetSaasUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// GetSaasUser request saas user from bcs-user-manager
func (c *UserManagerClient) GetSaasUser(userName string) (*GetSaasUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(getSaasUserUrl, userName), http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get saas user with '%s' failed", userName)
	}
	resp := new(GetSaasUserResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get saas user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CreateSaasUserResponse defines the response of create saas user
type CreateSaasUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// CreateSaasUser request saas user from bcs-user-manager
func (c *UserManagerClient) CreateSaasUser(userName string) (*CreateSaasUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(createSaasUserUrl, userName), http.MethodPost, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create saas user with '%s' failed", userName)
	}
	resp := new(CreateSaasUserResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create saas user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// RefreshSaasTokenResponse defines the response of refresh saas user token
type RefreshSaasTokenResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// RefreshSaasToken request refresh saas user token  from bcs-user-manager
func (c *UserManagerClient) RefreshSaasToken(userName string) (*RefreshSaasTokenResponse, error) {
	bs, err := c.do(fmt.Sprintf(refreshSaasTokenUrl, userName), http.MethodPut, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "refresh saas user token with '%s' failed", userName)
	}
	resp := new(RefreshSaasTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "refresh saas user token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// GetPlainUserResponse defines the response of get plain user
type GetPlainUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// GetPlainUser request Plain user from bcs-user-manager
func (c *UserManagerClient) GetPlainUser(userName string) (*GetPlainUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(getPlainUserUrl, userName), http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get Plain user with '%s' failed", userName)
	}
	resp := new(GetPlainUserResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get Plain user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CreatePlainUserResponse defines the response of create Plain user
type CreatePlainUserResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// CreatePlainUser request Plain user from bcs-user-manager
func (c *UserManagerClient) CreatePlainUser(userName string) (*CreatePlainUserResponse, error) {
	bs, err := c.do(fmt.Sprintf(createPlainUserUrl, userName), http.MethodPost, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "create Plain user with '%s' failed", userName)
	}
	resp := new(CreatePlainUserResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create Plain user unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// RefreshPlainTokenResponse defines the response of refresh Plain user token
type RefreshPlainTokenResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    *BcsUser `json:"data"`
}

// RefreshPlainToken request refresh Plain user token  from bcs-user-manager
func (c *UserManagerClient) RefreshPlainToken(userName, expireTime string) (*RefreshPlainTokenResponse, error) {
	bs, err := c.do(fmt.Sprintf(refreshPlainTokenUrl, userName, expireTime), http.MethodPut, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "refresh Plain user token with  userName '%s' ,expireTime '%s' failed",
			userName, expireTime)
	}
	resp := new(RefreshPlainTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "refresh Plain user token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// UserInfo for username
type UserInfo struct {
	UserName string `json:"username,omitempty"`
	URL      string `json:"avatar_url,omitempty"`
}

// UserInfoResponse defines the response of GetUserInfo
type UserInfoResponse struct {
	Result  bool      `json:"result"`
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    *UserInfo `json:"data"`
}

// GetUserInfo get user name from bcs-user-manager
func (c *UserManagerClient) GetUserInfo() (*UserInfoResponse, error) {
	data, err := c.do(getUserInfo, http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "validate user auth token failed: ") // nolint
	}
	resp := new(UserInfoResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, errors.Wrapf(err, "Get UserInfo response ", // nolint
			string(data))
	}
	return resp, nil
}
