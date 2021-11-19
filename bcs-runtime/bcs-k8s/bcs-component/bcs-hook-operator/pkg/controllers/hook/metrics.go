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

package hook

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "bkbcs"
var subsystem = "hook"

const initialMinVal = 999999

// metrics used to collect prom metrics for hook operator
type metrics struct {
	hookrunExecDurationMaxVal float64 //save the max execution duration(seconds) value of hookrun
	hookrunExecDurationMinVal float64 //save the min execution duration(seconds) value of hookrun
	metricExecDurationMaxVal  float64 //save the max execution duration(seconds) value of metric for a hookrun
	metricExecDurationMinVal  float64 //save the min execution duration(seconds) value of metric for a hookrun

	// errorTotalCount is total count of error
	errorTotalCount *prometheus.CounterVec

	// reconcileDuration is reconcile duration(seconds) for hook operator
	reconcileDuration *prometheus.HistogramVec

	// hookrunRequestCount is count of request for hookrun
	hookrunRequestCount *prometheus.CounterVec

	// hookrunExecDuration is execution duration(seconds) of each hookrun
	hookrunExecDuration *prometheus.HistogramVec

	// hookrunExecDurationMax is the max execution duration(seconds) of hookrun
	hookrunExecDurationMax *prometheus.GaugeVec

	// hookrunExecDurationMin is the min exection duration(seconds) of hookrun
	hookrunExecDurationMin *prometheus.GaugeVec

	// metricExecDuration is execution duration(seconds) of each metric belong to a hookrun
	metricExecDuration *prometheus.HistogramVec

	// metricExecDurationMax is the max execution duration(seconds) of metric belong to a hookrun
	metricExecDurationMax *prometheus.GaugeVec

	// metricExecDurationMin is the min execution duration(seconds) of metric belong to a hookrun
	metricExecDurationMin *prometheus.GaugeVec
}

// newMetrics new a metrics object for hook operator
func newMetrics() *metrics {

	m := new(metrics)
	// it will set to be a real min val once it collects a metric
	m.hookrunExecDurationMinVal = initialMinVal
	m.metricExecDurationMinVal = initialMinVal

	m.errorTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "error_total_count",
		Help:      "the total count of error",
	}, []string{"namespace"})
	prometheus.MustRegister(m.errorTotalCount)

	m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconcile_duration_seconds",
		Help:      "reconcile duration(seconds) for hook operator",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"namespace", "status"})
	prometheus.MustRegister(m.reconcileDuration)

	m.hookrunRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_request_count",
		Help:      "the count of request for hookrun",
	}, []string{"namespace", "status"})
	prometheus.MustRegister(m.hookrunRequestCount)

	m.hookrunExecDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds",
		Help:      "the execution duration(seconds) of every hookrun",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"namespace", "status"})
	prometheus.MustRegister(m.hookrunExecDuration)

	m.hookrunExecDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds_max",
		Help:      "the max execution duration(seconds) of hookrun",
	}, []string{"namespace", "status"})
	prometheus.MustRegister(m.hookrunExecDurationMax)

	m.hookrunExecDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds_min",
		Help:      "the min execution duration(seconds) of hookrun",
	}, []string{"namespace", "status"})
	prometheus.MustRegister(m.hookrunExecDurationMin)

	m.metricExecDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "metric_exec_duration_seconds",
		Help:      "the execution duration(seconds) of every metric belong to a hookrun",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{"namespace", "metric", "phase"})
	prometheus.MustRegister(m.metricExecDuration)

	m.metricExecDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "metric_exec_duration_seconds_max",
		Help:      "the max execution duration(seconds) of metric belong to a hookrun",
	}, []string{"namespace", "metric", "phase"})
	prometheus.MustRegister(m.metricExecDurationMax)

	m.metricExecDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "metric_exec_duration_seconds_min",
		Help:      "the min execution duration(seconds) of metric belong to a hookrun",
	}, []string{"namespace", "metric", "phase"})
	prometheus.MustRegister(m.metricExecDurationMin)

	return m
}

// collectErrorTotalCount collect error total count
func (m *metrics) collectErrorTotalCount(namespace string) {
	m.errorTotalCount.With(prometheus.Labels{"namespace": namespace}).Inc()
}

// collectReconcileDuration collect the reconcile duration(seconds) for gamedeployment operator
func (m *metrics) collectReconcileDuration(namespace, status string, d time.Duration) {
	m.reconcileDuration.With(prometheus.Labels{"namespace": namespace, "status": status}).Observe(d.Seconds())
}

// collectHookrunRequestCount collect the count of request for hookrun
func (m *metrics) collectHookrunRequestCount(namespace, status string) {
	m.hookrunRequestCount.With(prometheus.Labels{"namespace": namespace, "status": status}).Inc()
}

// collectHookrunExecDurations collect these metrics:
// 1.the execution duration(seconds) of every hookrun
// 2.the max execution duration(seconds) of hookrun
// 3.the min execution duration(seconds) of hookrun
func (m *metrics) collectHookrunExecDurations(namespace, status string, d time.Duration) {
	duration := d.Seconds()
	m.hookrunExecDuration.With(prometheus.Labels{"namespace": namespace, "status": status}).Observe(duration)
	if duration > m.hookrunExecDurationMaxVal {
		m.hookrunExecDurationMaxVal = duration
		m.hookrunExecDurationMax.With(prometheus.Labels{"namespace": namespace, "status": status}).Set(duration)
	}
	if duration < m.hookrunExecDurationMinVal {
		m.hookrunExecDurationMinVal = duration
		m.hookrunExecDurationMin.With(prometheus.Labels{"namespace": namespace, "status": status}).Set(duration)
	}
}

// collectMetricExecDurations collect these metrics:
// 1.the execution duration(seconds) of every metric belong to a hookrun
// 2.the max execution duration(seconds) of metric belong to a hookrun
// 3.the min execution duration(seconds) of metric belong to a hookrun
func (m *metrics) collectMetricExecDurations(namespace, metricName, phase string, d time.Duration) {
	duration := d.Seconds()
	m.metricExecDuration.With(prometheus.Labels{"namespace": namespace, "metric": metricName,
		"phase": phase}).Observe(duration)
	if duration > m.metricExecDurationMaxVal {
		m.metricExecDurationMaxVal = duration
		m.metricExecDurationMax.With(prometheus.Labels{"namespace": namespace, "metric": metricName,
			"phase": phase}).Set(duration)
	}
	if duration < m.metricExecDurationMinVal {
		m.metricExecDurationMinVal = duration
		m.metricExecDurationMin.With(prometheus.Labels{"namespace": namespace, "metric": metricName,
			"phase": phase}).Set(duration)
	}
}
