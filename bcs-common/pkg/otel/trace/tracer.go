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

package trace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/zipkin"
)

var (
	// errSwitchType switch type error
	errSwitchType = errors.New("error switch type, please input: [on or off]")
	// errTracingType tracing type error
	errTracingType = errors.New("error tracing type, please input: [zipkin or jaeger]")
	// errServiceName for service name is null
	errServiceName = errors.New("error service name is null")
)

const (
	// defaultSwitchType for default switch type
	defaultSwitchType = "off"
	// defaultTracingType for default tracing type
	defaultTracingType = "jaeger"
)

// Options set options for different tracing systems
type Options struct {
	// factory parameter
	TracingSwitch string `json:"tracingSwitch" value:"off" usage:"tracing switch"`
	TracingType   string `json:"tracingType" value:"jaeger" usage:"tracing type(default jaeger)"`

	ServiceName string `json:"serviceName" value:"bcs-common/pkg/otel" usage:"tracing serviceName"`

	ExporterURL string `json:"exporterURL" value:"" usage:"url of exporter"`

	OTLPEndpoint string `json:"otlpEndpoint" value:"" usage:"OpenTelemetry Collector service endpoint"`

	ResourceAttrs []attribute.KeyValue `json:"resourceAttrs" value:"" usage:"attributes of traced service"`
}

// Option for init Options
type Option func(f *Options)

// InitTracerProvider initialize an OTLP tracer provider
func InitTracerProvider(serviceName string, opt ...Option) (*sdktrace.TracerProvider, error) {
	defaultOptions := &Options{
		TracingSwitch: defaultSwitchType,
		TracingType:   defaultTracingType,
		ServiceName:   serviceName,
	}

	for _, o := range opt {
		o(defaultOptions)
	}

	err := validateTracingOptions(defaultOptions)
	if err != nil {
		klog.Errorf("validateTracingOptions failed: %v", err)
		return nil, err
	}

	if defaultOptions.TracingSwitch == "off" {
		return &sdktrace.TracerProvider{}, nil
	}

	if defaultOptions.TracingType == "jaeger" {
		return jaeger.NewTracerProvider(defaultOptions.ExporterURL, defaultOptions.ServiceName)
	}

	if defaultOptions.TracingType == "zipkin" {
		return zipkin.NewTracerProvider(defaultOptions.ExporterURL, defaultOptions.ServiceName)
	}
	return &sdktrace.TracerProvider{}, nil
}

func validateTracingOptions(opt *Options) error {
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
	if t == string(Jaeger) || t == string(Zipkin) {
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

// InitTracingProvider Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitTracingProvider(serviceName string, opt ...Option) (func(context.Context) error, error) {
	ctx := context.Background()

	defaultOptions := &Options{
		ServiceName: serviceName,
	}
	for _, o := range opt {
		o(defaultOptions)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(defaultOptions.ServiceName),
		),
		resource.WithAttributes(defaultOptions.ResourceAttrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %s", err.Error())
	}

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, defaultOptions.OTLPEndpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %s", err.Error())
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %s", err.Error())
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// InitTracing init tracing
func InitTracing(op *conf.TracingConfig, serviceName string) (func(context.Context) error, error) {
	if !op.Enabled {
		return nil, nil
	}
	opts := []Option{}

	if op.Endpoint != "" {
		opts = append(opts, OTLPEndpoint(op.Endpoint))
	}
	attrs := make([]attribute.KeyValue, 0)

	if op.Token != "" {
		attrs = append(attrs, attribute.String("bk.data.token", op.Token))
	}

	if op.ResourceAttrs != nil {
		attrs = append(attrs, newResource(op.ResourceAttrs)...)
	}

	opts = append(opts, ResourceAttrs(attrs))

	tracer, err := InitTracingProvider(serviceName, opts...)
	if err != nil {
		return nil, err
	}

	return tracer, nil
}

func newResource(attrs map[string]string) []attribute.KeyValue {
	attrValues := make([]attribute.KeyValue, 0)
	for k, v := range attrs {
		attrValues = append(attrValues, attribute.String(k, v))
	}
	return attrValues
}
