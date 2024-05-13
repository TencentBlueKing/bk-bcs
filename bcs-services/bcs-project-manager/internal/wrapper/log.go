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

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

// NewLogWrapper 记录流水
func NewLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		md, _ := metadata.FromContext(ctx)
		err := fn(ctx, req, rsp)
		if err != nil {
			logging.Error("method %s failed, request_id: %s, req: %v, err: %s, ctx: %v",
				req.Endpoint(), ctx.Value(ctxkey.RequestIDKey), req.Body(), err.Error(), md)
			return err
		}
		logging.Info("method %s, request_id: %s, req: %v", req.Method(), ctx.Value(ctxkey.RequestIDKey), req.Body())
		return nil
	}
}

// NewAuthLogWrapper 记录鉴权日志
func NewAuthLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		authUser, _ := middleware.GetUserFromContext(ctx)
		logging.Info("authUser: %v, method: %s, request_id: %s", authUser, req.Method(), ctx.Value(ctxkey.RequestIDKey))
		return fn(ctx, req, rsp)
	}
}
