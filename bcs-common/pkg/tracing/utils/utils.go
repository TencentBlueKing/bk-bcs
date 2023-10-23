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

// Package utils xxx
package utils

import (
	"context"
	"net/http"

	opentrace "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracinglog "github.com/opentracing/opentracing-go/log"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// StartSpanFromContext starts and returns a Span with `operationName`, using any Span found within `ctx`
// as a ChildOfRef. If no such parent could be found, StartSpanFromContext creates a root (parentless) Span.
func StartSpanFromContext(ctx context.Context, operationName string,
	opts ...opentrace.StartSpanOption) (opentrace.Span, context.Context) {
	span, ctx := opentrace.StartSpanFromContext(ctx, operationName, opts...)

	return span, ctx
}

// WithSpanForContext returns a new `context.Context` that holds a reference to the span
func WithSpanForContext(ctx context.Context, span opentrace.Span) context.Context {
	return opentrace.ContextWithSpan(ctx, span)
}

// GetSpanFromContext returns the `Span` previously associated with `ctx`, or `nil` if no such `Span` could be found.
func GetSpanFromContext(ctx context.Context) opentrace.Span {
	return opentrace.SpanFromContext(ctx)
}

// NewWrapHTTPClientSpan for init WrapHTTPClientSpan
func NewWrapHTTPClientSpan(ctx context.Context, span opentrace.Span, request *http.Request) *WrapHTTPClientSpan {
	return &WrapHTTPClientSpan{
		ctx:     ctx,
		span:    span,
		request: request,
	}
}

// WrapHTTPClientSpan wrap ctx and span for http client request
type WrapHTTPClientSpan struct {
	ctx     context.Context
	span    opentrace.Span
	request *http.Request
}

// GetRootContext get span context
func (wp *WrapHTTPClientSpan) GetRootContext() context.Context {
	if wp == nil {
		return nil
	}

	return wp.ctx
}

// GetRootSpan get root span
func (wp *WrapHTTPClientSpan) GetRootSpan() opentrace.Span {
	if wp == nil {
		return nil
	}

	return wp.span
}

// SetSpan set request span
func (wp *WrapHTTPClientSpan) SetSpan(span opentrace.Span) {
	if wp == nil {
		return
	}

	wp.span = span
}

// SetRequest set client request
func (wp *WrapHTTPClientSpan) SetRequest(request *http.Request) {
	if wp == nil {
		return
	}

	wp.request = request
}

// SpanFinish span finish
func (wp *WrapHTTPClientSpan) SpanFinish() {
	if wp == nil {
		return
	}

	wp.span.Finish()
}

// SetRequestSpanTag set request span tags
func (wp *WrapHTTPClientSpan) SetRequestSpanTag(peerService, peerHandler string) {
	if wp == nil {
		return
	}

	SetSpanKindTag(wp.span, ext.SpanKindRPCClientEnum)
	SetSpanHTTPTag(wp.span, HTTPUrl, wp.request.URL.Path)
	SetSpanHTTPTag(wp.span, HTTPMethod, wp.request.Method)

	SetSpanPeerTag(wp.span, PeerService, peerService)
	SetSpanPeerTag(wp.span, PeerHandler, peerHandler)
}

// InjectTracerIntoHeader inject span.Context into request http header
func (wp *WrapHTTPClientSpan) InjectTracerIntoHeader() error {
	if wp == nil {
		return nil
	}

	err := wp.span.Tracer().Inject(wp.span.Context(), opentrace.HTTPHeaders,
		opentrace.HTTPHeadersCarrier(wp.request.Header))
	if err != nil {
		blog.Errorf("InjectTracerIntoHeader failed: %v", err)
		return err
	}

	return nil
}

// SetRequestSpanResult set response result into span
func (wp *WrapHTTPClientSpan) SetRequestSpanResult(err error, response *http.Response, fields ...tracinglog.Field) {
	if wp == nil {
		return
	}

	if response != nil {
		SetSpanHTTPTag(wp.span, HTTPStatusCode, response.Status)
	}
	if err != nil {
		SetSpanLogTagError(wp.span, err, fields...)
	}
}

// SetRequestSpanTags set common tags
func (wp *WrapHTTPClientSpan) SetRequestSpanTags(key string, value interface{}) {
	if wp == nil {
		return
	}

	wp.span.SetTag(key, value)
}

// SetRequestSpanLogs set common logs
func (wp *WrapHTTPClientSpan) SetRequestSpanLogs(fields ...tracinglog.Field) {
	if wp == nil {
		return
	}

	wp.span.LogFields(fields...)
}
