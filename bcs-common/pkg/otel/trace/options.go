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

package trace

import (
	"go.opentelemetry.io/otel/attribute"
)

// TraceType tracing type
type TraceType string

const (
	// Jaeger show jaeger system
	Jaeger TraceType = "jaeger"
	// Zipkin show zipkin system
	Zipkin TraceType = "zipkin"
	// NullTrace show opentracing NoopTracer
	NullTrace TraceType = "null"
)

// TracerSwitch sets a factory tracing switch: on or off
func TracerSwitch(s string) Option {
	return func(o *Options) {
		o.TracingSwitch = s
	}
}

// TracerType sets a factory tracing type
func TracerType(t string) Option {
	return func(o *Options) {
		o.TracingType = string(t)
	}
}

// ServiceName sets a service name for a tracing system
func ServiceName(sn string) Option {
	return func(o *Options) {
		o.ServiceName = sn
	}
}

// ExporterURL sets a exporter url for tracing system
func ExporterURL(eu string) Option {
	return func(o *Options) {
		o.ExporterURL = eu
	}
}

// ResourceAttrs sets resource attributes
func ResourceAttrs(ra []attribute.KeyValue) Option {
	return func(o *Options) {
		o.ResourceAttrs = append(o.ResourceAttrs, ra...)
	}
}
