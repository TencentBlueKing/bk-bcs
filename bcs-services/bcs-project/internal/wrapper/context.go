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

	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// NewInjectRequestIDWrapper 生成 request id, 用于操作审计等便于跟踪
func NewInjectRequestIDWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		// generate uuid， e.g. 40a05290d67a4a39a04c705a0ee56add
		// TODO: trace id by opentelemetry
		uuid := util.GenUUID()
		ctx = context.WithValue(ctx, common.RequestIDKey, uuid)
		return fn(ctx, req, rsp)
	}
}

// NewLogWrapper 记录流水
func NewLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		requestIDKey := common.RequestIDKey
		md, _ := metadata.FromContext(ctx)
		logging.Info("request func %s, request_id: %s, ctx: %v", req.Endpoint(), ctx.Value(requestIDKey), md)
		if err := fn(ctx, req, rsp); err != nil {
			logging.Error("request func %s failed, request_id: %s, ctx: %v, body: %v", req.Endpoint(), ctx.Value(requestIDKey), md, req.Body())
			return err
		}
		return nil
	}
}

// NewResponseWrapper 添加request id, 统一处理返回
func NewResponseWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		err := fn(ctx, req, rsp)
		requestID := ctx.Value(common.RequestIDKey).(string)
		switch rsp.(type) {
		case *proto.ProjectResponse:
			if r, ok := rsp.(*proto.ProjectResponse); ok {
				r.RequestID = requestID
				if err != nil {
					r.Code = common.BcsInnerErr
					r.Data = nil
					r.Message = err.Error()
					return nil
				}
			}
		case *proto.ListProjectsResponse:
			if r, ok := rsp.(*proto.ListProjectsResponse); ok {
				r.RequestID = requestID
				if err != nil {
					r.Code = common.BcsInnerErr
					r.Data = nil
					r.Message = err.Error()
					return nil
				}
			}
		}
		return err
	}
}
