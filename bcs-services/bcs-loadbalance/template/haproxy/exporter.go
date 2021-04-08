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

package haproxy

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	haproxyPidMetricDesc          *prometheus.Desc
	haproxyUpTimeSecondMetricDesc *prometheus.Desc
	haproxyULimitMetricDesc       *prometheus.Desc
	haproxyMaxSock                *prometheus.Desc
	haproxyMaxConn                *prometheus.Desc
	haproxyMaxPipes               *prometheus.Desc
	haproxyCurrentConn            *prometheus.Desc
	haproxyCurrentPipes           *prometheus.Desc
	haproxyCurrentConnRate        *prometheus.Desc
	haproxyMaxConnRate            *prometheus.Desc
	frontendMetricDescArray       []*prometheus.Desc
	backendMetricDescArray        []*prometheus.Desc
	serverMetricDescArray         []*prometheus.Desc
	keysArray                     = map[int]string{
		0:  "CurrentQueue",
		1:  "MaxQueue",
		2:  "CurrentSession",
		3:  "MaxSession",
		4:  "SessionLimit",
		5:  "SessionTotal",
		6:  "BytesIn",
		7:  "BytesOut",
		8:  "RequestDeny",
		9:  "ResponseDeny",
		10: "RequestError",
		11: "ConnectError",
		12: "ResponseError",
		13: "ConnectRetry",
		14: "ConnectRedispatch",
		15: "Status",
		16: "Weight",
		17: "Active",
		18: "Backup",
		19: "CheckFail",
		20: "CheckDown",
		21: "DownTime",
		22: "DownTimeTotal",
		23: "QueueMaxLimit",
		24: "CurrentSessionRate",
		25: "MaxSessionRate",
		26: "SessionRateLimit",
		27: "CheckStatus",
		28: "RequestRate",
		29: "RequestMaxRate",
		30: "RequestTotal",
		31: "LastSessionSecond",
		32: "ConnectRate",
		33: "ConnectMaxRate",
	}
)

