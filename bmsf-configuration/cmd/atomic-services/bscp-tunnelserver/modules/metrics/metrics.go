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

// Collector is metrics collector for tunnelserver.
type Collector struct {
	endpoint string
	path     string

	// request total counter.
	reqCounter *prometheus.CounterVec

	// response time summary.
	respTimeSummary *prometheus.SummaryVec

	// session num gauge.
	sessionGauge prometheus.Gauge

	// gse plat service re-init event counter.
	gsePlatReInitCounter prometheus.Counter

	// gse plat service re-init error counter.
	gsePlatReInitErrCounter prometheus.Counter

	// gse plat service message channel size.
	gsePlatMessageChanGauge *prometheus.GaugeVec

	// gse plat service message counter.
	gsePlatMessageCounter *prometheus.CounterVec

	// gse plat service message processed counter.
	gsePlatProcessedCounter *prometheus.CounterVec

	// gse plat service fusing event counter.
	gsePlatFusingCounter *prometheus.CounterVec

	// recv and decode protocol cost time summary.
	processProtoTimeSummary *prometheus.SummaryVec

	// process cost time summary.
	processTimeSummary *prometheus.SummaryVec
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
			Subsystem: "tunnelserver",
			Name:      "request_total",
			Help:      "request total counter.",
		},
		[]string{"rpc", "errcode"},
	)

	c.respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)

	c.sessionGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "session_num",
			Help:      "session num gauge.",
		},
	)

	c.gsePlatReInitCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_reinit_total",
			Help:      "gse plat service re-init counter.",
		},
	)

	c.gsePlatReInitErrCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_reinit_error",
			Help:      "gse plat service re-init counter.",
		},
	)

	c.gsePlatMessageChanGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_message_chan_runtime_size",
			Help:      "gse plat service message channel runtime gauge.",
		},
		[]string{"id"},
	)

	c.gsePlatMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_message_total",
			Help:      "gse plat service message total.",
		},
		[]string{"id"},
	)

	c.gsePlatProcessedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_message_processed_total",
			Help:      "gse plat service message processed total.",
		},
		[]string{"id"},
	)

	c.gsePlatFusingCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "gse_platservice_message_fuse_total",
			Help:      "gse plat service message fuse total.",
		},
		[]string{"id"},
	)

	c.processProtoTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "process_proto_time",
			Help:      "process protocol time(ms) summary.",
		},
		[]string{"id"},
	)

	c.processTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "tunnelserver",
			Name:      "process_time",
			Help:      "process message time(ms) summary.",
		},
		[]string{"id"},
	)
}

// StatRequest stats metrics data for rpc requests.
func (c *Collector) StatRequest(rpc string, errcode pbcommon.ErrCode, inTime, outTime time.Time) int64 {
	c.reqCounter.With(prometheus.Labels{"rpc": rpc, "errcode": common.ToStr(int(errcode))}).Inc()

	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// StatProcessProtocol stats metrics data for processing protocol.
func (c *Collector) StatProcessProtocol(id string, inTime, outTime time.Time) int64 {
	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.processProtoTimeSummary.With(prometheus.Labels{"id": id}).Observe(float64(cost))
	return cost
}

// StatProcess stats metrics data for processing message.
func (c *Collector) StatProcess(id string, inTime, outTime time.Time) int64 {
	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.processTimeSummary.With(prometheus.Labels{"id": id}).Observe(float64(cost))
	return cost
}

// StatSessionNum stats metrics data for runtime session num.
func (c *Collector) StatSessionNum(n int64) {
	c.sessionGauge.Set(float64(n))
}

// StatGSEPlatServiceReInit stats gse plat service re-init counter.
func (c *Collector) StatGSEPlatServiceReInit(isSucc bool) {
	c.gsePlatReInitCounter.Inc()

	if !isSucc {
		c.gsePlatReInitErrCounter.Inc()
	}
}

// StatGSEPlatMessageChanRuntime stats metrics data for gse plat service message channel runtime.
func (c *Collector) StatGSEPlatMessageChanRuntime(id string, n int64) {
	c.gsePlatMessageChanGauge.With(prometheus.Labels{"id": id}).Set(float64(n))
}

// StatGSEPlatMessageCount stats gse plat service message total num.
func (c *Collector) StatGSEPlatMessageCount(id string) {
	c.gsePlatMessageCounter.With(prometheus.Labels{"id": id}).Inc()
}

// StatGSEPlatMessageProcessedCount stats gse plat service message processed total num.
func (c *Collector) StatGSEPlatMessageProcessedCount(id string) {
	c.gsePlatProcessedCounter.With(prometheus.Labels{"id": id}).Inc()
}

// StatGSEPlatMessageFuseCount stats gse plat service message fuse total num.
func (c *Collector) StatGSEPlatMessageFuseCount(id string) {
	c.gsePlatFusingCounter.With(prometheus.Labels{"id": id}).Inc()
}

// Setup setups the new Collector.
func (c *Collector) Setup() error {
	c.setup()

	prometheus.MustRegister(c.reqCounter, c.respTimeSummary, c.sessionGauge,
		c.gsePlatReInitCounter, c.gsePlatReInitErrCounter,
		c.gsePlatMessageChanGauge, c.gsePlatMessageCounter,
		c.gsePlatProcessedCounter, c.gsePlatFusingCounter,
		c.processProtoTimeSummary, c.processTimeSummary)

	http.Handle(c.path, promhttp.Handler())
	return http.ListenAndServe(c.endpoint, nil)
}
