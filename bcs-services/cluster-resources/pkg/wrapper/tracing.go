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
 *
 */

package wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"go-micro.dev/v4/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
)

func NewTracingWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			// 开始时间
			startTime := time.Now().UnixNano() / 1000000

			// 把request-id格式转成trace-id格式
			requestID := ctx.Value(ctxkey.RequestIDKey).(string)
			if requestID != "" {
				requestID = strings.Replace(requestID, "-", "", -1)
				tid, _ := trace.TraceIDFromHex(requestID)
				sid, _ := trace.SpanIDFromHex(requestID)
				sc := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    tid,
					SpanID:     sid,
					TraceFlags: trace.FlagsSampled,
					Remote:     true,
				})
				ctx = trace.ContextWithSpanContext(ctx, sc)
			}

			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

			tracer := otel.Tracer(req.Endpoint())
			commonAttrs := []attribute.KeyValue{
				attribute.String("component", "gRPC"),
				attribute.String("method", req.Method()),
				attribute.String("url", req.Endpoint()),
			}
			ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(commonAttrs...))
			defer span.End()

			reqData, _ := json.Marshal(req.Body())

			err = fn(ctx, req, rsp)

			rspData, _ := json.Marshal(rsp)
			// 结束时间
			endTime := time.Now().UnixNano() / 1000000
			costTime := fmt.Sprintf("%vms", endTime-startTime)

			reqBody := string(reqData)
			if len(reqBody) > 1024 {
				reqBody = fmt.Sprintf("%s...(Total %s)", reqBody[:1024], humanize.Bytes(uint64(len(reqBody))))
			}

			respBody := string(rspData)
			if len(respBody) > 1024 {
				respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
			}

			// 设置额外标签
			span.SetAttributes(attribute.Key("req").String(reqBody))
			span.SetAttributes(attribute.Key("cost_time").String(costTime))
			span.SetAttributes(attribute.Key("rsp").String(respBody))
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			return err
		}
	}
}
