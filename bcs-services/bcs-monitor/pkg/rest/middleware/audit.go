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

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// Audit middleware audit
func Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 开始时间
		startTime := time.Now()
		// 需要在审计操作记录中对body进行解析
		reqBody := getRequestBody(r)
		writer := responseWriter{
			w,
			bytes.NewBuffer([]byte{}),
		}
		w = writer
		next.ServeHTTP(w, r)

		restContext, err := rest.GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithUnauthorizedError(rest.InitRestContext(w, r), err))
			return
		}

		if _, ok := auditFuncMap[restContext.HandleName]; !ok {
			return
		}

		result := rest.Result{}
		err = json.Unmarshal(writer.b.Bytes(), &result)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithUnauthorizedError(rest.InitRestContext(w, r), err))
			return
		}
		endTime := time.Now()
		addAudit(restContext, reqBody, startTime, endTime, result.Code, restContext.HandleName, result.Message)
	})
}

// 审计中心新增操作记录
func addAudit(ctx *rest.Context, b []byte, startTime, endTime time.Time, code int, handleName, message string) {
	// get method audit func
	fn := auditFuncMap[handleName]

	res, act := fn(b, ctx)

	auditCtx := audit.RecorderContext{
		Username:  ctx.Username,
		RequestID: ctx.RequestId,
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  ctx.ProjectCode,
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
		Status:        audit.ActivityStatusSuccess,
		ResultCode:    code,
		ResultContent: message,
	}

	// code不为0的情况则为失败
	if code != 0 {
		result.Status = audit.ActivityStatusFailed
	}

	// add audit
	auditAction := component.GetAuditClient().R()
	// 查看类型不用记录activity
	if act.ActivityType == audit.ActivityTypeView {
		auditAction.DisableActivity()
	}
	_ = auditAction.SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}

var auditFuncMap = map[string]func(b []byte, ctx *rest.Context) (audit.Resource, audit.Action){
	"GetEntrypoints": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.ClusterID, ResourceName: res.ClusterID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"CreateLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		// resourceData解析
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "create_log_rule", ActivityType: audit.ActivityTypeCreate}
	},
	"GetLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "get_log_rule", ActivityType: audit.ActivityTypeView}
	},
	"UpdateLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "update_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"DeleteLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "delete_log_rule", ActivityType: audit.ActivityTypeDelete}
	},
	"RetryLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "retry_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"EnableLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "enable_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
	"DisableLogRule": func(
		b []byte, ctx *rest.Context) (audit.Resource, audit.Action) {
		res := getResourceID(b, ctx)
		return audit.Resource{
			ResourceType: audit.ResourceTypeLogRule, ResourceID: res.RuleID, ResourceName: res.RuleID,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "disable_log_rule", ActivityType: audit.ActivityTypeUpdate}
	},
}

type resource struct {
	ClusterID string `json:"cluster_id" yaml:"cluster_id"`
	ProjectID string `json:"project_id" yaml:"project_id"`
	Name      string `json:"name" yaml:"name"`
	RuleID    string `json:"-" yaml:"-"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)

	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}

	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}

	if r.Name != "" {
		result["Name"] = r.Name
	}

	if r.RuleID != "" {
		result["RuleID"] = r.RuleID
	}

	return result
}

// 获取resourceData 的资源
func getResourceID(b []byte, ctx *rest.Context) resource {
	resourceID := resource{}
	_ = json.Unmarshal(b, &resourceID)
	resourceID.ClusterID = ctx.ClusterId
	resourceID.ProjectID = ctx.ProjectId
	resourceID.RuleID = chi.URLParam(ctx.Request, "id")
	return resourceID
}
