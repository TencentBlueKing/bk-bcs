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
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "bkbcs"
var subsystem = "gamedeployment"

// Metrics used to collect prom metrics for gamedeployment operator
type Metrics struct {
	podCreateDurationMaxVal float64 //save the max create duration(seconds) value of pod
	podCreateDurationMinVal float64 //save the min create duration(seconds) value of pod
	podUpdateDurationMaxVal float64 //save the max update duration(seconds) value of pod
	podUpdateDurationMinVal float64 //save the min update duration(seconds) value of pod
	podDeleteDurationMaxVal float64 //save the max delete duration(seconds) value of pod
	podDeleteDurationMinVal float64 //save the min delete duration(seconds) value of pod

	// errorTotalCount is total count of error
	errorTotalCount *prometheus.CounterVec

	// reconcileDuration is reconcile duration(seconds) for gamedeployment operator
	reconcileDuration *prometheus.HistogramVec

	// podCreateDuration is create duration(seconds) of pod
	podCreateDuration *prometheus.HistogramVec

	// podUpdateDuration is update duration(seconds) of pod
	podUpdateDuration *prometheus.HistogramVec

	// podUpdateDuration is delete duration(seconds) of pod
	podDeleteDuration *prometheus.HistogramVec

	// podCreateDurationMax is max create duration(seconds) of pod
	podCreateDurationMax *prometheus.GaugeVec

	// podCreateDurationMin is min create duration(seconds) of pod
	podCreateDurationMin *prometheus.GaugeVec

	// podUpdateDurationMax is max update duration(seconds) of pod
	podUpdateDurationMax *prometheus.GaugeVec

	// podUpdateDurationMin is min update duration(seconds) of pod
	podUpdateDurationMin *prometheus.GaugeVec

	// podDeleteDurationMax is max delete duration(seconds) of pod
	podDeleteDurationMax *prometheus.GaugeVec

	// podDeleteDurationMin is min delete duration(seconds) of pod
	podDeleteDurationMin *prometheus.GaugeVec
}

// NewMetrics new a metrics object for gamedeployment operator
func NewMetrics() *Metrics {

	m := new(Metrics)
	m.podCreateDurationMinVal = float64(999999) // it will set to be a real min val once it collects a metric
	m.podUpdateDurationMinVal = float64(999999) // it will set to be a real min val once it collects a metric
	m.podDeleteDurationMinVal = float64(999999) // it will set to be a real min val once it collects a metric
	m.errorTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "error_total_count",
		Help:      "the total count of error",
	}, []string{"gdName"})
	prometheus.MustRegister(m.errorTotalCount)

	m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconcile_duration_seconds",
		Help:      "reconcile duration(seconds) for gamedeployment operator",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.reconcileDuration)

	m.podCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds",
		Help:      "create duration(seconds) of pod",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podCreateDuration)

	m.podUpdateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_update_duration_seconds",
		Help:      "update duration(seconds) of pod",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podUpdateDuration)

	m.podDeleteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds",
		Help:      "delete duration(seconds) of pod",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podDeleteDuration)

	m.podCreateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds_max",
		Help:      "the max create duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podCreateDurationMax)

	m.podCreateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds_min",
		Help:      "the min create duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podCreateDurationMin)

	m.podUpdateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_update_duration_seconds_max",
		Help:      "the max update duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podUpdateDurationMax)

	m.podUpdateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_update_duration_seconds_min",
		Help:      "the min update duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podUpdateDurationMin)

	m.podDeleteDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds_max",
		Help:      "the max delete duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podDeleteDurationMax)

	m.podDeleteDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds_min",
		Help:      "the max delete duration(seconds) of pod",
	}, []string{"gdName", "status"})
	prometheus.MustRegister(m.podDeleteDurationMin)

	return m
}

// CollectErrorTotalCount error total count
func (m *Metrics) CollectErrorTotalCount(gdName string) {
	m.errorTotalCount.With(prometheus.Labels{"gdName": gdName}).Inc()
}

// CollectReconcileDuration collect the reconcile duration(seconds) for gamedeployment operator
func (m *Metrics) CollectReconcileDuration(gdName, status string, d time.Duration) {
	m.reconcileDuration.With(prometheus.Labels{"gdName": gdName, "status": status}).Observe(d.Seconds())
}

// CollectPodCreateDurations collect these metrics:
// 1.the create duration(seconds) of each pod
// 2.the max create duration(seconds) of pods
// 3.the min create duration(seconds) of pods
func (m *Metrics) CollectPodCreateDurations(gdName, status string, d time.Duration) {
	duration := d.Seconds()
	m.podCreateDuration.With(prometheus.Labels{"gdName": gdName, "status": status}).Observe(duration)
	if duration > m.podCreateDurationMaxVal {
		m.podCreateDurationMaxVal = duration
		m.podCreateDurationMax.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
	if duration < m.podCreateDurationMinVal {
		m.podCreateDurationMinVal = duration
		m.podCreateDurationMin.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
}

// CollectPodUpdateDurations collect these metrics:
// 1.the update duration(seconds) of each pod
// 2.the max update duration(seconds) of pods
// 3.the min update duration(seconds) of pods
func (m *Metrics) CollectPodUpdateDurations(gdName, status string, d time.Duration) {
	duration := d.Seconds()
	m.podUpdateDuration.With(prometheus.Labels{"gdName": gdName, "status": status}).Observe(duration)
	if duration > m.podUpdateDurationMaxVal {
		m.podUpdateDurationMaxVal = duration
		m.podUpdateDurationMax.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
	if duration < m.podUpdateDurationMinVal {
		m.podUpdateDurationMinVal = duration
		m.podUpdateDurationMin.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
}

// CollectPodDeleteDurations collect these metrics:
// 1.the delete duration(seconds) of each pod
// 2.the max delete duration(seconds) of pods
// 3.the min delete duration(seconds) of pods
func (m *Metrics) CollectPodDeleteDurations(gdName, status string, d time.Duration) {
	duration := d.Seconds()
	m.podDeleteDuration.With(prometheus.Labels{"gdName": gdName, "status": status}).Observe(duration)
	if duration > m.podDeleteDurationMaxVal {
		m.podDeleteDurationMaxVal = duration
		m.podDeleteDurationMax.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
	if duration < m.podDeleteDurationMinVal {
		m.podDeleteDurationMinVal = duration
		m.podDeleteDurationMin.With(prometheus.Labels{"gdName": gdName, "status": status}).Set(duration)
	}
}
