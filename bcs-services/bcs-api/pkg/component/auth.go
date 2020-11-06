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

package component

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
)

const (
	appCodeKey   = "X-BK-APP-CODE"
	appSecretKey = "X-BK-APP-SECRET"
)

// PaaSAuth auth configuration for bk-app
type PaaSAuth struct {
	host      string
	appCode   string
	appSecret string
	version   string
}

// NewPaaSAuth construct Auth according configuration file/command line option
func NewPaaSAuth() *PaaSAuth {
	bKIamAuth := config.BKIamAuth
	return &PaaSAuth{
		host:      bKIamAuth.BKIamAuthHost,
		appCode:   bKIamAuth.BKIamAuthAppCode,
		appSecret: bKIamAuth.BKIamAuthAppSecret,
		version:   bKIamAuth.Version,
	}
}

// VerifyAccessTokenForIeod refresh access token
func (a *PaaSAuth) VerifyAccessTokenForIeod(accessToken string) (bool, map[string]interface{}, error) {

	url := fmt.Sprintf("%s/oauth/token", a.host)

	params := map[string]string{
		"access_token": accessToken,
	}

	result, err := HTTPGet(url, params)
	if err != nil {
		return false, nil, err
	}

	return true, result.Data, nil
}

// VerifyAccessTokenForEe verify access token according bk app info
func (a *PaaSAuth) VerifyAccessTokenForEe(accessToken string) (bool, map[string]interface{}, error) {
	var url string
	if a.version == "3" {
		url = fmt.Sprintf("%s/api/v1/auth/access-tokens/verify", a.host)
	} else {
		url = fmt.Sprintf("%s/bkiam/api/v1/auth/access-tokens/verify", a.host)
	}

	data := map[string]interface{}{
		"access_token": accessToken,
	}

	header := map[string]string{
		appCodeKey:   a.appCode,
		appSecretKey: a.appSecret,
	}

	result, err := HTTPPostToBkIamAuth(url, data, header)
	if err != nil {
		return false, nil, err
	}

	return true, result.Data.Identity, nil
}
