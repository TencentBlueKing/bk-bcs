/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	BkBcsHelmManager = "bkbcs_helmmanager"
)

var (
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsHelmManager,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for helm manager api",
	}, []string{"handler", "status"})

	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsHelmManager,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for helm manager api",
		Buckets:   []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"handler", "status"})
)

func init() {
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
}

// ReportAPIRequestMetric report api request metrics
func ReportAPIRequestMetric(handler, status string, started time.Time) {
	requestTotalAPI.WithLabelValues(handler, status).Inc()
	requestLatencyAPI.WithLabelValues(handler, status).Observe(time.Since(started).Seconds())
}
