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
	"sync"
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

	// reconcileDuration is reconcile duration(seconds) for hook operator
	reconcileDuration *prometheus.HistogramVec

	// hookrunSurviveTime is the survive time(seconds) of each hookrun
	hookrunSurviveTime *prometheus.GaugeVec

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

	// operatorImageVersion contains the image version of operator pods and CRD version
	operatorVersion *prometheus.GaugeVec
}

var (
	metricsInstance *metrics
	metricsOnce     sync.Once
)

// newMetrics new a metrics object for hook operator
func newMetrics() *metrics {

	metricsOnce.Do(func() {
		m := new(metrics)
		// it will set to be a real min val once it collects a metric
		m.hookrunExecDurationMinVal = initialMinVal
		m.metricExecDurationMinVal = initialMinVal

		m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "reconcile duration(seconds) for hook operator",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"namespace", "owner", "status"})
		prometheus.MustRegister(m.reconcileDuration)

		m.hookrunSurviveTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "hookrun_survive_time_seconds",
			Help:      "the survive time(seconds) of every hookrun until now",
		}, []string{"namespace", "owner", "name", "phase"})
		prometheus.MustRegister(m.hookrunSurviveTime)

		m.hookrunExecDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "hookrun_exec_duration_seconds",
			Help:      "the execution duration(seconds) of every hookrun",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"namespace", "owner", "status"})
		prometheus.MustRegister(m.hookrunExecDuration)

		m.hookrunExecDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "hookrun_exec_duration_seconds_max",
			Help:      "the max execution duration(seconds) of hookrun",
		}, []string{"namespace", "owner", "status"})
		prometheus.MustRegister(m.hookrunExecDurationMax)

		m.hookrunExecDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "hookrun_exec_duration_seconds_min",
			Help:      "the min execution duration(seconds) of hookrun",
		}, []string{"namespace", "owner", "status"})
		prometheus.MustRegister(m.hookrunExecDurationMin)

		m.metricExecDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "metric_exec_duration_seconds",
			Help:      "the execution duration(seconds) of every metric belong to a hookrun",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"namespace", "owner", "metric", "phase"})
		prometheus.MustRegister(m.metricExecDuration)

		m.metricExecDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "metric_exec_duration_seconds_max",
			Help:      "the max execution duration(seconds) of metric belong to a hookrun",
		}, []string{"namespace", "owner", "metric", "phase"})
		prometheus.MustRegister(m.metricExecDurationMax)

		m.metricExecDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "metric_exec_duration_seconds_min",
			Help:      "the min execution duration(seconds) of metric belong to a hookrun",
		}, []string{"namespace", "owner", "metric", "phase"})
		prometheus.MustRegister(m.metricExecDurationMin)

		m.operatorVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "operator_version",
			Help:      "operatorVersion contains the image version of hook operator pods and the version of CRD",
		}, []string{"name", "image_version", "hookrun_version", "hooktemplate_version",
			"git_version", "git_commit", "build_date"})
		prometheus.MustRegister(m.operatorVersion)

		metricsInstance = m
	})

	return metricsInstance
}

// collectReconcileDuration collect the reconcile duration(seconds) for gamedeployment operator
func (m *metrics) collectReconcileDuration(namespace, ownerRef, status string, d time.Duration) {
	m.reconcileDuration.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
		"status": status}).Observe(d.Seconds())
}

// collectHookrunExecDurations collect these metrics:
// 1.the execution duration(seconds) of every hookrun
// 2.the max execution duration(seconds) of hookrun
// 3.the min execution duration(seconds) of hookrun
func (m *metrics) collectHookrunExecDurations(namespace, ownerRef, status string, d time.Duration) {
	duration := d.Seconds()
	m.hookrunExecDuration.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
		"status": status}).Observe(duration)
	if duration > m.hookrunExecDurationMaxVal {
		m.hookrunExecDurationMaxVal = duration
		m.hookrunExecDurationMax.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
			"status": status}).Set(duration)
	}
	if duration < m.hookrunExecDurationMinVal {
		m.hookrunExecDurationMinVal = duration
		m.hookrunExecDurationMin.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
			"status": status}).Set(duration)
	}
}

// collectMetricExecDurations collect these metrics:
// 1.the execution duration(seconds) of every metric belong to a hookrun
// 2.the max execution duration(seconds) of metric belong to a hookrun
// 3.the min execution duration(seconds) of metric belong to a hookrun
func (m *metrics) collectMetricExecDurations(namespace, ownerRef, metricName, phase string, d time.Duration) {
	duration := d.Seconds()
	m.metricExecDuration.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
		"metric": metricName, "phase": phase}).Observe(duration)
	if duration > m.metricExecDurationMaxVal {
		m.metricExecDurationMaxVal = duration
		m.metricExecDurationMax.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
			"metric": metricName, "phase": phase}).Set(duration)
	}
	if duration < m.metricExecDurationMinVal {
		m.metricExecDurationMinVal = duration
		m.metricExecDurationMin.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef,
			"metric": metricName, "phase": phase}).Set(duration)
	}
}

// collectHookrunSurviveTime collect survive time of each hookrun:
func (m *metrics) collectHookrunSurviveTime(namespace, ownerRef, name, phase string, d time.Duration) {
	m.hookrunSurviveTime.With(prometheus.Labels{"namespace": namespace, "owner": ownerRef, "name": name,
		"phase": phase}).Set(d.Seconds())
}

// collectOperatorVersion collects the image version of gamestatefulset operator pods
func (m *metrics) collectOperatorVersion(imageVersion, hookrunVersion, hooktemplateVersion,
	gitVersion, gitCommit, buildDate string) {
	m.operatorVersion.WithLabelValues("Hook", imageVersion, hookrunVersion, hooktemplateVersion,
		gitVersion, gitCommit, buildDate).Set(float64(1))
}
