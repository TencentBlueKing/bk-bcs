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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/metric"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/utils"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

const (
	serviceName = "demo-http-client"
	tracerName  = "demo-http-tracer"
)

func main() {

	metricOpts := &metric.Options{
		MetricSwitchStatus:  "on",
		MetricType:          "prometheus",
		ProcessorWithMemory: true,
	}

	var metricOp []metric.Option
	if metricOpts.MetricSwitchStatus != "" {
		metricOp = append(metricOp, metric.SwitchStatus(metricOpts.MetricSwitchStatus))
	}
	if metricOpts.MetricType != "" {
		metricOp = append(metricOp, metric.TypeMetric(metricOpts.MetricType))
	}
	if !metricOpts.ProcessorWithMemory {
		metricOp = append(metricOp, metric.ProcessorWithMemory(metricOpts.ProcessorWithMemory))
	}

	mp, exp, err := metric.InitMeterProvider(metricOp...)
	if err != nil {
		log.Fatal(err)
	}
	meter := mp.Meter("demo-client-meter")

	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	commonLabels := []attribute.KeyValue{
		attribute.String("endpoint", "http_client"),
		attribute.String("bar", "foo"),
	}

	requestLatency := otelmetric.Must(meter).NewFloat64Histogram(
		"demo_client/request_latency",
		otelmetric.WithDescription("The latency of requests processed"),
	)

	requestCount := otelmetric.Must(meter).
		NewInt64Counter(
			"demo_client/request_counts",
			otelmetric.WithDescription("The number of requests processed"),
		)

	// start a http service for exposing metrics
	http.Handle("/", exp)
	go func() {
		log.Fatal(http.ListenAndServe(":30080", nil))
	}()

	traceOpts := trace.Options{
		TracingSwitch: "on",
		ServiceName:   serviceName,
		ExporterURL:   "http://localhost:14268/api/traces",
		ResourceAttrs: []attribute.KeyValue{
			attribute.String("endpoint", "http_client"),
		},
	}
	var traceOp []trace.Option
	traceOp = append(traceOp, trace.TracerSwitch(traceOpts.TracingSwitch))
	traceOp = append(traceOp, trace.ResourceAttrs(traceOpts.ResourceAttrs))
	traceOp = append(traceOp, trace.ExporterURL(traceOpts.ExporterURL))

	tp, err := trace.InitTracerProvider(traceOpts.ServiceName, traceOp...)
	tracer := tp.Tracer(tracerName)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	for {
		startTime := time.Now()
		ctx, span := tracer.Start(ctx, "Execute Request")
		log.Printf("traceID:%v, spanID:%v",
			span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
		commonLabels2 := append(commonLabels,
			attribute.String("traceID", span.SpanContext().TraceID().String()))
		makeRequest(ctx)
		span.End()
		latencyMs := float64(time.Since(startTime)) / 1e6

		requestCount.Add(ctx, 1, commonLabels...)
		requestLatency.Record(ctx, latencyMs, commonLabels2...)

		fmt.Printf("Latency: %.3fms\n", latencyMs)
		time.Sleep(time.Duration(300) * time.Second)
	}
}

func makeRequest(ctx context.Context) {
	// client.Transport.RoundTrip creates a Span and propagates its context via the provided request's headers
	// before handing the request to the configured base RoundTripper. The created span will
	// end when the response body is closed or when a read from the body returns io.EOF.
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9090/", nil)
	ctx, span := utils.Tracer(tracerName).Start(ctx, "HTTP DO")
	log.Printf("traceID:%v, spanID:%v",
		span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String())
	resp, err := client.Do(req)
	span.End()

	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return
	}

	// Get the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read from resp.Body failed, err:%v\n", err)
		return
	}
	fmt.Print(string(body))
}
