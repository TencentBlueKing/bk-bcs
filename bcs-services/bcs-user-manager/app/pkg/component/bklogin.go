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

package component

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	apputils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

// BKLoginAPIServer bk login api server
var BKLoginAPIServer = os.Getenv("BKLOGIN_API_SERVER")

// GetBKTokenUserInfoResp get bk token user info response
type GetBKTokenUserInfoResp struct {
	Data BKUserInfo `json:"data"`
}

// GetBKTokenUserInfoErrResp get bk token user info error response
type GetBKTokenUserInfoErrResp struct {
	Error BKTokenUserInfoErr `json:"error"`
}

// BKTokenUserInfoErr bk token user info error
type BKTokenUserInfoErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BKUserInfo bk user info
type BKUserInfo struct {
	BKUsername  string `json:"bk_username"`
	TenantID    string `json:"tenant_id"`
	DisplayName string `json:"display_name"`
	Language    string `json:"language"`
	TimeZone    string `json:"time_zone"`
}

// GetBKUserInfo get bk user info by bk token
func GetBKUserInfo(ctx context.Context, tenantID, bkToken string) (*BKUserInfo, error) {
	if bkToken == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/%s", BKLoginAPIServer, "login/api/v3/open/bk-tokens/userinfo/")
	authInfo, err := GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}
	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeader(apputils.HeaderTenantID, tenantID).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetQueryParam("bk_token", bkToken).
		Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		errResp := &GetBKTokenUserInfoErrResp{}
		if err = json.Unmarshal(resp.Body(), errResp); err == nil &&
			(errResp.Error.Code != "" || errResp.Error.Message != "") {
			return nil, errors.Errorf("get bk token userinfo failed, code: %s, message: %s",
				errResp.Error.Code, errResp.Error.Message)
		}
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &GetBKTokenUserInfoResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	return &result.Data, nil
}
