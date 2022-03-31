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

	"go-micro.dev/v4/errors"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
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
				r.Message, r.Code = getRespMsgCode(err)
				if err != nil {
					r.Data = nil
					// 返回 nil 避免框架重复处理 error
					return nil // nolint:nilerr
				}
			case *clusterRes.CommonListResp:
				r.RequestID = getRequestID(ctx)
				r.Message, r.Code = getRespMsgCode(err)
				if err != nil {
					r.Data = nil
					return nil // nolint:nilerr
				}
			}
			return err
		}
	}
}

// 获取 Context 中的 RequestID
func getRequestID(ctx context.Context) string {
	return fmt.Sprintf("%s", ctx.Value(ctxkey.RequestIDKey))
}

// 根据不同的错误类型，获取错误信息 & 错误码
func getRespMsgCode(err interface{}) (string, int32) {
	if err == nil {
		return "OK", errcode.NoErr
	}

	switch e := err.(type) {
	case *errorx.BaseError:
		return e.Error(), int32(e.Code())
	case *errors.Error:
		return e.Detail, errcode.General
	default:
		return fmt.Sprintf("%s", e), errcode.General
	}
}
