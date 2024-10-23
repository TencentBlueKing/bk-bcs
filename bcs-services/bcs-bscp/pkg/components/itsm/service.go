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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

var (
	getServiceDetailPath  = "/itsm/get_service_detail/"
	getWorkflowDetailPath = "/itsm/get_workflow_detail/"
	listServicesPath      = "/itsm/get_services/"
	importServicePath     = "/itsm/import_service/"
	updateServicePath     = "/itsm/update_service/"
)

// ListServicesResp itsm list services resp
type ListServicesResp struct {
	CommonResp
	Data []Service `json:"data"`
}

// Service ITSM get services item
type Service struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	ServiceType string `json:"service_type"`
}

// GetServiceDetailResp get itsm service detail resp
type GetServiceDetailResp struct {
	CommonResp
	Data ServiceDetail `json:"data"`
}

// ServiceDetail ITSM get service detail item
type ServiceDetail struct {
	WorkflowId int `json:"workflow_id"`
}

// GetWorkflowDetailResp get itsm workflow detail resp
type GetWorkflowDetailResp struct {
	CommonResp
	Data WorkflowDetail `json:"data"`
}

// WorkflowDetail ITSM get workflow detail item
type WorkflowDetail struct {
	States []States `json:"states"`
}

// States ITSM get workflow detail item states
type States struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

// ListServices list itsm services by catalog id
func ListServices(ctx context.Context, catalogID uint32) ([]Service, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s?catalog_id=%d", host, listServicesPath, catalogID)
	body, err := ItsmRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		logs.Errorf("request list itsm services in catalog %d failed, %s", catalogID, err.Error())
		return nil, fmt.Errorf("request list itsm services in catalog %d failed, %s", catalogID, err.Error())
	}
	// 解析返回的body
	resp := &ListServicesResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logs.Errorf("request list itsm services in catalog %d failed, msg: %s", catalogID, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

// GetWorkflowByService get itsm workfolw by service id
func GetWorkflowByService(ctx context.Context, serviceID int) (int, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}

	reqURL := fmt.Sprintf("%s%s?service_id=%d", host, getServiceDetailPath, serviceID)

	body, err := ItsmRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		logs.Errorf("request get itsm service detail in service %d failed, %s", serviceID, err.Error())
		return 0, fmt.Errorf("request get itsm service detail in service %d failed, %s", serviceID, err.Error())
	}
	// 解析返回的body
	resp := &GetServiceDetailResp{}
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return 0, fmt.Errorf("parse itsm body error, body: %v", body)
	}
	if resp.Code != 0 {
		logs.Errorf("request get itsm service in service %d failed, msg: %s", serviceID, resp.Message)
		return 0, errors.New(resp.Message)
	}
	return resp.Data.WorkflowId, nil
}

// GetStateApproveByWorkfolw get itsm state approve by workflow id
func GetStateApproveByWorkfolw(ctx context.Context, workflowID int) (int, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	// 暂时使用外部地址测试，后面废除
	host := itsmConf.Host
	// host := itsmConf.GatewayHost
	// if itsmConf.External {
	// 	host = itsmConf.Host
	// }

	reqURL := fmt.Sprintf("%s%s?workflow_id=%d", host, getWorkflowDetailPath, workflowID)

	body, err := ItsmRequest(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		logs.Errorf("request get itsm workflow detail in workflow %d failed, %s", workflowID, err.Error())
		return 0, fmt.Errorf("request get itsm workflow detail in workflow %d failed, %s", workflowID, err.Error())
	}
	// 解析返回的body
	resp := &GetWorkflowDetailResp{}
	if err := json.Unmarshal(body, &resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return 0, fmt.Errorf("parse itsm body error, body: %v", body)
	}
	if resp.Code != 0 {
		logs.Errorf("request get itsm workflow in workflow %d failed, msg: %s", workflowID, resp.Message)
		return 0, errors.New(resp.Message)
	}
	for _, v := range resp.Data.States {
		// 来自于创建的配置节点名称
		if v.Type == constant.ItsmApproveType {
			return v.Id, nil
		}
	}
	return 0, fmt.Errorf("workflow %d approve node not found", workflowID)
}

