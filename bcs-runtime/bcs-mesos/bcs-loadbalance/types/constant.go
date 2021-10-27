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

package types

const (
	// MetricLabelLoadbalance metric label name loadbalance
	MetricLabelLoadbalance = "loadbalance"
	// MetricLabelBackend metric label name for loadbalance backend
	MetricLabelBackend = "backend"
	// MetricLabelFrontent metric label name for loadbalance frontend
	MetricLabelFrontent = "frontend"
	// MetricLabelServer metric label name for loadbalance server
	MetricLabelServer = "server"
	// MetricLabelServerAddress metric label name for loadbalance server
	MetricLabelServerAddress = "address"
	// MetricLabelServiceName metric label name for service name
	MetricLabelServiceName = "serviceName"
	// MetricLabelNamespace metric label name for bcs namespace
	MetricLabelNamespace = "namespace"

	// EnvBcsLoadbalanceName env BCS_LOADBALANCE_NAME
	EnvBcsLoadbalanceName = "BCS_LOADBALANCE_NAME"
)