func (m *Manager) initMetric() {
	haproxyPidMetricDesc = m.newStatusMetricDesc("haproxy_pid", "pid of haproxy")
	haproxyUpTimeSecondMetricDesc = m.newStatusMetricDesc("up_time_second", "second from up time to now")
	haproxyULimitMetricDesc = m.newStatusMetricDesc("ulimit_n", "ulimit n of haproxy")
	haproxyMaxSock = m.newStatusMetricDesc("max_sock", "max sockets of haproxy")
	haproxyMaxConn = m.newStatusMetricDesc("max_conn", "max connection of haproxy")
	haproxyMaxPipes = m.newStatusMetricDesc("max_pipe", "max pipes of haproxy")
	haproxyCurrentConn = m.newStatusMetricDesc("current_conn", "current connection of haproxy")
	haproxyCurrentPipes = m.newStatusMetricDesc("current_pipe", "current pipes of haproxy")
	haproxyCurrentConnRate = m.newStatusMetricDesc("current_conn_rate", "current connection rate of haproxy")
	haproxyMaxConnRate = m.newStatusMetricDesc("max_conn_rate", "max connection rate")

	frontendMetricDescArray = []*prometheus.Desc{
		m.newFrontendMetricDesc("current_queue", "current queue for frontend"),
		m.newFrontendMetricDesc("max_queue", "max queue for frontend"),
		m.newFrontendMetricDesc("current_session", "current session for frontend"),
		m.newFrontendMetricDesc("max_session", "max session for frontend"),
		m.newFrontendMetricDesc("session_limit", "session limit for frontend"),
		m.newFrontendMetricDesc("session_total", "session total for frontend"),
		m.newFrontendMetricDesc("bytes_in", "bytes_in for frontend"),
		m.newFrontendMetricDesc("bytes_out", "bytes_out for frontend"),
		m.newFrontendMetricDesc("request_deny", "request_deny for frontend"),
		m.newFrontendMetricDesc("response_deny", "response_deny for frontend"),
		m.newFrontendMetricDesc("request_error", "request_error for frontend"),
		m.newFrontendMetricDesc("connect_error", "connect_error for frontend"),
		m.newFrontendMetricDesc("response_error", "response_error for frontend"),
		m.newFrontendMetricDesc("connect_retry", "connect_retry for frontend"),
		m.newFrontendMetricDesc("connect_redispatch", "connect_redispatch for frontend"),
		m.newFrontendMetricDesc("status", "status for frontend"),
		m.newFrontendMetricDesc("weight", "weight for frontend"),
		m.newFrontendMetricDesc("active_server_num", "active server num for frontend"),
		m.newFrontendMetricDesc("backup_server_num", "backup server num for frontend"),
		m.newFrontendMetricDesc("check_fail_num", "check fail num for frontend"),
		m.newFrontendMetricDesc("check_down_num", "check down num for frontend"),
		m.newFrontendMetricDesc("down_time_second", "down time second for frontend"),
		m.newFrontendMetricDesc("down_time_total", "down time total for frontend"),
		m.newFrontendMetricDesc("queue_limit", "queue limit for frontend"),
		m.newFrontendMetricDesc("current_session_rate", "current session_rate for frontend"),
		m.newFrontendMetricDesc("max_session_rate", "max session rate for frontend"),
		m.newFrontendMetricDesc("session_rate_limit", "session rate limit for frontend"),
		m.newFrontendMetricDesc("check_status", "check status for frontend"),
		m.newFrontendMetricDesc("request_rate", "request rate for frontend"),
		m.newFrontendMetricDesc("request_max_rate", "request max rate for frontend"),
		m.newFrontendMetricDesc("request_total", "request total for frontend"),
		m.newFrontendMetricDesc("last_session_second", "last session second for frontend"),
		m.newFrontendMetricDesc("connect_rate", "connect rate for frontend"),
		m.newFrontendMetricDesc("connect_max_rate", "connect max rate for frontend"),
	}

	backendMetricDescArray = []*prometheus.Desc{
		m.newBackendMetricDesc("current_queue", "current queue for backend"),
		m.newBackendMetricDesc("max_queue", "max queue for backend"),
		m.newBackendMetricDesc("current_session", "current session for backend"),
		m.newBackendMetricDesc("max_session", "max session for backend"),
		m.newBackendMetricDesc("session_limit", "session limit for backend"),
		m.newBackendMetricDesc("session_total", "session total for backend"),
		m.newBackendMetricDesc("bytes_in", "bytes_in for backend"),
		m.newBackendMetricDesc("bytes_out", "bytes_out for backend"),
		m.newBackendMetricDesc("request_deny", "request_deny for backend"),
		m.newBackendMetricDesc("response_deny", "response_deny for backend"),
		m.newBackendMetricDesc("request_error", "request_error for backend"),
		m.newBackendMetricDesc("connect_error", "connect_error for backend"),
		m.newBackendMetricDesc("response_error", "response_error for backend"),
		m.newBackendMetricDesc("connect_retry", "connect_retry for backend"),
		m.newBackendMetricDesc("connect_redispatch", "connect_redispatch for backend"),
		m.newBackendMetricDesc("status", "status for backend"),
		m.newBackendMetricDesc("weight", "weight for backend"),
		m.newBackendMetricDesc("active_server_num", "active server num for backend"),
		m.newBackendMetricDesc("backup_server_num", "backup server num for backend"),
		m.newBackendMetricDesc("check_fail_num", "check fail num for backend"),
		m.newBackendMetricDesc("check_down_num", "check down num for backend"),
		m.newBackendMetricDesc("down_time_second", "down time second for backend"),
		m.newBackendMetricDesc("down_time_total", "down time total for backend"),
		m.newBackendMetricDesc("queue_limit", "queue limit for backend"),
		m.newBackendMetricDesc("current_session_rate", "current session_rate for backend"),
		m.newBackendMetricDesc("max_session_rate", "max session rate for backend"),
		m.newBackendMetricDesc("session_rate_limit", "session rate limit for backend"),
		m.newBackendMetricDesc("check_status", "check status for backend"),
		m.newBackendMetricDesc("request_rate", "request rate for backend"),
		m.newBackendMetricDesc("request_max_rate", "request max rate for backend"),
		m.newBackendMetricDesc("request_total", "request total for backend"),
		m.newBackendMetricDesc("last_session_second", "last session second for backend"),
		m.newBackendMetricDesc("connect_rate", "connect rate for backend"),
		m.newBackendMetricDesc("connect_max_rate", "connect max rate for backend"),
	}

	serverMetricDescArray = []*prometheus.Desc{
		m.newServerMetricDesc("current_queue", "current queue for server"),
		m.newServerMetricDesc("max_queue", "max queue for server"),
		m.newServerMetricDesc("current_session", "current session for server"),
		m.newServerMetricDesc("max_session", "max session for server"),
		m.newServerMetricDesc("session_limit", "session limit for server"),
		m.newServerMetricDesc("session_total", "session total for server"),
		m.newServerMetricDesc("bytes_in", "bytes_in for server"),
		m.newServerMetricDesc("bytes_out", "bytes_out for server"),
		m.newServerMetricDesc("request_deny", "request_deny for server"),
		m.newServerMetricDesc("response_deny", "response_deny for server"),
		m.newServerMetricDesc("request_error", "request_error for server"),
		m.newServerMetricDesc("connect_error", "connect_error for server"),
		m.newServerMetricDesc("response_error", "response_error for server"),
		m.newServerMetricDesc("connect_retry", "connect_retry for server"),
		m.newServerMetricDesc("connect_redispatch", "connect_redispatch for server"),
		m.newServerMetricDesc("status", "status for server"),
		m.newServerMetricDesc("weight", "weight for server"),
		m.newServerMetricDesc("active_server_num", "active server num for server"),
		m.newServerMetricDesc("backup_server_num", "backup server num for server"),
		m.newServerMetricDesc("check_fail_num", "check fail num for server"),
		m.newServerMetricDesc("check_down_num", "check down num for server"),
		m.newServerMetricDesc("down_time_second", "down time second for server"),
		m.newServerMetricDesc("down_time_total", "down time total for server"),
		m.newServerMetricDesc("queue_limit", "queue limit for server"),
		m.newServerMetricDesc("current_session_rate", "current session_rate for server"),
		m.newServerMetricDesc("max_session_rate", "max session rate for server"),
		m.newServerMetricDesc("session_rate_limit", "session rate limit for server"),
		m.newServerMetricDesc("check_status", "check status for server"),
		m.newServerMetricDesc("request_rate", "request rate for server"),
		m.newServerMetricDesc("request_max_rate", "request max rate for server"),
		m.newServerMetricDesc("request_total", "request total for server"),
		m.newServerMetricDesc("last_session_second", "last session second for server"),
		m.newServerMetricDesc("connect_rate", "connect rate for server"),
		m.newServerMetricDesc("connect_max_rate", "connect max rate for server"),
	}
}

