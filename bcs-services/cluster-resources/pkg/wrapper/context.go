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

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common"
)

// NewContextInjectWrapper 创建 "向请求的 Context 注入信息" 装饰器
func NewContextInjectWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// 获取或生成 UUID，并作为 requestID 注入到 context
			uuid := uuid.New().String()
			ctx = context.WithValue(ctx, common.ContextKey("requestID"), uuid)
			// 实际执行业务逻辑，获取返回结果
			return fn(ctx, req, rsp)
		}
	}
}
