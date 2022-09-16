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
	"encoding/json"

	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// NewLogWrapper 记录流水
func NewLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		requestIDKey := ctxkey.RequestIDKey
		md, _ := metadata.FromContext(ctx)
		logging.Info("request func %s, request_id: %s, ctx: %v", req.Endpoint(), ctx.Value(requestIDKey), md)
		if err := fn(ctx, req, rsp); err != nil {
			logging.Error("request func %s failed, request_id: %s, ctx: %v, body: %v",
				req.Endpoint(), ctx.Value(requestIDKey), md, req.Body())
			return err
		}
		return nil
	}
}

// NewAuthLogWrapper 记录鉴权日志
func NewAuthLogWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		if authUser, err := middleware.GetUserFromContext(ctx); err == nil {
			if b, err := json.Marshal(authUser); err == nil {
				logging.Info("authUser: %s, method: %s", string(b), req.Method())
			}
		}
		return fn(ctx, req, rsp)
	}
}
