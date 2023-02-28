/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package itsm xxx
package itsm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	operateTicketPath = "/v2/itsm/operate_ticket/"
)

// OperateTicketResp itsm operate ticket resp
type OperateTicketResp struct {
	Code    int         `json:"code"`
	Result  bool        `json:"result"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// WithdrawTicket withdraw itsm ticket
func WithdrawTicket(username, sn string) error {
	itsmConf := config.GlobalConf.ITSM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", itsmConf.GatewayHost, operateTicketPath)
	headers := map[string]string{"Content-Type": "application/json"}
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"bk_app_code":    config.GlobalConf.App.Code,
			"bk_app_secret":  config.GlobalConf.App.Secret,
			"sn":             sn,
			"operator":       username,
			"action_type":    "WITHDRAW",
			"action_message": fmt.Sprintf("BCS 代理用户 %s 撤回", username),
		},
	}
	// 请求API
	proxy := ""
	body, err := component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("request itsm withdraw ticket %s failed, %s", sn, err.Error())
		return errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &OperateTicketResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return err
	}
	if resp.Code != 0 {
		logging.Error("itsm withdraw ticket %s failed, msg: %s", sn, resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}
