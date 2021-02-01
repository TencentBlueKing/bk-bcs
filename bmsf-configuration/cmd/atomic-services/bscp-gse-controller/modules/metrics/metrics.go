/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/common"
)

// Collector is metrics collector for gsecontroller.
type Collector struct {
	endpoint string
	path     string

	// request total counter.
	reqCounter *prometheus.CounterVec

	// response time summary.
	respTimeSummary *prometheus.SummaryVec
}

// NewCollector returns a new Collector.
func NewCollector(endpoint, path string) *Collector {
	if len(path) == 0 {
		path = "/metrics"
	}
	return &Collector{endpoint: endpoint, path: path}
}

func (c *Collector) setup() {
	c.reqCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "gsecontroller",
			Name:      "request_total",
			Help:      "request total counter.",
		},
		[]string{"rpc", "errcode"},
	)

	c.respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "gsecontroller",
			Name:      "response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)
}

// StatRequest stats metrics data for rpc requests.
func (c *Collector) StatRequest(rpc string, errcode pbcommon.ErrCode, inTime, outTime time.Time) int64 {
	c.reqCounter.With(prometheus.Labels{"rpc": rpc, "errcode": common.ToStr(int(errcode))}).Inc()

	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// Setup setups the new Collector.
func (c *Collector) Setup() error {
	c.setup()
	prometheus.MustRegister(c.reqCounter, c.respTimeSummary)

	http.Handle(c.path, promhttp.Handler())
	return http.ListenAndServe(c.endpoint, nil)
}
