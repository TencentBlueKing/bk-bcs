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

package gamestatefulset

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var namespace = "bkbcs"
var subsystem = "gamestatefulset"

const largeNumber = float64(999999)

// metrics used to collect prom metrics for gamestatefulset operator
type metrics struct {
	podCreateDurationMaxVal map[string]float64 //save the max create duration(seconds) value of pod
	podCreateDurationMinVal map[string]float64 //save the min create duration(seconds) value of pod
	podUpdateDurationMaxVal map[string]float64 //save the max update duration(seconds) value of pod
	podUpdateDurationMinVal map[string]float64 //save the min update duration(seconds) value of pod
	podDeleteDurationMaxVal map[string]float64 //save the max delete duration(seconds) value of pod
	podDeleteDurationMinVal map[string]float64 //save the min delete duration(seconds) value of pod

	// reconcileDuration is reconcile duration(seconds) for gamestatefulset operator
	reconcileDuration *prometheus.HistogramVec

	// podCreateDuration is update duration(seconds) of pod
	podCreateDuration *prometheus.HistogramVec

	// podUpdateDuration is update duration(seconds) of pod
	podUpdateDuration *prometheus.HistogramVec

	// podUpdateDuration is delete duration(seconds) of pod
	podDeleteDuration *prometheus.HistogramVec

	// podCreateDurationMax is max update duration(seconds) of pod
	podCreateDurationMax *prometheus.GaugeVec

	// podCreateDurationMin is max update duration(seconds) of pod
	podCreateDurationMin *prometheus.GaugeVec

	// podUpdateDurationMax is max update duration(seconds) of pod
	podUpdateDurationMax *prometheus.GaugeVec

	// podUpdateDurationMin is min update duration(seconds) of pod
	podUpdateDurationMin *prometheus.GaugeVec

	// podDeleteDurationMax is max delete duration(seconds) of pod
	podDeleteDurationMax *prometheus.GaugeVec

	// podDeleteDurationMin is min delete duration(seconds) of pod
	podDeleteDurationMin *prometheus.GaugeVec

	// replicas is the number of Pods created by the GameStatefulSet controller
	replicas *prometheus.GaugeVec

	// readyReplicas is the number of Pods created by the GameStatefulSet controller that have a Ready Condition
	readyReplicas *prometheus.GaugeVec

	// availableReplicas is the number of Pods created by the GameStatefulSet controller that have a Ready Condition
	// for at least minReadySeconds
	currentReplicas *prometheus.GaugeVec

	// updatedReplicas is the number of Pods created by the GameStatefulSet controller from the GameStatefulSet version
	// indicated by updateRevision
	updatedReplicas *prometheus.GaugeVec

	// updatedReadyReplicas is the number of Pods created by the GameStatefulSet controller from the
	// GameStatefulSet version indicated by updateRevision and have a Ready Condition
	updatedReadyReplicas *prometheus.GaugeVec

	// operatorImageVersion contains the image version of operator pods and CRD version
	operatorVersion *prometheus.GaugeVec
}

var (
	metricsInstance *metrics
	metricsOnce     sync.Once
)

var commonLabels = []string{"namespace", "name", "status", "action", "grace"}

