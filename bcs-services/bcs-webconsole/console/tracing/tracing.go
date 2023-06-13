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

package tracing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"
)

// MiddleWareTracing middleware tracing
func MiddleWareTracing() gin.HandlerFunc {
	cfg := NewConfig()

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		writer := responseWriter{
			c.Writer,
			bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer

		// get http X-Request-Id
		requestID := c.GetHeader(RequestIDHeaderKey)

		// 使用requestID当作TraceID
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		ctx = utils.ContextWithRequestID(ctx, requestID)

		// 记录额外的信息
		commonAttrs := semconv.NetAttributesFromHTTPRequest("tcp", c.Request)
		commonAttrs = append(commonAttrs, attribute.String("component", "http"))

		opts := []oteltrace.SpanStartOption{
			oteltrace.WithAttributes(commonAttrs...),
			oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request)...),
			oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(ServiceName, c.Request.URL.Path, c.Request)...),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		}

		spanName := c.Request.URL.Path
		if spanName == "" {
			spanName = fmt.Sprintf("HTTP %s route not found", c.Request.Method)
		}

		// tracerName是检测库的包名
		tracer := cfg.TracerProvider.Tracer(
			TracerName,
		)

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// 记录query参数
		query := c.Request.URL.Query().Encode()
		if len(query) > 1024 {
			query = fmt.Sprintf("%s...(Total %s)", query[:1024], humanize.Bytes(uint64(len(query))))
		}
		span.SetAttributes(attribute.Key("query").String(query))

		//记录body
		body := string(getRequestBody(c.Request))
		if len(body) > 1024 {
			body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
		}
		span.SetAttributes(attribute.Key("body").String(body))

		// pass the span through the request context
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		respBody := writer.b.String()
		if len(respBody) > 1024 {
			respBody = fmt.Sprintf("%s...(Total %s)", writer.b.String()[:1024], humanize.Bytes(uint64(len(writer.b.String()))))
		}
		span.SetAttributes(attribute.Key("rsp").String(respBody))

		elapsedTime := time.Since(startTime)
		span.SetAttributes(attribute.Key("elapsed_ime").String(elapsedTime.String()))

		status := middleware.NewWrapResponseWriter(c.Writer, c.Request.ProtoMajor).Status()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)
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

// responseWriter response writer
type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

// Write Rewrite Write of http.ResponseWriter
func (w responseWriter) Write(b []byte) (int, error) {
	// 向一个bytes.buffer中写一份数据来为获取body使用
	w.b.Write(b)
	// 完成http.ResponseWriter.Write()原有功能
	return w.ResponseWriter.Write(b)
}
