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
 */

// Package metric defines the metric info of gitops
package metric

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ManagerTunnelConnectStatus 定义 Tunnel 连接状态，0 --正常，1 --失败
	ManagerTunnelConnectStatus *prometheus.GaugeVec
	// ManagerTunnelConnectNum 定义 Tunnel 连接次数
	ManagerTunnelConnectNum *prometheus.CounterVec

	// ManagerReturnErrorNum 定义返回客户端 >= 500 状态码的数量
	ManagerReturnErrorNum *prometheus.CounterVec
	// ManagerHTTPRequestTotal 定义 HTTP 请求次数
	ManagerHTTPRequestTotal *prometheus.CounterVec
	// ManagerHTTPRequestDuration 定义 HTTP 请求时延
	ManagerHTTPRequestDuration *prometheus.HistogramVec
	// ManagerGRPCRequestTotal 定义 GRPC 请求次数
	ManagerGRPCRequestTotal *prometheus.CounterVec
	// ManagerGRPCRequestDuration 定义 GRPC 请求时延
	ManagerGRPCRequestDuration *prometheus.HistogramVec

	// ManagerArgoConnectionStatus 定义 Argo 长链状态，0 --正常，1 --失败
	ManagerArgoConnectionStatus *prometheus.GaugeVec
	// ManagerArgoConnectionNum 定义 Argo 连接次数
	ManagerArgoConnectionNum *prometheus.CounterVec

	// ManagerArgoOperateFailed 定义请求 Argo 失败的次数
	ManagerArgoOperateFailed *prometheus.CounterVec
	// ManagerArgoProxyFailed 定义 Proxy 到 Argo 失败的次数
	ManagerArgoProxyFailed *prometheus.CounterVec
	// ManagerSecretOperateFailed 定义请求 VaultPlugin 失败的次数
	ManagerSecretOperateFailed *prometheus.CounterVec
	// ManagerSecretProxyFailed 定义 Proxy 到 Secret 失败的次数
	ManagerSecretProxyFailed *prometheus.CounterVec

	once sync.Once
)

// nolint
func init() {
	once.Do(func() {
		ManagerTunnelConnectStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gitops_manager_tunnel_connect_status",
			Help: "defines the connection status of tunnel websocket",
		}, []string{})
		ManagerTunnelConnectNum = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_tunnel_connect_num",
			Help: "defines the connect num of tunnel websocket",
		}, []string{})
		ManagerReturnErrorNum = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_return_error_num",
			Help: "defines the return error more than 500 status code",
		}, []string{})
		ManagerHTTPRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_http_request_total",
			Help: "defines the http request total number",
		}, []string{})
		ManagerHTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gitops_manager_http_request_duration",
			Help:    "defines the http request duration",
			Buckets: []float64{0.1, 0.3, 1.2, 5, 10},
		}, []string{})
		ManagerGRPCRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_grpc_request_total",
			Help: "defines the grpc request total number",
		}, []string{})
		ManagerGRPCRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gitops_manager_grpc_request_duration",
			Help:    "defines the grpc request duration",
			Buckets: []float64{0.1, 0.3, 1.2, 5, 10},
		}, []string{})
		ManagerArgoConnectionStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gitops_manager_argo_connection_status",
			Help: "defines the argo connection status",
		}, []string{})
		ManagerArgoConnectionNum = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_argo_connection_num",
			Help: "defines the argo connection number",
		}, []string{})
		ManagerArgoOperateFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_argo_operate_failed",
			Help: "defines the argo operation failed number",
		}, []string{"operation"})
		ManagerArgoProxyFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_argo_proxy_failed",
			Help: "defines the argo proxy failed number",
		}, []string{})
		ManagerSecretOperateFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_secret_operate_failed",
			Help: "defines the secret operator failed",
		}, []string{})
		ManagerSecretProxyFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gitops_manager_secret_proxy_failed",
			Help: "defines the secret proxy failed",
		}, []string{})

		ManagerTunnelConnectStatus.WithLabelValues().Set(0)
		ManagerTunnelConnectNum.WithLabelValues().Add(0)
		ManagerReturnErrorNum.WithLabelValues().Add(0)
		ManagerHTTPRequestTotal.WithLabelValues().Add(0)
		ManagerGRPCRequestTotal.WithLabelValues().Add(0)
		ManagerArgoConnectionStatus.WithLabelValues().Set(0)
		ManagerArgoConnectionNum.WithLabelValues().Add(0)
		ManagerArgoOperateFailed.WithLabelValues("FAKE").Add(0)
		ManagerArgoProxyFailed.WithLabelValues().Add(0)
		ManagerSecretOperateFailed.WithLabelValues().Add(0)
		ManagerSecretProxyFailed.WithLabelValues().Add(0)

		prometheus.MustRegister(ManagerTunnelConnectStatus)
		prometheus.MustRegister(ManagerTunnelConnectNum)
		prometheus.MustRegister(ManagerReturnErrorNum)
		prometheus.MustRegister(ManagerHTTPRequestTotal)
		prometheus.MustRegister(ManagerHTTPRequestDuration)
		prometheus.MustRegister(ManagerGRPCRequestTotal)
		prometheus.MustRegister(ManagerGRPCRequestDuration)
		prometheus.MustRegister(ManagerArgoConnectionStatus)
		prometheus.MustRegister(ManagerArgoConnectionNum)
		prometheus.MustRegister(ManagerArgoOperateFailed)
		prometheus.MustRegister(ManagerArgoProxyFailed)
		prometheus.MustRegister(ManagerSecretOperateFailed)
		prometheus.MustRegister(ManagerSecretProxyFailed)
	})
}
