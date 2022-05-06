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

	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
)

// NewLogWrapper 记录访问日志流水装饰器
func NewLogWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			logging.Info(ctx, "called func: %s", req.Endpoint())
			if err := fn(ctx, req, rsp); err != nil {
				logging.Error(ctx, "call func %s failed, body: %v, error: %v", req.Endpoint(), req.Body(), err)
				return err
			}
			return nil
		}
	}
}
