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

	// db threads connected num gauge.
	dbThreadsConnectedGauge prometheus.Gauge

	// db questions counter.
	dbQuestionsGauge prometheus.Gauge

	// business num counter.
	bizNumGauge prometheus.Gauge

	// business app num counter.
	appNumGauge *prometheus.GaugeVec

	// business config num counter.
	configNumGauge *prometheus.GaugeVec

	// business app release num counter.
	releaseNumGauge *prometheus.GaugeVec

	// request total counter.
	reqCounter *prometheus.CounterVec

	// response time summary.
	respTimeSummary *prometheus.SummaryVec

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
	c.dbThreadsConnectedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "db_threads_connected_num",
			Help:      "db threads connected num gauge.",
		},
	)

	c.dbQuestionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "db_questions_total",
			Help:      "db questions total counter.",
		},
	)

	c.bizNumGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "business_num",
			Help:      "internal business num.",
		},
	)

	c.appNumGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "business_app_num",
			Help:      "business app num.",
		},
		[]string{"biz"},
	)

	c.configNumGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "business_config_num",
			Help:      "business config num.",
		},
		[]string{"biz", "app"},
	)

	c.releaseNumGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "bscp",
			Subsystem: "datamanager",
			Name:      "business_release_num",
			Help:      "business release num.",
		},
		[]string{"biz", "app", "interval"},
	)

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

// StatDBStatus stats metrics data for database status.
func (c *Collector) StatDBStatus(questionsTotal, threadsConnectedNum int64) {
	c.dbQuestionsGauge.Set(float64(questionsTotal))
	c.dbThreadsConnectedGauge.Set(float64(threadsConnectedNum))
}

// StatBusiness stats metrics data for internal business.
// bizNum: business num.
// appStat:     biz_id -> app count.
// configStat:  biz_id -> map[app_id -> config count].
// releaseStat: biz_id -> map[interval -> map[app_id -> release count]].
func (c *Collector) StatBusiness(bizNum int64,
	appStat map[string]int64,
	configStat map[string]map[string]int64,
	releaseStat map[string]map[string]map[string]int64) {

	// stat business.
	c.bizNumGauge.Set(float64(bizNum))

	// stat business app.
	for bizID, appCount := range appStat {
		c.appNumGauge.With(prometheus.Labels{"biz": bizID}).Set(float64(appCount))
	}

	// stat app config.
	for bizID, stat := range configStat {
		for appID, configCount := range stat {
			c.configNumGauge.With(prometheus.Labels{"biz": bizID, "app": appID}).Set(float64(configCount))
		}
	}

	// stat business app release.
	for bizID, stat := range releaseStat {

		// release stat for each app.
		for interval, subStat := range stat {

			// stat for each interval.
			for appID, releaseCount := range subStat {
				c.releaseNumGauge.With(prometheus.Labels{
					"biz":      bizID,
					"app":      appID,
					"interval": interval,
				}).Set(float64(releaseCount))
			}
		}
	}
}

// StatRequest stats metrics data for rpc requests.
func (c *Collector) StatRequest(rpc string, errcode pbcommon.ErrCode, inTime, outTime time.Time) int64 {
	c.reqCounter.With(prometheus.Labels{"rpc": rpc, "errcode": common.ToStr(int(errcode))}).Inc()

	cost := common.ToMSTimestamp(outTime) - common.ToMSTimestamp(inTime)
	c.respTimeSummary.With(prometheus.Labels{"rpc": rpc}).Observe(float64(cost))

	return cost
}

// StatAppInstanceRelease stats metrics data for release effect num.
func (c *Collector) StatAppInstanceRelease(isSucc bool) {
	c.appInstanceReleaseTotalCounter.Inc()

	if !isSucc {
		c.appInstanceReleaseErrCounter.Inc()
	}
}

// Setup setups the new Collector.
func (c *Collector) Setup() error {
	c.setup()
	prometheus.MustRegister(c.dbThreadsConnectedGauge, c.dbQuestionsGauge, c.bizNumGauge, c.appNumGauge,
		c.configNumGauge, c.releaseNumGauge, c.reqCounter, c.respTimeSummary,
		c.appInstanceReleaseTotalCounter, c.appInstanceReleaseErrCounter)

	http.Handle(c.path, promhttp.Handler())
	return http.ListenAndServe(c.endpoint, nil)
}
