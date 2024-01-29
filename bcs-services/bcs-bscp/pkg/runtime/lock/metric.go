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

package lock

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric() *metric {
	m := new(metric)
	labels := prometheus.Labels{}
	m.totalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.ResourceLockSubSys,
			Name:        "total_count",
			Help:        "the total request counts which are try to acquire the resource lock",
			ConstLabels: labels,
		}, []string{})
	metrics.Register().MustRegister(m.totalCounter)

	m.acquiredRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.ResourceLockSubSys,
		Name:        "acquired_rate",
		Help:        "the acquire rate is the value of acquired_count/total_count",
		ConstLabels: labels,
	}, []string{})
	metrics.Register().MustRegister(m.acquiredRate)

	m.acquiredCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.ResourceLockSubSys,
			Name:        "acquired_count",
			Help:        "the total request count which are are successfully acquire the resource lock",
			ConstLabels: labels,
		}, []string{})
	metrics.Register().MustRegister(m.acquiredCounter)

	return m
}

type metric struct {
	// totalCounter record all the try to lock request count.
	totalCounter *prometheus.CounterVec

	// acquiredCounter all the acquired lock request count.
	acquiredCounter *prometheus.CounterVec

	// acquiredRate record the acquired rate of the total request.
	acquiredRate *prometheus.GaugeVec
}
