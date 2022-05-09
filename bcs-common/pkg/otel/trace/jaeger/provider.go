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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/resource"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(exporterURL, serviceName string) (*sdktrace.TracerProvider, error) {
	// Create the Jaeger exporter
	traceExp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(exporterURL)))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(traceExp),
		// Record information about this application in an Resource.
		sdktrace.WithResource(resource.New(serviceName)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
