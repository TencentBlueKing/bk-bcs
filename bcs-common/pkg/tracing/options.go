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

// TracerType set factory tracing type
func TracerType(t TraceType) Option {
	return func(o *Options) {
		o.TracingType = string(t)
	}
}

// TracerSwitch set factory tracing switch: on or off
func TracerSwitch(s string) Option {
	return func(o *Options) {
		o.TracingSwitch = s
	}
}

// ServiceName set service name for tracing system
func ServiceName(sn string) Option {
	return func(o *Options) {
		o.ServiceName = sn
	}
}

// RPCMetrics set tracer rpcMetrics
func RPCMetrics(rm bool) Option {
	return func(opts *Options) {
		opts.RPCMetrics = rm
	}
}

// ReportMetrics set on prometheus report metrics
func ReportMetrics(rm bool) Option {
	return func(opts *Options) {
		opts.ReportMetrics = rm
	}
}

// ReportLog set report tracer/span info
func ReportLog(rl bool) Option {
	return func(opts *Options) {
		opts.ReportLog = rl
	}
}

// AgentFromEnv set report agent from env
func AgentFromEnv(af bool) Option {
	return func(opts *Options) {
		opts.AgentFromEnv = af
	}
}

// AgentHostPort set reporter agent server
func AgentHostPort(ah string) Option {
	return func(opts *Options) {
		opts.AgentHostPort = ah
	}
}

// SampleType set the jaeger samplerType
func SampleType(st string) Option {
	return func(opts *Options) {
		opts.SampleType = st
	}
}

// SampleParameter for sampler parameter
func SampleParameter(sp float64) Option {
	return func(opts *Options) {
		opts.SampleParameter = sp
	}
}

// SampleFromEnv set the jaeger samplerType
func SampleFromEnv(sf bool) Option {
	return func(opts *Options) {
		opts.SampleFromEnv = sf
	}
}

// SamplingServerURL for sampler parameter
func SamplingServerURL(su string) Option {
	return func(opts *Options) {
		opts.SamplingServerURL = su
	}
}
