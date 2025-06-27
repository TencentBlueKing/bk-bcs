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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	configm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
)

// SubmitCreateNamespaceTicket create new itsm create namespace ticket
func SubmitCreateNamespaceTicket(ctx context.Context, username string, projectCode, clusterID, namespace string,
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
	return CreateTicket(ctx, username, serviceID, fields)
}

// SubmitUpdateNamespaceTicket create new itsm update namespace ticket
func SubmitUpdateNamespaceTicket(ctx context.Context, username, projectCode, clusterID, namespace string,
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
	return CreateTicket(ctx, username, serviceID, fields)
}

// SubmitDeleteNamespaceTicket create new itsm delete namespace ticket
func SubmitDeleteNamespaceTicket(ctx context.Context, username,
	projectCode, clusterID, namespace string) (*CreateTicketData, error) {
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
	return CreateTicket(ctx, username, serviceID, fields)
}

// SubmitQuotaManagerCommonTicket create new itsm quota manager ticket 额度管理通用审批单据
func SubmitQuotaManagerCommonTicket(ctx context.Context, username,
	projectCode, clusterID, content string) (*CreateTicketData, error) {
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
	return CreateTicket(ctx, username, serviceID, fields)
}
