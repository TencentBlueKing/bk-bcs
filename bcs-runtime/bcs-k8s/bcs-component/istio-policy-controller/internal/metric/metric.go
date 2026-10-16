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
// package metric is used to collect metrics for controller
package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// ControllerName controller name
	ControllerName = "istio_policy_controller"
)

// declare metrics
var (
	// PolicyGeneratedTotal 策略生成数量
	PolicyGeneratedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: ControllerName,
			Name:      "policy_generated_total",
			Help:      "Total number of policies generated.",
		},
	)
	// PolicySuccessTotal 策略下发成功数量
	PolicySuccessTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: ControllerName,
			Name:      "policy_applied_success_total",
			Help:      "Total number of policies successfully applied.",
		},
	)
	// PolicyConflictTotal 策略冲突数量
	PolicyConflictTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Subsystem: ControllerName,
			Name:      "policy_conflict_total",
			Help:      "Total number of policy conflicts detected.",
		},
	)
)

func init() {
	metrics.Registry.MustRegister(PolicyGeneratedTotal)
	metrics.Registry.MustRegister(PolicySuccessTotal)
	metrics.Registry.MustRegister(PolicyConflictTotal)
}
