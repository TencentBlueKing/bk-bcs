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
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/micro/go-micro/v2/errors"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NewResponseWrapper 添加request id, 统一处理返回
func NewResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		startTime := time.Now()
		err := fn(ctx, req, rsp)
		requestID := ctx.Value(ctxkey.RequestIDKey).(string)
		endTime := time.Now()
		go addAudit(ctx, req, rsp, startTime, endTime)
		return RenderResponse(rsp, requestID, err)
	}
}

// RenderResponse 处理返回数据，用于返回统一结构
func RenderResponse(rsp interface{}, requestID string, err error) error {
	// support for data type string,slice and empty,haven't test for map and so on
	msg, code := getMsgCode(err)
	v := reflect.ValueOf(rsp)
	if v.Elem().FieldByName("RequestID").IsValid() {
		v.Elem().FieldByName("RequestID").SetString(requestID)
	}
	v.Elem().FieldByName("Message").SetString(msg)
	v.Elem().FieldByName("Code").SetUint(uint64(code))
	if err == nil {
		return nil
	}

	// handle error
	switch e := err.(type) {
	case *authutils.PermDeniedError:
		if v.Elem().FieldByName("WebAnnotations").IsValid() {
			perms := &proto.Perms{}
			permsMap := map[string]interface{}{}
			permsMap["apply_url"] = e.Perms.ApplyURL
			actionList := []map[string]string{}
			for _, actions := range e.Perms.ActionList {
				actionList = append(actionList, map[string]string{
					"action_id":     actions.Action,
					"resource_type": actions.Type,
				})
			}
			permsMap["action_list"] = actionList
			perms.Perms = convert.Map2pbStruct(permsMap)
			v.Elem().FieldByName("WebAnnotations").Set(reflect.ValueOf(perms))
		}
		return nil
	default:
		dataField := v.Elem().FieldByName("Data")
		if !dataField.IsValid() {
			return nil
		}
		switch dataField.Kind() {
		case reflect.Interface, reflect.Ptr:
			if dataField.Elem().CanSet() {
				tp := reflect.TypeOf(dataField.Elem().Interface())
				dataField.Elem().Set(reflect.Zero(tp))
			}
		default:
			tp := reflect.TypeOf(dataField.Interface())
			dataField.Set(reflect.Zero(tp))
		}
		return nil
	}
}

// getMsgCode 根据不同的错误类型，获取错误信息 & 错误码
func getMsgCode(err interface{}) (string, uint32) {
	if err == nil {
		return errorx.SuccessMsg, errorx.Success
	}
	switch e := err.(type) {
	case *errorx.ProjectError:
		return e.Error(), e.Code()
	case *errors.Error:
		return e.Detail, errorx.InnerErr
	case *authutils.PermDeniedError:
		return err.(*authutils.PermDeniedError).Error(), errorx.NoPermissionErr
	default:
		return fmt.Sprintf("%s", e), errorx.InnerErr
	}
}

// resource ResourceData struct
type resource struct {
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	Namespace   string `json:"namespace" yaml:"namespace"`
	Name        string `json:"name" yaml:"name"`
	Key         string `json:"key" yaml:"key"`
	IDs         string `json:"idList" yaml:"idList"`
	VariableID  string `json:"variableID" yaml:"variableID"`
	ProjectID   string `json:"projectID" yaml:"projectID"`
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
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
	return resourceID
}

// NOCC: golint/unparam(设计如此:)
// nolint
var auditFuncMap = map[string]func(req server.Request, rsp interface{}) (audit.Resource, audit.Action){
	"BCSProject.CreateProject": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeProject, ResourceID: res.ProjectCode, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "project_create", ActivityType: audit.ActivityTypeCreate}
	},
	"BCSProject.UpdateProject": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		// ProjectID 代替 ProjectCode
		if res.ProjectCode == "" {
			res.ProjectCode = res.ProjectID
		}
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeProject, ResourceID: res.ProjectID, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "project_edit", ActivityType: audit.ActivityTypeUpdate}
	},
	"Namespace.CreateNamespace": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_create", ActivityType: audit.ActivityTypeCreate}
	},
	"Namespace.UpdateNamespace": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_update", ActivityType: audit.ActivityTypeUpdate}
	},
	"Namespace.DeleteNamespace": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeNamespace, ResourceID: res.Namespace, ResourceName: res.Namespace,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "namespace_delete", ActivityType: audit.ActivityTypeDelete}
	},
	"Variable.CreateVariable": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.Key, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "create_variable", ActivityType: audit.ActivityTypeCreate}
	},
	"Variable.UpdateVariable": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.Key, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_variable", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.DeleteVariableDefinitions": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.IDs, ResourceName: res.IDs,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "deleteVariable_definitions", ActivityType: audit.ActivityTypeDelete}
	},
	"Variable.UpdateClustersVariables": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.VariableID, ResourceName: res.VariableID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_clusters_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateNamespacesVariables": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.VariableID, ResourceName: res.VariableID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_namespaces_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateClusterVariables": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ProjectCode:  res.ProjectCode,
			ResourceType: audit.ResourceTypeVariable, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_cluster_variables", ActivityType: audit.ActivityTypeUpdate}
	},
	"Variable.UpdateNamespaceVariables": func(req server.Request, rsp interface{}) (audit.Resource, audit.Action) {
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

	res, act := fn(req, rsp)

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
	_ = component.GetAuditClient().R().
		SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}
