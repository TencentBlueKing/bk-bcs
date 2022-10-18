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
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	//tokenModels "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/token"
)

const (
	createTokenUrl       = "/v1/tokens"
	getTokenUrl          = "/v1/users/%s/tokens"
	deleteTokenUrl       = "/v1/tokens/%s"
	updateTokenUrl       = "/v1/tokens/%s"
	createTempTokenUrl   = "/v1/tokens/temp"
	createClientTokenUrl = "/v1/tokens/client"
)

// CreateTokenResponse defines the response of create token
type CreateTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// CreateToken request create token from bcs-user-manager
func (c *UserManagerClient) CreateToken(reqBody string) (*CreateTokenResponse, error) {
	bs, err := c.do(createTokenUrl, http.MethodPost, nil, []byte(reqBody))
	if err != nil {
		return nil, errors.Wrapf(err, "create token with '%s' failed", reqBody)
	}
	resp := new(CreateTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// GetTokenResponse defines the response of grant permission
type GetTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// GetToken request get token from bcs-user-manager
func (c *UserManagerClient) GetToken(userName string) (*GetTokenResponse, error) {
	bs, err := c.do(fmt.Sprintf(getTokenUrl, userName), http.MethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get token with userName '%s' failed", userName)
	}
	resp := new(GetTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// DeleteTokenResponse defines the response of delete token
type DeleteTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// DeleteToken request delete token from bcs-user-manager
func (c *UserManagerClient) DeleteToken(token string) (*DeleteTokenResponse, error) {
	bs, err := c.do(fmt.Sprintf(deleteTokenUrl, token), http.MethodDelete, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "delete token with '%s' failed", token)
	}
	resp := new(DeleteTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "delete token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// UpdateTokenResponse defines the response of update token
type UpdateTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// UpdateToken request update token from bcs-user-manager
func (c *UserManagerClient) UpdateToken(token string) (*UpdateTokenResponse, error) {
	bs, err := c.do(fmt.Sprintf(updateTokenUrl, token), http.MethodPut, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "update token with '%s' failed", token)
	}
	resp := new(UpdateTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "update token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CreateTempTokenResponse defines the response of create temp token
type CreateTempTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// CreateTempToken request create temp token from bcs-user-manager
func (c *UserManagerClient) CreateTempToken(reqBody string) (*CreateTempTokenResponse, error) {
	bs, err := c.do(createTempTokenUrl, http.MethodPost, nil, []byte(reqBody))
	if err != nil {
		return nil, errors.Wrapf(err, "create temp token with '%s' failed", reqBody)
	}
	resp := new(CreateTempTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create temp token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}

// CreateClientTokenResponse defines the response of create client token
type CreateClientTokenResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	//Data    *tokenModels.TokenResp `json:"data"`
}

// CreateClientToken request create client token from bcs-user-manager
func (c *UserManagerClient) CreateClientToken(reqBody string) (*CreateClientTokenResponse, error) {
	bs, err := c.do(createClientTokenUrl, http.MethodPost, nil, []byte(reqBody))
	if err != nil {
		return nil, errors.Wrapf(err, "create client token with '%s' failed", reqBody)
	}
	resp := new(CreateClientTokenResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create client token unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}
