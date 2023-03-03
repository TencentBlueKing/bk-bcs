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

package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TracingConf web 相关配置
type TracingConf struct {
	TracingSwitch string               `yaml:"tracing_switch" usage:"tracing switch"`
	TracingType   string               `yaml:"tracing_type" usage:"tracing type(default jaeger)"`
	ServiceName   string               `yaml:"service_name" usage:"tracing serviceName"`
	ExporterURL   string               `yaml:"exporter_url" usage:"url of exporter"`
	ResourceAttrs []attribute.KeyValue `yaml:"resourceAttrs" usage:"attributes of traced service"`
}

// init 初始化
func (c *TracingConf) init() error {
	return nil
}

func InitTracingInstance(c *TracingConf) (*sdktrace.TracerProvider, error) {
	opts := []trace.Option{}
	if c.TracingSwitch != "" {
		opts = append(opts, trace.TracerSwitch(c.TracingSwitch))
	}
	if c.TracingType != "" {
		opts = append(opts, trace.TracerType(c.TracingType))
	}

	if c.ExporterURL != "" {
		opts = append(opts, trace.ExporterURL(c.ExporterURL))
	}
	tp, err := trace.InitTracerProvider(c.ServiceName, opts...)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

// defaultTracingConf 默认配置
func defaultTracingConf() *TracingConf {
	values := make([]attribute.KeyValue, 0)
	c := &TracingConf{
		TracingSwitch: "on",
		TracingType:   "jaeger",
		ServiceName:   "bcs-monitor",
		ExporterURL:   "http://localhost:14268/api/traces",
		ResourceAttrs: values,
	}
	return c
}
