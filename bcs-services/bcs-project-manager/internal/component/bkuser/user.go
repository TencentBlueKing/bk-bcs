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

// Package bkuser xxx
package bkuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/parnurzeal/gorequest"
	gocache "github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

var (
	defaultTimeout        = 10
	queryVirtualUsersPath = "/api/v3/open/tenant/virtual-users/-/lookup/"
)

// LookupVirtualUserRsp resp xxx
type LookupVirtualUserRsp struct {
	Data []VirtualUserData `json:"data"`
}

// VirtualUserData virtual user data
type VirtualUserData struct {
	BkUsername  string `json:"bk_username"`
	LoginName   string `json:"login_name"`
	DisplayName string `json:"display_name"`
}

// QueryUserInfoByTenantLoginName query bkUserName by tenant login name
func QueryUserInfoByTenantLoginName(ctx context.Context, tenantId, loginNames string) ([]VirtualUserData, error) {
	// 请求超时时间
	timeout := defaultTimeout
	if config.GlobalConf.BkUser.Timeout != 0 {
		timeout = config.GlobalConf.BkUser.Timeout
	}
	path := fmt.Sprintf("%s%s", config.GlobalConf.BkUser.Host, queryVirtualUsersPath)
	proxy := ""

	req := gorequest.New().Get(path).
		Set(headerkey.TenantIdKey, tenantId).
		Query(fmt.Sprintf("%s=%s", "lookup_field", "login_name")).
		Query(fmt.Sprintf("%s=%s", "lookups", loginNames)).
		SetDebug(config.GlobalConf.BkUser.Debug)

	// 获取返回数据
	body, err := component.Request(*req, timeout, proxy, component.GetAuthAppHeader())
	if err != nil {
		return nil, errorx.NewRequestBkUserErr(err.Error())
	}
	// 解析返回的body
	var resp LookupVirtualUserRsp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse bkUser body error, body: %v", body)
		return nil, err
	}

	return resp.Data, nil
}

const (
	cacheBkUserTenantInfo = "cached_bkuser_tenant"
)

func buildCacheName(keyPrefix string, tenant, name string) string {
	return fmt.Sprintf("%s_%v_%v", keyPrefix, tenant, name)
}

// GetBkUserNameByTenantLoginName get bkUserName by tenant login name
func GetBkUserNameByTenantLoginName(ctx context.Context, tenantId, loginName string, useCache bool) (string, error) {
	if !tenant.IsMultiTenantEnabled() {
		return loginName, nil
	}

	cacheName := buildCacheName(cacheBkUserTenantInfo, tenantId, loginName)
	if useCache {
		val, ok := cache.GetCache().Get(cacheName)
		if ok && val != "" {
			logging.Info("GetBkUserNameByTenantLoginName cacheName:%s, cache exist %+v", cacheName, val)
			if bkUserName, ok1 := val.(string); ok1 {
				return bkUserName, nil
			}
		}
	}

	data, err := QueryUserInfoByTenantLoginName(ctx, tenantId, loginName)
	if err != nil {
		logging.Error("GetBkUserNameByTenantLoginName QueryUserInfoByTenantLoginName failed, err: %v", err)
		return "", err
	}
	if len(data) == 0 {
		logging.Error("GetBkUserNameByTenantLoginName QueryUserInfoByTenantLoginName[%s:%s] failed, data is empty",
			tenantId, loginName)
		return "", fmt.Errorf("data is empty")
	}

	if useCache {
		err = cache.GetCache().Add(cacheName, data[0].BkUsername, gocache.DefaultExpiration)
		if err != nil {
			logging.Error("GetBkUserNameByTenantLoginName cacheName:%s, cache failed %+v", cacheName, err)
		}
	}

	return data[0].BkUsername, nil
}

// GetAuthHeader 获取蓝鲸网关通用认证头
func GetAuthHeader(ctx context.Context) (map[string]string, error) {
	var (
		tenantId   = tenant.GetTenantIdFromContext(ctx)
		bkUserName = config.GlobalConf.App.BkUsername
	)

	// 多租户模式下 通过loginName获取租户下的bkUserName
	if tenant.IsMultiTenantEnabled() && bkUserName != "" {
		userName, err := GetBkUserNameByTenantLoginName(ctx, tenantId, bkUserName, true)
		if err != nil {
			return nil, fmt.Errorf("get bkUserName by tenant failed: %v", err)
		}
		bkUserName = userName

		logging.Info("GetAuthHeader get bkUserName by tenant, tenantId: %s, loginName: %s, bkUserName: %s",
			tenantId, config.GlobalConf.App.BkUsername, bkUserName)
	}

	return map[string]string{
		"Content-Type": "application/json",
		"X-Bkapi-Authorization": fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s", "bk_username": "%s"}`,
			config.GlobalConf.App.Code, config.GlobalConf.App.Secret, bkUserName),
		headerkey.TenantIdKey: tenantId,
	}, nil
}
