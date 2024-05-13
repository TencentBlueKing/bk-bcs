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

package apimetric

import (
	"net/http"
	"strconv"
	"time"

	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector metric collector for cloud netservice
type Collector struct {
	endpoint string
	path     string

	// request total counter
	reqCounter *prometheus.CounterVec

	// response time summary
	respTimeSummary *prometheus.SummaryVec
}

// NewCollector returns a new Collector
func NewCollector(endpoint, path string) *Collector {
	if len(path) == 0 {
		path = "/metrics"
	}
	return &Collector{endpoint: endpoint, path: path}
}

// Init init metrics
func (c *Collector) Init(ns, subsystem string) {
	c.reqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "request_total",
			Help:      "request total counter",
		},
		[]string{"rpc", "errcode"},
	)

	c.respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)

	prometheus.MustRegister(c.reqCounter, c.respTimeSummary)
}

// RegisterMux register handler to mux
func (c *Collector) RegisterMux(mux *http.ServeMux) {
	mux.Handle(c.path, promhttp.Handler())
}

// StatRequest report metrics for rpc requests
func (c *Collector) StatRequest(rpc string, errcode pbcommon.ErrCode, inTime, outTime time.Time) int64 {
	c.reqCounter.With(prometheus.Labels{
		"rpc":     rpc,
		"errcode": strconv.Itoa(int(errcode)),
	}).Inc()

	cost := toMSTimestamp(outTime) - toMSTimestamp(inTime)
	c.respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// toMSTimestamp converts time.Time to millisecond timestamp.
func toMSTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}
