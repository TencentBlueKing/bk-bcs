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
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Success         = "success"
	Failure         = "failure"
	AddOperation    = "add"
	UpdateOperation = "update"
	DeleteOperation = "delete"
)

// Metrics the bcsscheduler plugin exports.
var (
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "request_count_total",
		Help:      "Counter of requests to plugin bcsscheduler.",
	}, []string{"status"})

	RequestOutProxyCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "request_out_proxy_count_total",
		Help:      "Counter of requests to external dns.",
	}, []string{"status"})

	RequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "request_latency_seconds",
		Buckets:   plugin.TimeBuckets,
		Help:      "Histogram of the time (in seconds) each request took.",
	}, []string{"status"})

	DnsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "dns_total",
		Help:      "Total counter of dns in plugin bcsscheduler.",
	})

	ZkNotifyTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "zookeeper_notify_total",
		Help:      "counter of zookeeper notify.",
	}, []string{"operator"})

	StorageOperatorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "storage_operator_total",
		Help:      "counter of storage operator.",
	}, []string{"operator"})

	StorageOperatorLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: plugin.Namespace,
		Subsystem: "bcsscheduler",
		Name:      "storage_operator_latency_seconds",
		Buckets:   plugin.TimeBuckets,
		Help:      "Histogram of the time (in seconds) each storage operator.",
	}, []string{"status"})
)
