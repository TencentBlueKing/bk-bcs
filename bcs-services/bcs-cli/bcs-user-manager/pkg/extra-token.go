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
	"time"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

const (
	getTokenByUserAndClusterIDUrl = "/v1/tokens/extra/getClusterUserToken" // nolint
)

// TokenStatus is a enum for token status.
type TokenStatus uint8

// ExtraTokenResponse is the response of extra token
type ExtraTokenResponse struct {
	UserName  string       `json:"username"`
	Token     string       `json:"token"`
	Status    *TokenStatus `json:"status,omitempty"`
	ExpiredAt *time.Time   `json:"expired_at"` // nil means never expired
}

// GetTokenByUserAndClusterIDRequest defines the request of GetTokenByUserAndClusterID
type GetTokenByUserAndClusterIDRequest struct {
	UserName   string `json:"username"`
	ClusterID  string `json:"cluster_id"`
	BusinessID string `json:"business_id"`
}

// GetTokenByUserAndClusterIDResponse defines the response of GetTokenByUserAndClusterID
type GetTokenByUserAndClusterIDResponse struct {
	Result  bool                `json:"result"`
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *ExtraTokenResponse `json:"data"`
}

// GetTokenByUserAndClusterID request get token by user and cluster id from bcs-user-manager
func (c *UserManagerClient) GetTokenByUserAndClusterID(request GetTokenByUserAndClusterIDRequest) (*GetTokenByUserAndClusterIDResponse, error) {

	queryParams, err := query.Values(request)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := c.do(getTokenByUserAndClusterIDUrl, http.MethodGet, queryParams, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get token by '%v' failed", queryParams)
	}
	resp := new(GetTokenByUserAndClusterIDResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get token by user and clusterID unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}
