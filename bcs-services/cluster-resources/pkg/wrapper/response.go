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

	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// NewResponseFormatWrapper 创建 "格式化返回结果" 装饰器
func NewResponseFormatWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			err := fn(ctx, req, rsp)
			// 若返回结构是标准结构，则这里将错误信息捕获，按照规范格式化到结构体中
			if r, ok := rsp.(*clusterRes.CommonResp); ok {
				r.RequestID = fmt.Sprintf("%s", ctx.Value(common.ContextKey("requestID")))
				if err != nil {
					r.Code = 500
					r.Data = nil
					r.Message = fmt.Sprintf("%s", err)
					// 返回 nil 避免框架重复处理 error
					return nil
				}
				r.Message = "OK"
			}
			return err
		}
	}
}
