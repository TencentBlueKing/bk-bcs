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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "bkbcs"
var subsystem = "gamedeployment"
var IsGrace string

const (
	SuccessStatus          = "success"
	FailureStatus          = "failure"
	InplaceUpdateStrategy  = "inplaceUpdate"
	HotPatchUpdateStrategy = "hotPatchUpdate"
	DeletePodAction        = "deletePod"
)

const initialMinVal = 999999

// Metrics used to collect prom metrics for gamedeployment operator
type Metrics struct {
	podCreateDurationMaxVal float64 //save the max create duration(seconds) value of pod
	podCreateDurationMinVal float64 //save the min create duration(seconds) value of pod
	podUpdateDurationMaxVal float64 //save the max update duration(seconds) value of pod
	podUpdateDurationMinVal float64 //save the min update duration(seconds) value of pod
	podDeleteDurationMaxVal float64 //save the max delete duration(seconds) value of pod
	podDeleteDurationMinVal float64 //save the min delete duration(seconds) value of pod

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

	// replicas is the number of Pods created by the GameDeployment controller
	replicas *prometheus.GaugeVec

	// readyReplicas is the number of Pods created by the GameDeployment controller that have a Ready Condition
	readyReplicas *prometheus.GaugeVec

	// availableReplicas is the number of Pods created by the GameDeployment controller that have a Ready Condition
	// for at least minReadySeconds
	availableReplicas *prometheus.GaugeVec

	// updatedReplicas is the number of Pods created by the GameDeployment controller from the GameDeployment version
	// indicated by updateRevision
	updatedReplicas *prometheus.GaugeVec

	// updatedReadyReplicas is the number of Pods created by the GameDeployment controller from the
	// GameDeployment version indicated by updateRevision and have a Ready Condition
	updatedReadyReplicas *prometheus.GaugeVec

	// operatorImageVersion contains the image version of operator pods and CRD version
	operatorVersion *prometheus.GaugeVec
}

var metrics *Metrics
var metricsOnce sync.Once

// NewMetrics new a metrics object for gamedeployment operator
func NewMetrics() *Metrics {

	metricsOnce.Do(func() {
		m := new(Metrics)
		// it will set to be a real min val once it collects a metric
		m.podCreateDurationMinVal = initialMinVal
		m.podUpdateDurationMinVal = initialMinVal
		m.podDeleteDurationMinVal = initialMinVal

		m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "reconcile duration(seconds) for gamedeployment operator",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"gd", "status"})
		prometheus.MustRegister(m.reconcileDuration)

		m.podCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds",
			Help:      "create duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"gd", "status"})
		prometheus.MustRegister(m.podCreateDuration)

		m.podUpdateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds",
			Help:      "update duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podUpdateDuration)

		m.podDeleteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds",
			Help:      "delete duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podDeleteDuration)

		m.podCreateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds_max",
			Help:      "the max create duration(seconds) of pod",
		}, []string{"gd", "status"})
		prometheus.MustRegister(m.podCreateDurationMax)

		m.podCreateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds_min",
			Help:      "the min create duration(seconds) of pod",
		}, []string{"gd", "status"})
		prometheus.MustRegister(m.podCreateDurationMin)

		m.podUpdateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds_max",
			Help:      "the max update duration(seconds) of pod",
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podUpdateDurationMax)

		m.podUpdateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds_min",
			Help:      "the min update duration(seconds) of pod",
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podUpdateDurationMin)

		m.podDeleteDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds_max",
			Help:      "the max delete duration(seconds) of pod",
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podDeleteDurationMax)

		m.podDeleteDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds_min",
			Help:      "the max delete duration(seconds) of pod",
		}, []string{"gd", "status", "grace", "action"})
		prometheus.MustRegister(m.podDeleteDurationMin)

		m.replicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "replicas",
			Help:      "the number of Pods created by the GameDeployment controller",
		}, []string{"gd"})
		prometheus.MustRegister(m.replicas)

		m.readyReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "ready_replicas",
			Help:      "the number of Pods created by the GameDeployment controller that have a Ready Condition",
		}, []string{"gd"})
		prometheus.MustRegister(m.readyReplicas)

		m.availableReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "available_replicas",
			Help: "availableReplicas is the number of Pods created by the GameDeployment controller that have a " +
				"Ready Condition for at least minReadySeconds",
		}, []string{"gd"})
		prometheus.MustRegister(m.availableReplicas)

		m.updatedReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "updated_replicas",
			Help: "the number of Pods created by the GameDeployment controller from the GameDeployment version" +
				"indicated by updateRevision",
		}, []string{"gd"})
		prometheus.MustRegister(m.updatedReplicas)

		m.updatedReadyReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "updated_ready_replicas",
			Help: "updatedReadyReplicas is the number of Pods created by the GameDeployment controller from the" +
				"GameDeployment version indicated by updateRevision and have a Ready Condition",
		}, []string{"gd"})
		prometheus.MustRegister(m.updatedReadyReplicas)

		m.operatorVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "operator_version",
			Help:      "operatorVersion contains the image version of gamedeployment operator pods and the version of CRD",
		}, []string{"name", "image_version", "crd_version", "git_version", "git_commit", "build_date"})
		prometheus.MustRegister(m.operatorVersion)

		metrics = m
	})

	return metrics
}

