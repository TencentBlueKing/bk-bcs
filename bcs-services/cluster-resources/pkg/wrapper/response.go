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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/server"
	"google.golang.org/protobuf/types/known/structpb"

	audit2 "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// NewResponseFormatWrapper 创建 "格式化返回结果" 装饰器
func NewResponseFormatWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			startTime := time.Now()
			err := fn(ctx, req, rsp)
			endTime := time.Now()
			// 添加审计
			go addAudit(ctx, req, rsp, startTime, endTime)
			// 若返回结构是标准结构，则这里将错误信息捕获，按照规范格式化到结构体中
			switch r := rsp.(type) {
			case *clusterRes.CommonResp:
				r.RequestID = getRequestID(ctx)
				r.Message, r.Code = getRespMsgCode(err)
				if err != nil {
					r.Data = genNewRespData(ctx, err)
					// 返回 nil 避免框架重复处理 error
					return nil
				}
			case *clusterRes.CommonListResp:
				r.RequestID = getRequestID(ctx)
				r.Message, r.Code = getRespMsgCode(err)
				if err != nil {
					r.Data = nil
					return nil // nolint:nilerr
				}
			}
			return err
		}
	}
}

// getRequestID 获取 Context 中的 RequestID
func getRequestID(ctx context.Context) string {
	return fmt.Sprintf("%s", ctx.Value(ctxkey.RequestIDKey))
}

// getRespMsgCode 根据不同的错误类型，获取错误信息 & 错误码
func getRespMsgCode(err interface{}) (string, int32) {
	if err == nil {
		return "OK", errcode.NoErr
	}

	switch e := err.(type) {
	case *perm.IAMPermError:
		return e.Error(), int32(e.Code)
	case *errorx.BaseError:
		return e.Error(), int32(e.Code())
	case *errors.Error:
		return e.Detail, errcode.General
	default:
		return fmt.Sprintf("%s", e), errcode.General
	}
}

// genNewRespData 根据不同错误类型，更新 Data 字段信息
func genNewRespData(ctx context.Context, err interface{}) *structpb.Struct {
	switch e := err.(type) {
	case *perm.IAMPermError:
		perms, genPermErr := e.Perms()
		if genPermErr != nil {
			log.Warn(ctx, "generate iam perm apply url failed: %v", genPermErr)
		}
		spbPerms, _ := pbstruct.Map2pbStruct(perms)
		return spbPerms
	default:
		return nil
	}
}

type reqResource struct {
	ProjectID   string `json:"projectID" yaml:"projectID"`
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
	Namespace   string `json:"namespace" yaml:"namespace"`
	Name        string `json:"name" yaml:"name"`
	Kind        string `json:"kind" yaml:"kind"`
	Version     string `json:"apiVersion" yaml:"apiVersion"`
	RawData     *struct {
		Version  string `json:"apiVersion" yaml:"apiVersion"`
		Kind     string `json:"kind,omitempty" yaml:"kind"`
		Metadata *struct {
			Namespace string `json:"namespace" yaml:"namespace"`
			Name      string `json:"name" yaml:"name"`
		} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	} `json:"rawData" yaml:"rawData"`
}

// resource to map
func (r reqResource) toMap() map[string]any {
	result := make(map[string]any, 0)
	if r.ProjectID != "" {
		result["ProjectID"] = r.ProjectID
	}
	if r.ProjectCode != "" {
		result["ProjectCode"] = r.ProjectCode
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
	if r.Kind != "" {
		result["Kind"] = r.Kind
	}
	if r.Version != "" {
		result["Version"] = r.Version
	}
	return result
}

func getReqResource(req server.Request) reqResource {
	body := req.Body()
	b, _ := json.Marshal(body)

	resourceID := reqResource{}
	_ = json.Unmarshal(b, &resourceID)

	// 防止rawData数据不存或者格式错误
	if resourceID.RawData != nil && resourceID.RawData.Metadata != nil {
		if resourceID.RawData.Metadata.Name != "" {
			resourceID.Name = resourceID.RawData.Metadata.Name
		}
		if resourceID.RawData.Metadata.Namespace != "" {
			resourceID.Namespace = resourceID.RawData.Metadata.Namespace
		}
		if resourceID.RawData.Kind != "" {
			resourceID.Kind = resourceID.RawData.Kind
		}
		if resourceID.RawData.Version != "" {
			resourceID.Version = resourceID.RawData.Version
		}
	}
	// name没有的情况下使用ProjectCode代替
	if resourceID.Name == "" {
		resourceID.Name = resourceID.ProjectCode
	}
	return resourceID
}

// nolint
var auditFuncMap = map[string]func(req server.Request) (audit.Resource, audit.Action){
	"Get": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeView}
	},
	"Create": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeCreate}
	},
	"Update": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
	"Delete": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeDelete}
	},
	"Restart": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
	"PauseOrResume": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
	"Scale": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
	"Rollout": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
	"Reschedule": func(req server.Request) (audit.Resource, audit.Action) {
		res := getReqResource(req)
		return audit.Resource{
			ResourceType: audit.ResourceTypeK8SResource, ResourceID: res.Name, ResourceName: res.Name,
			ResourceData: res.toMap(),
		}, audit.Action{ActionID: req.Method(), ActivityType: audit.ActivityTypeUpdate}
	},
}

func addAudit(ctx context.Context, req server.Request, rsp interface{}, startTime, endTime time.Time) {
	method := req.Method()
	if req.Method() != "" {
		arr := strings.Split(req.Method(), ".")
		if len(arr) >= 2 {
			if strings.Contains(arr[1], "Get") {
				method = "Get"
			}
			if strings.Contains(arr[1], "Create") {
				method = "Create"
			}
			if strings.Contains(arr[1], "Update") {
				method = "Update"
			}
			if strings.Contains(arr[1], "Delete") {
				method = "Delete"
			}
			if strings.Contains(arr[1], "Restart") {
				method = "Restart"
			}
			if strings.Contains(arr[1], "PauseOrResume") {
				method = "PauseOrResume"
			}
			if strings.Contains(arr[1], "Scale") {
				method = "Scale"
			}
			if strings.Contains(arr[1], "Reschedule") {
				method = "Reschedule"
			}
			if strings.Contains(arr[1], "Rollout") {
				method = "Rollout"
			}
		}
	}

	// get method audit func
	fn, ok := auditFuncMap[method]
	if !ok {
		return
	}

	res, act := fn(req)

	auditCtx := audit.RecorderContext{
		Username:  GetUserFromCtx(ctx),
		SourceIP:  GetSourceIPFromCtx(ctx),
		UserAgent: GetUserAgentFromCtx(ctx),
		RequestID: getRequestID(ctx),
		StartTime: startTime,
		EndTime:   endTime,
	}
	resource := audit.Resource{
		ProjectCode:  GetProjectCodeFromCtx(ctx),
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
		code := int(codeField.Interface().(int32))
		result.ResultCode = code
	}
	if messageField.CanInterface() {
		message := messageField.Interface().(string)
		result.ResultContent = message
	}
	if result.ResultCode != errcode.NoErr {
		result.Status = audit.ActivityStatusFailed
	}

	// add audit
	auditAction := audit2.GetAuditClient().R()
	// 查看类型不用记录 activity
	if act.ActivityType == audit.ActivityTypeView {
		auditAction.DisableActivity()
	}
	_ = auditAction.SetContext(auditCtx).SetResource(resource).SetAction(action).SetResult(result).Do()
}
