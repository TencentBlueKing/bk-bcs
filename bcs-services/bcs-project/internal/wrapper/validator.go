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

	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/ctxkey"
)

type validator interface {
	Validate() error
}

// NewValidatorWrapper 参数校验
func NewValidatorWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		if v, ok := req.Body().(validator); ok {
			if err := v.Validate(); err != nil {
				requestID := ctx.Value(ctxkey.RequestIDKey).(string)
				return RenderResponse(rsp, requestID, err)
			}
		}
		return fn(ctx, req, rsp)
	}
}