// CollectReconcileDuration collect the reconcile duration(seconds) for gamedeployment operator
func (m *Metrics) CollectReconcileDuration(gdName, status string, d time.Duration) {
	m.reconcileDuration.With(prometheus.Labels{"gd": gdName, "status": status}).Observe(d.Seconds())
}

// CollectPodCreateDurations collect these metrics:
// 1.the create duration(seconds) of each pod
// 2.the max create duration(seconds) of pods
// 3.the min create duration(seconds) of pods
func (m *Metrics) CollectPodCreateDurations(gdName, status string, d time.Duration) {
	duration := d.Seconds()
	m.podCreateDuration.WithLabelValues(gdName, status).Observe(duration)
	if duration > m.podCreateDurationMaxVal {
		m.podCreateDurationMaxVal = duration
		m.podCreateDurationMax.WithLabelValues(gdName, status).Set(duration)
	}
	if duration < m.podCreateDurationMinVal {
		m.podCreateDurationMinVal = duration
		m.podCreateDurationMin.WithLabelValues(gdName, status).Set(duration)
	}
}

// CollectPodUpdateDurations collect these metrics:
// 1.the update duration(seconds) of each pod
// 2.the max update duration(seconds) of pods
// 3.the min update duration(seconds) of pods
func (m *Metrics) CollectPodUpdateDurations(gdName, status, action, grace string, d time.Duration) {
	duration := d.Seconds()
	m.podUpdateDuration.WithLabelValues(gdName, status, grace, action).Observe(duration)

	if duration > m.podUpdateDurationMaxVal {
		m.podUpdateDurationMaxVal = duration
		m.podUpdateDurationMax.WithLabelValues(gdName, status, grace, action).Set(duration)
	}
	if duration < m.podUpdateDurationMinVal {
		m.podUpdateDurationMinVal = duration
		m.podUpdateDurationMin.WithLabelValues(gdName, status, grace, action).Set(duration)
	}
}

// CollectPodDeleteDurations collect these metrics:
// 1.the delete duration(seconds) of each pod
// 2.the max delete duration(seconds) of pods
// 3.the min delete duration(seconds) of pods
func (m *Metrics) CollectPodDeleteDurations(gdName, status string, action, grace string, d time.Duration) {
	duration := d.Seconds()
	m.podDeleteDuration.WithLabelValues(gdName, status, grace, action).Observe(duration)
	if duration > m.podDeleteDurationMaxVal {
		m.podDeleteDurationMaxVal = duration
		m.podDeleteDurationMax.WithLabelValues(gdName, status, grace, action).Set(duration)
	}
	if duration < m.podDeleteDurationMinVal {
		m.podDeleteDurationMinVal = duration
		m.podDeleteDurationMin.WithLabelValues(gdName, status, grace, action).Set(duration)
	}
}

// CollectRelatedReplicas collect replicas, readyReplicas, availableReplicas, updatedReplicas, updatedReadyReplicas
func (m *Metrics) CollectRelatedReplicas(gdName string,
	replicas, readyReplicas, availableReplicas, updatedReplicas, updatedReadyReplicas int32) {
	m.replicas.With(prometheus.Labels{"gd": gdName}).Set(float64(replicas))
	m.readyReplicas.With(prometheus.Labels{"gd": gdName}).Set(float64(readyReplicas))
	m.availableReplicas.With(prometheus.Labels{"gd": gdName}).Set(float64(availableReplicas))
	m.updatedReplicas.With(prometheus.Labels{"gd": gdName}).Set(float64(updatedReplicas))
	m.updatedReadyReplicas.With(prometheus.Labels{"gd": gdName}).Set(float64(updatedReadyReplicas))
}

// CollectOperatorVersion collects the image version of GameDeployment operator pods
func (m *Metrics) CollectOperatorVersion(imageVersion, CRDVersion, gitVersion, gitCommit, buildDate string) {
	m.operatorVersion.WithLabelValues("GameDeployment", imageVersion, CRDVersion,
		gitVersion, gitCommit, buildDate).Set(float64(1))
}
