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

package lcache

import (
	"github.com/prometheus/client_golang/prometheus"
	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

func initMetric() *metric {
	m := new(metric)
	labels := prm.Labels{}
	m.hitCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.FSLocalCacheSubSys,
			Name:        "total_hit_cache_count",
			Help:        "the total hit count to the local cache",
			ConstLabels: labels,
		}, []string{"resource", "biz"})
	metrics.Register().MustRegister(m.hitCounter)

	m.hitRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.FSLocalCacheSubSys,
		Name:        "hit_rate",
		Help:        "record the rate of hit the local cache with different resources",
		ConstLabels: labels,
	}, []string{"resource"})
	metrics.Register().MustRegister(m.hitRate)

	m.refreshLagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.FSLocalCacheSubSys,
		Name:        "refresh_lag_milliseconds",
		Help:        "the lags(milliseconds) to refresh the local cache",
		ConstLabels: labels,
		Buckets:     []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 25, 50, 100, 150, 200},
	}, []string{"resource", "biz"})
	metrics.Register().MustRegister(m.refreshLagMS)

	m.errCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.FSLocalCacheSubSys,
			Name:        "total_err_count",
			Help:        "the total error count when refresh the local cache",
			ConstLabels: labels,
		}, []string{"resource", "biz"})
	metrics.Register().MustRegister(m.errCounter)

	m.evictCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.FSLocalCacheSubSys,
			Name:        "total_evict_count",
			Help:        "the total request count to the local cache",
			ConstLabels: labels,
		}, []string{"resource"})
	metrics.Register().MustRegister(m.evictCounter)

	return m
}

type metric struct {
	// hitCounter record the total count to hit the cache.
	hitCounter *prometheus.CounterVec

	// hitRate record the rate of hit the local cache with different resources.
	hitRate *prometheus.GaugeVec

	// record the cost time in a milliseconds of refresh the cache.
	refreshLagMS *prometheus.HistogramVec

	// errCounter record the total error count when refresh the cache.
	errCounter *prometheus.CounterVec

	// evictCounter record the total evict cache.
	evictCounter *prometheus.CounterVec
}
