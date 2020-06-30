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
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TimeBuckets is based on Prometheus client_golang prometheus.DefBuckets
var timeBuckets = prometheus.ExponentialBuckets(0.00025, 2, 16) // from 0.25ms to 8 seconds

// Metrics the bcs-api exports.
var (
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bcs_api",
		Name:      "request_count_total",
		Help:      "Counter of requests to bcs-api.",
	}, []string{"type", "method"})

	RequestErrorCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bcs_api",
		Name:      "request_err_count_total",
		Help:      "Counter of error requests to bcs-api.",
	}, []string{"type", "method"})

	RequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bcs_api",
		Name:      "request_latency_seconds",
		Buckets:   timeBuckets,
		Help:      "Histogram of the time (in seconds) each request took.",
	}, []string{"type", "method"})

	RequestErrorLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bcs_api",
		Name:      "request_error_latency_seconds",
		Buckets:   timeBuckets,
		Help:      "Histogram of the time (in seconds) each error request took.",
	}, []string{"type", "method"})
)

func RunMetric(conf *config.ApiServConfig, err error) {

	blog.Infof("run metric: port(%d)", conf.MetricPort)
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(RequestErrorCount)
	prometheus.MustRegister(RequestLatency)
	prometheus.MustRegister(RequestErrorLatency)
	http.Handle("/metrics", promhttp.Handler())
	addr := conf.Address + ":" + strconv.Itoa(int(conf.MetricPort))
	go http.ListenAndServe(addr, nil)

	blog.Infof("run metric ok")
}
