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

// Package gin xxx
package gin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"
)

// Middleware returns middleware that will trace incoming requests.
// The service parameter should describe the name of the (virtual)
// server handling the request.
func Middleware(server string, opts ...Option) gin.HandlerFunc { // nolint
	cfg := option{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}

	// tracerName是检测库的包名
	tracer := cfg.TracerProvider.Tracer(
		constants.TracerName,
		oteltrace.WithInstrumentationVersion(SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		writer := responseWriter{
			c.Writer,
			bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		for _, f := range cfg.Filters {
			if !f(c.Request) {
				// Serve the request to the next middleware
				// if a filter rejects the request.
				c.Next()
				return
			}
		}
		c.Set(constants.TracerKey, tracer)

		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		// 判断Header 是否有放置Transparent
		traceparent := c.Request.Header.Get(constants.Traceparent)
		if traceparent != "" {
			// 有则从上游解析Transparent
			savedCtx = cfg.Propagators.Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
		} else {
			// 没有则从request id截取生成
			requestID := c.Request.Header.Get(constants.RequestIDHeaderKey)
			// 使用requestID当作TraceID

			savedCtx = context.WithValue(savedCtx, constants.RequestIDKey, requestID)
			savedCtx = utils.ContextWithRequestID(savedCtx, requestID)
		}

		// 记录额外的信息
		commonAttrs := semconv.NetAttributesFromHTTPRequest("tcp", c.Request)
		commonAttrs = append(commonAttrs, attribute.String("component", "http"))

		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(commonAttrs...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(server, c.FullPath(), c.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		spanName := c.FullPath()
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		ctx, span := tracer.Start(savedCtx, spanName, opts...)
		defer span.End()

		// 记录query参数
		query := c.Request.URL.Query().Encode()
		if len(query) > 1024 {
			query = fmt.Sprintf("%s...(Total %s)", query[:1024], humanize.Bytes(uint64(len(query))))
		}
		// 以utf-8方式合法截取字符串
		query = strings.ToValidUTF8(query, "")
		span.SetAttributes(attribute.Key("query").String(query))

		// 记录body
		body := string(getRequestBody(c.Request))
		if len(body) > 1024 {
			body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
		}
		// 以utf-8方式合法截取字符串
		body = strings.ToValidUTF8(body, "")
		span.SetAttributes(attribute.Key("body").String(body))

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)

		// 返回写入traceparent
		cfg.Propagators.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		// serve the request to the next middleware
		c.Next()

		// 记录响应参数和响应时间
		respBody := writer.b.String()
		if len(respBody) > 1024 {
			respBody = fmt.Sprintf("%s...(Total %s)", writer.b.String()[:1024], humanize.Bytes(uint64(len(writer.b.String()))))
		}
		// 以utf-8方式合法截取字符串
		respBody = strings.ToValidUTF8(respBody, "")
		span.SetAttributes(attribute.Key("rsp").String(respBody))
		elapsedTime := time.Since(startTime)
		span.SetAttributes(attribute.Key("elapsed_ime").String(elapsedTime.String()))

		status := c.Writer.Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
		}
	}

}

// 获取请求体
func getRequestBody(r *http.Request) []byte {
	// 读取请求体
	body, _ := io.ReadAll(r.Body)
	// 恢复请求体
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	// 向一个bytes.buffer中写一份数据来为获取body使用
	w.b.Write(b)
	// 完成gin.Context.Writer.Write()原有功能
	return w.ResponseWriter.Write(b)
}
