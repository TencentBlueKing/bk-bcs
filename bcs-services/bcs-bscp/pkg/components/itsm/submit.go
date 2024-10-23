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
	createTicketPath  = "/itsm/create_ticket/"
	approveTicketPath = "/itsm/approve/"
)

// CreateTicketResp itsm create ticket resp
type CreateTicketResp struct {
	CommonResp
	RequestID string           `json:"request_id"`
	Data      CreateTicketData `json:"data"`
}

// CreateTicketData itsm create ticket data
type CreateTicketData struct {
	SN        string `json:"sn"`
	ID        int    `json:"id"`
	TicketURL string `json:"ticket_url"`
	StateID   int    `json:"state_id"`
}

// CreateTicket create itsm ticket
func CreateTicket(ctx context.Context, reqData map[string]interface{}) (*CreateTicketData, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, createTicketPath)

	// 请求API
	body, err := ItsmRequest(context.Background(), http.MethodPost, reqURL, reqData)
	if err != nil {
		logs.Errorf("request itsm create ticket failed, %s", err.Error())
		return nil, fmt.Errorf("request itsm create ticket failed, %s", err.Error())
	}
	// 解析返回的body
	resp := &CreateTicketResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logs.Errorf("itsm create ticket failed, msg: %s", resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}

// UpdateTicketByApporver update itsm ticket by approver
func UpdateTicketByApporver(ctx context.Context, reqData map[string]interface{}) error {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, approveTicketPath)

	// 请求API
	body, err := ItsmRequest(ctx, http.MethodPost, reqURL, reqData)
	if err != nil {
		logs.Errorf("request itsm update ticket failed, %s", err.Error())
		return fmt.Errorf("request itsm update ticket failed, %s", err.Error())
	}
	// 解析返回的body
	resp := &CommonResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return err
	}
	if resp.Code != 0 {
		logs.Errorf("itsm update ticket failed, msg: %s", resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}
