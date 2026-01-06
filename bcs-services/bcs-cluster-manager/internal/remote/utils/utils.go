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

// Package utils xxx
package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	gocache "github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	bkuser "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/bk_user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
)

// GetGatewayAuthAndTenantInfo generate blueking gateway auth and tenant info, user is bkUserName
func GetGatewayAuthAndTenantInfo(ctx context.Context, auth *types.AuthInfo, user string) (string, string, error) {
	tenantId := tenant.GetTenantIdFromContext(ctx)

	if options.GetGlobalCMOptions().TenantConfig.EnableMultiTenantMode {
		// 多租户模式下，优先使用传入的user，否则根据租户获取用户名
		if user != "" {
			auth.BkUserName = user
		} else if auth.BkUserName != "" {
			bkUserName, err := GetBkUserNameByTenantLoginName(ctx, tenantId, auth.BkUserName, true)
			if err != nil {
				return "", "", fmt.Errorf("get bkUserName by tenant failed: %v", err)
			}
			auth.BkUserName = bkUserName
		}
	} else {
		// 非多租户模式，直接使用传入的user.否则直接使用auth bkUserName
		if user != "" {
			auth.BkUserName = user
		}
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", "", err
	}
	return string(userAuth), tenantId, nil
}

const (
	cacheBkUserTenantInfo = "cached_bkuser_tenant"
)

func buildCacheName(keyPrefix string, tenant, name string) string {
	return fmt.Sprintf("%s_%v_%v", keyPrefix, tenant, name)
}

// GetBkUserNameByTenantLoginName get bkUserName by tenant login name
func GetBkUserNameByTenantLoginName(ctx context.Context, tenantId, loginName string, useCache bool) (string, error) {
	if !options.GetGlobalCMOptions().TenantConfig.EnableMultiTenantMode {
		return loginName, nil
	}

	cacheName := buildCacheName(cacheBkUserTenantInfo, tenantId, loginName)
	if useCache {
		val, ok := cache.GetCache().Get(cacheName)
		if ok && val != "" {
			blog.Infof("GetBkUserNameByTenantLoginName cacheName:%s, cache exist %+v", cacheName, val)
			if bkUserName, ok1 := val.(string); ok1 {
				return bkUserName, nil
			}
		}
	}

	data, err := bkuser.GetBkUserClient().QueryUserInfoByTenantLoginName(ctx, tenantId, loginName)
	if err != nil {
		blog.Errorf("GetBkUserNameByTenantLoginName QueryUserInfoByTenantLoginName failed, err: %v", err)
		return "", err
	}

	if len(data) == 0 {
		return "", fmt.Errorf("GetBkUserNameByTenantLoginName QueryUserInfoByTenantLoginName empty")
	}

	if useCache {
		err = cache.GetCache().Add(cacheName, data[0].BkUsername, gocache.DefaultExpiration)
		if err != nil {
			blog.Errorf("GetBkUserNameByTenantLoginName cacheName:%s, cache failed %+v", cacheName, err)
		}
	}

	return data[0].BkUsername, nil
}