func getValue(s *Server, key string) float64 {
	r := reflect.ValueOf(s)
	f := reflect.Indirect(r).FieldByName(key)
	if f.Kind() != reflect.Int64 {
		blog.Errorf("key %s is kind %d, not Int64, cannot be parsed", key, f.Kind())
		return 0
	}
	return float64(f.Int())
}

func (m *Manager) newStatusMetricDesc(metricName, metricDoc string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("loadbalance", "haproxy", metricName),
		metricDoc, nil,
		prometheus.Labels{
			types.MetricLabelLoadbalance: m.LoadbalanceName,
		},
	)
}

func (m *Manager) newFrontendMetricDesc(metricName, metricDoc string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("loadbalance", "haproxy", "frontend_"+metricName),
		metricDoc, []string{types.MetricLabelFrontent},
		prometheus.Labels{
			types.MetricLabelLoadbalance: m.LoadbalanceName,
		},
	)
}

func (m *Manager) newBackendMetricDesc(metricName, metricDoc string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("loadbalance", "haproxy", "backend_"+metricName),
		metricDoc, []string{types.MetricLabelBackend},
		prometheus.Labels{
			types.MetricLabelLoadbalance: m.LoadbalanceName,
		},
	)
}

func (m *Manager) newServerMetricDesc(metricName, metricDoc string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("loadbalance", "haproxy", "server_"+metricName),
		metricDoc, []string{types.MetricLabelServer, types.MetricLabelBackend, types.MetricLabelServerAddress},
		prometheus.Labels{
			types.MetricLabelLoadbalance: m.LoadbalanceName,
		},
	)
}

func convertStatus(status string) float64 {
	switch status {
	case "DOWN", "DOWN 1/2":
		return 0
	case "UP", "UP 1/3", "UP 2/3", "OPEN":
		return 1
	case "MAINT":
		return 2
	case "no check":
		return 3
	case "NOLB":
		return 4
	}
	return 0
}

func convertCheckStatus(status string) float64 {
	switch status {
	case "L4OK", "L6OK", "L7OK":
		return 1
	}
	return 0
}

