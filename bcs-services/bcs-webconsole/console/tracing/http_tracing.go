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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingTransport  used by http Transport
type TracingTransport struct {
	Transport      http.RoundTripper
	TracerProvider trace.TracerProvider
	Propagators    propagation.TextMapPropagator
}

// RoundTrip tracing Transport
func (t *TracingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// 开始时间
	st := time.Now()

	// 解析traceparent
	ctx := r.Context()
	ctx = t.Propagators.Extract(ctx, propagation.HeaderCarrier(r.Header))

	// 记录额外的信息
	commonAttrs := semconv.NetAttributesFromHTTPRequest("tcp", r)
	commonAttrs = append(commonAttrs, attribute.String("component", "http"))

	opts := []trace.SpanStartOption{
		trace.WithAttributes(commonAttrs...),
		trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
		trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(ServiceName, r.URL.Path, r)...),
		trace.WithSpanKind(trace.SpanKindServer),
	}

	spanName := r.URL.Path
	if spanName == "" {
		spanName = fmt.Sprintf("HTTP %s route not found", r.Method)
	}

	// tracerName是检测库的包名
	tracer := t.TracerProvider.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, spanName, opts...)
	defer span.End()

	// 记录query参数
	query := r.URL.Query().Encode()
	if len(query) > 1024 {
		query = fmt.Sprintf("%s...(Total %s)", query[:1024], humanize.Bytes(uint64(len(query))))
	}
	span.SetAttributes(attribute.Key("query").String(query))

	//记录body
	body := string(getRequestBody(r))
	if len(body) > 1024 {
		body = fmt.Sprintf("%s...(Total %s)", body[:1024], humanize.Bytes(uint64(len(body))))
	}
	span.SetAttributes(attribute.Key("body").String(body))

	// 传往下游，traceparent
	t.Propagators.Inject(ctx, propagation.HeaderCarrier(r.Header))
	// pass the span through the request context
	r = r.WithContext(ctx)
	resp, err := t.transport().RoundTrip(r)
	if err != nil {
		span.SetAttributes(attribute.String("errors", err.Error()))
		return resp, err
	}

	respBody := string(getResponseBody(resp))
	if len(respBody) > 1024 {
		respBody = fmt.Sprintf("%s...(Total %s)", respBody[:1024], humanize.Bytes(uint64(len(respBody))))
	}
	span.SetAttributes(attribute.Key("rsp").String(respBody))

	elapsedTime := time.Since(st)
	span.SetAttributes(attribute.Key("elapsed_ime").String(elapsedTime.String()))

	status := resp.StatusCode
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(status)
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(status)
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)

	return resp, err
}

// transport return RoundTripper
func (t *TracingTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// NewTracingTransport make a new tracing transport, default transport can be nil
func NewTracingTransport(transport http.RoundTripper) *TracingTransport {
	return &TracingTransport{
		Transport:      transport,
		TracerProvider: otel.GetTracerProvider(),
		Propagators:    otel.GetTextMapPropagator(),
	}
}

// getRequestBody 获取http请求体
func getRequestBody(r *http.Request) []byte {
	// nil的情况下直接返回
	if r.Body == nil {
		return []byte{}
	}
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return []byte{}
	}
	// 恢复请求体
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

// getResponseBody 获取http请求体
func getResponseBody(r *http.Response) []byte {
	// nil的情况下直接返回
	if r.Body == nil {
		return []byte{}
	}
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return []byte{}
	}
	// 恢复请求体
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}
