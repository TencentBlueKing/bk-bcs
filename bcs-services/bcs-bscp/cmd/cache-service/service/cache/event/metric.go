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

package event

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric() *metric {
	m := new(metric)
	labels := prometheus.Labels{}
	m.eventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.CSEventSubSys,
			Name:        "total_event_count",
			Help:        "record the total consumed event counts",
			ConstLabels: labels,
		}, []string{"type", "biz"})
	metrics.Register().MustRegister(m.eventCounter)

	m.loopLagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CSEventSubSys,
		Name:        "loop_lag_milliseconds",
		Help:        "the lags(milliseconds) to consumes one batch of events",
		ConstLabels: labels,
		Buckets:     []float64{30, 50, 100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000, 7000, 10000},
	}, []string{})
	metrics.Register().MustRegister(m.loopLagMS)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.CSEventSubSys,
			Name:        "total_err_count",
			Help:        "the total error count when loop or consume the events",
			ConstLabels: labels,
		}, []string{})
	metrics.Register().MustRegister(m.errCounter)

	m.lastCursor = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CSEventSubSys,
		Name:        "last_cursor",
		Help:        "record the last consumed cursor id by cache service",
		ConstLabels: labels,
	}, []string{})
	metrics.Register().MustRegister(m.lastCursor)

	return m
}

type metric struct {
	// eventCounter record the total consumed event counts.
	eventCounter *prometheus.CounterVec

	// record the cost time of consumes one batch of events.
	loopLagMS *prometheus.HistogramVec

	// errCounter record the total error count when consume the events.
	errCounter *prometheus.CounterVec

	// lastCursor record the last consumed cursor id.
	lastCursor *prometheus.GaugeVec
}