// newMetrics new a metrics object for gamestatefulset operator
func newMetrics() *metrics {

	metricsOnce.Do(func() {
		m := new(metrics)
		m.podCreateDurationMaxVal = map[string]float64{}
		m.podCreateDurationMinVal = map[string]float64{}
		m.podUpdateDurationMaxVal = map[string]float64{}
		m.podUpdateDurationMinVal = map[string]float64{}
		m.podDeleteDurationMaxVal = map[string]float64{}
		m.podDeleteDurationMinVal = map[string]float64{}

		m.reconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "reconcile_duration_seconds",
			Help:      "reconcile duration(seconds) for gamestatefulset operator",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"namespace", "name", "status"})
		prometheus.MustRegister(m.reconcileDuration)

		m.podCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds",
			Help:      "create duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, []string{"namespace", "name", "status"})
		prometheus.MustRegister(m.podCreateDuration)

		m.podUpdateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds",
			Help:      "update duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, commonLabels)
		prometheus.MustRegister(m.podUpdateDuration)

		m.podDeleteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds",
			Help:      "delete duration(seconds) of pod",
			Buckets:   []float64{0.001, 0.01, 0.1, 0.5, 1, 5, 10, 20, 30, 60, 120},
		}, commonLabels)
		prometheus.MustRegister(m.podDeleteDuration)

		m.podCreateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds_max",
			Help:      "the max create duration(seconds) of pod",
		}, []string{"namespace", "name", "status"})
		prometheus.MustRegister(m.podCreateDurationMax)

		m.podCreateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_create_duration_seconds_min",
			Help:      "the min create duration(seconds) of pod",
		}, []string{"namespace", "name", "status"})
		prometheus.MustRegister(m.podCreateDurationMin)

		m.podUpdateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds_max",
			Help:      "the max update duration(seconds) of pod",
		}, commonLabels)
		prometheus.MustRegister(m.podUpdateDurationMax)

		m.podUpdateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_update_duration_seconds_min",
			Help:      "the min update duration(seconds) of pod",
		}, commonLabels)
		prometheus.MustRegister(m.podUpdateDurationMin)

		m.podDeleteDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds_max",
			Help:      "the max delete duration(seconds) of pod",
		}, commonLabels)
		prometheus.MustRegister(m.podDeleteDurationMax)

		m.podDeleteDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pod_delete_duration_seconds_min",
			Help:      "the min delete duration(seconds) of pod",
		}, commonLabels)
		prometheus.MustRegister(m.podDeleteDurationMin)

		m.replicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "replicas",
			Help:      "the number of Pods created by the GameStatefulSet controller",
		}, []string{"namespace", "name"})
		prometheus.MustRegister(m.replicas)

		m.readyReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "ready_replicas",
			Help:      "the number of Pods created by the GameStatefulSet controller that have a Ready Condition",
		}, []string{"namespace", "name"})
		prometheus.MustRegister(m.readyReplicas)

		m.currentReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "current_replicas",
			Help: "currentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version" +
				"indicated by currentRevision",
		}, []string{"namespace", "name"})
		prometheus.MustRegister(m.currentReplicas)

		m.updatedReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "updated_replicas",
			Help: "the number of Pods created by the GameStatefulSet controller from the GameStatefulSet version" +
				"indicated by updateRevision",
		}, []string{"namespace", "name"})
		prometheus.MustRegister(m.updatedReplicas)

		m.updatedReadyReplicas = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "updated_ready_replicas",
			Help: "updatedReadyReplicas is the number of Pods created by the GameStatefulSet controller from the" +
				"GameStatefulSet version indicated by updateRevision and have a Ready Condition",
		}, []string{"namespace", "name"})
		prometheus.MustRegister(m.updatedReadyReplicas)

		m.operatorVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "operator_version",
			Help:      "operatorVersion contains the image version of gamestatefulset operator pods and the version of CRD",
		}, []string{"name", "image_version", "crd_version", "git_version", "git_commit", "build_date"})
		prometheus.MustRegister(m.operatorVersion)

		metricsInstance = m
	})

	return metricsInstance
}

// collectReconcileDuration collect the reconcile duration(seconds) for gamestatefulset operator
func (m *metrics) collectReconcileDuration(namespace, name, status string, d time.Duration) {
	m.reconcileDuration.WithLabelValues(namespace, name, status).Observe(d.Seconds())
}

// collectPodCreateDurations collect below metrics:
// 1.the create duration(seconds) of each pod
// 2.the max create duration(seconds) of pod
// 3.the min create duration(seconds) of pod
func (m *metrics) collectPodCreateDurations(namespace, name, status string, d time.Duration) {
	duration := d.Seconds()
	key := fmt.Sprintf("%s/%s", namespace, name)

	m.podCreateDuration.WithLabelValues(namespace, name, status).Observe(duration)
	if duration > m.podCreateDurationMaxVal[key] {
		m.podCreateDurationMaxVal[key] = duration
		m.podCreateDurationMax.WithLabelValues(namespace, name, status).Set(duration)
	}

	if m.podCreateDurationMinVal[key] == float64(0) {
		m.podCreateDurationMinVal[key] = largeNumber
		if duration < m.podCreateDurationMinVal[key] {
			m.podCreateDurationMinVal[key] = duration
			m.podCreateDurationMin.WithLabelValues(namespace, name, status).Set(duration)
		}
	} else {
		if duration < m.podCreateDurationMinVal[key] {
			m.podCreateDurationMinVal[key] = duration
			m.podCreateDurationMin.WithLabelValues(namespace, name, status).Set(duration)
		}
	}
}

