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
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/otlp/otlpgrpctrace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/otlp/otlphttptrace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/resource"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"

	"go.opentelemetry.io/otel/attribute"
	oteljaeger "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

var (
	// errSwitchType switch type error
	errSwitchType error = errors.New("error switch type, please input: [on or off]")
	// errTracingType tracing type error
	errTracingType error = errors.New("error tracing type, please input: [jaeger, zipkin, otlpgrpc or otlphttp]")
	// errServiceName for service name is null
	errServiceName error = errors.New("error service name is null")
)

const (
	// defaultSwitchType for default switch type
	defaultSwitchType = "off"
	// defaultTracingType for default tracing type
	defaultTracingType = "jaeger"
)

// TracerProviderConfig set TracerProviderConfig for different tracing systems
type TracerProviderConfig struct {
	TracingSwitch   string                `json:"tracingSwitch" value:"off" usage:"tracing switch"`
	TracingType     string                `json:"tracingType" value:"jaeger" usage:"tracing type"`
	ServiceName     string                `json:"serviceName" usage:"tracing serviceName"`
	JaegerConfig    JaegerConfig          `json:"jaegerConfig,omitempty"`
	OTLPConfig      OTLPConfig            `json:"otlpConfig,omitempty"`
	Sampler         SamplerType           `json:"sampler,omitempty"`
	IDGenerator     sdktrace.IDGenerator  `json:"idGenerator,omitempty" usage:"idGenerator to generate trace id and span id"`
	ResourceAttrs   []attribute.KeyValue  `json:"resourceAttrs,omitempty" usage:"attributes for the traced service"`
	ResourceOptions []otelresource.Option `json:"-"`
}

type JaegerConfig struct {
	CollectorEndpoint jaeger.CollectorEndpoint `json:"collectorEndpoint,omitempty"`
	AgentEndpoint     jaeger.AgentEndpoint     `json:"agentEndpoint,omitempty"`
}

type OTLPConfig struct {
	GRPCConfig otlpgrpctrace.GRPCConfig `json:"grpcConfig,omitempty"`
	HTTPConfig otlphttptrace.HTTPConfig `json:"httpConfig,omitempty"`
}

type SamplerType struct {
	AlwaysOnSampler   bool    `json:"alwaysOnSampler,omitempty" value:"false" usage:"alwaysOnSampler will always sample"`
	AlwaysOffSampler  bool    `json:"alwaysOffSampler,omitempty" value:"false" usage:"alwaysOffSampler will never sample"`
	RatioBasedSampler float64 `json:"ratioBasedSampler,omitempty" value:"0" usage:"ratioBasedSampler will sample base on a ratio"`
	DefaultOnSampler  bool    `json:"defaultOnSampler,omitempty" value:"false" usage:"defaultOnSampler set a parent based always on sample"`
	DefaultOffSampler bool    `json:"defaultOffSampler,omitempty" value:"false" usage:"defaultOffSampler set a parent based always off sample"`
}