// ImportServiceReq itsm import service req
type ImportServiceReq struct {
	// Key 服务类型
	Key string `json:"key"`
	// Name 服务名称
	Name string `json:"name"`
	// CatelogID 服务关联的服务目录ID
	CatelogID uint32 `json:"catalog_id"`
	// Desc 服务描述
	Desc string `json:"desc"`
	// Owners 服务负责人
	Owners string `json:"owner"`
	// CanTicketAgency 是否允许代提单
	CanTicketAgency bool `json:"can_ticket_agency"`
	// IsValid 是否启用，不传默认为 false
	IsValid bool `json:"is_valid"`
	// DisplayType 显示类型
	DisplayType string `json:"display_type"`
	// DisplayRole 显示角色，display_type 为 open 时，值为空
	DisplayRole string `json:"display_role"`
	// Source 服务来源
	Source string `json:"source"`
	// ProjectKey 项目key
	ProjectKey string `json:"project_key"`
	// Workflow 流程数据
	Workflow interface{} `json:"workflow"`
}

// ImportServiceResp itsm import service resp
type ImportServiceResp struct {
	CommonResp
	Data ImportServiceData `json:"data"`
}

// ImportServiceData itsm import service data
type ImportServiceData struct {
	ID int `json:"id"`
}

// ImportService import itsm service
func ImportService(ctx context.Context, data ImportServiceReq) (int, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, importServicePath)
	reqData := map[string]interface{}{
		"key":               data.Key,
		"name":              data.Name,
		"desc":              data.Desc,
		"catalog_id":        data.CatelogID,
		"owner":             data.Owners,
		"can_ticket_agency": data.CanTicketAgency,
		"is_valid":          data.IsValid,
		"display_type":      data.DisplayType,
		"display_role":      data.DisplayRole,
		"source":            data.Source,
		"project_key":       data.ProjectKey,
		"workflow":          data.Workflow,
	}
	// 请求API
	body, err := ItsmRequest(ctx, http.MethodPost, reqURL, reqData)
	if err != nil {
		logs.Errorf("request import service %s failed, %s", data.Name, err.Error())
		return 0, fmt.Errorf("request bk itsm api error: %s", err)
	}
	// 解析返回的body
	resp := &ImportServiceResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return 0, err
	}
	if resp.Code != 0 {
		logs.Errorf("request import service %s failed, msg: %s", data.Name, resp.Message)
		return 0, errors.New(resp.Message)
	}
	return resp.Data.ID, nil
}

// UpdateServiceReq itsm update service req
type UpdateServiceReq struct {
	ID int `json:"id"`
	ImportServiceReq
}

// UpdateService update itsm service
func UpdateService(ctx context.Context, data UpdateServiceReq) error {

	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, updateServicePath)
	reqData := map[string]interface{}{
		"id":                data.ID,
		"key":               data.Key,
		"name":              data.Name,
		"desc":              data.Desc,
		"catalog_id":        data.CatelogID,
		"owner":             data.Owners,
		"can_ticket_agency": data.CanTicketAgency,
		"is_valid":          data.IsValid,
		"display_type":      data.DisplayType,
		"display_role":      data.DisplayRole,
		"source":            data.Source,
		"project_key":       data.ProjectKey,
		"workflow":          data.Workflow,
	}
	// 请求API
	body, err := ItsmRequest(ctx, http.MethodPost, reqURL, reqData)
	if err != nil {
		logs.Errorf("request update service %s failed, %s", data.Name, err.Error())
		return fmt.Errorf("request bk itsm api error: %s", err)
	}
	// 解析返回的body
	resp := &CommonResp{}
	if err := json.Unmarshal(body, resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return err
	}
	if resp.Code != 0 {
		logs.Errorf("request update service %s failed, msg: %s", data.Name, resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}
