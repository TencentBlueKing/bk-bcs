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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	listTenants = "/api/v3/open/tenants/"
)

// ListTenantsRsp resp xxx
type ListTenantsRsp struct {
	Data []TenantData `json:"data"`
}

// TenantData tenant data
type TenantData struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ListTenants list system tenants
func ListTenants(ctx context.Context, tenantId string) ([]TenantData, error) {
	// 请求超时时间
	timeout := defaultTimeout
	if config.GlobalConf.BkUser.Timeout != 0 {
		timeout = config.GlobalConf.BkUser.Timeout
	}
	path := fmt.Sprintf("%s%s", config.GlobalConf.BkUser.Host, listTenants)
	proxy := ""

	req := gorequest.New().Get(path).
		Set(headerkey.TenantIdKey, tenantId).
		SetDebug(config.GlobalConf.BkUser.Debug)

	// 获取返回数据
	body, err := component.Request(*req, timeout, proxy, component.GetAuthAppHeader())
	if err != nil {
		return nil, errorx.NewRequestBkUserErr(err.Error())
	}
	// 解析返回的body
	var resp ListTenantsRsp
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		logging.Error("parse bkUser body error, body: %v", body)
		return nil, err
	}

	return resp.Data, nil
}

// 租户管理审批人，初始化也可以