// InitTracerProvider initialize an OTLP tracer provider with processors and exporters.
func InitTracerProvider(serviceName string, options ...TracerProviderOption) (context.Context, *sdktrace.TracerProvider, error) {
	defaultOptions := &TracerProviderConfig{
		TracingSwitch: defaultSwitchType,
		TracingType:   defaultTracingType,
		ServiceName:   serviceName,
	}

	for _, o := range options {
		o(defaultOptions)
	}

	err := validateTracingOptions(defaultOptions)
	if err != nil {
		blog.Errorf("validateTracingOptions failed: %v", err)
		return nil, nil, err
	}

	ctx := context.Background()
	if defaultOptions.TracingSwitch == "off" {
		return ctx, &sdktrace.TracerProvider{}, err
	}

	resource, err := initResource(ctx, defaultOptions)
	if err != nil {
		return ctx, &sdktrace.TracerProvider{}, err
	}
	sampler := initSampler(defaultOptions)

	switch defaultOptions.TracingType {
	case string(Jaeger):
		blog.Info("Creating jaeger exporter...")
		if defaultOptions.JaegerConfig.AgentEndpoint.Host != "" && defaultOptions.JaegerConfig.AgentEndpoint.Port != "" {
			opts := append(initAgentEndpointOptions(defaultOptions),
				defaultOptions.JaegerConfig.AgentEndpoint.AgentOptions...)
			jaegerExporter, err := jaeger.NewAgentExporter(opts...)
			if err != nil {
				blog.Errorf("%s: %v", "failed to create jaeger exporter", err)
				return ctx, &sdktrace.TracerProvider{}, err
			}
			processors := initProcessors(jaegerExporter)
			return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
		}
		if defaultOptions.JaegerConfig.CollectorEndpoint.Endpoint != "" {
			opts := append(initCollectorEndpointOptions(defaultOptions),
				defaultOptions.JaegerConfig.CollectorEndpoint.CollectorOptions...)
			jaegerExporter, err := jaeger.NewCollectorExporter(opts...)
			if err != nil {
				blog.Errorf("%s: %v", "failed to create jaeger exporter", err)
				return ctx, &sdktrace.TracerProvider{}, err
			}
			processors := initProcessors(jaegerExporter)
			return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
		}

		blog.Info("Jaeger agent and collector are not set, trying default agent endpoint %s:%v first...",
			DefaultJaegerAgentEndpointHost, DefaultJaegerAgentEndpointPort)
		defaultOptions.JaegerConfig.AgentEndpoint.Host = DefaultJaegerAgentEndpointHost
		defaultOptions.JaegerConfig.AgentEndpoint.Port = DefaultJaegerAgentEndpointPort
		agentOpts := initAgentEndpointOptions(defaultOptions)
		jaegerExporter, err := jaeger.NewAgentExporter(agentOpts...)

		if err != nil {
			blog.Info("failed to connect default jaeger agent, trying default jaeger collector endpoint %s...",
				DefaultJaegerCollectorEndpoint)
			defaultOptions.JaegerConfig.CollectorEndpoint.Endpoint = DefaultJaegerCollectorEndpoint
			collectorOpts := initCollectorEndpointOptions(defaultOptions)
			jaegerExporter, err = jaeger.NewCollectorExporter(collectorOpts...)
			if err != nil {
				blog.Errorf("%s: %v", "failed to create jaeger exporter", err)
				return ctx, &sdktrace.TracerProvider{}, err
			}
		}
		processors := initProcessors(jaegerExporter)
		return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
	case string(OTLP_GRPC):
		blog.Info("Using otlpgrpc exporter...")
		if defaultOptions.OTLPConfig.GRPCConfig.GRPCEndpoint != "" {
			if defaultOptions.OTLPConfig.GRPCConfig.GRPCURLPath == "" {
				defaultOptions.OTLPConfig.GRPCConfig.GRPCURLPath = DefaultOTLPColTracesPath
			}
			opts := append(initGRPCConfigOptions(defaultOptions), defaultOptions.OTLPConfig.GRPCConfig.GRPCOptions...)
			traceClient := otlpgrpctrace.NewClient(opts...)
			grpcExporter, err := otlpgrpctrace.New(ctx, traceClient)
			if err != nil {
				blog.Errorf("%s: %v", "failed to create otelgrpc exporter", err)
				return ctx, &sdktrace.TracerProvider{}, err
			}
			processors := initProcessors(grpcExporter)
			return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
		}

		blog.Info("Using default OTLPGrpc endpoint: %s:%v", DefaultOTLPCollectorHost, DefaultOTLPCollectorPort)
		if defaultOptions.OTLPConfig.GRPCConfig.GRPCURLPath == "" {
			defaultOptions.OTLPConfig.GRPCConfig.GRPCURLPath = DefaultOTLPColTracesPath
		}
		opts := initGRPCConfigOptions(defaultOptions)
		traceClient := otlpgrpctrace.NewClient(opts...)
		grpcExporter, err := otlpgrpctrace.New(ctx, traceClient)
		if err != nil {
			blog.Errorf("%s: %v", "failed to create otelgrpc exporter", err)
			return ctx, &sdktrace.TracerProvider{}, err
		}
		processors := initProcessors(grpcExporter)
		return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
	case string(OTLP_HTTP):
		blog.Info("Using otlphttp exporter...")
		if defaultOptions.OTLPConfig.HTTPConfig.HTTPEndpoint != "" {
			if defaultOptions.OTLPConfig.HTTPConfig.HTTPURLPath == "" {
				defaultOptions.OTLPConfig.HTTPConfig.HTTPURLPath = DefaultOTLPColTracesPath
			}
			opts := append(initHTTPConfigOptions(defaultOptions), defaultOptions.OTLPConfig.HTTPConfig.HTTPOptions...)
			httpExporter, err := otlphttptrace.New(ctx, opts...)
			if err != nil {
				blog.Errorf("%s: %v", "failed to create otlphttp exporter", err)
				return ctx, &sdktrace.TracerProvider{}, err
			}
			processors := initProcessors(httpExporter)
			return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
		}

		blog.Info("Using default OTLPHttp endpoint: %s:%v", DefaultOTLPCollectorHost, DefaultOTLPCollectorPort)
		defaultOptions.OTLPConfig.HTTPConfig.HTTPEndpoint =
			fmt.Sprintf("%s:%d", DefaultOTLPCollectorHost, DefaultOTLPCollectorPort)
		if defaultOptions.OTLPConfig.HTTPConfig.HTTPURLPath == "" {
			defaultOptions.OTLPConfig.HTTPConfig.HTTPURLPath = DefaultOTLPColTracesPath
		}
		opts := initHTTPConfigOptions(defaultOptions)
		grpcExporter, err := otlphttptrace.New(ctx, opts...)
		if err != nil {
			blog.Errorf("%s: %v", "failed to create otlphttp exporter", err)
			return ctx, &sdktrace.TracerProvider{}, err
		}
		processors := initProcessors(grpcExporter)
		return newTracerProvider(ctx, processors, resource, sampler, defaultOptions.IDGenerator)
	case string(Zipkin):
	}
	return ctx, &sdktrace.TracerProvider{}, nil
}

