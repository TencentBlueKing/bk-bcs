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
	"time"

	"github.com/pkg/errors"
)

const (
	getTokenByUserAndClusterIDUrl = "/v1/tokens/extra/getClusterUserToken"
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

// GetTokenByUserAndClusterIDResponse defines the response of GetTokenByUserAndClusterID
type GetTokenByUserAndClusterIDResponse struct {
	Result  bool                `json:"result"`
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *ExtraTokenResponse `json:"data"`
}

// GetTokenByUserAndClusterID request get token by user and cluster id from bcs-user-manager
func (c *UserManagerClient) GetTokenByUserAndClusterID(userName, clusterId, businessId string) (*GetTokenByUserAndClusterIDResponse, error) {
	queryURL := make(map[string]string)
	queryURL["username"] = userName
	queryURL["cluster_id"] = clusterId
	queryURL["business_id"] = businessId
	bs, err := c.do(getTokenByUserAndClusterIDUrl, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get token by userName = %s and clusterID = %s and businessId=%s, failed", userName, clusterId, businessId)
	}
	resp := new(GetTokenByUserAndClusterIDResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "get token by user and clusterID unmarshal failed with response '%s'", string(bs))
	}
	return resp, nil
}
