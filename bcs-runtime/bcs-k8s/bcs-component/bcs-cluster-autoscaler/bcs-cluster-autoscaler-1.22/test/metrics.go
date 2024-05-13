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

// Package main defines prometheus metrics
package main

import (
	"github.com/prometheus/client_golang/prometheus" // for client-go metrics registration
)

const (
	caNamespace = "cluster_autoscaler_e2e"
)

var (
	failedScaleUpCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_up_count",
			Help:      "failed scale up count.",
		},
	)

	scaleUpCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scale_up_count",
			Help:      "scale up count.",
		},
	)

	failedScaleDownCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_down_count",
			Help:      "failed scale down count.",
		},
	)

	scaleDownCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scale_down_count",
			Help:      "scale down count.",
		},
	)

	scaleUpSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: caNamespace,
			Name:      "scale_up_seconds",
			Buckets:   []float64{60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		},
	)

	scaleDownSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: caNamespace,
			Name:      "scale_down_seconds",
			Buckets:   []float64{60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		},
	)

	scaleUpSuccessRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_up_success_rate",
		},
	)
	scaleDownSuccessRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_down_success_rate",
		},
	)
)

// registerAll registers all metrics.
func registerAll() {
	prometheus.MustRegister(failedScaleUpCount)
	prometheus.MustRegister(failedScaleDownCount)
	prometheus.MustRegister(scaleUpCount)
	prometheus.MustRegister(scaleDownCount)
	prometheus.MustRegister(scaleUpSeconds)
	prometheus.MustRegister(scaleDownSeconds)
	prometheus.MustRegister(scaleUpSuccessRate)
	prometheus.MustRegister(scaleDownSuccessRate)
}
