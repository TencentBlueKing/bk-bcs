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
	listServicesPath  = "/itsm/get_services/"
	importServicePath = "/itsm/import_service/"
	updateServicePath = "/itsm/update_service/"
)

// ListServicesResp itsm list services resp
type ListServicesResp struct {
	component.CommonResp
	Data []Service `json:"data"`
}

// Service ITSM get services item
type Service struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	ServiceType string `json:"service_type"`
}

// ListServices list itsm services by catalog id
func ListServices(ctx context.Context, catalogID uint32) ([]Service, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s?catalog_id=%d", host, listServicesPath, catalogID)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "GET",
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("ListServices get auth header failed, %s", err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	proxy := ""
	body, err := component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("request list itsm services in catalog %d failed, %s", catalogID, err.Error())
		return nil, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &ListServicesResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return nil, err
	}
	if resp.Code != 0 {
		logging.Error("request list itsm services in catalog %d failed, msg: %s", catalogID, resp.Message)
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
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
	component.CommonResp
	Data ImportServiceData `json:"data"`
}

// ImportServiceData itsm import service data
type ImportServiceData struct {
	ID int `json:"id"`
}

// ImportService import itsm service
func ImportService(ctx context.Context, data ImportServiceReq) (int, error) {
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, importServicePath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
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
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("ImportService get auth header failed, %s", err.Error())
		return 0, errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request import service %s failed, %s", data.Name, err.Error())
		return 0, errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &ImportServiceResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return 0, err
	}
	if resp.Code != 0 {
		logging.Error("request import service %s failed, msg: %s", data.Name, resp.Message)
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
	itsmConf := config.GlobalConf.ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s", host, updateServicePath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
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
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("ImportService get auth header failed, %s", err.Error())
		return errorx.NewRequestITSMErr(err.Error())
	}

	// 请求API
	body, err := component.Request(req, timeout, "", headers)
	if err != nil {
		logging.Error("request update service %s failed, %s", data.Name, err.Error())
		return errorx.NewRequestITSMErr(err.Error())
	}
	// 解析返回的body
	resp := &component.CommonResp{}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		logging.Error("parse itsm body error, body: %v", body)
		return err
	}
	if resp.Code != 0 {
		logging.Error("request update service %s failed, msg: %s", data.Name, resp.Message)
		return errors.New(resp.Message)
	}
	return nil
}
