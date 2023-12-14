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

package eventc

import (
	"github.com/prometheus/client_golang/prometheus"
	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric(name string) *metric {
	m := new(metric)
	labels := prm.Labels{"name": name}
	m.consumerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.FSEventc,
		Name:        "total_consumer_count",
		Help:        "record the total of consumer count from sidecar with watch",
		ConstLabels: labels,
	}, []string{"biz", "app"})
	metrics.Register().MustRegister(m.consumerCount)

	m.retryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.FSEventc,
			Name:        "retry_list_count",
			Help:        "record the total retry list count",
			ConstLabels: labels,
		}, []string{"biz", "app"})
	metrics.Register().MustRegister(m.retryCounter)

	return m
}

type metric struct {
	// consumerCount record the total of consumer count from sidecar with watch
	consumerCount *prometheus.GaugeVec
	// retryCounter record the total retry list count.
	retryCounter *prometheus.CounterVec
}
