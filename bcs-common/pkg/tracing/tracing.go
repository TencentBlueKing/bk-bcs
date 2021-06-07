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
	"errors"
	"io"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/jaeger"

	"github.com/opentracing/opentracing-go"
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

// InitTracing for init tracing system interface
type InitTracing interface {
	Init() (io.Closer, error)
}

// Options set options for different tracing system
type Options struct {
	// factory parameter
	TracingSwitch string `json:"tracingSwitch" value:"off" usage:"tracing switch"`
	TracingType   string `json:"tracingType" value:"jaeger" usage:"tracing type(default jaeger)"`

	// jaeger
	ServiceName   string `json:"serviceName" value:"bcs-common/pkg/tracing" usage:"tracing serviceName"`
	RPCMetrics    bool   `json:"rPCMetrics" value:"false"`
	ReportMetrics bool   `json:"reportMetrics" value:"false"`

	// reporter
	ReportLog     bool   `json:"reportLog" value:"true"`
	AgentFromEnv  bool   `json:"agentFromEnv"`
	AgentHostPort string `json:"agentHostPort"`

	// sampler
	SampleType        string  `json:"sampleType" value:"const"`
	SampleParameter   float64 `json:"sampleParameter" value:"1"`
	SampleFromEnv     bool    `json:"sampleFromEnv"`
	SamplingServerURL string  `json:"samplingServerURL"`
}

// Option for init Options
type Option func(f *Options)

// NewInitTracing init tracing system
func NewInitTracing(serviceName string, opts ...Option) (InitTracing, error) {
	defaultOptions := &Options{
		TracingSwitch: defaultSwitchType,
		TracingType:   defaultTracingType,
		ServiceName:   serviceName,
	}

	for _, o := range opts {
		o(defaultOptions)
	}

	err := validateTracingOptions(defaultOptions)
	if err != nil {
		blog.Errorf("validateTracingOptions failed: %v", err)
		return nil, err
	}

	if defaultOptions.TracingSwitch == "off" {
		return &nullTracer{tracer: opentracing.NoopTracer{}}, nil
	}

	if defaultOptions.TracingType == string(Jaeger) {
		opts := []jaeger.JaeOption{}
		opts = append(opts, jaeger.ServiceName(defaultOptions.ServiceName))

		if defaultOptions.RPCMetrics {
			opts = append(opts, jaeger.RPCMetrics(defaultOptions.RPCMetrics))
		}

		if defaultOptions.ReportLog {
			opts = append(opts, jaeger.ReportLog(defaultOptions.ReportLog))
		}

		if defaultOptions.ReportMetrics {
			opts = append(opts, jaeger.ReportMetrics(defaultOptions.ReportMetrics))
		}

		if defaultOptions.SampleType != "" {
			opts = append(opts, jaeger.SamplerConfigInit(jaeger.SamplerConfig{
				SampleType:        defaultOptions.SampleType,
				SampleParameter:   defaultOptions.SampleParameter,
				FromEnv:           defaultOptions.SampleFromEnv,
				SamplingServerURL: defaultOptions.SamplingServerURL,
			}))
		}

		if defaultOptions.AgentFromEnv {
			opts = append(opts, jaeger.FromEnv(defaultOptions.AgentFromEnv))
		}

		if defaultOptions.AgentHostPort != "" {
			opts = append(opts, jaeger.AgentHostPort(defaultOptions.AgentHostPort))
		}

		return jaeger.NewJaegerServer(opts...)
	}

	// zipkin init
	if defaultOptions.TracingType == string(Zipkin) {
	}

	return &nullTracer{tracer: opentracing.NoopTracer{}}, nil
}

type nullTracer struct {
	tracer opentracing.NoopTracer
}

// Init init nullTracer
func (nt nullTracer) Init() (io.Closer, error) {
	opentracing.SetGlobalTracer(nt.tracer)
	return &nullCloser{}, nil
}

type nullCloser struct{}

// Close realize nullCloser
func (*nullCloser) Close() error { return nil }

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

func validateServiceName(sm string) error {
	if sm == "" {
		return errServiceName
	}

	return nil
}
