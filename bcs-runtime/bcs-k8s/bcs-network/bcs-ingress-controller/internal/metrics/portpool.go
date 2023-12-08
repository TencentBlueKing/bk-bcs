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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	portBindLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "portpool",
		Name:      "bind_latency_seconds",
		Help:      "port bind latency for bcs ingress controller port-pool",
		Buckets:   []float64{1, 5, 10, 30, 60, 120, 180, 300, 600, 1200},
	}, []string{})

	portAllocateFailedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "portpool",
		Name:      "allocate_failed_gauge",
		Help:      "port allocate failed gauge",
	}, []string{"name", "namespace"})
)

func init() {
	metrics.Registry.MustRegister(portBindLatency)
	metrics.Registry.MustRegister(portAllocateFailedGauge)
}

// ReportPortBindMetric report port bind metrics
func ReportPortBindMetric(started time.Time) {
	portBindLatency.WithLabelValues().Observe(time.Since(started).Seconds())
}

// ReportPortAllocate report port allocate failed
func ReportPortAllocate(name, namespace string, allocateSuccess bool) {
	if allocateSuccess == false {
		portAllocateFailedGauge.WithLabelValues(name, namespace).Set(1)
	} else {
		portAllocateFailedGauge.WithLabelValues(name, namespace).Set(0)
	}
}

func CleanPortAllocateMetric(name, namespace string) {
	portAllocateFailedGauge.Delete(prometheus.Labels{"name": name, "namespace": namespace})
}
