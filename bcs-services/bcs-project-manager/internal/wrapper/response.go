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
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/audit"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/ctxkey"
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
		result := audit.ActionResult{
			Status: audit.ActivityStatusSuccess,
		}
		v := reflect.ValueOf(rsp)
		codeField := v.Elem().FieldByName("Code")
		messageField := v.Elem().FieldByName("Message")
		if codeField.CanInterface() {
			if c, ok := codeField.Interface().(uint32); ok {
				code := int(c)
				result.ResultCode = code
			}
		}
		if messageField.CanInterface() {
			if m, ok := messageField.Interface().(string); ok {
				message := m
				result.ResultContent = message
			}
		}
		if result.ResultCode != errorx.Success {
			result.Status = audit.ActivityStatusFailed
		}
		go addAudit(ctx, req, result, startTime, endTime)
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
	if v.Elem().FieldByName("Message").IsValid() {
		v.Elem().FieldByName("Message").SetString(msg)
	}
	if v.Elem().FieldByName("Code").IsValid() {
		v.Elem().FieldByName("Code").SetUint(uint64(code))
	}
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
