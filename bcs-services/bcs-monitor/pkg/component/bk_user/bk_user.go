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

// Package bkuser user
package bkuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// GetTenantAPIAuthorization generate bk user api auth header, X-Bkapi-Authorization
func GetTenantAPIAuthorization(ctx context.Context, username string) (string, error) {
	if username == "" {
		username = config.G.Base.BKUsername
	}
	if !config.G.Base.EnableMultiTenant {
		return component.GetBKAPIAuthorization(username)
	}

	// get bk_username from bk user api
	username = "bk_admin"
	user, err := LookupVirtualUsers(ctx, username)
	if err != nil {
		return "", err
	}
	if len(user) != 1 {
		return "", errors.New("user not found")
	}
	auth := &component.AuthInfo{
		BkAppCode:   config.G.Base.AppCode,
		BkAppSecret: config.G.Base.AppSecret,
		BkUserName:  user[0].BKUsername,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// LookupVirtualUsersResp lookup virtual users response
type LookupVirtualUsersResp struct {
	Data []BKUser `json:"data"`
}

// BKUser bk user
type BKUser struct {
	BKUsername  string `json:"bk_username"`
	LoginName   string `json:"login_name"`
	DisplayName string `json:"display_name"`
}

// LookupVirtualUsers lookup virtual users
func LookupVirtualUsers(ctx context.Context, username string) ([]BKUser, error) {
	url := fmt.Sprintf("%s/%s", config.G.BKUser.APIServer, "api/v3/open/tenant/virtual-users/-/lookup/")
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader(utils.TenantIDHeaderKey, utils.GetTenantIDFromContext(ctx)).
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		SetQueryParam("lookups", username).
		SetQueryParam("lookup_field", "login_name").
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &LookupVirtualUsersResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	return result.Data, nil
}

// ListTenantResp list tenant response
type ListTenantResp struct {
	Data []BkTenant `json:"data"`
}

// BkTenant bk tenant
type BkTenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ListTenant list bk tenant
func ListTenant(ctx context.Context) ([]BkTenant, error) {
	url := fmt.Sprintf("%s/%s", config.G.BKUser.APIServer, "api/v3/open/tenants/")
	// generate bk api auth header, X-Bkapi-Authorization
	authInfo, err := component.GetBKAPIAuthorization("")
	if err != nil {
		return nil, err
	}
	resp, err := component.GetClient().R().
		SetContext(ctx).
		SetHeader(utils.TenantIDHeaderKey, "system").
		SetHeader("X-Bkapi-Authorization", authInfo).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		Get(url)

	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := &ListTenantResp{}
	err = json.Unmarshal(resp.Body(), result)
	if err != nil {
		return nil, errors.Errorf("unmarshal resp body error, %s", err.Error())
	}

	return result.Data, nil
}
