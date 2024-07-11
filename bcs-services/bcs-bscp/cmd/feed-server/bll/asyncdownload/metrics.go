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

package asyncdownload

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

// InitMetric init the async doenload related prometheus metrics
func InitMetric() *metric {
	m := new(metric)
	labels := prometheus.Labels{}

	m.jobDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "job_duration_seconds",
		Help:        "the duration(seconds) to precess async download job",
		ConstLabels: labels,
		Buckets:     []float64{1, 2, 4, 6, 10, 15, 30, 50, 100, 150, 200, 400, 600},
	}, []string{"biz", "app", "file", "targets", "status"})
	metrics.Register().MustRegister(m.jobDurationSeconds)

	m.jobCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "job_count",
		Help:        "the count of the async download job",
		ConstLabels: labels,
	}, []string{"biz", "app", "file", "targets", "status"})
	metrics.Register().MustRegister(m.jobCounter)

	m.taskDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "task_duration_seconds",
		Help:        "the duration(seconds) to precess async download task",
		ConstLabels: labels,
		Buckets:     []float64{1, 2, 4, 6, 10, 15, 30, 50, 100, 150, 200, 400, 600},
	}, []string{"biz", "app", "file", "status"})
	metrics.Register().MustRegister(m.taskDurationSeconds)

	m.taskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "task_count",
		Help:        "the count of the async download task",
		ConstLabels: labels,
	}, []string{"biz", "app", "file", "status"})
	metrics.Register().MustRegister(m.taskCounter)

	m.sourceFilesSizeBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "source_files_size_bytes",
		Help:        "the size of the source files cache size in bytes",
		ConstLabels: labels,
	})
	metrics.Register().MustRegister(m.sourceFilesSizeBytes)

	m.sourceFilesCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Subsystem:   metrics.AsyncDownload,
		Name:        "source_files_count",
		Help:        "the count of the source files count",
		ConstLabels: labels,
	})
	metrics.Register().MustRegister(m.sourceFilesCounter)

	return m
}

type metric struct {

	// jobDurationSeconds record the duration of the async download job.
	jobDurationSeconds *prometheus.HistogramVec

	// jobCounter record the count of the async download job.
	jobCounter *prometheus.CounterVec

	// taskDurationSeconds record the duration of the async download task.
	taskDurationSeconds *prometheus.HistogramVec

	// taskCounter record the count of the async download task.
	taskCounter *prometheus.CounterVec

	// sourceFilesSizeBytes record the size of the source files cache size in bytes.
	sourceFilesSizeBytes prometheus.Gauge

	// sourceFilesCounter record the count of the source files count.
	sourceFilesCounter prometheus.Gauge
}
