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

package workqueue

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
)

// Metrics subsystem and keys used by the workqueue.
const (
	BkBcsWorkQueueSystem       = "bkbcs_workqueue"
	DepthKey                   = "depth"
	AddsKey                    = "adds_total"
	QueueLatencyKey            = "queue_duration_seconds"
	WorkDurationKey            = "work_duration_seconds"
	UnfinishedWorkKey          = "unfinished_work_seconds"
	LongestRunningProcessorKey = "longest_running_processor_seconds"
	RetriesKey                 = "retries_total"
)

var (
	depth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      DepthKey,
		Help:      "Current depth of workqueue",
	}, []string{"name"})

	adds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      AddsKey,
		Help:      "Total number of adds handled by workqueue",
	}, []string{"name"})

	latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      QueueLatencyKey,
		Help:      "How long in seconds an item stays in workqueue before being requested.",
		Buckets:   prometheus.ExponentialBuckets(10e-9, 10, 10),
	}, []string{"name"})

	workDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      WorkDurationKey,
		Help:      "How long in seconds processing an item from workqueue takes.",
		Buckets:   prometheus.ExponentialBuckets(10e-9, 10, 10),
	}, []string{"name"})

	unfinished = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      UnfinishedWorkKey,
		Help: "How many seconds of work has done that " +
			"is in progress and hasn't been observed by work_duration. Large " +
			"values indicate stuck threads. One can deduce the number of stuck " +
			"threads by observing the rate at which this increases.",
	}, []string{"name"})

	longestRunningProcessor = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      LongestRunningProcessorKey,
		Help: "How many seconds has the longest running " +
			"processor for workqueue been running.",
	}, []string{"name"})

	retries = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsWorkQueueSystem,
		Name:      RetriesKey,
		Help:      "Total number of retries handled by workqueue",
	}, []string{"name"})

	metrics = []prometheus.Collector{
		depth, adds, latency, workDuration, unfinished, longestRunningProcessor, retries,
	}
)

func init() {
	for _, m := range metrics {
		prometheus.MustRegister(m)
	}
	workqueue.SetProvider(prometheusMetricsProvider{})
}

type prometheusMetricsProvider struct {
}

func (prometheusMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	return depth.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	return adds.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	return latency.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	return workDuration.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return unfinished.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return longestRunningProcessor.WithLabelValues(name)
}

func (prometheusMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	return retries.WithLabelValues(name)
}
