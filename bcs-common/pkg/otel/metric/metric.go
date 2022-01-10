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

package metric

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/metric/prometheus"

	prom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
)

const (
	defaultSwitchStatus = "off"
	defaultMetricType   = "prometheus"
)

var (
	// errSwitchStatus switch type error
	errSwitchStatus error = errors.New("error switch status, please input: [on or off]")

	// errMetricType metric type error
	errMetricType error = errors.New("error metric type, please input: [prometheus]")

	// errProcessorWithMemory processor with memory error
	errProcessorWithMemory = errors.New("error value of ProcessorWithMemory, please input: [true or false]")
)

// Options set options for different tracing systems
type Options struct {
	// factory parameter
	MetricSwitchStatus string `json:"meterSwitchStatus" value:"off" usage:"meter switch status"`
	MetricType         string `json:"meterType" value:"prometheus" usage:"meter type(default prometheus)"`

	// ProcessorWithMemory controls whether the processor remembers metric
	// instruments and label sets that were previously reported.
	// When Memory is true, Reader.ForEach() will visit
	// metrics that were not updated in the most recent interval.
	ProcessorWithMemory prometheus.MemoryOption `json:"processorWithMemory" value:"false" usage:"processor memory policy"`

	// ControllerOption is the interface type slice that applies the value to a configuration option.
	ControllerOption []controller.Option `json:"controllerOption" value:"" usage:"applies the value to a configuration option"`
}

// Option for init Options
type Option func(f *Options)

// InitMeterProvider initialize a meter
func InitMeterProvider(op ...Option) (metric.MeterProvider, *prom.Exporter, error) {
	defaultOptions := &Options{
		MetricSwitchStatus:  defaultSwitchStatus,
		MetricType:          defaultMetricType,
		ProcessorWithMemory: true,
	}

	for _, o := range op {
		o(defaultOptions)
	}

	err := validateMetricOptions(defaultOptions)
	if err != nil {
		blog.Errorf("validateMetricOptions failed: %v", err)
		return nil, nil, err
	}

	if defaultOptions.MetricSwitchStatus == defaultSwitchStatus {
		return &controller.Controller{}, nil, nil
	}
	if defaultOptions.MetricType == defaultMetricType {
		return prometheus.NewMeterProvider(defaultOptions.ProcessorWithMemory, defaultOptions.ControllerOption...)
	}
	return &controller.Controller{}, nil, nil
}

func validateMetricOptions(opts *Options) error {
	err := validateMetricSwitch(opts.MetricSwitchStatus)
	if err != nil {
		return err
	}

	err = validateMetricType(opts.MetricType)
	if err != nil {
		return err
	}

	err = validateProcessorWithMemory(opts.ProcessorWithMemory)
	if err != nil {
		return err
	}
	return nil
}

func validateMetricSwitch(ms string) error {
	if ms == "on" || ms == "off" {
		return nil
	}
	return errSwitchStatus
}

func validateMetricType(mt string) error {
	if mt == "prometheus" {
		return nil
	}
	return errMetricType
}

func validateProcessorWithMemory(p prometheus.MemoryOption) error {
	if p == false || p == true {
		return nil
	}
	return errProcessorWithMemory
}
