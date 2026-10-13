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

// Package v4 xxx
package v4

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
	createTicketPath  = "/api/v1/system/create"
	listTicketPath    = "/api/v1/ticket/list"
	ticketDetailPath  = "/api/v1/ticket/detail"
	revokedTicketPath = "/api/v1/tickets/revoked"

	systemCode = "bk_bcs"
)

// CreateTicketReq xxx
type CreateTicketReq struct {
	// WorkFlowKey 流程pk
	WorkFlowKey string `json:"workflow_key"`
	// ServiceID 服务pk
	ServiceID string `json:"service_id"`
	// FormData schema的实例化数据
	FormData map[string]interface{} `json:"form_data"`
	// CallbackUrl 回调url，post请求
	CallbackUrl string `json:"callback_url"`
	// CallbackToken 回调时作为post参数传入，由业务系统自己控制鉴权
	CallbackToken string `json:"callback_token"`
	// Options xxx
	Options Options `json:"options"`
	// SystemID 如果传入system_id，需要在请求头加入SYSTEM-TOKEN
	SystemID string `json:"system_id"`
	// Operator 实际提单人
	Operator string `json:"operator"`
}

// Options options
type Options struct {
}

// CreateTicketResp xxx
type CreateTicketResp struct {
	component.CommonResp
	Data CreateTicketData `json:"data"`
}

// CreateTicketData xxx
type CreateTicketData struct {
	// Name 系统名称
	Name string `json:"name"`
	// Code 系统标识
	Code uint32 `json:"code"`
	// Desc 系统描述
	Desc string `json:"desc"`
	// FrontendUrl xxx
	FrontendUrl string `json:"frontend_url"`
}

// CallbackResp xxx
type CallbackResp struct {
	component.CommonResp
	Data CallbackData `json:"data"`
}

// CallbackData xxx
type CallbackData struct {
	BkAppCode     string `json:"bk_app_code"`
	BkAppSecret   string `json:"bk_app_secret"`
	BkUsername    string `json:"bk_username"`
	Ticket        Ticket `json:"ticket"`
	CallbackToken string `json:"callback_token"`
}

