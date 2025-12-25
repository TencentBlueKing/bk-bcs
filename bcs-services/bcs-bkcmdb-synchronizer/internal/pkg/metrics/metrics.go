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

// Package metrics define metrics for cmdb synchronizer
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// BkBcsCmdbSynchronizer metrics namespace
	BkBcsCmdbSynchronizer = "bkbcs_cmdbsynchronizer"
)

var (
	// CMDBRequestsTotal CMDB请求总数计数器
	CMDBRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: BkBcsCmdbSynchronizer,
			Name:      "cmdb_requests_total",
			Help:      "Total number of CMDB requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// CMDBRequestDuration CMDB请求延迟直方图
	CMDBRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: BkBcsCmdbSynchronizer,
			Name:      "cmdb_request_duration_seconds",
			Help:      "CMDB request duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// CMDBRequestsInFlight 当前正在处理的CMDB请求数
	CMDBRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: BkBcsCmdbSynchronizer,
			Name:      "cmdb_requests_in_flight",
			Help:      "Number of CMDB requests currently being processed",
		},
		[]string{"endpoint"},
	)
)