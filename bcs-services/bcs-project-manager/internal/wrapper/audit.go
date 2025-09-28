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

package wrapper

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

// NewAuditWrapper 审计
func NewAuditWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		startTime := time.Now()
		err := fn(ctx, req, rsp)
		endTime := time.Now()
		go addAudit(ctx, req, rsp, startTime, endTime)
		return err
	}
}

// resource ResourceData struct
type resource struct {
	ClusterID       string `json:"clusterID" yaml:"clusterID"`
	Namespace       string `json:"namespace" yaml:"namespace"`
	Name            string `json:"name" yaml:"name"`
	Key             string `json:"key" yaml:"key"`
	IDs             string `json:"idList" yaml:"idList"`
	VariableID      string `json:"variableID" yaml:"variableID"`
	ProjectID       string `json:"projectID" yaml:"projectID"`
	ProjectCode     string `json:"projectCode" yaml:"projectCode"`
	ProjectIDOrCode string `json:"projectIDOrCode" yaml:"projectIDOrCode"`
}

// resource to map
func (r resource) toMap() map[string]interface{} {
	result := make(map[string]interface{}, 0)
	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}
	if r.Namespace != "" {
		result["Namespace"] = r.Namespace
	}
	if r.Name != "" {
		result["Name"] = r.Name
	}
	if r.Key != "" {
		result["Key"] = r.Key
	}
	if r.VariableID != "" {
		result["VariableID"] = r.VariableID
	}
	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}
	if r.ProjectCode != "" {
		result["ProjectCode"] = r.ProjectCode
	}
	return result
}

// get resource
func getResourceID(req server.Request) resource {
	body := req.Body()
	b, _ := json.Marshal(body)

	resourceID := resource{}
	_ = json.Unmarshal(b, &resourceID)
	// ProjectCode为空的情况下使用ProjectID或者ProjectIDOrCode代替
	if resourceID.ProjectCode == "" {
		if resourceID.ProjectID != "" {
			resourceID.ProjectCode = resourceID.ProjectID
		} else {
			resourceID.ProjectCode = resourceID.ProjectIDOrCode
		}
	}

	return resourceID
}

// NOCC: golint/unparam(设计如此:)
// nolint
var auditFuncMap = map[string]func(req server.Request) (audit.Resource, audit.Action){
	"BCSProject.CreateProject": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeProject, ResourceID: res.ProjectCode, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "project_create", ActivityType: audit.ActivityTypeCreate}
	},
	"BCSProject.GetProject": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeProject, ResourceID: res.ProjectCode, ResourceName: res.ProjectCode,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "project_get", ActivityType: audit.ActivityTypeView}
	},
	"BCSProject.UpdateProject": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeProject, ResourceID: res.ProjectID, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "project_edit", ActivityType: audit.ActivityTypeUpdate}
	},
	"Namespace.CreateNamespace": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_create", ActivityType: audit.ActivityTypeCreate}
	},
	"Namespace.UpdateNamespace": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_update", ActivityType: audit.ActivityTypeUpdate}
	},
	"Namespace.GetNamespace": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_get", ActivityType: audit.ActivityTypeView}
	},
	"Namespace.DeleteNamespace": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_delete", ActivityType: audit.ActivityTypeDelete}
	},
	"Variable.CreateVariable": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.Key, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "create_variable", ActivityType: audit.ActivityTypeCreate}
	},
	"Variable.UpdateVariable": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.Key, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_variable", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.DeleteVariableDefinitions": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.IDs, ResourceName: res.IDs,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "deleteVariable_definitions", ActivityType: audit.ActivityTypeDelete}
	},
	"Variable.UpdateClustersVariables": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.VariableID, ResourceName: res.VariableID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_clusters_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateNamespacesVariables": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.VariableID, ResourceName: res.VariableID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_namespaces_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateClusterVariables": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_cluster_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateNamespaceVariables": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_namespace_variables", ActivityType: audit.ActivityTypeUpdate}
	},
}

// addAudit 添加审计
func addAudit(ctx context.Context, req server.Request, rsp interface{}, startTime, endTime time.Time) {
	// get method audit func
	fn, ok := auditFuncMap[req.Method()]
	if !ok {
		return
	}

	res, act := fn(req)

	// get audit context
	auditCtx := audit.RecorderContext{
		Username:  auth.GetUserFromCtx(ctx),
		SourceIP:  contextx.GetSourceIPFromCtx(ctx),
		UserAgent: contextx.GetUserAgentFromCtx(ctx),
		RequestID: contextx.GetRequestIDFromCtx(ctx),
		StartTime: startTime,
		EndTime:   endTime,
	}
	// get resource & action
	resource := audit.Resource{
		ProjectCode:  res.ProjectCode,
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		ResourceName: res.ResourceName,
		ResourceData: res.ResourceData,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	// get action result
	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
	}

	// get handle result
	v := reflect.ValueOf(rsp)
	codeField := v.Elem().FieldByName("Code")
	messageField := v.Elem().FieldByName("Message")
	if codeField.CanInterface() {
		code := int(codeField.Interface().(uint32))
		result.ResultCode = code
	}
	if messageField.CanInterface() {
		message := messageField.Interface().(string)
		result.ResultContent = message
	}
	if result.ResultCode != errorx.Success {
		result.Status = audit.ActivityStatusFailed
	}

	// add audit
	auditAction := component.GetAuditClient().R()
	if act.ActivityType == audit.ActivityTypeView {
		// 查看类型不用记录 activity
		auditAction.DisableActivity()
	}
	logging.Info("add audit, auditCtx: %+v, resource:%+v, action: %+v, result: %+v", auditCtx, resource, action, result)
	err := auditAction.SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
	logging.Error("add audit failed, err: %v", err)
}
