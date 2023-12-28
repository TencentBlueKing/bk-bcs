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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	ListenerTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "listener",
		Name:      "total",
		Help:      "The total listener managed by controller",
	}, []string{"status", "target_group_type"})

	PortBindingTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "portbinding",
		Name:      "total",
		Help:      "The total portBinding managed by controller",
	}, []string{"status"})

	PortBindingConflictGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "bkbcs_ingressctrl",
		Subsystem: "portbinding",
		Name:      "conflict",
		Help:      "The conflict portbinding gauge",
	}, []string{"conflict_key"})
)

func init() {
	metrics.Registry.MustRegister(ListenerTotal)
	metrics.Registry.MustRegister(PortBindingTotal)
	metrics.Registry.MustRegister(PortBindingConflictGauge)
}
