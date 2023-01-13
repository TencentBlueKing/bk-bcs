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

package jaeger

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	TracingType = "jaeger"
	ServiceName = "bcs-cluster-resources"
)

func InitTracingInstance(op *config.TracingConf) (*sdktrace.TracerProvider, error) {

	opts := []trace.Option{}
	if op.TracingEnabled != false {
		opts = append(opts, trace.TracerSwitch("on"))
	}

	if TracingType != "" {
		opts = append(opts, trace.TracerType(TracingType))
	}

	if op.ExporterURL != "" {
		opts = append(opts, trace.ExporterURL(op.ExporterURL))
	}
	tracer, err := trace.InitTracerProvider(ServiceName, opts...)
	if err != nil {
		return nil, err
	}

	return tracer, nil
}
