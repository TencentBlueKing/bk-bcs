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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

var (
	operateTicketPath = "/itsm/operate_ticket/"
)

// OperateTicketResp itsm operate ticket resp
type OperateTicketResp struct {
	CommonResp
	Data interface{} `json:"data"`
}

// WithdrawTicket withdraw itsm ticket
func WithdrawTicket(ctx context.Context, reqData map[string]interface{}) error {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, operateTicketPath)
	// 请求API
	body, err := ItsmRequest(ctx, http.MethodPost, reqURL, reqData)
	if err != nil {
		logs.Errorf("request itsm withdraw ticket %s failed, %s", reqData["sn"], err.Error())
		return fmt.Errorf("request itsm withdraw ticket %s failed, %s", reqData["sn"], err.Error())
	}
	// 解析返回的body
	resp := &OperateTicketResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return err
	}
	if resp.Code != 0 {
		logs.Errorf("itsm withdraw ticket %s failed, msg: %s", reqData["sn"], resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}
