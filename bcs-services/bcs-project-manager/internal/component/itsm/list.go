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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

const (
	// FINISHED FINISHED已结束
	FINISHED = "FINISHED"
	// TERMINATED TERMINATED被终止
	TERMINATED = "TERMINATED"
	// SUSPENDED SUSPENDED被挂起
	SUSPENDED = "SUSPENDED"
	// RUNNING RUNNING处理中
	RUNNING = "RUNNING"
	// RESOLVED RESOLVED已解决
	RESOLVED = "RESOLVED"
	// CONFIRMED CONFIRMED待解决
	CONFIRMED = "CONFIRMED"
	// REVOKED REVOKED已撤销
	REVOKED = "REVOKED"
)

var (
	listTicketsPath          = "/itsm/get_tickets/"
	ticketApprovalResultPath = "/itsm/ticket_approval_result/"
	limit                    = 100
)

// ListTicketsResp itsm list tickets resp
type ListTicketsResp struct {
	component.CommonResp
	Data ListTicketsData `json:"data"`
}

// ListTicketsData list tickets data
type ListTicketsData struct {
	Page      int           `json:"page"`
	TotalPage int           `json:"total_page"`
	Count     int           `json:"count"`
	Next      string        `json:"next"`
	Previous  string        `json:"previous"`
	Items     []TicketsItem `json:"items"`
}

// TicketsItem ITSM list tickets item
type TicketsItem struct {
	ID            int    `json:"id"`
	SN            string `json:"sn"`
	Title         string `json:"title"`
	CatalogID     int    `json:"catalog_id"`
	ServiceID     int    `json:"service_id"`
	ServiceType   string `json:"service_type"`
	FlowID        int    `json:"flow_id"`
	CurrentStatus string `json:"current_status"`
	CommentID     string `json:"comment_id"`
	IsCommented   bool   `json:"is_commented"`
	UpdatedBy     string `json:"updated_by"`
	UpdateAt      string `json:"update_at"`
	EndAt         string `json:"ent_at"`
	Creator       string `json:"creator"`
	CreateAt      string `json:"creat_at"`
	BkBizID       int    `json:"bk_biz_id"`
	TicketURL     string `json:"ticket_url"`
}

// ListTickets list itsm tickets by sn list
func ListTickets(snList []string) ([]TicketsItem, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	tickets := []TicketsItem{}
	var page = 1
	for {
		reqURL := fmt.Sprintf("%s%s?page=%d&page_size=%d", host, listTicketsPath, page, limit)
		req := gorequest.SuperAgent{
			Url:    reqURL,
			Method: "POST",
			Data: map[string]interface{}{
				"sns": snList,
			},
		}
		// 请求API
		proxy := ""
		body, err := component.Request(req, timeout, proxy, component.GetAuthHeader())
		if err != nil {
			logging.Error("request list itsm tickets %v failed, %s", snList, err.Error())
			return nil, errorx.NewRequestITSMErr(err.Error())
		}
		// 解析返回的body
		resp := &ListTicketsResp{}
		if err := json.Unmarshal([]byte(body), resp); err != nil {
			logging.Error("parse itsm body error, body: %v", body)
			return nil, err
		}
		if resp.Code != 0 {
			logging.Error("list itsm tickets %v failed, msg: %s", snList, resp.Message)
			return nil, errors.New(resp.Message)
		}
		tickets = append(tickets, resp.Data.Items...)
		if page >= resp.Data.TotalPage {
			break
		}
		page++
	}
	return tickets, nil
}

// ListTicketsApprovalResp itsm list tickets approval resp
type ListTicketsApprovalResp struct {
	component.CommonResp
	Data []TicketApprovalItem `json:"data"`
}

// TicketApprovalItem ITSM ticket approval result
type TicketApprovalItem struct {
	SN             string `json:"sn"`
	Title          string `json:"title"`
	CurrentStatus  string `json:"current_status"`
	TicketURL      string `json:"ticket_url"`
	CommentID      string `json:"comment_id"`
	UpdatedBy      string `json:"updated_by"`
	UpdateAt       string `json:"update_at"`
	ApprovalResult bool   `json:"approve_result"`
}

// ListTicketsApprovalResult list itsm tickets approval result by sn list
func ListTicketsApprovalResult(snList []string) ([]TicketApprovalItem, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}

	reqURL := fmt.Sprintf("%s%s", host, ticketApprovalResultPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"sn": snList,
		},
	}
	// 请求API
	proxy := ""
	body, err := component.Request(req, timeout, proxy, component.GetAuthHeader())
	if err != nil {
		logging.Error("request list itsm tickets approval %v failed, %s", snList, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &ListTicketsApprovalResp{}
	if err = json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("list itsm tickets approval %v failed, msg: %s", snList, resp.Message)
		return nil, errors.New(resp.Message)
	}

	return resp.Data, nil
}
