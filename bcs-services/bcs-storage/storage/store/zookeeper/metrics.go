/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zookeeper

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	operatorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs_storage",
		Subsystem: "driver",
		Name:      "zookeeper_total",
		Help:      "The total number of operation to zookeeper",
	}, []string{"method", "status"})
	operatorLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_storage",
		Subsystem: "driver",
		Name:      "zookeeper_latency_seconds",
		Help:      "BCS storage zookeeper operation latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"method", "status"})
)

func init() {
	prometheus.MustRegister(operatorTotal)
	prometheus.MustRegister(operatorLatency)
}

//reportAPIMetrics report all zookeeper operation metrics
func reportZKMetrics(method, status string, started time.Time) {
	operatorTotal.WithLabelValues(method, status).Inc()
	go operatorLatency.WithLabelValues(method, status).Observe(time.Since(started).Seconds())
}
