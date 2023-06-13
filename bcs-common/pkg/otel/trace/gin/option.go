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

package gin

import (
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type option struct {
	TracerProvider    oteltrace.TracerProvider
	Propagators       propagation.TextMapPropagator
	Filters           []Filter
	SpanNameFormatter SpanNameFormatter
}

// Filter is a predicate used to determine whether a given http.request should
// be traced. A Filter must return true if the request should be traced.
type Filter func(*http.Request) bool

// SpanNameFormatter is used to set span name by http.request.
type SpanNameFormatter func(r *http.Request) string

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*option)
}

type optionFunc func(*option)

func (o optionFunc) apply(c *option) {
	o(c)
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *option) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return optionFunc(func(cfg *option) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithFilter adds a filter to the list of filters used by the handler.
// If any filter indicates to exclude a request then the request will not be
// traced. All filters must allow a request to be traced for a Span to be created.
// If no filters are provided then all requests are traced.
// Filters will be invoked for each processed request, it is advised to make them
// simple and fast.
func WithFilter(f ...Filter) Option {
	return optionFunc(func(c *option) {
		c.Filters = append(c.Filters, f...)
	})
}

// WithSpanNameFormatter takes a function that will be called on every
// request and the returned string will become the Span Name.
func WithSpanNameFormatter(f func(r *http.Request) string) Option {
	return optionFunc(func(c *option) {
		c.SpanNameFormatter = f
	})
}
