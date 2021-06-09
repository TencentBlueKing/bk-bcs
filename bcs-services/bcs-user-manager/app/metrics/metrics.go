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

package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// ErrStatus for success status
	ErrStatus = "failure"
	// SucStatus for failure status
	SucStatus = "success"
)

const (
	// BkBcsUserManager for module bcs-user-manager metrics prefix
	BkBcsUserManager = "bkbcs_usermanager"
)

// TimeBuckets is based on Prometheus client_golang prometheus.DefBuckets
var timeBuckets = prometheus.ExponentialBuckets(0.00025, 2, 16) // from 0.25ms to 8 seconds

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

)

// RunMetric metric entrypoint
func RunMetric(conf *config.UserMgrConfig) {
	blog.Infof("run metric: port(%d)", conf.MetricPort)

	// prometheus register collector
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestLatency)

	// prometheus metrics server
	http.Handle("/metrics", promhttp.Handler())
	addr := conf.Address + ":" + strconv.Itoa(int(conf.MetricPort))
	go http.ListenAndServe(addr, nil)

	blog.Infof("run metric ok")
}

//ReportRequestAPIMetrics report API request metrics
func ReportRequestAPIMetrics(handler, method, status string, started time.Time) {
	requestCount.WithLabelValues(handler, method, status).Inc()
	requestLatency.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}
