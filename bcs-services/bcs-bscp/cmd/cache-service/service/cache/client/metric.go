/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bscp.io/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func initMetric() *metric {
	m := new(metric)
	labels := prometheus.Labels{}
	m.hitCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.CSCacheSubSys,
			Name:        "total_hit_cache_count",
			Help:        "the total hit count to the bedis cache",
			ConstLabels: labels,
		}, []string{"rsc", "biz"})
	metrics.Register().MustRegister(m.hitCounter)

	m.refreshLagMS = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CSCacheSubSys,
		Name:        "refresh_lag_milliseconds",
		Help:        "the lags(milliseconds) to refresh the bedis cache",
		ConstLabels: labels,
		Buckets:     []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 40, 60, 80, 100, 150, 200},
	}, []string{"rsc", "biz"})
	metrics.Register().MustRegister(m.refreshLagMS)

	m.strategyByteSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CSCacheSubSys,
		Name:        "strategy_size_bytes",
		Help:        "the total strategy size of an app's instance to be matched in bytes",
		ConstLabels: labels,
		Buckets:     []float64{400, 600, 800, 1000, 1200, 1400, 1800, 2000, 2500, 3000, 3500, 4000, 5000, 6000},
	}, []string{"rsc", "biz"})
	metrics.Register().MustRegister(m.strategyByteSize)

	m.releasedCIByteSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.CSCacheSubSys,
		Name:        "released_ci_size_bytes",
		Help:        "the total configure items size of an app's release in bytes",
		ConstLabels: labels,
		Buckets:     []float64{500, 600, 800, 1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500, 5000, 7000, 10000},
	}, []string{"rsc", "biz"})
	metrics.Register().MustRegister(m.releasedCIByteSize)

	return m
}

type metric struct {
	// hitCounter record the total count to hit the cache.
	hitCounter *prometheus.CounterVec

	// record the cost time in a milliseconds of refresh the cache.
	refreshLagMS *prometheus.HistogramVec

	// record the total strategy size of an app's instance to be matched with bytes.
	strategyByteSize *prometheus.HistogramVec

	// record the total size of an app's all the configure-items of one release with bytes.
	releasedCIByteSize *prometheus.HistogramVec
}

const (
	amRes         = "app-meta"
	instRes       = "instance"
	releasedCIRes = "release-ci"
	strategyRes   = "strategy"
)