// Ticket 工单数据结构
type Ticket struct {
	// ID 工单ID
	ID string `json:"id"`
	// SN 工单单号
	SN string `json:"sn"`
	// Title 工单标题
	Title string `json:"title"`
	// CreatedAt 提单时间
	CreatedAt string `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt bool `json:"updated_at"`
	// EndAt 结束时间
	EndAt string `json:"end_at"`
	// Status 状态标识
	Status string `json:"status"`
	// StatusDisplay 状态展示名
	StatusDisplay string `json:"status_display"`
	// WorkflowID 流程ID
	WorkflowID string `json:"workflow_id"`
	// ServiceID 服务ID
	ServiceID string `json:"service_id"`
	// PortalID 门户ID
	PortalID string `json:"portal_id"`
	// CurrentProcessors 当前处理人列表
	CurrentProcessors []Processor `json:"current_processors"`
	// CurrentSteps 当前步骤列表
	CurrentSteps []Step `json:"current_steps"`
	// FrontendURL 工单前端访问地址
	FrontendURL string `json:"frontend_url"`
	// FormData 工单表单实例化数据
	FormData json.RawMessage `json:"form_data"`
	// ApproveResult 审批结果
	ApproveResult bool `json:"approve_result"`
	// CallbackResult 回调结果
	CallbackResult CallbackResult `json:"callback_result"`
}

// Step 步骤信息
type Step struct {
	TicketID string `json:"ticket_id"`
	// Name 步骤名称
	Name string `json:"name"`
}

// CallbackResult 回调结果
type CallbackResult struct {
	// Result 回调接口最外层的result信息
	Result bool `json:"result"`
	// Message 回调报错信息或者回调接口最外层的message信息
	Message string `json:"message"`
}

// Processor 处理人信息
type Processor struct {
	TicketID string `json:"ticket_id"`
	// Processor 类型处理人标识列表字符串
	Processor string `json:"processor"`
	// ProcessorType 处理人类型: user/group/organization
	ProcessorType string `json:"processor_type"`
}

// CreateTicket xxx
func CreateTicket(ctx context.Context, data CreateTicketReq) (*CreateTicketData, error) {
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
			"workflow_key":   data.WorkFlowKey,
			"service_id":     data.ServiceID,
			"form_data":      data.FormData,
			"callback_url":   data.CallbackUrl,
			"callback_token": data.CallbackToken,
			"system_id":      data.SystemID,
			"operator":       data.Operator,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("CreateTicket get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request create itsm ticket %v failed, error: %s", data, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 解析返回的body
	resp := &CreateTicketResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request create itsm ticket %v failed, msg: %s", data, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}

// ListTicketReq 工单列表请求参数
type ListTicketReq struct {
	// ViewType 视图类型，默认"all"
	ViewType string `json:"view_type"`
	// Page 页码，默认1
	Page int `json:"page"`
	// PageSize 页大小，默认10，最大50
	PageSize int `json:"page_size"`
	// WorkflowKeyIn 逗号隔开的多个流程id
	WorkflowKeyIn string `json:"workflow_key__in"`
	// CurrentProcessorsIn 逗号隔开的多个用户对象
	CurrentProcessorsIn string `json:"current_processors__in"`
	// SnContains 单号模糊查询
	SnContains string `json:"sn__contains"`
	// TitleContains 标题模糊查询
	TitleContains string `json:"title__contains"`
	// CreatorIn 逗号隔开的多个username
	CreatorIn string `json:"creator__in"`
	// StatusDisplayIn 逗号隔开的多个状态名
	StatusDisplayIn string `json:"status_display__in"`
	// CreatedAtRange 提单时间范围
	CreatedAtRange string `json:"created_at__range"`
	// SystemIdIn 逗号隔开的多个系统标识
	SystemIdIn string `json:"system_id__in"`
	// IdIn 逗号隔开的多个工单id
	IdIn string `json:"id__in"`
}

// ListTicketResp xxx
type ListTicketResp struct {
	component.CommonResp
	Data []*ListTicketData `json:"data"`
}

// ListTicketData 工单列表响应数据
type ListTicketData struct {
	// Results 工单详情数据
	Results []Ticket `json:"results"`
	// Page 页码
	Page string `json:"page"`
	// PageSize 页大小
	PageSize string `json:"page_size"`
	// Count 总数
	Count int `json:"count"`
}

// ListTicket xxx
func ListTicket(ctx context.Context, data ListTicketReq) ([]*ListTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, listTicketPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "GET",
		Data: map[string]interface{}{
			"view_type":              data.ViewType,
			"page":                   data.Page,
			"page_size":              data.PageSize,
			"workflow_key__in":       data.WorkflowKeyIn,
			"current_processors__in": data.CurrentProcessorsIn,
			"sn__contains":           data.SnContains,
			"title__contains":        data.TitleContains,
			"creator__in":            data.CreatorIn,
			"status_display__in":     data.StatusDisplayIn,
			"created_at__range":      data.CreatedAtRange,
			"system_id__in":          data.SystemIdIn,
			"id__in":                 data.IdIn,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("ListTicket get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request get itsm ticket %v list failed, error: %s", data, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 解析返回的body
	resp := &ListTicketResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request get itsm ticket %v list failed, msg: %s", data, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

// TicketDetailReq xxx
type TicketDetailReq struct {
	// ID 工单id
	ID string `json:"id"`
}

// TicketDetailResp xxx
type TicketDetailResp struct {
	component.CommonResp
	Data Ticket `json:"data"`
}

// TicketDetail xxx
func TicketDetail(ctx context.Context, data TicketDetailReq) (*Ticket, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, ticketDetailPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "GET",
		Data: map[string]interface{}{
			"id": data.ID,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("TicketDetail get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request get itsm ticket %v detail failed, error: %s", data.ID, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 解析返回的body
	resp := &TicketDetailResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request get itsm ticket %v detail failed, msg: %s", data.ID, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}

// RevokedTicketReq 撤销工单请求参数
type RevokedTicketReq struct {
	// SystemID 系统标识
	SystemID string `json:"system_id"`
	// TicketID 工单标识
	TicketID string `json:"ticket_id"`
}

// RevokedTicketResp 撤销工单返回参数
type RevokedTicketResp struct {
	component.CommonResp
	Data RevokedTicketData `json:"data"`
}

// RevokedTicketData 撤销工单返回数据
type RevokedTicketData struct {
	// Result 	是否撤销成功
	Result bool `json:"result"`
}

// RevokedTicket xxx
func RevokedTicket(ctx context.Context, data RevokedTicketReq) (*RevokedTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, revokedTicketPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"system_id": data.SystemID,
			"ticket_id": data.TicketID,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("RevokedTicket get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request revoke itsm ticket %v failed, error: %s", data, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 解析返回的body
	resp := &RevokedTicketResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request revoke itsm ticket %v failed, msg: %s", data, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return &resp.Data, nil
}
