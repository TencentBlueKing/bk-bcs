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

	userManagerModels "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

const (
	getAdminUserUrl = "/v1/users/admin/%s"
)

// GetAdminUserResponse defines the response of get admin user
type GetAdminUserResponse struct {
	Result  bool                       `json:"result"`
	Code    int                        `json:"code"`
	Message string                     `json:"message"`
	Data    *userManagerModels.BcsUser `json:"data"`
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
