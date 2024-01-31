/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package metric_manager xxx
// nolint
package metric_manager

import (
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog"
)

var (
	requestsTotalLib = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lib_request_total_num",
		Help: "The total number of requests for cluster manager to call other system api",
	}, []string{"system", "handler", "method", "status"})
	requestLatencyLib = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "lib_request_latency_time",
		Help:    "api request latency statistic for cluster manager to call other system",
		Buckets: []float64{0.01, 0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0},
	}, []string{"system", "handler", "method", "status"})

	requestsTotalAPI = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_request_total_num",
		Help: "The total number of requests for cluster manager api",
	}, []string{"handler", "method", "status"})
	requestLatencyAPI = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "plugin_latency_time",
		Help:    "plugin latency statistic ",
		Buckets: []float64{0.1, 0.5, 0.75, 1.0, 2.0, 3.0, 5.0, 10.0, 20.0},
	}, []string{"plugin", "condition1", "condition2", "condition3"})

	// MM
	MM *MetricManger
)

func init() {
	prometheus.MustRegister(requestsTotalLib)
	prometheus.MustRegister(requestLatencyLib)
	prometheus.MustRegister(requestsTotalAPI)
	prometheus.MustRegister(requestLatencyAPI)
	MM = NewMetricManger()
}

// MetricManger metric manager
type MetricManger struct {
	registryMap     map[string]*prometheus.Registry
	registryMapLock sync.Mutex
}

// NewMetricManger init metric manager
func NewMetricManger() *MetricManger {
	return &MetricManger{
		registryMap: make(map[string]*prometheus.Registry),
	}
}

// SetSeperatedMetric 将指标暴露在独立于/metrics的其他路径上 /path/metrics
func (mm *MetricManger) SetSeperatedMetric(path string) {
	if _, ok := mm.registryMap[path]; !ok {
		mm.registryMap[path] = prometheus.NewRegistry()

		componentAHandler := promhttp.HandlerFor(
			prometheus.Gatherers{mm.registryMap[path]},
			promhttp.HandlerOpts{},
		)

		http.Handle("/"+path+"/metrics", componentAHandler)
	}
}

// RegisterSeperatedMetric register metrics
func (mm *MetricManger) RegisterSeperatedMetric(path string, vec *prometheus.GaugeVec) {
	if _, ok := mm.registryMap[path]; !ok {
		mm.SetSeperatedMetric(path)
	}
	mm.registryMap[path].MustRegister(vec)
}

// Register collector
func Register(collector prometheus.Collector) {
	prometheus.MustRegister(collector)
}

// GaugeVecSet xxx
type GaugeVecSet struct {
	Labels []string
	Value  float64
}

// SetMetric xxx
func SetMetric(metricVec *prometheus.GaugeVec, gaugeVecSetList []*GaugeVecSet) {
	metricVec.Reset()

	for _, gaugeVecSet := range gaugeVecSetList {
		metricVec.WithLabelValues(gaugeVecSet.Labels...).Set(gaugeVecSet.Value)

		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("SetMetric failed: %s, stack: %v\n", r, string(debug.Stack()))
				// klog.Errorf("SetMetric failed: %s", gaugeVecSet)
				for _, gaugeVecSet1 := range gaugeVecSetList {
					if gaugeVecSet1 != nil {
						klog.Errorf("SetMetric failed: %s", strings.Join(gaugeVecSet1.Labels, ";"))
						break
					}
				}
			}
		}()
	}

}

// SetCommonDurationMetric xxx
func SetCommonDurationMetric(labels []string, started time.Time) {
	requestLatencyAPI.WithLabelValues(labels...).Observe(time.Since(started).Seconds())
}

// RunPrometheusMetricsServer run metrics server
func (mm *MetricManger) RunPrometheusMetricsServer() {
	// register prometheus server
	http.Handle("/metrics", promhttp.Handler())
	addr := "0.0.0.0:6216"
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			klog.Fatalf("Failed to listen and serve metric server, err %s", err.Error())
		}
	}()
	klog.Infof("run prometheus server ok")
}
