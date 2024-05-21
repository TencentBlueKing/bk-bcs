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

package bcssystem

// AvailableNodeMetrics 蓝鲸监控节点的metrics
var AvailableNodeMetrics = []string{
	"bcs:cluster:cpu:usage",
	"bcs:cluster:cpu:total",
	"bcs:cluster:cpu:used",
	"bcs:cluster:pod:usage",
	"bcs:cluster:pod:total",
	"bcs:cluster:pod:used",
	"bcs:cluster:cpu:request",
	"bcs:cluster:cpu_request:usage",
	"bcs:cluster:memory:total",
	"bcs:cluster:memory:used",
	"bcs:cluster:memory:usage",
	"bcs:cluster:memory:request",
	"bcs:cluster:memory_request:usage",
	"bcs:cluster:disk:total",
	"bcs:cluster:disk:used",
	"bcs:cluster:disk:usage",
	"bcs:cluster:diskio:usage",
	"bcs:cluster:diskio:used",
	"bcs:cluster:diskio:total",
	"bcs:cluster:group:node_num",
	"bcs:cluster:group:max_node_num",
	"bcs:node:info",
	"bcs:node:cpu:total",
	"bcs:node:cpu:used",
	"bcs:node:cpu:request",
	"bcs:node:cpu:usage",
	"bcs:node:cpu_request:usage",
	"bcs:node:disk:usage",
	"bcs:node:disk:used",
	"bcs:node:disk:total",
	"bcs:node:diskio:usage",
	"bcs:node:memory:total",
	"bcs:node:memory:used",
	"bcs:node:memory:request",
	"bcs:node:memory:usage",
	"bcs:node:memory_request:usage",
	"bcs:node:container_count",
	"bcs:node:pod_count",
	"bcs:node:pod_total",
	"bcs:node:network_transmit",
	"bcs:node:network_receive",
	"bcs:pod:cpu_usage",
	"bcs:pod:cpu_limit_usage",
	"bcs:pod:cpu_request_usage",
	"bcs:pod:memory_used",
	"bcs:pod:network_transmit",
	"bcs:pod:network_receive",
	"bcs:container:cpu_usage",
	"bcs:container:memory_used",
	"bcs:container:cpu_limit",
	"bcs:container:memory_limit",
	"bcs:container:gpu_memory_usage",
	"bcs:container:gpu_used",
	"bcs:container:gpu_usage",
	"bcs:container:disk_read_total",
	"bcs:container:disk_write_total",
}
