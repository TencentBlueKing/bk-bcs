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

package utils

import (
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
)

// Authenticate only authenticate admin user token
func Authenticate(req *http.Request) bool {
	var token string
	authHeaderList := req.Header["Authorization"]
	if len(authHeaderList) > 0 {
		authHeader := strings.Split(authHeaderList[0], " ")
		if len(authHeader) == 2 && authHeader[0] == "Bearer" {
			token = strings.TrimSpace(authHeader[1])
		}
	}
	// if not specified token, authenticate failed
	if token == "" {
		return false
	}

	u := models.BcsUser{
		UserToken: token,
	}
	user := sqlstore.GetUserByCondition(&u)
	if user == nil {
		return false
	}

	// only authenticate admin user
	if user.UserType == sqlstore.AdminUser && !user.HasExpired() {
		return true
	}
	return false
}
