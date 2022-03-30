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
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func welcomePage(w http.ResponseWriter, r *http.Request) {
	_, span := utils.Tracer("server").Start(r.Context(), "WelcomePage")
	defer span.End()
	w.Write([]byte("Welcome to my website!"))
}

func main() {
	opts := trace.TracerProviderConfig{
		TracingSwitch: "on",
		TracingType:   "jaeger",
		JaegerConfig: &jaeger.EndpointConfig{
			CollectorEndpoint: &jaeger.CollectorEndpoint{
				Endpoint: "http://localhost:14268/api/traces",
			},
		},
		ResourceAttrs: []attribute.KeyValue{
			attribute.String("EndPoint", "HttpServer"),
		},
		Sampler: &trace.SamplerType{
			DefaultOnSampler: true,
		},
	}
	op := trace.ValidateTracerProviderOption(&opts)

	ctx, tp, err := trace.InitTracerProvider("http-server", op...)
	otel.SetTextMapPropagator(propagation.TraceContext{})
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

	wrappedHandler := otelhttp.NewHandler(http.HandlerFunc(welcomePage), "/")
	http.Handle("/", wrappedHandler)
	log.Fatal(http.ListenAndServe("localhost:9090", nil))
}
