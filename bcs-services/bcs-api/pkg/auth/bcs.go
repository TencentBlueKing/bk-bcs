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

package auth

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/component"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
)

// VerifyAccessTokenAndCreateUser verifies the given access_token, if the token is valid, create and return the user
// object, otherwise return error message instead.
func VerifyAccessTokenAndCreateUser(accessToken string) (*m.User, error) {
	// 1. validate accessToken
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken can not be empty")
	}

	paasAuth := component.NewPaaSAuth()

	var userID, userType interface{}
	var ok bool
	edition := config.Edition

	if edition == "ieod" {
		isValid, userInfo, err := paasAuth.VerifyAccessTokenForIeod(accessToken)
		if err != nil {
			return nil, fmt.Errorf("validate access_token failed: %s", err.Error())
		}
		if !isValid {
			return nil, fmt.Errorf("access_token is invalid: %s", err.Error())
		}
		userID, ok = userInfo["user_id"]
		if !ok {
			return nil, fmt.Errorf("get user_id failed")
		}
		userType, ok = userInfo["user_type"]
		if !ok {
			return nil, fmt.Errorf("get user_type failed")
		}
	} else if edition == "ee" {
		isValid, userInfo, err := paasAuth.VerifyAccessTokenForEe(accessToken)
		if err != nil {
			return nil, fmt.Errorf("validate access_token failed: %s", err.Error())
		}
		if !isValid {
			return nil, fmt.Errorf("access_token is invalid: %s", err.Error())
		}
		userID, ok = userInfo["username"]
		if !ok {
			return nil, fmt.Errorf("get user_id failed")
		}
		userType, ok = userInfo["user_type"]
		if !ok {
			return nil, fmt.Errorf("get user_type failed")
		}
	} else {
		return nil, fmt.Errorf("invalid bcs-api edition")
	}

	if userType.(string) == "" || userID.(string) == "" {
		return nil, fmt.Errorf("access_token is invalid, userType or userId is empty")
	}

	// 2. get or create user
	user, err := sqlstore.GetOrCreateUser(m.ExternalUserSourceTypeBCS, userID.(string), userType.(string))
	if err != nil {
		return nil, fmt.Errorf("create user failed: %s", err.Error())
	}
	return user, nil
}
