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

package metric

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// RequestResultSuccess request successfully
	RequestResultSuccess = "ok"
	// RequestResultFailed request return error
	RequestResultFailed = "failed"
	// RequestResultPartialFailed partial failed result
	RequestResultPartialFailed = "partialfailed"
)

// declare metrics
var (
	cliReqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_network",
			Subsystem: "cloudnetcontroller",
			Name:      "cli_request_total",
			Help:      "total request counter as client",
		},
		[]string{"module", "rpc", "errcode", "result"},
	)
	cliRespSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bcs_network",
			Subsystem: "cloudnetcontroller",
			Name:      "cli_response_time",
			Help:      "response time(ms) of other module summary.",
		},
		[]string{"module", "rpc"},
	)
)

func init() {
	metrics.Registry.MustRegister(cliReqCounter)
	metrics.Registry.MustRegister(cliRespSummary)
}

// StatClientRequest report client request metrics
func StatClientRequest(module, rpc string, respCode int, result string, inTime, outTime time.Time) {
	cliReqCounter.With(prometheus.Labels{
		"module":  module,
		"rpc":     rpc,
		"errcode": strconv.Itoa(respCode),
		"result":  result,
	}).Inc()

	cost := toMSTimestamp(outTime) - toMSTimestamp(inTime)
	cliRespSummary.With(prometheus.Labels{"module": module, "rpc": rpc}).Observe(float64(cost))
}

// toMSTimestamp converts time.Time to millisecond timestamp.
func toMSTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}
