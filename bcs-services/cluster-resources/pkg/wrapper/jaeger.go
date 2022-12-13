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
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
)

type MDReaderWriter struct {
	metadata.MD
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

func NewHandlerWrapper(tracer opentracing.Tracer) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			// 开始时间
			startTime := time.Now().UnixNano() / 1000000
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}

			spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{MD: md})
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				grpclog.Errorf("extract from metadata err: %v", err)
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span := tracer.StartSpan(
				name,
				opentracing.ChildOf(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)

			defer span.Finish()

			ext.HTTPMethod.Set(span, req.Method())
			ext.HTTPUrl.Set(span, req.Endpoint())

			ctx = opentracing.ContextWithSpan(ctx, span)

			ctx = context.WithValue(ctx, "Tracer", tracer)
			ctx = context.WithValue(ctx, "ParentSpanContext", span.Context())

			err = fn(ctx, req, rsp)
			// 结束时间
			endTime := time.Now().UnixNano() / 1000000
			costTime := fmt.Sprintf("%vms", endTime-startTime)
			// 设置额外标签
			span.SetTag("request", req.Body())
			span.SetTag("reply", rsp)
			span.SetTag("cost_time", costTime)
			span.SetTag("parent-id", ctx.Value(ctxkey.RequestIDKey))
			if err != nil {
				span.LogFields(opentracinglog.String("error", err.Error()))
				span.SetTag("error", true)
			}

			return err
		}
	}
}
