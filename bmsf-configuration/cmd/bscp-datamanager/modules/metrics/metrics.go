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

// Collector is metrics collector for datamanager.
type Collector struct {
	endpoint string
	path     string

	// request total counter.
	reqCounter *prometheus.CounterVec

	// response time summary.
	respTimeSummary *prometheus.SummaryVec

	// commit cache get total counter.
	commitCacheTotalCounter prometheus.Counter

	// commit cache hit counter.
	commitCacheHitCounter prometheus.Counter

	// release cache get total counter.
	releaseCacheTotalCounter prometheus.Counter

	// release cache hit counter.
	releaseCacheHitCounter prometheus.Counter

	// configs cache get total counter.
	configsCacheTotalCounter prometheus.Counter

	// configs cache hit counter.
	configsCacheHitCounter prometheus.Counter

	// create app instance release total counter.
	appInstanceReleaseTotalCounter prometheus.Counter

	// create app instance release error counter.
	appInstanceReleaseErrCounter prometheus.Counter
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
			Subsystem: "datamanager",
			Name:      "request_total",
			Help:      "request total counter.",
		},
		[]string{"rpc", "errcode"},
	)

	c.respTimeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "response_time",
			Help:      "response time(ms) summary.",
		},
		[]string{"rpc"},
	)

	c.commitCacheTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "commitcache_total",
			Help:      "commit cache get total counter.",
		},
	)

	c.commitCacheHitCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "commitcache_hit",
			Help:      "commit cache hit counter.",
		},
	)

	c.releaseCacheTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "releasecache_total",
			Help:      "release cache get total counter.",
		},
	)

	c.releaseCacheHitCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "releasecache_hit",
			Help:      "release cache hit counter.",
		},
	)

	c.configsCacheTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "configscache_total",
			Help:      "configs cache get total counter.",
		},
	)

	c.configsCacheHitCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "configscache_hit",
			Help:      "configs cache hit counter.",
		},
	)

	c.appInstanceReleaseTotalCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "instance_release_total_num",
			Help:      "create app instance release total counter.",
		},
	)

	c.appInstanceReleaseErrCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "instance_release_err_num",
			Help:      "create app instance release error counter.",
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

// StatCommitCache stats metrics data for commit cache.
func (c *Collector) StatCommitCache(isHit bool) {
	c.commitCacheTotalCounter.Inc()

	if isHit {
		c.commitCacheHitCounter.Inc()
	}
}

// StatReleaseCache stats metrics data for release cache.
func (c *Collector) StatReleaseCache(isHit bool) {
	c.releaseCacheTotalCounter.Inc()

	if isHit {
		c.releaseCacheHitCounter.Inc()
	}
}

// StatConfigsCache stats metrics data for configs cache.
func (c *Collector) StatConfigsCache(isHit bool) {
	c.configsCacheTotalCounter.Inc()

	if isHit {
		c.configsCacheHitCounter.Inc()
	}
}

func (c *Collector) StatAppInstanceRelease(isSucc bool) {
	c.appInstanceReleaseTotalCounter.Inc()

	if !isSucc {
		c.appInstanceReleaseErrCounter.Inc()
	}
}

// Setup setups the new Collector.
func (c *Collector) Setup() error {
	c.setup()
	prometheus.MustRegister(c.reqCounter, c.respTimeSummary, c.releaseCacheTotalCounter, c.releaseCacheHitCounter,
		c.configsCacheTotalCounter, c.configsCacheHitCounter, c.appInstanceReleaseTotalCounter, c.appInstanceReleaseErrCounter)

	http.Handle(c.path, promhttp.Handler())
	return http.ListenAndServe(c.endpoint, nil)
}
