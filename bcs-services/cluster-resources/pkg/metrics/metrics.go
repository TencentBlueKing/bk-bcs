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
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const bkBcsClusterResources = "bkbcs_clusterresources"

var (
	// api 请求总量
	apiRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: bkBcsClusterResources,
		Name:      "api_request_total_num",
		Help:      "The total number of requests for template set api",
	}, []string{"handler", "code"})

	// api 请求耗时, 包含页面返回, API请求
	apiRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: bkBcsClusterResources,
		Name:      "api_request_duration_seconds",
		Help:      "Histogram of latencies for template set api",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60},
	}, []string{"handler", "code"})
)

func init() {
	prometheus.MustRegister(apiRequestsTotal)
	prometheus.MustRegister(apiRequestDuration)
}

// collectAPIRequestMetric api metrics 处理
func collectAPIRequestMetric(handler, code string, started time.Time) {
	apiRequestsTotal.WithLabelValues(handler, code).Inc()
	apiRequestDuration.WithLabelValues(handler, code).Observe(time.Since(started).Seconds())
}

// RecordTemplateMetrics 记录模板集metrics
func RecordTemplateMetrics(method string, code int32, started time.Time) {
	methods := strings.Split(method, ".")
	if len(methods) != 2 {
		return
	}
	// 只记录模板集
	if methods[0] != "TemplateSet" {
		return
	}
	collectAPIRequestMetric(methods[1], strconv.Itoa(int(code)), started)
}
