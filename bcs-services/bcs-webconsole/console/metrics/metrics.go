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
	// http 请求总量
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_requests_total",
		Help:      "Counter of requests to bcs-webconsole.",
	}, []string{"handler", "method", "status", "code"})

	// http 请求耗时, 包含页面返回, API请求, WebSocket(去掉pod_create耗时)
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request took.",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 5, 10, 30, 60},
	}, []string{"handler", "method", "status", "code"})

	// 创建pod数量
	podCreateTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_total",
		Help:      "Counter of pod create to bcs-webconsole.",
	}, []string{"namespace", "name", "status"})

	// 创建pod延迟指标
	podCreateDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_create_duration_seconds",
		Help:      "create duration(seconds) of pod",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"namespace", "name", "status"})

	// 删除pod数量
	podDeleteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_total",
		Help:      "Counter of pod delete total to bcs-webconsole.",
	}, []string{"namespace", "name", "status", "reason"})

	// 删除pod延迟指标
	podDeleteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_delete_duration_seconds",
		Help:      "delete duration(seconds) of pod",
		Buckets:   []float64{0.1, 1, 5, 10, 30, 60},
	}, []string{"namespace", "name", "status", "reason"})

	// ws连接
	wsConnectionTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_total",
		Help:      "The total number of websocket connection",
	}, []string{"username", "tg_cluster_id", "tg_namespace", "tg_pod_name", "tg_container_name"})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(podCreateTotal)
	prometheus.MustRegister(podDeleteTotal)
	prometheus.MustRegister(podCreateDuration)
	prometheus.MustRegister(podDeleteDuration)
	prometheus.MustRegister(wsConnectionTotal)
}

// RegisterWsConnection
func RegisterWsConnection(loader func() float64) {
	wsConnectionOnlineCount := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_online_count",
		Help:      "The number of websocket current connections",
	}, loader)
	prometheus.MustRegister(wsConnectionOnlineCount)
}

func HandlerFunc() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// collectHTTPRequestMetric http metrics 处理
func collectHTTPRequestMetric(handler, method, status, code string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(handler, method, status, code).Inc()
	httpRequestDuration.WithLabelValues(handler, method, status, code).Observe(duration.Seconds())
}

// CollectPodDeleteDurations collect below metrics:
// 1.the delete duration(seconds) of each pod
// 2.the max delete duration(seconds) of pod
// 3.the min delete duration(seconds) of pod
func CollectPodDeleteDurations(namespace, name, status, podName string, started time.Time) {
	duration := time.Since(started).Seconds()

	podDeleteTotal.WithLabelValues(namespace, name, status).Inc()
	podDeleteDuration.WithLabelValues(namespace, name, status).Observe(duration)

}

// PodCollect Pod拉起耗时统计
func PodCollect(namespace, podName, status string, duration time.Duration) {
	podCreateTotal.WithLabelValues(namespace, podName, status).Inc()
	podCreateDuration.WithLabelValues(namespace, podName, status).Observe(duration.Seconds())
}

// CollectWsConnection Websocket 长链接统计
func CollectWsConnection(username, targetClusterId, namespace, podName, containerName string) {
	wsConnectionTotal.WithLabelValues(username, targetClusterId, namespace, podName, containerName).Inc()
}
