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

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	podCreateDurationMaxVal = map[string]float64{}
	podCreateDurationMinVal = map[string]float64{}
	podDeleteDurationMaxVal = map[string]float64{}
	podDeleteDurationMinVal = map[string]float64{}
)

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "api_request_total_num",
		Help:      "Counter of requests to bcs-web-console.",
	}, []string{"handler", "method", "status"})

	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "api_request_latency_time",
		Help:      "Histogram of the time (in seconds) each request took.",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"handler", "method", "status"})

	// 创建pod数量
	podCreateTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_total_num",
		Help:      "Counter of pod create to bcs-web-console.",
	}, []string{"namespace", "name", "status"})

	// 删除pod数量
	podDeleteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_total_num",
		Help:      "Counter of pod delete total to bcs-web-console.",
	}, []string{"namespace", "name", "status"})

	// 创建pod延迟指标
	podCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds",
		Help:      "create duration(seconds) of pod",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"namespace", "name", "status"})

	// 删除pod延迟指标
	podDeleteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds",
		Help:      "delete duration(seconds) of pod",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"namespace", "name", "status"})

	// 创建pod最大时间
	podCreateDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds_max",
		Help:      "the max create duration(seconds) of pod",
	}, []string{"namespace", "name", "status"})

	// 创建pod最小时间
	podCreateDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds_min",
		Help:      "the min create duration(seconds) of pod",
	}, []string{"namespace", "name", "status"})

	// 删除pod最大时间
	podDeleteDurationMax = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds_max",
		Help:      "the max delete duration(seconds) of pod",
	}, []string{"namespace", "name", "status"})

	// 删除pod最小时间
	podDeleteDurationMin = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds_min",
		Help:      "the min delete duration(seconds) of pod",
	}, []string{"namespace", "name", "status"})

	// ws连接
	wsConnectionTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_total_num",
		Help:      "The total number of websocket connection",
	}, []string{"namespace", "name"})

	// 断开ws连接
	wsCloseTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_close_total_num",
		Help:      "The total number of websocket disconnections",
	}, []string{"namespace", "name"})

	// ws连接延迟
	wsConnectionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_duration_seconds",
		Help:      "Counter of websocket connection to bcs-web-console.",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"namespace", "name"})
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(podCreateTotal)
	prometheus.MustRegister(podDeleteTotal)
	prometheus.MustRegister(podCreateDuration)
	prometheus.MustRegister(podDeleteDuration)
	prometheus.MustRegister(podCreateDurationMax)
	prometheus.MustRegister(podCreateDurationMin)
	prometheus.MustRegister(podDeleteDurationMax)
	prometheus.MustRegister(podDeleteDurationMin)
	prometheus.MustRegister(wsConnectionTotal)
	prometheus.MustRegister(wsCloseTotal)
	prometheus.MustRegister(wsConnectionDuration)
}

func HandlerFunc() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func ReportAPIRequestMetric(handler, method, status string, started time.Time) {
	requestsTotalLib.WithLabelValues(handler, method, status).Inc()
	requestLatencyLib.WithLabelValues(handler, method, status).Observe(time.Since(started).Seconds())
}

// CollectPodCreateDurations collect below metrics:
// 1.the create duration(seconds) of each pod
// 2.the max create duration(seconds) of pod
// 3.the min create duration(seconds) of pod
func CollectPodCreateDurations(namespace, podName, status string, started time.Time) {
	duration := time.Since(started).Seconds()

	podCreateTotal.WithLabelValues(namespace, podName, status).Inc()
	podCreateDuration.WithLabelValues(namespace, podName, status).Observe(duration)
	if duration > podCreateDurationMaxVal[podName] {
		podCreateDurationMaxVal[podName] = duration
		podCreateDurationMax.WithLabelValues(namespace, podName, status).Set(duration)
	}

	if podCreateDurationMinVal[podName] == float64(0) {
		podCreateDurationMinVal[podName] = duration
		podCreateDurationMin.WithLabelValues(namespace, podName, status).Set(duration)
	} else if duration < podCreateDurationMinVal[podName] {
		podCreateDurationMinVal[podName] = duration
		podCreateDurationMin.WithLabelValues(namespace, podName, status).Set(duration)
	}
}

// CollectPodDeleteDurations collect below metrics:
// 1.the delete duration(seconds) of each pod
// 2.the max delete duration(seconds) of pod
// 3.the min delete duration(seconds) of pod
func CollectPodDeleteDurations(namespace, name, status, podName string, started time.Time) {
	duration := time.Since(started).Seconds()

	podDeleteTotal.WithLabelValues(namespace, name, status).Inc()
	podDeleteDuration.WithLabelValues(namespace, name, status).Observe(duration)

	if duration > podDeleteDurationMaxVal[podName] {
		podDeleteDurationMaxVal[podName] = duration
		podDeleteDurationMax.WithLabelValues(namespace, name, status).Set(duration)
	}

	if podDeleteDurationMinVal[podName] == float64(0) {
		podDeleteDurationMinVal[podName] = duration
		podCreateDurationMin.WithLabelValues(namespace, name, status).Set(duration)
	} else if duration < podDeleteDurationMinVal[podName] {
		podDeleteDurationMinVal[podName] = duration
		podCreateDurationMin.WithLabelValues(namespace, name, status).Set(duration)
	}
}

func CollectWsConnection(namespace, name string, started time.Time) {
	wsConnectionTotal.WithLabelValues(namespace, name).Inc()
	wsConnectionDuration.WithLabelValues(namespace, name).Observe(time.Since(started).Seconds())
}

// CollectCloseWs 断开ws连接
func CollectCloseWs(namespace, name string) {
	wsCloseTotal.WithLabelValues(namespace, name).Inc()
}