func newTracerProvider(ctx context.Context, processors []sdktrace.SpanProcessor,
	resource *otelresource.Resource, sampler sdktrace.Sampler, gen sdktrace.IDGenerator) (
	context.Context, *sdktrace.TracerProvider, error) {
	var tpos []sdktrace.TracerProviderOption
	for i := 0; i < len(processors); i++ {
		tpos = append(tpos, utils.WithSpanProcessor(processors[i]))
	}
	tpos = append(tpos, utils.WithResource(resource), utils.WithSampler(sampler))
	if gen != nil {
		tpos = append(tpos, utils.WithIDGenerator(gen))
	}
	tp := utils.NewTracerProvider(tpos...)
	utils.SetTracerProvider(tp)
	return ctx, tp, nil
}

// ValidateTracerProviderOption sets a slice of TracerProviderOption based on a tracer provider configuration.
func ValidateTracerProviderOption(config *TracerProviderConfig) []TracerProviderOption {
	var tpos []TracerProviderOption
	if config.TracingSwitch != "" {
		tpos = append(tpos, TracerSwitch(config.TracingSwitch))
	}
	if config.TracingType != "" {
		tpos = append(tpos, TracerType(config.TracingType))
	}
	if config.ServiceName != "" {
		tpos = append(tpos, ServiceName(config.ServiceName))
	}
	if config.JaegerConfig.CollectorEndpoint.Endpoint != "" {
		tpos = append(tpos, JaegerCollectorEndpoint(config.JaegerConfig.CollectorEndpoint.Endpoint))
	}
	if config.JaegerConfig.CollectorEndpoint.Username != "" {
		tpos = append(tpos, JaegerCollectorUsername(config.JaegerConfig.CollectorEndpoint.Username))
	}
	if config.JaegerConfig.CollectorEndpoint.Password != "" {
		tpos = append(tpos, JaegerCollectorPassword(config.JaegerConfig.CollectorEndpoint.Password))
	}
	if config.JaegerConfig.AgentEndpoint.Host != "" {
		tpos = append(tpos, JaegerAgentHost(config.JaegerConfig.AgentEndpoint.Host))
	}
	if config.JaegerConfig.AgentEndpoint.Port != "" {
		tpos = append(tpos, JaegerAgentPort(config.JaegerConfig.AgentEndpoint.Port))
	}
	if config.OTLPConfig.GRPCConfig.GRPCEndpoint != "" {
		tpos = append(tpos, WithOTLPGRPCEndpoint(config.OTLPConfig.GRPCConfig.GRPCEndpoint))
	}
	if config.OTLPConfig.GRPCConfig.GRPCURLPath != "" {
		tpos = append(tpos, WithOTLPGRPCURLPath(config.OTLPConfig.GRPCConfig.GRPCURLPath))
	}
	if config.OTLPConfig.GRPCConfig.GRPCInsecure {
		tpos = append(tpos, WithOTLPGRPCInsecure())
	}
	if config.OTLPConfig.HTTPConfig.HTTPEndpoint != "" {
		tpos = append(tpos, WithOTLPHTTPEndpoint(config.OTLPConfig.HTTPConfig.HTTPEndpoint))
	}
	if config.OTLPConfig.HTTPConfig.HTTPURLPath != "" {
		tpos = append(tpos, WithOTLPHTTPURLPath(config.OTLPConfig.HTTPConfig.HTTPURLPath))
	}
	if config.OTLPConfig.HTTPConfig.HTTPInsecure {
		tpos = append(tpos, WithOTLPHTTPInsecure())
	}
	if config.ResourceAttrs != nil {
		tpos = append(tpos, ResourceAttrs(config.ResourceAttrs))
	}
	if config.Sampler.AlwaysOnSampler {
		tpos = append(tpos, WithAlwaysOnSampler())
	}
	if config.Sampler.AlwaysOffSampler {
		tpos = append(tpos, WithAlwaysOffSampler())
	}
	if fmt.Sprint(config.Sampler.RatioBasedSampler) != "0" {
		tpos = append(tpos, WithRatioBasedSampler(config.Sampler.RatioBasedSampler))
	}
	if config.Sampler.DefaultOnSampler {
		tpos = append(tpos, WithDefaultOnSampler())
	}
	if config.Sampler.DefaultOffSampler {
		tpos = append(tpos, WithDefaultOffSampler())
	}
	return tpos
}

