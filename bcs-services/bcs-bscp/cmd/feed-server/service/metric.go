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

package service

import (
	"github.com/prometheus/client_golang/prometheus"
	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric(name string) *metric {
	m := new(metric)
	labels := prm.Labels{"name": name}
	m.watchTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.FSConfigConsume,
		Name:        "current_watch_count",
		Help:        "record the current total connection count of sidecars with watch",
		ConstLabels: labels,
	}, []string{"biz"})
	metrics.Register().MustRegister(m.watchTotal)

	m.watchCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.FSConfigConsume,
			Name:      "total_watch_count",
			Help: "record the total connection count of sidecars with watch, that used to get the new connection count " +
				"within a specified time range",
			ConstLabels: labels,
		}, []string{"biz"})
	metrics.Register().MustRegister(m.watchCounter)

	return m
}

type metric struct {
	// watchTotal record the current total connection count of sidecars with watch.
	watchTotal *prometheus.GaugeVec
	// watchCounter record the total connection count of sidecars with watch, used to get the new the connection count
	// within a specified time range.
	watchCounter *prometheus.CounterVec
}
