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
	"strconv"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	configm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
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

// CreateTicket create itsm ticket
func CreateTicket(username string, serviceID int, fields []map[string]interface{}) (*CreateTicketData, error) {
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
	// 请求API
	proxy := ""
	body, err := component.Request(req, timeout, proxy, component.GetAuthHeader())
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
	var serviceID int
	itsmConf := config.GlobalConf.ITSM
	if itsmConf.AutoRegister {
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(),
			configm.ConfigKeyCreateNamespaceItsmServiceID)
		if err != nil {
			return nil, err
		}
		serviceID, err = strconv.Atoi(serviceIDStr)
		if err != nil {
			return nil, err
		}
	} else {
		serviceID = itsmConf.CreateNamespaceServiceID
	}
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
	return CreateTicket(username, serviceID, fields)
}

// SubmitUpdateNamespaceTicket create new itsm update namespace ticket
func SubmitUpdateNamespaceTicket(username, projectCode, clusterID, namespace string,
	cpuLimits, memoryLimits, oldCPULimits, oldMemoryLimits int) (*CreateTicketData, error) {
	var serviceID int
	itsmConf := config.GlobalConf.ITSM
	if itsmConf.AutoRegister {
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(),
			configm.ConfigKeyUpdateNamespaceItsmServiceID)
		if err != nil {
			return nil, err
		}
		serviceID, err = strconv.Atoi(serviceIDStr)
		if err != nil {
			return nil, err
		}
	} else {
		serviceID = itsmConf.UpdateNamespaceServiceID
	}
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
			"value": oldMemoryLimits,
		},
	}
	return CreateTicket(username, serviceID, fields)
}

// SubmitDeleteNamespaceTicket create new itsm delete namespace ticket
func SubmitDeleteNamespaceTicket(username, projectCode, clusterID, namespace string) (*CreateTicketData, error) {
	var serviceID int
	itsmConf := config.GlobalConf.ITSM
	if itsmConf.AutoRegister {
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(),
			configm.ConfigKeyDeleteNamespaceItsmServiceID)
		if err != nil {
			return nil, err
		}
		serviceID, err = strconv.Atoi(serviceIDStr)
		if err != nil {
			return nil, err
		}
	} else {
		serviceID = itsmConf.DeleteNamespaceServiceID
	}
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
	return CreateTicket(username, serviceID, fields)
}

// SubmitQuotaManagerCommonTicket create new itsm quota manager ticket 额度管理通用审批单据
func SubmitQuotaManagerCommonTicket(username, projectCode, clusterID, content string) (*CreateTicketData, error) {
	var (
		serviceID int
		itsmConf  = config.GlobalConf.ITSM
	)

	if itsmConf.AutoRegister {
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(),
			configm.QuotaManagerCommonItsmServiceID)
		if err != nil {
			return nil, err
		}
		serviceID, err = strconv.Atoi(serviceIDStr)
		if err != nil {
			return nil, err
		}
	} else {
		serviceID = itsmConf.QuotaManagerCommonServiceID
	}

	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "额度管理通用审批单据",
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
			"key":   "apply_content",
			"value": content,
		},
	}
	return CreateTicket(username, serviceID, fields)
}
