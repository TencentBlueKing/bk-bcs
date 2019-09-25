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

package lib

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_storage",
		Subsystem: "api",
		Name:      "request_total",
		Help:      "The total number of requests to bcs-storage api",
	}, []string{"handler", "method", "status"})
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_storage",
		Subsystem: "api",
		Name:      "request_latency_seconds",
		Help:      "BCS storage api request latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "method", "status"})
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestLatency)
}

//reportAPIMetrics report all api action metrics
func reportAPIMetrics(handler, method, status string, started time.Time) {
	pathList := strings.Split(handler, "/")
	shortPath := handler
	//a large amount of URL due to due to multiple cluster , namespace , resource
	//reduce URL numbers for metrics collection
	if len(pathList) > 4 {
		shortPath = strings.Join(pathList[0:5], "/")
	}
	requestsTotal.WithLabelValues(shortPath, method, status).Inc()
	requestLatency.WithLabelValues(shortPath, method, status).Observe(time.Since(started).Seconds())
}
