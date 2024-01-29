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

// Package bknotice provides bknotice client.
package bknotice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
)

type registerSystemResp struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RegisterSystem 注册系统到通知中心
func RegisterSystem(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/register/", cc.ApiServer().BKNotice.Host)

	authHeader := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\"}",
		cc.ApiServer().Esb.AppCode, cc.ApiServer().Esb.AppSecret)

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetHeader("X-Bkapi-Authorization", authHeader).
		Post(url)

	if err != nil {
		return err
	}

	resigerResp := &registerSystemResp{}
	if err := json.Unmarshal(resp.Body(), resigerResp); err != nil {
		return err
	}

	if resigerResp.Code != 0 {
		return fmt.Errorf("register system to bknotice failed, code: %d, message: %s",
			resigerResp.Code, resigerResp.Message)
	}
	return nil
}
