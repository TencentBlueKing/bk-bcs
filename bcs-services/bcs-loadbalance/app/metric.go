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

package app

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// LoadbalanceZookeeperStateMetric loadbalance metric for zookeeper connection
	LoadbalanceZookeeperStateMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "state",
			Help:      "the state for zookeeper connection, 0 for abnormal, 1 for normal",
		},
	)
	// LoadbalanceZookeeperEventMetric loadbalance metric for zookeeper event
	LoadbalanceZookeeperEventMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "export_service_event",
			Help:      "event of exported service record in zookeeper",
		},
		[]string{"kind", "name", "namespace"},
	)
)

func init() {
	prometheus.Register(LoadbalanceZookeeperStateMetric)
	prometheus.Register(LoadbalanceZookeeperEventMetric)
}
