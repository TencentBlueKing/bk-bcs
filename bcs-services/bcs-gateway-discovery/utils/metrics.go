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

package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	// ErrStatus for success status
	ErrStatus = "failure"
	// SucStatus for failure status
	SucStatus = "success"
	// OtherStatus for other status
	OtherStatus = "999"
)

// APIMetricsMeta for metrics metadata
type APIMetricsMeta struct {
	System  string
	Handler string
	Method  string
	Status  string
	Started time.Time
}

const (
	// BkBcsGatewayDiscovery for bcs-gateway-discovery module metrics prefix
	BkBcsGatewayDiscovery = "bkbcs_gatewaydiscovery"
)

var (
	// bcs-gateway-discovery request action metrics
	requestTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsGatewayDiscovery,
		Name:      "api_request_total_num",
		Help:      "The total num of requests for bcs-gateway-discovery api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsGatewayDiscovery,
		Name:      "api_request_latency_time",
		Help:      "api request latency statistic for bcs-gateway-discovery api",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"system", "handler", "method", "status"})

	// bcs-gateway-discovery register metrics
	requestTotalRegister = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: BkBcsGatewayDiscovery,
		Name:      "register_request_total_num",
		Help:      "The total num of requests for bcs-gateway-discovery register",
	}, []string{"system", "handler", "status"})
	requestLatencyRegister = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: BkBcsGatewayDiscovery,
		Name:      "register_request_latency_time",
		Help:      "api request latency statistic for bcs-gateway-discovery register",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"system", "handler", "status"})

	// bcs-gateway-discovery eventChan length
	discoveryEventChanLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: BkBcsGatewayDiscovery,
		Name:      "eventchan_length",
		Help:      "bcs-gateway-discovery discovery module eventchan length metrics",
	})
)

func init() {
	// bcs-gateway-discovery request api
	prometheus.MustRegister(requestTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)

	// bcs-gateway-discovery register kong or apisix interface metrics
	prometheus.MustRegister(requestTotalRegister)
	prometheus.MustRegister(requestLatencyRegister)

	// bcs-gateway-discovery
	prometheus.MustRegister(discoveryEventChanLength)
}

// ReportDiscoveryEventChanLengthInc report eventChan inc
func ReportDiscoveryEventChanLengthInc() {
	discoveryEventChanLength.Inc()
}

// ReportDiscoveryEventChanLengthDec report eventChan dec
func ReportDiscoveryEventChanLengthDec() {
	discoveryEventChanLength.Dec()
}

//ReportBcsGatewayAPIMetrics report all api action metrics
func ReportBcsGatewayAPIMetrics(metricData APIMetricsMeta) {
	requestTotalAPI.WithLabelValues(metricData.System, metricData.Handler, metricData.Method, metricData.Status).Inc()
	requestLatencyAPI.WithLabelValues(metricData.System, metricData.Handler, metricData.Method, metricData.Status).
		Observe(time.Since(metricData.Started).Seconds())
}

//ReportBcsGatewayRegistryMetrics report all register interface metrics
func ReportBcsGatewayRegistryMetrics(metricData APIMetricsMeta) {
	requestTotalRegister.WithLabelValues(metricData.System, metricData.Handler, metricData.Status).Inc()
	requestLatencyRegister.WithLabelValues(metricData.System, metricData.Handler, metricData.Status).
		Observe(time.Since(metricData.Started).Seconds())
}
