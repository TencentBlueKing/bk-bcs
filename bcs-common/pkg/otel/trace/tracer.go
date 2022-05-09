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
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/zipkin"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	// errSwitchType switch type error
	errSwitchType error = errors.New("error switch type, please input: [on or off]")
	// errTracingType tracing type error
	errTracingType error = errors.New("error tracing type, please input: [zipkin or jaeger]")
	// errServiceName for service name is null
	errServiceName error = errors.New("error service name is null")
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
		blog.Errorf("validateTracingOptions failed: %v", err)
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
