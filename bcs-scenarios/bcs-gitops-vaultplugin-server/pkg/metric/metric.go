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

// Package metric defines the metric info of vaultplugin
package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// RequestTotal used to sum task create num
	RequestTotal *prometheus.CounterVec
	// RequestDuration the latency of task handled
	RequestDuration *prometheus.HistogramVec
	// RequestFailed defines the failed request num
	RequestFailed *prometheus.CounterVec
)

func init() {
	RequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gitops_vaultpluginserver_request_total",
		Help: "number of server received",
	}, []string{})
	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gitops_vaultpluginserver_request_duration",
		Help:    "the time took of handle requests",
		Buckets: []float64{0.1, 0.3, 1.2, 5, 10},
	}, []string{})
	RequestFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gitops_vaultpluginserver_request_failed",
		Help: "failed number of request",
	}, []string{})

	prometheus.MustRegister(RequestTotal)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(RequestFailed)
}
