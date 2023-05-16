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
	createTicketPath = "/v2/itsm/create_ticket/"
	timeout          = 10
)

// CreateTicketResp itsm create ticket resp
type CreateTicketResp struct {
	Code      int              `json:"code"`
	Result    bool             `json:"result"`
	Message   string           `json:"message"`
	RequestID string           `json:"request_id"`
	Data      CreateTicketData `json:"data"`
}

// CreateTicketData itsm create ticket data
type CreateTicketData struct {
	SN        string `json:"sn"`
	ID        int    `json:"id"`
	TicketURL string `json:"ticket_url"`
}

// CreateTicket create itsm ticket
func CreateTicket(username string, serviceID int, fields []map[string]interface{}) (*CreateTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", itsmConf.GatewayHost, createTicketPath)
	headers := map[string]string{"Content-Type": "application/json"}
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"bk_app_code":   config.GlobalConf.App.Code,
			"bk_app_secret": config.GlobalConf.App.Secret,
			"creator":       username,
			"service_id":    serviceID,
			"fields":        fields,
		},
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

// SubmitCreateNamespaceTicket create new itsm create namespace ticket
func SubmitCreateNamespaceTicket(username, projectCode, clusterID, namespace string,
	cpuLimits, memoryLimits int) (*CreateTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "创建命名空间",
		},
		{
			"key":   "PROJECT_CODE",
			"value": projectCode,
		},
		{
			"key":   "CLUSTER_ID",
			"value": clusterID,
		},
		{
			"key":   "NAMESPACE",
			"value": namespace,
		},
		{
			"key":   "CPU_LIMITS",
			"value": cpuLimits,
		},
		{
			"key":   "MEMORY_LIMITS",
			"value": memoryLimits,
		},
	}
	return CreateTicket(username, itsmConf.CreateNamespaceServiceID, fields)
}

// SubmitUpdateNamespaceTicket create new itsm update namespace ticket
func SubmitUpdateNamespaceTicket(username, projectCode, clusterID, namespace string,
	cpuLimits, memoryLimits, oldCPULimits, oldMemoryLimits int) (*CreateTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "更新命名空间",
		},
		{
			"key":   "PROJECT_CODE",
			"value": projectCode,
		},
		{
			"key":   "CLUSTER_ID",
			"value": clusterID,
		},
		{
			"key":   "NAMESPACE",
			"value": namespace,
		},
		{
			"key":   "CPU_LIMITS",
			"value": cpuLimits,
		},
		{
			"key":   "MEMORY_LIMITS",
			"value": memoryLimits,
		},
		{
			"key":   "OLD_CPU_LIMITS",
			"value": oldCPULimits,
		},
		{
			"key":   "OLD_MEMORY_LIMITS",
			"value": oldCPULimits,
		},
	}
	return CreateTicket(username, itsmConf.UpdateNamespaceServiceID, fields)
}

// SubmitDeleteNamespaceTicket create new itsm delete namespace ticket
func SubmitDeleteNamespaceTicket(username, projectCode, clusterID, namespace string) (*CreateTicketData, error) {
	itsmConf := config.GlobalConf.ITSM
	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "删除命名空间",
		},
		{
			"key":   "PROJECT_CODE",
			"value": projectCode,
		},
		{
			"key":   "CLUSTER_ID",
			"value": clusterID,
		},
		{
			"key":   "NAMESPACE",
			"value": namespace,
		},
	}
	return CreateTicket(username, itsmConf.DeleteNamespaceServiceID, fields)
}
