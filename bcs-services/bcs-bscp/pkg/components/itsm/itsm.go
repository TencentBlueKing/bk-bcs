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

// Package itsm xxx
package itsm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
)

// GetAuthHeader 获取蓝鲸网关通用认证头
func GetAuthHeader() string {
	return fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s", "bk_username": "%s"}`,
		cc.DataService().Esb.AppCode, cc.DataService().Esb.AppSecret, cc.DataService().Esb.User)
}

// ItsmRequest itsm request
// nolint revive
func ItsmRequest(ctx context.Context, method, reqURL string, data interface{}) ([]byte, error) {

	client := components.GetClient().R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Bkapi-Authorization", GetAuthHeader())

	switch method {
	case http.MethodGet:
		resp, err := client.Get(reqURL)
		if err != nil {
			return nil, err
		}
		return resp.Body(), nil
	case http.MethodPost:
		resp, err := client.SetBody(data).Post(reqURL)
		if err != nil {
			return nil, err
		}
		return resp.Body(), nil
	default:
		return nil, fmt.Errorf("invalid method: %s", method)
	}
}
