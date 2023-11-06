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

// Package micro xxx
package micro

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/micro/go-micro/v2/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	grpc_codes "google.golang.org/grpc/codes"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"
)

// NewTracingWrapper :
func NewTracingWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		// 开始时间
		startTime := time.Now()

		// 获取或生成 request id 注入到 context
		requestID := utils.GetOrCreateReqID(ctx)
		ctx = context.WithValue(ctx, constants.RequestIDKey, requestID)
		ctx = utils.ContextWithRequestID(ctx, requestID)

		name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

		tracer := otel.Tracer(req.Service())
		commonAttrs := []attribute.KeyValue{
			attribute.String("component", "gRPC"),
			attribute.String("method", req.Method()),
			attribute.String("url", req.Endpoint()),
		}
		ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(commonAttrs...))
		defer span.End()

		reqData, _ := json.Marshal(req.Body())

		err = fn(ctx, req, rsp)

		rspData, _ := json.Marshal(rsp)
		elapsedTime := time.Since(startTime)

		reqBody := string(reqData)
		if len(reqBody) > 1024 {
			reqBody = fmt.Sprintf("%s...(Total %s)", reqBody[:1024], humanize.Bytes(uint64(len(reqBody))))
		}

		respBody := string(rspData)
		if len(respBody) > 1024 {
			respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
		}

		// 以utf-8方式合法截取字符串
		reqBody = strings.ToValidUTF8(reqBody, "")
		respBody = strings.ToValidUTF8(respBody, "")
		// 设置额外标签
		span.SetAttributes(attribute.Key("req").String(reqBody))
		span.SetAttributes(attribute.Key("elapsed_ime").String(elapsedTime.String()))
		span.SetAttributes(attribute.Key("rsp").String(respBody))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.SetAttributes(constants.GRPCStatusCodeKey.Int(int(codes.Error)))
		} else {
			span.SetAttributes(constants.GRPCStatusCodeKey.Int(int(grpc_codes.OK)))
		}

		return err
	}
}
