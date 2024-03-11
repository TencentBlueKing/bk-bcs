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

package cloudcollector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	namespaceForCloudBalance = "cloudloadbalance"
	backendHealthy           = 1
	backendUnhealthy         = 0
)

var (
	fetchBackendStatusMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "bkbcs_ingressctrl",
			Subsystem: "cloudloadbalance",
			Name:      "fetchbackend_flag",
			Help:      "status flag for fetch cloud loadbalance status",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(fetchBackendStatusMetric)
}

func newBackendHealthMetric(namespace string, metricName string,
	docString string, constLabels map[string]string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "backendhealth", metricName),
		docString, []string{"lbid", "listenerid", "port", "protocol", "host", "path", "rsip", "rsport", "namespace"},
		constLabels)
}

func getMapKeys(m map[string]struct{}) []string {
	var retList []string
	for k := range m {
		retList = append(retList, k)
	}
	return retList
}

func mergeLbMap(m map[string]map[string]map[string]struct{}, region, ns, value string) {
	_, ok := m[region]
	if !ok {
		m[region] = make(map[string]map[string]struct{})
	}
	_, ok = m[region][ns]
	if !ok {
		m[region][ns] = make(map[string]struct{})
	}
	m[region][ns][value] = struct{}{}
}
