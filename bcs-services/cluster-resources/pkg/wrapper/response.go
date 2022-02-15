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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/types"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// NewResponseFormatWrapper 创建 "格式化返回结果" 装饰器
func NewResponseFormatWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			err := fn(ctx, req, rsp)
			// 若返回结构是标准结构，则这里将错误信息捕获，按照规范格式化到结构体中
			switch r := rsp.(type) {
			case *clusterRes.CommonResp:
				r.RequestID = getRequestID(ctx)
				r.Message = getRespMessage(err)
				if err != nil {
					// 若出现错误，但未特殊指定错误码，则设置为默认值
					if r.Code == 0 {
						r.Code = errcode.DefaultErrCode
					}
					r.Data = nil
					// 返回 nil 避免框架重复处理 error
					return nil
				}
			case *clusterRes.CommonListResp:
				r.RequestID = getRequestID(ctx)
				r.Message = getRespMessage(err)
				if err != nil {
					if r.Code == 0 {
						r.Code = errcode.DefaultErrCode
					}
					r.Data = nil
					return nil
				}
			}
			return err
		}
	}
}

// 获取 Context 中的 RequestID
func getRequestID(ctx context.Context) string {
	return fmt.Sprintf("%s", ctx.Value(types.ContextKey("requestID")))
}

// 根据不同的错误类型，获取错误信息
func getRespMessage(err interface{}) string {
	if err == nil {
		return "OK"
	}

	switch e := err.(type) {
	case *errors.Error:
		return e.Detail
	default:
		return fmt.Sprintf("%s", e)
	}
}
