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

package main

import (
	"context"
	"log"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

const (
	service     = "otel-trace-demo"
	environment = "development"
	id          = 0
)

func main() {
	opts := trace.TracerProviderConfig{
		TracingSwitch: "on",
		TracingType:   "jaeger",
		ServiceName:   service,
		JaegerConfig: trace.JaegerConfig{
			CollectorEndpoint: jaeger.CollectorEndpoint{
				Endpoint: "http://localhost:14268/api/traces",
			},
		},
		ResourceAttrs: []attribute.KeyValue{
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		},
		Sampler: trace.SamplerType{
			DefaultOnSampler: true,
		},
	}
	op := trace.ValidateTracerProviderOption(&opts)
	op = append(op, trace.WithResourceOption(resource.WithFromEnv()))

	ctx, tp, err := trace.InitTracerProvider(opts.ServiceName, op...)
	if err != nil {
		log.Fatal(err)
	}

	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx2)

	tr := tp.Tracer("component-main")

	ctx3, span := tr.Start(context.Background(), "foo")
	span.SetAttributes(attribute.String("testkey", "testvalue"))
	defer span.End()
	bar(ctx3)
}

func bar(ctx context.Context) {
	// Use the global TracerProvider.
	tr := utils.Tracer("component-bar")
	_, span := tr.Start(ctx, "bar")
	defer span.End()
}
