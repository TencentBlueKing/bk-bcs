/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wrapper

import (
	"context"
	"fmt"

	"github.com/micro/go-micro/v2/errors"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// NewResponseWrapper 添加request id, 统一处理返回
func NewResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		err := fn(ctx, req, rsp)
		requestID := ctx.Value(ctxkey.RequestIDKey).(string)
		return RenderResponse(rsp, requestID, err)
	}
}

// RenderResponse 处理返回数据，用于返回统一结构
func RenderResponse(rsp interface{}, requestID string, err error) error {
	switch rsp.(type) {
	case *proto.ProjectResponse:
		if r, ok := rsp.(*proto.ProjectResponse); ok {
			r.RequestID = requestID
			var perm map[string]interface{}
			r.Message, r.Code, perm = getMsgCodePerm(err)
			r.WebAnnotations = &proto.Perms{Perms: convert.Map2pbStruct(perm)}
			if err != nil {
				r.Data = nil
				return nil
			}
		}
	case *proto.ListProjectsResponse:
		if r, ok := rsp.(*proto.ListProjectsResponse); ok {
			r.RequestID = requestID
			r.Message, r.Code = getMsgCode(err)
			if err != nil {
				r.Data = nil
				return nil
			}
		}
	case *proto.ListAuthorizedProjResp:
		if r, ok := rsp.(*proto.ListAuthorizedProjResp); ok {
			r.RequestID = requestID
			r.Message, r.Code = getMsgCode(err)
			if err != nil {
				r.Data = nil
				return nil
			}
		}
	}
	return err
}

// 根据不同的错误类型，获取错误信息 & 错误码
func getMsgCode(err interface{}) (string, uint32) {
	if err == nil {
		return errorx.SuccessMsg, errorx.Success
	}
	switch e := err.(type) {
	case *errorx.ProjectError:
		return e.Error(), e.Code()
	case *errors.Error:
		return e.Detail, errorx.InnerErr
	default:
		return fmt.Sprintf("%s", e), errorx.InnerErr
	}
}

// 获取错误信息 & 错误码 & 权限信息
func getMsgCodePerm(err interface{}) (string, uint32, map[string]interface{}) {
	if err != nil {
		if e, ok := err.(*errorx.PermissionDeniedError); ok {
			return e.Error(), e.Code(), map[string]interface{}{
				"applyUrl":   e.ApplyUrl(),
				e.ActionID(): e.HasPerm(),
			}
		}
	}
	msg, code := getMsgCode(err)
	return msg, code, nil
}
