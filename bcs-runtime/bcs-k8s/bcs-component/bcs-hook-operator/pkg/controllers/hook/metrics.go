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

// metrics used to collect prom metrics for hook operator
type metrics struct {
	hookrunExecDuraationMaxVal float64 //save the max execution duration(seconds) value of hookrun
	hookrunExecDuraationMinVal float64 //save the min execution duration(seconds) value of hookrun

	// errorTotalCount is total count of error
	errorTotalCount *prometheus.CounterVec

	// reconcileDuration is reconcile duration(seconds) for hook operator
	reconcileDuration *prometheus.HistogramVec

	// hookrunRequestTotalCount is total count of request for hookrun
	hookrunRequestTotalCount *prometheus.CounterVec

	// hookrunRequestSuccessCount is success count of request for hookrun
	hookrunRequestSuccessCount *prometheus.CounterVec

	// hookrunRequestFailCount is fail count of request for hookrun
	hookrunRequestFailCount *prometheus.CounterVec

	// hookrunExecDuration is execution duration(seconds) of each hookrun
	hookrunExecDuration *prometheus.HistogramVec

	// hookrunExecDurationMax is the max execution duration(seconds) of hookrun
	hookrunExecDurationMax *prometheus.GaugeVec

	// hookrunExecDurationMin is the min exection duration(seconds) of hookrun
	hookrunExecDurationMin *prometheus.GaugeVec
}

// newMetrics new a metrics object for hook operator
func newMetrics() *metrics {

	m := new(metrics)
	m.hookrunExecDuraationMinVal = float64(999999) // it will set to be a real min val once it collects a metric

	m.errorTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "error_total_count",
		Help:      "the total count of error",
	}, []string{})
	prometheus.MustRegister(m.errorTotalCount)

	m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "reconcile_duration_seconds",
		Help:      "reconcile duration(seconds) for hook operator",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{})
	prometheus.MustRegister(m.reconcileDuration)

	m.hookrunRequestTotalCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_request_total_count",
		Help:      "the total count of request for hookrun",
	}, []string{})
	prometheus.MustRegister(m.hookrunRequestTotalCount)

	m.hookrunRequestSuccessCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_request_success_count",
		Help:      "the success count of request for hookrun",
	}, []string{})
	prometheus.MustRegister(m.hookrunRequestSuccessCount)

	m.hookrunRequestFailCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_request_fail_count",
		Help:      "the fail count of request for hookrun",
	}, []string{})
	prometheus.MustRegister(m.hookrunRequestFailCount)

	m.hookrunExecDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds",
		Help:      "the execution duration(seconds) of every hookrun",
		Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
	}, []string{})
	prometheus.MustRegister(m.hookrunExecDuration)

	m.hookrunExecDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds_max",
		Help:      "the max execution duration(seconds) of hookrun",
	}, []string{})
	prometheus.MustRegister(m.hookrunExecDurationMax)

	m.hookrunExecDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hookrun_exec_duration_seconds_min",
		Help:      "the min execution duration(seconds) of hookrun",
	}, []string{})
	prometheus.MustRegister(m.hookrunExecDurationMin)

	return m
}

// collectErrorTotalCount error total count
func (m *metrics) collectErrorTotalCount() {
	m.errorTotalCount.With(prometheus.Labels{}).Inc()
}

// collectReconcileDuration collect the reconcile duration(seconds) for gamedeployment operator
func (m *metrics) collectReconcileDuration(d time.Duration) {
	m.reconcileDuration.With(prometheus.Labels{}).Observe(d.Seconds())
}

// collectHookrunRequestTotalCount collect the total count of request for hookrun
func (m *metrics) collectHookrunRequestTotalCount() {
	m.hookrunRequestTotalCount.With(prometheus.Labels{}).Inc()
}

// collectHookrunRequestSuccessCount collect the success count of request for hookrun
func (m *metrics) collectHookrunRequestSuccessCount() {
	m.hookrunRequestSuccessCount.With(prometheus.Labels{}).Inc()
}

// collecthookRunRequestFailCount collect the fail count of request for hookrun
func (m *metrics) collecthookRunRequestFailCount() {
	m.hookrunRequestFailCount.With(prometheus.Labels{}).Inc()
}

// collectHookrunExecDurations collect these metrics:
// 1.the execution duration(seconds) of every hookrun
// 2.the max execution duration(seconds) of hookrun
// 3.the min execution duration(seconds) of hookrun
func (m *metrics) collectHookrunExecDurations(d time.Duration) {
	duration := d.Seconds()
	m.hookrunExecDuration.With(prometheus.Labels{}).Observe(duration)
	if duration > m.hookrunExecDuraationMaxVal {
		m.hookrunExecDuraationMaxVal = duration
		m.hookrunExecDurationMax.With(prometheus.Labels{}).Set(duration)
	}
	if duration < m.hookrunExecDuraationMinVal {
		m.hookrunExecDuraationMinVal = duration
		m.hookrunExecDurationMin.With(prometheus.Labels{}).Set(duration)
	}
}