// initSampler returns a default off sampler by default
func initSampler(tpc *TracerProviderConfig) sdktrace.Sampler {
	if tpc.Sampler.AlwaysOnSampler {
		return sdktrace.AlwaysSample()
	}
	if tpc.Sampler.AlwaysOffSampler {
		return sdktrace.NeverSample()
	}
	if fmt.Sprint(tpc.Sampler.RatioBasedSampler) != "0" {
		return sdktrace.TraceIDRatioBased(tpc.Sampler.RatioBasedSampler)
	}
	if tpc.Sampler.DefaultOnSampler {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
	if tpc.Sampler.DefaultOffSampler {
		return sdktrace.ParentBased(sdktrace.NeverSample())
	}
	return sdktrace.ParentBased(sdktrace.NeverSample())
}

func initResource(ctx context.Context, tpc *TracerProviderConfig) (*otelresource.Resource, error) {
	tpc.ResourceOptions = append(tpc.ResourceOptions,
		otelresource.WithSchemaURL(semconv.SchemaURL),
		otelresource.WithAttributes(resource.ServiceNameKey.String(tpc.ServiceName)))
	r, err := otelresource.New(ctx, tpc.ResourceOptions...)
	if err != nil {
		blog.Errorf("%s: %v", "failed to create resource", err)
		return &otelresource.Resource{}, err
	}
	if tpc.ResourceAttrs != nil {
		for _, a := range tpc.ResourceAttrs {
			r, _ = otelresource.Merge(r, otelresource.NewSchemaless(a))
		}
	}
	tpc.ResourceAttrs = append([]attribute.KeyValue{resource.ServiceNameKey.String(tpc.ServiceName)}, tpc.ResourceAttrs...)
	return r, nil
}

func initCollectorEndpointOptions(config *TracerProviderConfig) []oteljaeger.CollectorEndpointOption {
	var op []oteljaeger.CollectorEndpointOption
	if config.JaegerConfig.CollectorEndpoint.Endpoint != "" {
		op = append(op, oteljaeger.WithEndpoint(config.JaegerConfig.CollectorEndpoint.Endpoint))
	}
	if config.JaegerConfig.CollectorEndpoint.Username != "" {
		op = append(op, oteljaeger.WithUsername(config.JaegerConfig.CollectorEndpoint.Username))
	}
	if config.JaegerConfig.CollectorEndpoint.Password != "" {
		op = append(op, oteljaeger.WithPassword(config.JaegerConfig.CollectorEndpoint.Password))
	}
	return op
}

func initAgentEndpointOptions(config *TracerProviderConfig) []oteljaeger.AgentEndpointOption {
	var op []oteljaeger.AgentEndpointOption
	if config.JaegerConfig.AgentEndpoint.Host != "" {
		op = append(op, oteljaeger.WithAgentHost(config.JaegerConfig.AgentEndpoint.Host))
	}
	if config.JaegerConfig.AgentEndpoint.Port != "" {
		op = append(op, oteljaeger.WithAgentPort(config.JaegerConfig.AgentEndpoint.Port))
	}
	return op
}

func initGRPCConfigOptions(config *TracerProviderConfig) []otlptracegrpc.Option {
	var op []otlptracegrpc.Option
	if config.OTLPConfig.GRPCConfig.GRPCEndpoint != "" {
		op = append(op, otlptracegrpc.WithEndpoint(config.OTLPConfig.GRPCConfig.GRPCEndpoint))
	}
	if config.OTLPConfig.GRPCConfig.GRPCInsecure {
		op = append(op, otlptracegrpc.WithInsecure())
	}
	return op
}

func initHTTPConfigOptions(config *TracerProviderConfig) []otlptracehttp.Option {
	var op []otlptracehttp.Option
	if config.OTLPConfig.HTTPConfig.HTTPEndpoint != "" {
		op = append(op, otlptracehttp.WithEndpoint(config.OTLPConfig.HTTPConfig.HTTPEndpoint))
	}
	if config.OTLPConfig.HTTPConfig.HTTPURLPath != "" {
		op = append(op, otlptracehttp.WithURLPath(config.OTLPConfig.HTTPConfig.HTTPURLPath))
	}
	if config.OTLPConfig.HTTPConfig.HTTPInsecure {
		op = append(op, otlptracehttp.WithInsecure())
	}
	return op
}

// initProcessors sets processors for OTEL.
func initProcessors(exporter sdktrace.SpanExporter) (sps []sdktrace.SpanProcessor) {
	// Processors must be enabled for every data source. Always be sure to batch in production.
	sp := utils.NewBatchSpanProcessor(exporter)
	sps = append(sps, sp)
	return sps
}

func validateTracingOptions(opt *TracerProviderConfig) error {
	err := validateTracingSwitch(opt.TracingSwitch)
	if err != nil {
		return err
	}

	err = validateTracingType(opt.TracingType)
	if err != nil {
		return err
	}

	err = validateServiceName(opt.ServiceName)
	if err != nil {
		return err
	}
	return nil
}

func validateTracingSwitch(s string) error {
	if s == "on" || s == "off" {
		return nil
	}
	return errSwitchType
}

func validateTracingType(t string) error {
	if t == string(Jaeger) || t == string(Zipkin) || t == string(OTLP_GRPC) || t == string(OTLP_HTTP) {
		return nil
	}
	return errTracingType
}

func validateServiceName(sn string) error {
	if sn == "" {
		return errServiceName
	}
	return nil
}
