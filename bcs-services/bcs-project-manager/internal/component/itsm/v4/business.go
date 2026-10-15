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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	configm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

// SubmitCreateNamespaceTicket create new itsm create namespace ticket
func SubmitCreateNamespaceTicket(ctx context.Context, username, projectCode, clusterID, namespace string,
	cpuLimits, memoryLimits int, approvers []string) (*CreateTicketData, error) {

	var serviceSn string

	tenantId := tenant.GetTenantIdFromContext(ctx)
	itsmConf := config.GlobalConf.ITSM

	// AutoRegister true community version
	if itsmConf.AutoRegister {
		serviceKey := fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyCreateNamespaceItsmServiceID)
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(), serviceKey)
		if err != nil {
			return nil, err
		}
		serviceSn = serviceIDStr
	} else {
		serviceSn = fmt.Sprintf("%v", itsmConf.CreateNamespaceServiceID)
	}

	fields := map[string]interface{}{
		"ticket__title": "创建命名空间",
		"PROJECT_CODE":  projectCode,
		"CLUSTER_ID":    clusterID,
		"NAMESPACE":     namespace,
		"CPU_LIMITS":    cpuLimits,
		"MEMORY_LIMITS": memoryLimits,
		"approver":      approvers,
	}

	return CreateTicket(ctx, CreateTicketReq{
		WorkFlowKey: serviceSn,
		FormData:    fields,
		// CallbackUrl 工单审批通过或者拒绝后执行回调,后续配置在配置文件中通过新的回调接口支持
		CallbackUrl: "",
		SystemID:    systemCode,
		Operator:    username,
	})
}

// SubmitUpdateNamespaceTicket create new itsm update namespace ticket
func SubmitUpdateNamespaceTicket(ctx context.Context, username, projectCode, clusterID, namespace string,
	cpuLimits, memoryLimits int, approvers []string) (*CreateTicketData, error) {
	var serviceSn string

	tenantId := tenant.GetTenantIdFromContext(ctx)
	itsmConf := config.GlobalConf.ITSM

	// AutoRegister true community version
	if itsmConf.AutoRegister {
		serviceKey := fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyUpdateNamespaceItsmServiceID)
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(), serviceKey)
		if err != nil {
			return nil, err
		}
		serviceSn = serviceIDStr
	} else {
		serviceSn = fmt.Sprintf("%v", itsmConf.CreateNamespaceServiceID)
	}

	fields := map[string]interface{}{
		"ticket__title": "更新命名空间",
		"PROJECT_CODE":  projectCode,
		"CLUSTER_ID":    clusterID,
		"NAMESPACE":     namespace,
		"CPU_LIMITS":    cpuLimits,
		"MEMORY_LIMITS": memoryLimits,
		"approver":      approvers,
	}

	return CreateTicket(ctx, CreateTicketReq{
		WorkFlowKey: serviceSn,
		FormData:    fields,
		// CallbackUrl 工单审批通过或者拒绝后执行回调,后续配置在配置文件中通过新的回调接口支持
		CallbackUrl: "",
		SystemID:    systemCode,
		Operator:    username,
	})
}

// SubmitDeleteNamespaceTicket create new itsm delete namespace ticket
func SubmitDeleteNamespaceTicket(ctx context.Context, username,
	projectCode, clusterID, namespace string, approvers []string) (*CreateTicketData, error) {
	var serviceSn string

	tenantId := tenant.GetTenantIdFromContext(ctx)
	itsmConf := config.GlobalConf.ITSM

	// AutoRegister true community version
	if itsmConf.AutoRegister {
		serviceKey := fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyDeleteNamespaceItsmServiceID)
		serviceIDStr, err := store.GetModel().GetConfig(context.Background(), serviceKey)
		if err != nil {
			return nil, err
		}
		serviceSn = serviceIDStr
	} else {
		serviceSn = fmt.Sprintf("%v", itsmConf.CreateNamespaceServiceID)
	}

	fields := map[string]interface{}{
		"ticket__title": "删除命名空间",
		"PROJECT_CODE":  projectCode,
		"CLUSTER_ID":    clusterID,
		"NAMESPACE":     namespace,
		"approver":      approvers,
	}

	return CreateTicket(ctx, CreateTicketReq{
		WorkFlowKey: serviceSn,
		FormData:    fields,
		// CallbackUrl 工单审批通过或者拒绝后执行回调,后续配置在配置文件中通过新的回调接口支持
		CallbackUrl: "",
		SystemID:    systemCode,
		Operator:    username,
	})
}

// ItsmV4TemplateRender xxx
type ItsmV4TemplateRender struct {
	FormModel                     string `json:"FormModel"`
	WorkflowCategories            string `json:"WorkflowCategories"`
	WorkflowDeleteSharedNamespace string `json:"WorkflowDeleteSharedNamespace"`
	WorkflowCreateSharedNamespace string `json:"WorkflowCreateSharedNamespace"`
	WorkflowUpdateSharedNamespace string `json:"WorkflowUpdateSharedNamespace"`
}

const (
	formModel        = "formmodel"
	workflowCategory = "workflowcategory"
	workflow         = "workflow"
)

func generateTemplateId(tenant string, systemCode string, category string) string {
	return stringx.RandomString(fmt.Sprintf("%s_%s_%s_", tenant, systemCode, category), 8)
}

// TenantWorkflowData tenant workflow data
type TenantWorkflowData struct {
	CreateSharedNamespace entity.KeyValue
	DeleteSharedNamespace entity.KeyValue
	UpdateSharedNamespace entity.KeyValue
}

// ItsmV4SystemMigrate xxx
func ItsmV4SystemMigrate(ctx context.Context) (*TenantWorkflowData, error) {
	tenantId := tenant.GetTenantIdFromContext(ctx)

	// 读取模板文件内容
	templateContent, err := ioutil.ReadFile(migrateItsm)
	if err != nil {
		return nil, err
	}
	// 解析模板
	tmpl, err := template.New("json").Parse(string(templateContent))
	if err != nil {
		return nil, err
	}

	values := ItsmV4TemplateRender{
		FormModel:                     generateTemplateId(tenantId, systemCode, formModel),
		WorkflowCategories:            generateTemplateId(tenantId, systemCode, workflowCategory),
		WorkflowDeleteSharedNamespace: generateTemplateId(tenantId, systemCode, workflow),
		WorkflowCreateSharedNamespace: generateTemplateId(tenantId, systemCode, workflow),
		WorkflowUpdateSharedNamespace: generateTemplateId(tenantId, systemCode, workflow),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, values)
	if err != nil {
		return nil, err
	}
	content := buf.String()

	err = MigrateSystem(ctx, []byte(content))
	if err != nil {
		return nil, err
	}

	return &TenantWorkflowData{
		CreateSharedNamespace: entity.KeyValue{
			Key:   fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyCreateNamespaceItsmServiceID),
			Value: values.WorkflowCreateSharedNamespace,
		},
		DeleteSharedNamespace: entity.KeyValue{
			Key:   fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyDeleteNamespaceItsmServiceID),
			Value: values.WorkflowDeleteSharedNamespace,
		},
		UpdateSharedNamespace: entity.KeyValue{
			Key:   fmt.Sprintf("%s-%s", tenantId, configm.ConfigKeyUpdateNamespaceItsmServiceID),
			Value: values.WorkflowUpdateSharedNamespace,
		},
	}, nil
}
