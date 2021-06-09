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

package grpc

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
)

// Option instances may be used in OpenTracing(Server|Client)Interceptor
// initialization.
//
// See this post about the "functional options" pattern:
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(o *options)

// LogPayloads returns an Option that tells the OpenTracing instrumentation to
// try to log application payloads in both directions.
func LogPayloads() Option {
	return func(o *options) {
		o.logPayloads = true
	}
}

// SpanInclusionFunc provides an optional mechanism to decide whether or not
// to trace a given gRPC call. Return true to create a Span and initiate
// tracing, false to not create a Span and not trace.
//
// parentSpanCtx may be nil if no parent could be extraction from either the Go
// context.Context (on the client) or the RPC (on the server).
type SpanInclusionFunc func(
	parentSpanCtx opentracing.SpanContext,
	method string,
	req, resp interface{}) bool

// IncludingSpans binds a IncludeSpanFunc to the options
func IncludingSpans(inclusionFunc SpanInclusionFunc) Option {
	return func(o *options) {
		o.inclusionFunc = inclusionFunc
	}
}

// SpanDecoratorFunc provides an (optional) mechanism for otgrpc users to add
// arbitrary tags/logs/etc to the opentracing.Span associated with client
// and/or server RPCs.
type SpanDecoratorFunc func(
	ctx context.Context,
	span opentracing.Span,
	method string,
	req, resp interface{},
	grpcError error)

// SpanDecorator binds a function that decorates gRPC Spans.
func SpanDecorator(decorator SpanDecoratorFunc) Option {
	return func(o *options) {
		o.decorator = decorator
	}
}

// The internal-only options struct. Obviously overkill at the moment; but will
// scale well as production use dictates other configuration and tuning
// parameters.
type options struct {
	logPayloads bool
	decorator   SpanDecoratorFunc
	// May be nil.
	inclusionFunc SpanInclusionFunc
}

// newOptions returns the default options.
func newOptions() *options {
	return &options{
		logPayloads:   false,
		inclusionFunc: nil,
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}
