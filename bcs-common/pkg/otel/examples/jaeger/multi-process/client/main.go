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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/exporter/jaeger"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	opts := &trace.TracerProviderConfig{
		TracingSwitch: "on",
		TracingType:   "jaeger",
		JaegerConfig: trace.JaegerConfig{
			AgentEndpoint: jaeger.AgentEndpoint{
				Host: "localhost",
				Port: "6831",
			},
		},
		ResourceAttrs: []attribute.KeyValue{
			attribute.String("EndPoint", "HttpClient"),
		},
		Sampler: trace.SamplerType{
			DefaultOnSampler: true,
		},
	}
	op := trace.ValidateTracerProviderOption(opts)
	op = append(op, trace.WithAlwaysOnSampler())

	ctx, tp, err := trace.InitTracerProvider("http-client", op...)
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

	ctx3, span := tp.Tracer("client").Start(context.Background(), "HTTP Request")
	makeRequest(ctx3)
	defer span.End()

}

func makeRequest(ctx context.Context) {
	// 创建GET请求
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9090/", nil)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return
	}

	// 获取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read from resp.Body failed, err:%v\n", err)
		return
	}
	// 打印响应体
	fmt.Print(string(body))
}
