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

// Collector is metrics collector for connserver.
type Collector struct {
	endpoint string
	path     string

	// request total counter.
	reqCounter *prometheus.CounterVec

	// response time summary.
	respTimeSummary *prometheus.SummaryVec

	// connection num gauge.
	connGauge prometheus.Gauge

	// access node num gauge.
	nodeGauge prometheus.Gauge

	// publishing total counter.
	pubTotalCounter prometheus.Counter

	// publishing error counter.
	pubErrCounter prometheus.Counter

	// configs content cache get total counter.
	configsCacheTotalCounter prometheus.Counter

	// configs content cache hit counter.
	configsCacheHitCounter prometheus.Counter

	// publishing task num gauge.
	publishingTaskGauge prometheus.Gauge
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
			Subsystem: "connserver",
			Name:      "request_total",
			Help:      "request total counter.",
		},
		[]string{"rpc", "errcode"},
	)

	c.respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)

	c.connGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "connection_num",
			Help:      "connection num gauge.",
		},
	)

	c.nodeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "access_node_num",
			Help:      "access node num gauge.",
		},
	)

	c.pubTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "publishing_total",
			Help:      "publishing total counter.",
		},
	)

	c.pubErrCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "publishing_error",
			Help:      "publishing error counter.",
		},
	)

	c.configsCacheTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "configscache_total",
			Help:      "configs content cache get total counter.",
		},
	)

	c.configsCacheHitCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "configscache_hit",
			Help:      "configs content cache hit counter.",
		},
	)

	c.publishingTaskGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "connserver",
			Name:      "publishing_task_num",
			Help:      "publishing task num gauge.",
		},
	)
}

// StatRequest stats metrics data for rpc requests.
func (c *Collector) StatRequest(rpc string, errcode pbcommon.ErrCode, inTime, outTime time.Time) int64 {
	c.reqCounter.With(prometheus.Labels{"rpc": rpc, "errcode": common.ToStr(int(errcode))}).Inc()

	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// StatConnNum stats metrics data for runtime connection num.
func (c *Collector) StatConnNum(n int64) {
	c.connGauge.Set(float64(n))
}

// StatAccessNodeNum stats metrics data for runtime access node num.
func (c *Collector) StatAccessNodeNum(n int64) {
	c.nodeGauge.Set(float64(n))
}

// StatPublishing stats metrics data for publishing.
func (c *Collector) StatPublishing(isSucc bool) {
	c.pubTotalCounter.Inc()

	if !isSucc {
		c.pubErrCounter.Inc()
	}
}

// StatConfigsCache stats metrics data for configs cache.
func (c *Collector) StatConfigsCache(isHit bool) {
	c.configsCacheTotalCounter.Inc()

	if isHit {
		c.configsCacheHitCounter.Inc()
	}
}

// StatPublishingTask stats metrics data for publishing tasks.
func (c *Collector) StatPublishingTask(isInc bool) {
	if isInc {
		c.publishingTaskGauge.Inc()
	} else {
		c.publishingTaskGauge.Dec()
	}
}

// Setup setups the new Collector.
func (c *Collector) Setup() error {
	c.setup()

	prometheus.MustRegister(c.reqCounter, c.respTimeSummary, c.connGauge, c.nodeGauge, c.pubTotalCounter,
		c.pubErrCounter, c.configsCacheTotalCounter, c.configsCacheHitCounter, c.publishingTaskGauge)

	http.Handle(c.path, promhttp.Handler())
	return http.ListenAndServe(c.endpoint, nil)
}
