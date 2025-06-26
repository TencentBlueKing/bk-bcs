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

// Package v2 xxx
package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	createTicketPath = "/itsm/create_ticket/"
	timeout          = 10
)

// CreateTicketResp itsm create ticket resp
type CreateTicketResp struct {
	component.CommonResp
	RequestID string           `json:"request_id"`
	Data      CreateTicketData `json:"data"`
}

// CreateTicketData itsm create ticket data
type CreateTicketData struct {
	SN        string `json:"sn"`
	ID        int    `json:"id"`
	TicketURL string `json:"ticket_url"`
}

// CreateTicket create itsm ticket, username 单据创建者
func CreateTicket(ctx context.Context, username string, serviceID int,
	fields []map[string]interface{}) (*CreateTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, createTicketPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"creator":    username,
			"service_id": serviceID,
			"fields":     fields,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("CreateTicket get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	proxy := ""
	body, err := component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("request itsm create ticket failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &CreateTicketResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("itsm create ticket failed, msg: %s", resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}
