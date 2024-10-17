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

package service

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
)

var (
	metricInstance *metric
	once           sync.Once
)

func initMetric() *metric {
	once.Do(func() {
		m := new(metric)
		labels := prometheus.Labels{}
		m.syncQueueLen = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.RepoSyncSubSys,
			Name:        "sync_queue_len",
			Help:        "the length of sync queue for repo sync",
			ConstLabels: labels,
		})
		metrics.Register().MustRegister(m.syncQueueLen)

		m.ackQueueLen = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   metrics.Namespace,
			Subsystem:   metrics.RepoSyncSubSys,
			Name:        "ack_queue_len",
			Help:        "the length of ack queue for repo sync",
			ConstLabels: labels,
		})
		metrics.Register().MustRegister(m.ackQueueLen)

		metricInstance = m

	})
	return metricInstance
}

type metric struct {
	// syncQueueLen records the length of sync queue for repo sync
	syncQueueLen prometheus.Gauge

	// ackQueueLen records the length of ack queue for repo sync
	ackQueueLen prometheus.Gauge
}
