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

package metrics

import (
	gprm "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version"
)

// globalRegister is a global register which is used to collect metrics we need.
// it will be initialized when process is up for safe usage.
// and then be revised later when service is initialized.
var globalRegister prometheus.Registerer

func init() {
	// set default global register
	globalRegister = prometheus.DefaultRegisterer
}

const (
	// Namespace is the root namespace of the bscp metric
	Namespace = "bscp"

	// CSCacheSubSys defines cache service's bedis cache metric sub system.
	CSCacheSubSys = "remote_cache"

	// CSEventSubSys defines cache service's event metric sub system.
	CSEventSubSys = "event"

	// BedisCmdSubSys defines all the bedis command related sub system.
	BedisCmdSubSys = "bedis_cmd"

	// OrmCmdSubSys defines all the sharding database command related sub system.
	OrmCmdSubSys = "orm"

	// ResourceLockSubSys defines all the runtime lock related sub system.
	ResourceLockSubSys = "resource_lock"

	// FSLocalCacheSubSys defines feed server's local cache metric sub system.
	FSLocalCacheSubSys = "local_cache"

	// FSObserver defines feed server's observer sub system
	FSObserver = "observer"

	// FSEventc defines feed server's eventc sub system.
	FSEventc = "eventc"

	// FSConfigConsume defines feed server's config consume sub system.
	FSConfigConsume = "config_consume"

	// RestfulSubSys defines rest server's sub system
	RestfulSubSys = "restful"
)

// labels
const (
	LabelProcessName = "process_name"
	LabelHost        = "host"
)

var (
	// GrpcBuckets defines the grpc server's metric buckets.
	GrpcBuckets = gprm.WithHistogramBuckets([]float64{0.001, 0.003, 0.005, 0.007, 0.01, 0.015, 0.02, 0.025, 0.03,
		0.04, 0.05, 0.075, 0.1, 0.2, 0.3, 0.4, 0.5, 1, 2.5, 5, 10})

	// http 请求总量
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Counter of HTTP requests to bscp",
	}, []string{"handler", "method", "code"})

	// http 请求耗时
	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Histogram of latencies for HTTP requests to bscp.",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1},
	}, []string{"handler", "method", "code"})

	// BSCPServerHandledTotal 请求数
	BSCPServerHandledTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bscp_server_handled_total",
		Help: "Total number of platform user operations",
	}, []string{"service", "handler", "code", "biz", "username"})
)

// InitMetrics init metrics registerer and http handler
func InitMetrics(endpoint string) {
	registry := prometheus.NewRegistry()

	processName := string(cc.ServiceName())
	label := prometheus.Labels{LabelProcessName: processName, LabelHost: endpoint}

	register := prometheus.WrapRegistererWith(label, registry)

	// set up global register
	globalRegister = register

	register.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	register.MustRegister(collectors.NewGoCollector())

	// metric current service version.
	versionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   Namespace,
			Subsystem:   "version",
			Name:        "info",
			Help:        "The version info of the current service",
			ConstLabels: prometheus.Labels{},
		},
		[]string{"version", "build_time", "git_hash"},
	)
	register.MustRegister(versionGauge)
	versionGauge.With(prometheus.Labels{
		"version":    version.VERSION,
		"build_time": version.BUILDTIME,
		"git_hash":   version.GITHASH,
	}).Set(1)

	// set up metrics http handler
	httpHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
