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
	"log"
	"net/http"
	"sync"
	"time"

	tmetric "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/metric"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	lemonsKey = attribute.Key("ex.com/lemons")
)

func main() {
	opts := &tmetric.Options{
		MetricSwitchStatus:  "on",
		MetricType:          "prometheus",
		ProcessorWithMemory: true,
	}

	op := []tmetric.Option{}
	if opts.MetricSwitchStatus != "" {
		op = append(op, tmetric.SwitchStatus(opts.MetricSwitchStatus))
	}
	if opts.MetricType != "" {
		op = append(op, tmetric.TypeMetric(opts.MetricType))
	}
	if !opts.ProcessorWithMemory {
		op = append(op, tmetric.ProcessorWithMemory(opts.ProcessorWithMemory))
	}

	mp, exp, err := tmetric.InitMeterProvider(op...)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", exp)
	go func() {
		log.Fatal(http.ListenAndServe(":30080", nil))
	}()

	meter := mp.Meter("ex.com/basic")
	observerLock := new(sync.RWMutex)
	observerValueToReport := new(float64)
	observerLabelsToReport := new([]attribute.KeyValue)
	cb := func(_ context.Context, result metric.Float64ObserverResult) {
		(*observerLock).RLock()
		value := *observerValueToReport
		labels := *observerLabelsToReport
		(*observerLock).RUnlock()
		result.Observe(value, labels...)
	}
	_ = metric.Must(meter).NewFloat64GaugeObserver("ex.com.one", cb,
		metric.WithDescription("A GaugeObserver set to 1.0"),
	)

	histogram := metric.Must(meter).NewFloat64Histogram("ex.com.two",
		metric.WithDescription("A Histogram set to 1.0"))
	counter := metric.Must(meter).NewFloat64Counter("ex.com.three",
		metric.WithDescription("A Counter set to 1.0"))

	commonLabels := []attribute.KeyValue{
		lemonsKey.Int(10),
		attribute.String("A", "1"),
		attribute.String("B", "2"),
		attribute.String("C", "3")}
	notSoCommonLabels := []attribute.KeyValue{lemonsKey.Int(13)}

	ctx := context.Background()

	(*observerLock).Lock()
	*observerValueToReport = 1.0
	*observerLabelsToReport = commonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		commonLabels,
		histogram.Measurement(2.0),
		counter.Measurement(12.0),
	)

	time.Sleep(2 * time.Second)

	(*observerLock).Lock()
	*observerValueToReport = 1.0
	*observerLabelsToReport = notSoCommonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		notSoCommonLabels,
		histogram.Measurement(2.0),
		counter.Measurement(22.0),
	)

	time.Sleep(2 * time.Second)

	(*observerLock).Lock()
	*observerValueToReport = 13.0
	*observerLabelsToReport = commonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		commonLabels,
		histogram.Measurement(12.0),
		counter.Measurement(13.0),
	)

	fmt.Println("Example finished updating, please visit :30080/metrics")

	select {}
}