// Describe implements prometheus exporter Describe interface
func (m *Manager) Describe(ch chan<- *prometheus.Desc) {

	ch <- haproxyPidMetricDesc
	ch <- haproxyUpTimeSecondMetricDesc
	ch <- haproxyULimitMetricDesc
	ch <- haproxyMaxSock
	ch <- haproxyMaxConn
	ch <- haproxyMaxPipes
	ch <- haproxyCurrentConn
	ch <- haproxyCurrentPipes
	ch <- haproxyCurrentConnRate
	ch <- haproxyMaxConnRate

	for _, frontendMetricDesc := range frontendMetricDescArray {
		ch <- frontendMetricDesc
	}
	for _, backendMetricDesc := range backendMetricDescArray {
		ch <- backendMetricDesc
	}
	for _, serverMetricDesc := range serverMetricDescArray {
		ch <- serverMetricDesc
	}
}

// Collect implements prometheus exporter Collect interface
func (m *Manager) Collect(ch chan<- prometheus.Metric) {
	var status *Status
	m.statsMutex.Lock()
	status = m.stats
	m.statsMutex.Unlock()

	if status == nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		haproxyPidMetricDesc, prometheus.GaugeValue, float64(status.HaproxyPid), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyUpTimeSecondMetricDesc, prometheus.GaugeValue, float64(status.UpTimeSecond), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyULimitMetricDesc, prometheus.GaugeValue, float64(status.ULimitN), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyMaxSock, prometheus.GaugeValue, float64(status.MaxSock), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyMaxConn, prometheus.GaugeValue, float64(status.MaxConn), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyMaxPipes, prometheus.GaugeValue, float64(status.MaxPipes), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyCurrentConn, prometheus.GaugeValue, float64(status.CurrentConn), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyCurrentConnRate, prometheus.GaugeValue, float64(status.ConnRate), []string{}...)
	ch <- prometheus.MustNewConstMetric(
		haproxyMaxConnRate, prometheus.GaugeValue, float64(status.ConnMaxRate), []string{}...)

	// Attentions: frontend or backend may be empty occasionally
	for _, service := range status.Services {
		frontend := service.Frontend
		if frontend != nil {
			for i := 0; i <= 33; i++ {
				if i == 15 {
					ch <- prometheus.MustNewConstMetric(frontendMetricDescArray[i],
						prometheus.GaugeValue, convertStatus(frontend.Status), frontend.Name)
				} else if i == 27 {
					ch <- prometheus.MustNewConstMetric(frontendMetricDescArray[i],
						prometheus.GaugeValue, convertCheckStatus(frontend.CheckStatus), frontend.Name)
				} else {
					ch <- prometheus.MustNewConstMetric(frontendMetricDescArray[i],
						prometheus.GaugeValue, getValue(frontend, keysArray[i]), frontend.Name)
				}
			}
		}

		backend := service.Backend
		if backend != nil {
			for i := 0; i <= 33; i++ {
				if i == 15 {
					ch <- prometheus.MustNewConstMetric(backendMetricDescArray[i],
						prometheus.GaugeValue, convertStatus(backend.Status), backend.Name)
				} else if i == 27 {
					ch <- prometheus.MustNewConstMetric(backendMetricDescArray[i],
						prometheus.GaugeValue, convertCheckStatus(backend.CheckStatus), backend.Name)
				} else {
					ch <- prometheus.MustNewConstMetric(backendMetricDescArray[i],
						prometheus.GaugeValue, getValue(backend, keysArray[i]), backend.Name)
				}
			}
		}

		servers := service.Servers
		if len(servers) != 0 {
			for _, server := range servers {
				for i := 0; i <= 33; i++ {
					if i == 15 {
						ch <- prometheus.MustNewConstMetric(serverMetricDescArray[i],
							prometheus.GaugeValue, convertStatus(server.Status),
							server.ServerName, server.Name, server.Address)
					} else if i == 27 {
						ch <- prometheus.MustNewConstMetric(serverMetricDescArray[i],
							prometheus.GaugeValue, convertCheckStatus(server.CheckStatus),
							server.ServerName, server.Name, server.Address)
					} else {
						ch <- prometheus.MustNewConstMetric(serverMetricDescArray[i],
							prometheus.GaugeValue, getValue(server, keysArray[i]),
							server.ServerName, server.Name, server.Address)
					}
				}
			}
		}
	}
}
