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

package utils

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer creates a named tracer that implements Tracer interface.
// If the name is an empty string then provider uses default name.
func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

// GetTracerProvider returns the registered global trace provider.
// If none is registered then an instance of NoopTracerProvider is returned.
//
// Use the trace provider to create a named tracer. E.g.
//     tracer := utils.GetTracerProvider().Tracer("example.com/foo")
// or
//     tracer := utils.Tracer("example.com/foo")
func GetTracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

// SetTracerProvider registers `tp` as the global trace provider.
func SetTracerProvider(tp trace.TracerProvider) {
	otel.SetTracerProvider(tp)
}

// NewTracerProvider returns a new and configured TracerProvider.
//
// By default the returned TracerProvider is configured with:
//  - a ParentBased(AlwaysSample) Sampler
//  - a random number IDGenerator
//  - the resource.Default() Resource
//  - the default SpanLimits.
//
// The passed opts are used to override these default values and configure the
// returned TracerProvider appropriately.
func NewTracerProvider(opts ...sdktrace.TracerProviderOption) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(opts...)
}

// WithSyncer registers the exporter with the TracerProvider using a
// SimpleSpanProcessor.
//
// This is not recommended for production use. The synchronous nature of the
// SimpleSpanProcessor that will wrap the exporter make it good for testing,
// debugging, or showing examples of other feature, but it will be slow and
// have a high computation resource usage overhead. The WithBatcher option is
// recommended for production use instead.
func WithSyncer(e sdktrace.SpanExporter) sdktrace.TracerProviderOption {
	return sdktrace.WithSyncer(e)
}

// WithBatcher registers the exporter with the TracerProvider using a
// BatchSpanProcessor configured with the passed opts.
func WithBatcher(e sdktrace.SpanExporter, opts ...sdktrace.BatchSpanProcessorOption) sdktrace.TracerProviderOption {
	return sdktrace.WithBatcher(e, opts...)
}

// WithSpanProcessor registers the SpanProcessor with a TracerProvider.
func WithSpanProcessor(sp sdktrace.SpanProcessor) sdktrace.TracerProviderOption {
	return sdktrace.WithSpanProcessor(sp)
}

// WithResource returns a TracerProviderOption that will configure the
// Resource r as a TracerProvider's Resource. The configured Resource is
// referenced by all the Tracers the TracerProvider creates. It represents the
// entity producing telemetry.
//
// If this option is not used, the TracerProvider will use the
// resource.Default() Resource by default.
func WithResource(r *resource.Resource) sdktrace.TracerProviderOption {
	return sdktrace.WithResource(r)
}

// WithIDGenerator returns a TracerProviderOption that will configure the
// IDGenerator g as a TracerProvider's IDGenerator. The configured IDGenerator
// is used by the Tracers the TracerProvider creates to generate new Span and
// Trace IDs.
//
// If this option is not used, the TracerProvider will use a random number
// IDGenerator by default.
func WithIDGenerator(g sdktrace.IDGenerator) sdktrace.TracerProviderOption {
	return sdktrace.WithIDGenerator(g)
}

// WithSampler returns a TracerProviderOption that will configure the Sampler
// s as a TracerProvider's Sampler. The configured Sampler is used by the
// Tracers the TracerProvider creates to make their sampling decisions for the
// Spans they create.
//
// If this option is not used, the TracerProvider will use a
// ParentBased(AlwaysSample) Sampler by default.
func WithSampler(s sdktrace.Sampler) sdktrace.TracerProviderOption {
	return sdktrace.WithSampler(s)
}

// WithSpanLimits returns a TracerProviderOption that will configure the
// SpanLimits sl as a TracerProvider's SpanLimits. The configured SpanLimits
// are used used by the Tracers the TracerProvider and the Spans they create
// to limit tracing resources used.
//
// If this option is not used, the TracerProvider will use the default
// SpanLimits.
func WithSpanLimits(sl sdktrace.SpanLimits) sdktrace.TracerProviderOption {
	return sdktrace.WithSpanLimits(sl)
}