// collectPodUpdateDurations collect below metrics:
// 1.the update duration(seconds) of each pod
// 2.the max update duration(seconds) of pod
// 3.the min update duration(seconds) of pod
func (m *metrics) collectPodUpdateDurations(namespace, name, status, action, grace string, d time.Duration) {
	duration := d.Seconds()
	key := fmt.Sprintf("%s/%s", namespace, name)
	m.podUpdateDuration.WithLabelValues(namespace, name, status, action, grace).Observe(duration)
	if duration > m.podUpdateDurationMaxVal[key] {
		m.podUpdateDurationMaxVal[key] = duration
		m.podUpdateDurationMax.WithLabelValues(namespace, name, status, action, grace).Set(duration)
	}

	if m.podUpdateDurationMinVal[key] == float64(0) {
		m.podUpdateDurationMinVal[key] = largeNumber
		if duration < m.podUpdateDurationMinVal[key] {
			m.podUpdateDurationMinVal[key] = duration
			m.podUpdateDurationMin.WithLabelValues(namespace, name, status, action, grace).Set(duration)
		}
	} else {
		if duration < m.podUpdateDurationMinVal[key] {
			m.podUpdateDurationMinVal[key] = duration
			m.podUpdateDurationMin.WithLabelValues(namespace, name, status, action, grace).Set(duration)
		}
	}
}

// collectPodDeleteDurations collect these metrics:
// 1.the delete duration(seconds) of each pod
// 2.the max delete duration(seconds) of pod
// 3.the min delete duration(seconds) of pod
func (m *metrics) collectPodDeleteDurations(namespace, name, status, action, grace string, d time.Duration) {
	duration := d.Seconds()
	key := fmt.Sprintf("%s/%s", namespace, name)

	m.podDeleteDuration.WithLabelValues(namespace, name, status, action, grace).Observe(duration)
	if duration > m.podDeleteDurationMaxVal[key] {
		m.podDeleteDurationMaxVal[key] = duration
		m.podDeleteDurationMax.WithLabelValues(namespace, name, status, action, grace).Set(duration)
	}

	if m.podDeleteDurationMinVal[key] == float64(0) {
		m.podDeleteDurationMinVal[key] = largeNumber
		if duration < m.podDeleteDurationMinVal[key] {
			m.podDeleteDurationMinVal[key] = duration
			m.podDeleteDurationMin.WithLabelValues(namespace, name, status, action, grace).Set(duration)
		}
	} else {
		if duration < m.podDeleteDurationMinVal[key] {
			m.podDeleteDurationMinVal[key] = duration
			m.podDeleteDurationMin.WithLabelValues(namespace, name, status, action, grace).Set(duration)
		}
	}
}

// CollectRelatedReplicas collect replicas, readyReplicas, availableReplicas, updatedReplicas, updatedReadyReplicas
func (m *metrics) collectRelatedReplicas(namespace, name string,
	replicas, readyReplicas, availableReplicas, updatedReplicas, updatedReadyReplicas int32) {
	m.replicas.WithLabelValues(namespace, name).Set(float64(replicas))
	m.readyReplicas.WithLabelValues(namespace, name).Set(float64(readyReplicas))
	m.currentReplicas.WithLabelValues(namespace, name).Set(float64(availableReplicas))
	m.updatedReplicas.WithLabelValues(namespace, name).Set(float64(updatedReplicas))
	m.updatedReadyReplicas.WithLabelValues(namespace, name).Set(float64(updatedReadyReplicas))
}

// collectOperatorVersion collects the image version of gamestatefulset operator pods
func (m *metrics) collectOperatorVersion(imageVersion, CRDVersion, gitVersion, gitCommit, buildDate string) {
	m.operatorVersion.WithLabelValues("GameStatefulSet", imageVersion, CRDVersion,
		gitVersion, gitCommit, buildDate).Set(float64(1))
}
