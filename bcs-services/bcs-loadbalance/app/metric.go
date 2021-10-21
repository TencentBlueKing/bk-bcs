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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// LoadbalanceZookeeperStateMetric loadbalance metric for zookeeper connection
	LoadbalanceZookeeperStateMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "state",
			Help:      "the state for zookeeper connection, 0 for abnormal, 1 for normal",
		},
		[]string{types.MetricLabelLoadbalance},
	)
	// LoadbalanceZookeeperEventAddMetric loadbalance metric for zookeeper event add
	LoadbalanceZookeeperEventAddMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "export_service_event_add",
			Help:      "add event of exported service record in zookeeper",
		},
		[]string{types.MetricLabelLoadbalance},
	)
	// LoadbalanceZookeeperEventUpdateMetric loadbalance metric for zookeeper event update
	LoadbalanceZookeeperEventUpdateMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "export_service_event_update",
			Help:      "update event of exported service record in zookeeper",
		},
		[]string{types.MetricLabelLoadbalance},
	)
	// LoadbalanceZookeeperEventDeleteMetric loadbalance metric for zookeeper event delete
	LoadbalanceZookeeperEventDeleteMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadbalance",
			Subsystem: "zookeeper",
			Name:      "export_service_event_delete",
			Help:      "delete event of exported service record in zookeeper",
		},
		[]string{types.MetricLabelLoadbalance},
	)
	// LoadbalanceServiceConflictMetric loadbalance metric for service conflict about port or [port + domain + url]
	LoadbalanceServiceConflictMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadbalance",
			Subsystem: "service",
			Name:      "conflict_total",
			Help:      "conflict counter of service",
		},
		[]string{types.MetricLabelLoadbalance, types.MetricLabelServiceName},
	)
)
