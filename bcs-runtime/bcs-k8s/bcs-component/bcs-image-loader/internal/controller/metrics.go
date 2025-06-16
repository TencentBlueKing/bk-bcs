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

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	imageLoaderCompletedSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "image_loader_completed_seconds",
			Help: "The time spent to complete an image loader",
		},
		[]string{"namespace", "name", "status"},
	)

	imageLoaderFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "image_loader_failed_total",
			Help: "Number of failed image loaders",
		},
		[]string{"namespace", "name", "node"},
	)
	imageLoaderRuningSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "image_loader_running_seconds",
			Help: "The time spent to run an image loader",
		},
		[]string{"namespace", "name"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(imageLoaderCompletedSeconds, imageLoaderFailed, imageLoaderRuningSeconds)
}
