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

// Package bkuser bk user related functions
package bkuser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
)

// bkUserClient global bkUser client
var bkUserClient *Client

// SetBkUserClient set bkUser client
func SetBkUserClient(options Options) error {
	cli, err := NewBkUserClient(options)
	if err != nil {
		return err
	}

	bkUserClient = cli
	return nil
}

// GetBkUserClient get bkUser client
func GetBkUserClient() *Client {
	return bkUserClient
}

// Client for bkUser
type Client struct {
	types.CommonClient
	userAuth string
}

// NewBkUserClient create bkUser client
func NewBkUserClient(options Options) (*Client, error) {
	c := &Client{
		CommonClient: types.CommonClient{
			AppCode:   options.AppCode,
			AppSecret: options.AppSecret,
			Server:    options.Server,
			Debug:     options.Debug,
		},
	}

	auth := &AuthInfo{
		BkAppCode:   options.AppCode,
		BkAppSecret: options.AppSecret,
	}
	userAuth, err := json.Marshal(auth)
	if err != nil {
		return nil, err
	}
	c.userAuth = string(userAuth)

	return c, nil
}

// QueryUserInfoByTenantLoginName query user info by tenant login name
func (c *Client) QueryUserInfoByTenantLoginName(ctx context.Context, tenantId,
	loginNames string) ([]VirtualUserData, error) {
	if c == nil {
		return nil, fmt.Errorf("bkUser client not init")
	}

	var (
		handler = "QueryUserInfoByTenantLoginName"
		path    = fmt.Sprintf("%s/api/v3/open/tenant/virtual-users/-/lookup/", c.Server)
	)

	data := &LookupVirtualUserRsp{}

	start := time.Now()
	rsp, _, errs := gorequest.New().
		Timeout(types.DefaultTimeOut).
		Get(path).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		Set(common.BkTenantIdHeaderKey, tenantId).
		Query(fmt.Sprintf("%s=%s", "lookup_field", "login_name")).
		Query(fmt.Sprintf("%s=%s", "lookups", loginNames)).
		SetDebug(c.Debug).
		EndStruct(&data)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("bkuser", handler, "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api blueking BkUser QueryUserInfoByTenantLoginName failed: %v", errs[0])
		return nil, errs[0]
	}

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		metrics.ReportLibRequestMetric("bkuser", handler, "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api blueking BkUser QueryUserInfoByTenantLoginName failed with status: %d",
			rsp.StatusCode)
		return nil, fmt.Errorf("api failed return statusCode: %d", rsp.StatusCode)
	}
	metrics.ReportLibRequestMetric("bkuser", handler, "http", metrics.LibCallStatusOK, start)
	blog.Infof("call api blueking BkUser QueryUserInfoByTenantLoginName with url(%s) successfully", path)

	return data.Data, nil
}
