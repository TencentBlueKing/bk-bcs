/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package route xxx
package route

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	consoleAudit "github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// resource
type resource struct {
	ClusterID   string `json:"cluster_id" yaml:"cluster_id"`
	ProjectID   string `json:"project_id" yaml:"project_id"`
	ProjectCode string `json:"project_code" yaml:"project_code"`
	RequestID   string `json:"request_id" yaml:"request_id"`
}

// resource to map
func (r resource) toMap() map[string]any {
	result := make(map[string]any, 0)

	if r.ClusterID != "" {
		result["ClusterID"] = r.ClusterID
	}

	if r.ProjectCode != "" {
		result["ProjectCode"] = r.ProjectCode
	}
	if r.RequestID != "" {
		result["RequestID"] = r.RequestID
	}

	return result
}

var auditFuncMap = map[string]func(c *gin.Context) (audit.Resource, audit.Action){
	"GET./api/projects/:projectId/clusters/:clusterId/session/": func(c *gin.Context) (audit.Resource, audit.Action) {
		res := getResourceID(c)
		return audit.Resource{ResourceType: audit.ResourceTypeWebConsole, ProjectCode: res.ProjectCode,
				ResourceID: res.ProjectID, ResourceData: res.toMap()},
			audit.Action{ActionID: "web_console_start", ActivityType: audit.ActivityTypeStart}
	},
}

// AuditHandler 操作记录中间件
func AuditHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		recorder := &respRecorder{ResponseWriter: c.Writer}
		// 自定义respRecorder 替换原始ResponseWriter
		c.Writer = recorder
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		body := recorder.data
		addAudit(c, startTime, endTime, body)
	}
}

// 获取resourceData 的资源
func getResourceID(ctx *gin.Context) resource {
	authCtx := MustGetAuthContext(ctx)
	return resource{
		ClusterID:   authCtx.ClusterId,
		ProjectID:   authCtx.ProjectId,
		RequestID:   authCtx.RequestId,
		ProjectCode: authCtx.ProjectCode,
	}
}

// addAudit
func addAudit(c *gin.Context, startTime, endTime time.Time, data *types.APIResponse) {
	authCtx := MustGetAuthContext(c)
	method := c.Request.Method
	path := c.FullPath()
	s := method + "." + path
	// get method audit func
	fn, ok := auditFuncMap[s]
	if !ok {
		return
	}

	res, act := fn(c)

	auditCtx := audit.RecorderContext{
		Username:  authCtx.Username,
		RequestID: authCtx.RequestId,
		StartTime: startTime,
		EndTime:   endTime,
	}
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

	result := audit.ActionResult{
		Status: audit.ActivityStatusSuccess,
	}
	result.ResultCode = data.Code
	result.ExtraData["Message"] = data.Message

	consoleAudit.GetAuditClient().R().
		SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}

// 自定义respRecorder, 实现http.ResponseWriter 接口
type respRecorder struct {
	gin.ResponseWriter
	data *types.APIResponse
}

// Write 实现Write方法, 捕获响应体
func (r *respRecorder) Write(data []byte) (int, error) {
	d := &types.APIResponse{}
	_ = json.Unmarshal(data, d)
	r.data = d
	return r.ResponseWriter.Write(data)
}
