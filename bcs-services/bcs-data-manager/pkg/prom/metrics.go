/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package prom

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// StatusErr error shows call error
	StatusErr = "failure"
	// StatusOK call successfully
	StatusOK = "success"
)

const (
	// BkBcsDataManager for prometheus namespace
	BkBcsDataManager = "bcs_data_manager"
	// BkBcsMonitor bcs monitor
	BkBcsMonitor = "bcs_monitor"
	// BkBcsStorage bcs storage
	BkBcsStorage = "bcs_storage"
	// BkBcsClusterManager bcs cluster manager
	BkBcsClusterManager = "bcs_cluster_manager"
)

var InstanceIP string

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsDataManager,
		Name:      "lib_request_total_num",
		Help:      "The total number of requests for data manager to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsDataManager,
		Name:      "lib_request_latency_time",
		Help:      "api request latency statistic for data manager to call other system",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsDataManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for data manager api",
	}, []string{"handler", "method", "status", "instance"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsDataManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for data manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "method", "status", "instance"})

	consumeJobTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsDataManager,
		Name:      "consume_job_total_num",
		Help:      "The total number of consume job",
	}, []string{"jobType", "dimension", "status", "instance"})
	consumeJobLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsDataManager,
		Name:      "consume_job_latency_time",
		Help:      "time for consume a job",
		Buckets:   []float64{0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0, 15.0},
	}, []string{"jobType", "dimension", "status", "instance"})
	jobLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsDataManager,
		Name:      "job_latency_time",
		Help:      "time from the time it record to the time finished ",
		Buckets:   []float64{0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0, 15.0, 20.0, 30.0, 60.0, 120.0, 300.0},
	}, []string{"jobType", "dimension", "status", "instance"})

	produceJobTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsDataManager,
		Name:      "produce_job_total_num",
		Help:      "The total number of produce job",
	}, []string{"jobType", "dimension", "status", "instance"})
	produceJobLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsDataManager,
		Name:      "produce_job_latency_time",
		Help:      "time for produce job",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0, 15.0, 30.0, 60.0, 120.0, 300.0},
	}, []string{"jobType", "dimension", "status", "instance"})

	consumerConcurrency = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: BkBcsDataManager,
		Name:      "consumer_concurrency",
		Help:      "concurrency for consumer",
	}, []string{"instance"})
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(requestsTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	prometheus.MustRegister(consumeJobTotal)
	prometheus.MustRegister(consumeJobLatency)
	prometheus.MustRegister(jobLatency)
	prometheus.MustRegister(produceJobTotal)
	prometheus.MustRegister(produceJobLatency)
	prometheus.MustRegister(consumerConcurrency)
	InstanceIP = os.Getenv("localIp")
}

// ReportLibRequestMetric report lib call metrics
func ReportLibRequestMetric(system, handler, method string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	requestsTotalLib.WithLabelValues(system, handler, method, status).Inc()
	requestLatencyLib.WithLabelValues(system, handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, method, status string, started time.Time) {
	requestsTotalAPI.WithLabelValues(handler, method, status, InstanceIP).Inc()
	requestLatencyAPI.WithLabelValues(handler, method, status, InstanceIP).Observe(time.Since(started).Seconds())
}

// ReportConsumeJobMetric report consume job metrics
func ReportConsumeJobMetric(jobType, dimension string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	consumeJobTotal.WithLabelValues(jobType, dimension, status, InstanceIP).Inc()
	consumeJobLatency.WithLabelValues(jobType, dimension, status, InstanceIP).Observe(time.Since(started).Seconds())
}

// ReportJobMetric report job metrics
func ReportJobMetric(jobType, dimension string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	jobLatency.WithLabelValues(jobType, dimension, status, InstanceIP).Observe(time.Since(started).Seconds())
}

// ReportProduceJobLatencyMetric report produce job latency
func ReportProduceJobLatencyMetric(jobType, dimension string, err error, started time.Time) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	produceJobLatency.WithLabelValues(jobType, dimension, status, InstanceIP).Observe(time.Since(started).Seconds())
}

// ReportProduceJobTotalMetric report produce job total
func ReportProduceJobTotalMetric(jobType, dimension string, err error) {
	status := StatusOK
	if err != nil {
		status = StatusErr
	}
	produceJobTotal.WithLabelValues(jobType, dimension, status, InstanceIP).Inc()
}

// ReportConsumeConcurrency report consume concurrency
func ReportConsumeConcurrency(concurrency int) {
	consumerConcurrency.WithLabelValues(InstanceIP).Set(float64(concurrency))
}
