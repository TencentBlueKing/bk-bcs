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
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ResponseWrapper 处理返回
func ResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		requestID := getRequestID(ctx)
		startTime := time.Now()
		ctx = context.WithValue(ctx, contextx.RequestIDContextKey, requestID)
		err = fn(ctx, req, rsp)
		endTime := time.Now()

		// 添加审计
		go addAudit(ctx, req, rsp, startTime, endTime)
		return renderResponse(rsp, requestID, err)
	}
}

func renderResponse(rsp interface{}, requestID string, err error) error {
	v := reflect.ValueOf(rsp)
	if v.Elem().FieldByName("RequestID") != (reflect.Value{}) {
		v.Elem().FieldByName("RequestID").Set(reflect.ValueOf(&requestID))
	}
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *authutils.PermDeniedError:
		errCode := uint32(common.NoPermissionErr)
		errMsg := err.(*authutils.PermDeniedError).Error()
		if v.Elem().FieldByName("Code") != (reflect.Value{}) {
			v.Elem().FieldByName("Code").Set(reflect.ValueOf(&errCode))
		}
		if v.Elem().FieldByName("Message") != (reflect.Value{}) {
			v.Elem().FieldByName("Message").Set(reflect.ValueOf(&errMsg))
		}
		if v.Elem().FieldByName("WebAnnotations").IsValid() {
			perms := &proto.WebAnnotations{}
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
			perms.Perms = common.Map2pbStruct(permsMap)
			v.Elem().FieldByName("WebAnnotations").Set(reflect.ValueOf(perms))
			return nil
		}
		return err
	default:
		return err
	}
}

// getRequestID 获取 request id
func getRequestID(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return stringx.GenUUID()
	}
	// 当request id不存在或者为空时，生成id
	requestID, ok := md.Get(contextx.RequestIDHeaderKey)
	if !ok || requestID == "" {
		return stringx.GenUUID()
	}

	return requestID
}

// actionDesc 操作描述
type actionDesc string // nolint

// String string
func (a actionDesc) String() string { // nolint
	return string(a)
}

type resource struct {
	RepoName  string `json:"repoName" yaml:"repoName"`
	ClusterID string `json:"clusterID" yaml:"clusterID"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Version   string `json:"version" yaml:"version"`
	Chart     string `json:"chart" yaml:"chart"`
	Revision  uint32 `json:"revision" yaml:"revision"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)
	if r.RepoName != "" {
		result["RepoName"] = r.RepoName
	}
	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}
	if r.Namespace != "" {
		result["Namespace"] = r.Namespace
	}
	if r.Name != "" {
		result["Name"] = r.Name
	}
	if r.Version != "" {
		result["Version"] = r.Version
	}
	if r.Chart != "" {
		result["Chart"] = r.Chart
	}
	if r.Revision != 0 {
		result["Revision"] = r.Revision
	}
	return result
}

func getResourceID(req server.Request) resource {
	body := req.Body()
	b, _ := json.Marshal(body)

	resourceID := resource{}
	_ = json.Unmarshal(b, &resourceID)
	return resourceID
}

var auditFuncMap = map[string]func(req server.Request) (audit.Resource, audit.Action){
	"HelmManager.GetChartDetailV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_chart_detail", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetVersionDetailV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_version_detail", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.DeleteChart": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "delete_chart", ActivityType: audit.ActivityTypeDelete}
	},
	"HelmManager.DeleteChartVersion": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "delete_chart_version", ActivityType: audit.ActivityTypeDelete}
	},
	"HelmManager.GetChartRelease": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_chart_release", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetReleaseDetailV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_release_detail", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.InstallReleaseV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "install_release", ActivityType: audit.ActivityTypeCreate}
	},
	"HelmManager.UninstallReleaseV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "uninstall_release", ActivityType: audit.ActivityTypeDelete}
	},
	"HelmManager.UpgradeReleaseV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "upgrade_release", ActivityType: audit.ActivityTypeUpdate}
	},
	"HelmManager.RollbackReleaseV1": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "rollback_release", ActivityType: audit.ActivityTypeUpdate}
	},
	"HelmManager.ReleasePreview": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "release_preview", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetReleaseHistory": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_release_history", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetReleaseManifest": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_release_manifest", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetReleaseStatus": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_release_status", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetReleasePods": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeHelm, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_release_pods", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.GetAddonsDetail": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeAddons, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_addons_detail", ActivityType: audit.ActivityTypeView}
	},
	"HelmManager.InstallAddons": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeAddons, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "install_addons", ActivityType: audit.ActivityTypeCreate}
	},
	"HelmManager.UpgradeAddons": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeAddons, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "upgrade_addons", ActivityType: audit.ActivityTypeUpdate}
	},
	"HelmManager.UninstallAddons": func(req server.Request) (audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeAddons, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "uninstall_addons", ActivityType: audit.ActivityTypeDelete}
	},
}

func addAudit(ctx context.Context, req server.Request, rsp interface{}, startTime, endTime time.Time) {
	// get method audit func
	fn, ok := auditFuncMap[req.Method()]
	if !ok {
		return
	}

	res, act := fn(req)

	auditCtx := audit.RecorderContext{
		Username:  auth.GetUserFromCtx(ctx),
		SourceIP:  contextx.GetSourceIPFromCtx(ctx),
		UserAgent: contextx.GetUserAgentFromCtx(ctx),
		RequestID: contextx.GetRequestIDFromCtx(ctx),
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  contextx.GetProjectCodeFromCtx(ctx),
		ResourceType: res.ResourceType,
		ResourceID:   res.ResourceID,
		ResourceName: res.ResourceName,
		ResourceData: res.ResourceData,
	}
	action := audit.Action{
		ActionID:     act.ActionID,
		ActivityType: act.ActivityType,
	}

	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
	}

	// get handle result
	v := reflect.ValueOf(rsp)
	codeField := v.Elem().FieldByName("Code")
	messageField := v.Elem().FieldByName("Message")
	if codeField.CanInterface() {
		code := int(*codeField.Interface().(*uint32))
		result.ResultCode = code
	}
	if messageField.CanInterface() {
		message := *messageField.Interface().(*string)
		result.ResultContent = message
	}
	if result.ResultCode != int(common.ErrHelmManagerSuccess) {
		result.Status = audit.ActivityStatusFailed
	}

	// add audit
	auditAction := component.GetAuditClient().R()
	// 查看类型不用记录 activity
	if act.ActivityType == audit.ActivityTypeView {
		auditAction.DisableActivity()
	}
	_ = auditAction.SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}
