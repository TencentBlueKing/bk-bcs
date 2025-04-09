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
 */

// Package metrics xxx
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

const (
	// ErrStatus for success status
	ErrStatus = "failure"
	// SucStatus for failure status
	SucStatus = "success"

	// Query for mysql query
	Query = "query"
	// Create for mysql create
	Create = "create"
	// Delete for mysql delete
	Delete = "delete"
	// Update for mysql update
	Update = "update"
)

const (
	// BkBcsUserManager for module bcs-user-manager metrics prefix
	BkBcsUserManager = "bkbcs_usermanager"
)

// TimeBuckets is based on Prometheus client_golang prometheus.DefBuckets
var timeBuckets = prometheus.ExponentialBuckets(0.00025, 2, 32) // from 0.25ms to 16 seconds

// Metrics the bcs-user-manager exports.
var (
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsUserManager,
		Name:      "api_request_total_num",
		Help:      "Counter of requests to bcs-user-manager.",
	}, []string{"handler", "method", "status"})

	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsUserManager,
		Name:      "api_request_latency_time",
		Buckets:   timeBuckets,
		Help:      "Histogram of the time (in seconds) each request took.",
	}, []string{"handler", "method", "status"})

	mysqlSlowQueryCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsUserManager,
		Name:      "mysq_slow_query_total_num",
		Help:      "Counter of mysql slow query to bcs-user-manager.",
	}, []string{"handler", "method", "status"})

	mysqlSlowQueryLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsUserManager,
		Name:      "mysq_slow_query_latency_time",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60},
		Help:      "Histogram of the time (in seconds) each mysql slow query.",
	}, []string{"handler", "method", "status"})
)

// RunMetric metric entrypoint
func RunMetric(conf *config.UserMgrConfig) {
	blog.Infof("run metric: port(%d)", conf.MetricPort)

	// prometheus register collector
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(mysqlSlowQueryCount)
	prometheus.MustRegister(mysqlSlowQueryLatency)

	// prometheus metrics server
	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", promhttp.Handler())

	// server address
	addresses := []string{conf.Address}
	if len(conf.IPv6Address) > 0 {
		addresses = append(addresses, conf.IPv6Address)
	}
	metricServer := ipv6server.NewIPv6Server(addresses, strconv.Itoa(int(conf.MetricPort)), "", metricMux)
	// nolint
	go metricServer.ListenAndServe()

	blog.Infof("run metric ok")
}

// ReportRequestAPIMetrics report API request metrics
func ReportRequestAPIMetrics(handler, method, status string, started time.Time) {
	requestCount.WithLabelValues(handler, method, status).Inc()
	requestLatency.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

// ReportMysqlSlowQueryMetrics report mysql slow query metrics
func ReportMysqlSlowQueryMetrics(handler, method, status string, started time.Time) {
	// 记录大于200ms的慢查询
	latency := time.Since(started).Milliseconds()
	if latency > int64(config.GetGlobalConfig().MysqlSlowRecord) {
		klog.Infof("slow query handler: %s, method: %s, latency: %dms", handler, method, latency)
	}
	mysqlSlowQueryCount.WithLabelValues(handler, method, status).Inc()
	mysqlSlowQueryLatency.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}
