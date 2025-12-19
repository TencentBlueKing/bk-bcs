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

package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	traceUtil "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"
	"github.com/dustin/go-humanize"
	restful "github.com/emicklei/go-restful/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/tracing"
)

// TracingFilter add tracing
func TracingFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	cfg := tracing.NewConfig()
	// 开始时间
	startTime := time.Now()
	writer := responseWriter{
		response.ResponseWriter,
		bytes.NewBuffer([]byte{}),
	}
	response.ResponseWriter = writer

	ctx := request.Request.Context()
	// 上游transparent的打通，判断Header 是否有放置Transparent
	traceparent := request.Request.Header.Get(constant.Traceparent)
	if traceparent != "" {
		// 有则从上游解析Transparent
		ctx = cfg.Propagators.Extract(ctx, propagation.HeaderCarrier(request.Request.Header))
	} else {
		// 没有则从request id截取生成
		// get http X-Request-Id
		requestID := request.Request.Header.Get(constant.RequestIDHeaderKey)
		// 使用requestID当作TraceID
		// NOCC:golint/staticcheck(设计如此:)
		// nolint
		ctx = context.WithValue(ctx, constant.RequestIDKey, requestID)
		ctx = traceUtil.ContextWithRequestID(ctx, requestID)
	}

	// 记录额外的信息
	commonAttrs := semconv.NetAttributesFromHTTPRequest("tcp", request.Request)
	commonAttrs = append(commonAttrs, attribute.String("component", "http"))

	opts := []trace.SpanStartOption{
		trace.WithAttributes(commonAttrs...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(request.Request)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(tracing.ServiceName,
			request.Request.URL.Path, request.Request)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	spanName := request.Request.URL.Path
	if spanName == "" {
		spanName = fmt.Sprintf("HTTP %s route not found", request.Request.Method)
	}

	// tracerName是检测库的包名
	tracer := cfg.TracerProvider.Tracer(
		constant.TracerName,
	)

	ctx, span := tracer.Start(ctx, spanName, opts...)
	defer span.End()

	// 记录query参数
	query := request.Request.URL.Query().Encode()
	if len(query) > 1024 {
		query = fmt.Sprintf("%s...(Total %s)", query[:1024], humanize.Bytes(uint64(len(query))))
	}
	span.SetAttributes(attribute.Key("query").String(query))

	// 记录body
	body := string(getRequestBody(request.Request))
	if len(body) > 1024 {
		body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
	}
	span.SetAttributes(attribute.Key("body").String(body))

	// pass the span through the request context
	request.Request = request.Request.WithContext(ctx)

	// 返回Header添加Traceparent
	cfg.Propagators.Inject(ctx, propagation.HeaderCarrier(response.Header()))

	chain.ProcessFilter(request, response)

	respBody := writer.b.String()
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", writer.b.String()[:1024], humanize.Bytes(uint64(len(writer.b.String()))))
	}
	span.SetAttributes(attribute.Key("rsp").String(respBody))

	elapsedTime := time.Since(startTime)
	span.SetAttributes(attribute.Key("elapsed_ime").String(elapsedTime.String()))

	status := response.StatusCode()
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)
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
	http.ResponseWriter
	b *bytes.Buffer
}

// Write Rewrite Write of restful.Response
func (w responseWriter) Write(b []byte) (int, error) {
	// 向一个bytes.buffer中写一份数据来为获取body使用
	w.b.Write(b)
	// 完成http.ResponseWriter.Write()原有功能
	return w.ResponseWriter.Write(b)
}
