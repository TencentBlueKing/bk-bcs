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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bcs_networkpolicy"
)

var (
	// ControllerIPTablesSyncTime Time it took for controller to sync iptables
	ControllerIPTablesSyncTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "controller_iptables_sync_time",
		Help:      "Time it took for controller to sync iptables",
	})
	// ControllerPolicyChainsSyncTime Time it took for controller to sync policys
	ControllerPolicyChainsSyncTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "controller_policy_chains_sync_time",
		Help:      "Time it took for controller to sync policy chains",
	})
	// ControllerIPTablesSyncError status for network policy controller
	ControllerIPTablesSyncError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "controller_iptables_error_counter",
			Help:      "controller iptables sync error counter",
		},
	)
)
