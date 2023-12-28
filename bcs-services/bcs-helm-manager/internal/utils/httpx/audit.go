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

// Package httpx xxx
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

// AuditMiddleware 审计
func AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		// 获取请求的返回内容
		writer := responseWriter{
			w,
			bytes.NewBuffer([]byte{}),
		}
		next.ServeHTTP(writer, r)
		endTime := time.Now()
		// 返回的内容解析成结构体
		baseResponse := BaseResponse{}
		err := json.Unmarshal(writer.b.Bytes(), &baseResponse)
		if err != nil {
			// 有错误不return， 继续审计，只是内容是空
			blog.Errorf("AuditMiddleware Unmarshal failed: %v", baseResponse)
		}
		// 添加审计
		go addAudit(r.Context(), r, baseResponse, startTime, endTime)
	})
}

// resource 资源
type resource struct {
	RepoName    string `json:"repoName" yaml:"repoName"`
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)
	if r.RepoName != "" {
		result["RepoName"] = r.RepoName
	}

	if r.ProjectCode != "" {
		result["ProjectCode"] = r.ProjectCode
	}
	return result
}

// 获资源ID
func getResourceID(req *http.Request) resource {
	resourceID := resource{}
	vars := mux.Vars(req)
	if vars != nil {
		resourceID.RepoName = vars["repoName"]
		resourceID.ProjectCode = vars["projectCode"]
	}
	return resourceID
}

// 返回审计资源
var auditFuncMap = map[string]func(req *http.Request) (audit.Resource, audit.Action){
	"POST./helmmanager/api/v1/projects/{projectCode}/repos/{repoName}/charts/upload": func(req *http.Request) (
		audit.Resource, audit.Action) {
		res := getResourceID(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeChart, ResourceID: res.RepoName, ResourceName: res.RepoName,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: "charts_upload", ActivityType: audit.ActivityTypeCreate}
	},
}

// 添加审计
func addAudit(ctx context.Context, req *http.Request, baseResponse BaseResponse, startTime, endTime time.Time) {

	// 获取原始的请求url path
	path, err := mux.CurrentRoute(req).GetPathTemplate()
	if err != nil {
		blog.Errorf("get current path failed: %v", err)
		return
	}

	// get method audit func
	fn, ok := auditFuncMap[req.Method+"."+path]
	if !ok {
		return
	}

	res, act := fn(req)

	auditCtx := audit.RecorderContext{
		Username:  "auth.GetUserFromCtx(ctx)",
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
		Status:        audit.ActivityStatusSuccess,
		ResultCode:    baseResponse.Code,
		ResultContent: baseResponse.Message,
	}

	// 非0的情况下返回失败状态
	if baseResponse.Code != 0 {
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

// responseWriter 返回ResponseWriter重写
type responseWriter struct {
	http.ResponseWriter
	b *bytes.Buffer
}

// 输出的内容备一份用于审计
func (w responseWriter) Write(b []byte) (int, error) {
	// 向一个bytes.buffer中写一份数据来为获取body使用
	w.b.Write(b)
	// 完成gin.Context.Writer.Write()原有功能
	return w.ResponseWriter.Write(b)
}
