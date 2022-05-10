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
		Help:      "Counter of HTTP requests to bcs-webconsole.",
	}, []string{"handler", "method", "code"})

	// http 请求耗时, 包含页面返回, API请求, WebSocket(去掉pod_create耗时)
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "http_request_duration_seconds",
		Help:      "Histogram of latencies for HTTP requests to bcs-webconsole.",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60},
	}, []string{"handler", "method", "code"})

	// 创建/等待 pod Ready 数量
	podReadyTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_ready_total",
		Help:      "Counter of create/wait pod ready.",
	}, []string{"tg_cluster_id", "tg_namespace", "tg_pod_name", "status"})

	// 创建/等待 pod Ready 延迟指标
	podReadyDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_ready_duration_seconds",
		Help:      "Histogram of latencies for create/wait pod ready.",
		Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60},
	}, []string{"tg_cluster_id", "tg_namespace", "tg_pod_name", "status"})

	// pod 存活数量
	podCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "pod_count",
		Help:      "The number of current pod in namespace",
	}, []string{"tg_cluster_id", "tg_namespace"})

	// ws连接
	wsConnectionTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_total",
		Help:      "Counter of websocket connection.",
	}, []string{"username", "tg_cluster_id", "tg_namespace", "tg_pod_name", "tg_container_name"})

	// 连接数统计
	wsConnectionOnlineCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "ws_connection_online_count",
		Help:      "The number of websocket current connections",
	}, []string{})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(wsConnectionTotal)
	prometheus.MustRegister(wsConnectionOnlineCount)
	prometheus.MustRegister(podReadyTotal)
	prometheus.MustRegister(podReadyDuration)
	prometheus.MustRegister(podCount)
}

// PromMetricHandler prometheus handler 转换为 Gin Handler
func PromMetricHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// collectHTTPRequestMetric http metrics 处理
func collectHTTPRequestMetric(handler, method, code string, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(handler, method, code).Inc()
	httpRequestDuration.WithLabelValues(handler, method, code).Observe(duration.Seconds())
}

// CollectWsConnection Websocket 请求统计
func CollectWsConnection(username, targetClusterId, namespace, podName, containerName string) {
	wsConnectionTotal.WithLabelValues(username, targetClusterId, namespace, podName, containerName).Inc()
}

// CollectWsConnectionOnline  Websocket 长链接统计
func CollectWsConnectionOnline(value float64) {
	wsConnectionOnlineCount.WithLabelValues().Add(value)
}

// CollectPodReady Pod 拉起耗时统计
func CollectPodReady(clusterId, namespace, podName string, err error, duration time.Duration) {
	podReadyTotal.WithLabelValues(clusterId, namespace, podName, makePodStatus(err)).Inc()
	podReadyDuration.WithLabelValues(clusterId, namespace, podName, makePodStatus(err)).Observe(duration.Seconds())
}

// makePodStatus Pod 状态
func makePodStatus(err error) string {
	if err != nil {
		return ErrStatus
	}
	return SucStatus
}

// CollectPodCount
func CollectPodCount(clusterId, namespace string, count float64) {
	podCount.WithLabelValues(clusterId, namespace).Set(count)
}
